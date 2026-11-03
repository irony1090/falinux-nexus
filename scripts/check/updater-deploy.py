#!/usr/bin/env python3
# 업데이트 app P3(배포) 확인 — 로컬 updater를 임시 DB·5051에 붙여 배포·자동 되돌리기·수동 되돌리기·취소·잠그기·updater 재시작 중단을 API로. REF-update-p3.md 단위-7(확인)
# 실행: BASE=http://127.0.0.1:5051 CHECK_DB=<임시 DB> UPDATER=<updater 실행 파일> python3 scripts/check/updater-deploy.py
import http.cookiejar, json, os, shutil, signal, subprocess, sys, tempfile, time, urllib.error, urllib.request

BASE = os.environ.get("BASE", "http://127.0.0.1:5051")
API = BASE + "/api"
DB = os.environ.get("CHECK_DB", "")
UPDATER = os.environ["UPDATER"]
fails = 0
work = tempfile.mkdtemp(prefix="upd-dep-")
procs = []


def ok(name, cond, extra=""):
    global fails
    print(("PASS " if cond else "FAIL ") + name + (f"  -- {extra}" if extra and not cond else ""))
    if not cond:
        fails += 1


def call(op, method, path, body=None):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(API + path, data=data, method=method)
    if data is not None:
        req.add_header("Content-Type", "application/json")
    try:
        with op.open(req) as r:
            b = r.read()
            return r.status, (json.loads(b) if b else None)
    except urllib.error.HTTPError as e:
        b = e.read()
        try:
            return e.code, json.loads(b)
        except Exception:
            return e.code, b


def psql(sql):
    subprocess.run(["docker", "exec", "postgres15", "psql", "-U", "irony", "-d", DB, "-v", "ON_ERROR_STOP=1", "-qAtc", sql], check=True)


def login():
    op = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    ident = f"dep-admin-{int(time.time()*1000)}"
    call(op, "POST", "/users", {"identification": ident, "nickname": ident, "password": "pw-1234-check"})
    psql(f"UPDATE users SET perms = '{{admin}}' WHERE nickname = '{ident}'")
    s, b = call(op, "POST", "/users/session", {"identification": ident, "password": "pw-1234-check"})
    assert s == 200, (s, b)
    return op


def wait(cond, timeout=20.0, step=0.15):
    end = time.time() + timeout
    while time.time() < end:
        v = cond()
        if v:
            return v
        time.sleep(step)
    return None


def text(op, ref, path, content):
    s, b = call(op, "PUT", ref + "/text", {"path": path, "content": content})
    assert s == 200, (path, s, b)


def start_updater(cfg_dir):
    log = open(os.path.join(cfg_dir, "out.log"), "ab")
    p = subprocess.Popen([UPDATER, "-c", os.path.join(cfg_dir, "updater.json")], stdout=log, stderr=log)
    procs.append(p)
    return p


def instance(op, main):
    s, b = call(op, "GET", f"/admin/updates/devices/{main}")
    return b["instances"][0] if s == 200 and b["instances"] else {}


def dep(op, did):
    return call(op, "GET", f"/admin/updates/deployments/{did}")[1]


def done(op, did, timeout=20.0):
    return wait(lambda: (lambda d: d if d["result"] != "running" and d["rollbackResult"] != "running" and not d["busy"] else None)(dep(op, did)), timeout)


def deploy(op, main, ver, subs):
    s, b = call(op, "POST", f"/admin/updates/bundles/build/{main}/{ver}/deploy", {"subKeys": subs})
    return s, (b or {}).get("items", b)


def runs_of(d, lst):
    return [(r["stepId"], r["result"]) for r in d["runs"] if r["list"] == lst]


def read(path):
    try:
        return open(path).read()
    except FileNotFoundError:
        return None


try:
    admin = login()
    MAIN = "dep-jig"
    s, b = call(admin, "POST", "/admin/updates/devices", {"mainKey": MAIN})
    key = b["apiKey"]

    # 공용 묶음 tools@c1 + 빌드 v1(uses tools@c1)
    C = "/admin/updates/bundles/common/tools/c1"
    call(admin, "POST", "/admin/updates/bundles/common/tools", {"version": "c1"})
    text(admin, C, "tool.conf", "t=1\n")
    B1 = f"/admin/updates/bundles/build/{MAIN}/v1"
    call(admin, "POST", f"/admin/updates/bundles/build/{MAIN}", {"version": "v1"})
    text(admin, B1, "app.conf", "ver=1\n")
    text(admin, B1, "hello.sh", "#!/bin/sh\necho \"hello $VERSION $DEPLOY_ID\"\n")
    text(admin, B1, "after.sh", "#!/bin/sh\necho \"after rollback $(cat $TARGET_DIR/app.conf)\"\n")
    text(admin, B1, "prune.sh", "#!/bin/sh\necho prune\n")
    steps1 = {"uses": ["tools@c1"], "deploy": [
        {"id": "backup", "builtin": "backup"},
        {"id": "copy", "builtin": "copy"},
        {"id": "hello", "script": "hello.sh"},
        {"id": "prune", "script": "prune.sh", "manual": True},
    ], "rollback": [{"id": "restore", "builtin": "restore"}, {"id": "after", "script": "after.sh"}]}
    text(admin, B1, "steps.json", json.dumps(steps1, indent=2) + "\n")

    target = os.path.join(work, "target")
    os.makedirs(target)
    with open(os.path.join(target, "app.conf"), "w") as f:
        f.write("ver=0\n")
    cfg_dir = os.path.join(work, "u1")
    os.makedirs(cfg_dir)
    with open(os.path.join(cfg_dir, "updater.json"), "w") as f:
        json.dump({"supervisor": BASE.replace("http", "ws"), "mainKey": MAIN, "apiKey": key, "targetDir": target}, f)
    p = start_updater(cfg_dir)
    ins = wait(lambda: (lambda i: i if i.get("online") else None)(instance(admin, MAIN)))
    SUB = ins["subKey"]

    # ---- 1. 성공 배포 ----
    s, items = deploy(admin, MAIN, "v1", [SUB])
    ok("배포 시작 = 배포 id", s == 200 and items[0].get("deployId") and not items[0].get("error"), items)
    D1 = items[0]["deployId"]
    d = done(admin, D1)
    ok("배포 성공 = ok, 단계 3개(manual 뺌) ok", d and d["result"] == "ok" and runs_of(d, "deploy") == [("backup", "ok"), ("copy", "ok"), ("hello", "ok")], d and d["runs"])
    ok("자동 되돌리기 꺼짐 = rollback 기록 없음", d and runs_of(d, "rollback") == [] and d["rollbackResult"] == "", d)
    ok("단계 정의(deploy 4개·rollback 2개, manual 표시)", d and len(d["deploy"]) == 4 and d["deploy"][3]["manual"] and len(d["rollback"]) == 2, d and d["deploy"])
    hello = [r for r in d["runs"] if r["stepId"] == "hello"][0]
    s, r = call(admin, "GET", f"/admin/updates/runs/{hello['id']}")
    ok("script 로그 = DEPLOY_ID d<id>", f"hello v1 d{D1}" in r["log"], r["log"])
    ok("설치 경로에 깔림(공용 포함)", read(os.path.join(target, "app.conf")) == "ver=1\n" and read(os.path.join(target, "tool.conf")) == "t=1\n")
    ok("설치본 이름 = v1", wait(lambda: instance(admin, MAIN).get("installedVersion") == "v1"), instance(admin, MAIN))
    ok("state.json installed = v1", json.load(open(os.path.join(cfg_dir, "state.json"))).get("installed") == "v1")
    ok("잠글 수 있음(lockable)", d and d["lockable"] and not d["buildLocked"], d)
    s, b = call(admin, "GET", f"/admin/updates/devices/{MAIN}")
    ok("빌드 행 lastDeployId", s == 200 and b["builds"][0]["lastDeployId"] == D1, b["builds"])

    # 개별 실행(manual 단계) — 배포 문맥(d<id>) · Run·Log의 개별 실행 목록
    s, b = call(admin, "POST", f"/admin/updates/devices/{MAIN}/instances/{SUB}/runs", {"version": "v1", "stepId": "prune", "deployId": f"d{D1}"})
    ok("manual 단계 개별 실행", s == 200, (s, b))
    wait(lambda: dep(admin, D1)["manual"] and dep(admin, D1)["manual"][0]["result"] == "ok")
    ok("Run·Log 개별 실행 목록", [m["stepId"] for m in dep(admin, D1)["manual"]] == ["prune"], dep(admin, D1)["manual"])

    # ---- 2. 잠그기 ----
    s, b = call(admin, "POST", B1 + "/lock")
    ok("잠그기 = 200", s == 200, (s, b))
    s, b = call(admin, "GET", B1)
    s2, c = call(admin, "GET", C)
    ok("빌드·공용 묶음 같이 잠김", b["locked"] and c["locked"], (b.get("locked"), c.get("locked")))
    s, b = call(admin, "POST", B1 + "/lock")
    ok("다시 잠그기 = 409", s == 409, s)

    # ---- 3. 실패 배포 + 자동 되돌리기(v2 = v1 복사 + 실패 단계) ----
    call(admin, "POST", f"/admin/updates/bundles/build/{MAIN}", {"version": "v2", "from": {"owner": MAIN, "version": "v1"}})
    B2 = f"/admin/updates/bundles/build/{MAIN}/v2"
    text(admin, B2, "app.conf", "ver=2\n")
    text(admin, B2, "bad.sh", "#!/bin/sh\necho 망가짐 >&2\nexit 2\n")
    steps2 = dict(steps1, autoRollback=True)
    steps2["deploy"] = steps1["deploy"][:2] + [{"id": "bad", "script": "bad.sh"}, {"id": "hello", "script": "hello.sh"}]
    text(admin, B2, "steps.json", json.dumps(steps2, indent=2) + "\n")
    s, items = deploy(admin, MAIN, "v2", [SUB])
    D2 = items[0]["deployId"]
    d = done(admin, D2)
    ok("실패 배포 = fail, 남은 단계 skip", d and d["result"] == "fail" and runs_of(d, "deploy") == [("backup", "ok"), ("copy", "ok"), ("bad", "fail"), ("hello", "skip")], d and d["runs"])
    ok("자동 되돌리기 = ok, restore·after 실행", d and d["rollbackResult"] == "ok" and runs_of(d, "rollback") == [("restore", "ok"), ("after", "ok")], d and (d["rollbackResult"], d["runs"]))
    ok("실패 이유 msg", d and "bad" in d["msg"], d and d["msg"])
    ok("되돌린 뒤 설치 경로 = v1 내용", read(os.path.join(target, "app.conf")) == "ver=1\n" and read(os.path.join(target, "bad.sh")) is None)
    ok("설치본 이름 = 백업 때 이름(v1)", wait(lambda: instance(admin, MAIN).get("installedVersion") == "v1"), instance(admin, MAIN))
    ok("실패한 빌드 = 잠글 수 없음", d and not d["lockable"], d)
    s, b = call(admin, "POST", B2 + "/lock")
    ok("잠그기 409(성공 기록 없음)", s == 409, (s, b))

    # ---- 4. 수동 되돌리기(v1 배포) ----
    s, b = call(admin, "POST", f"/admin/updates/deployments/{D1}/rollback", {})
    ok("되돌리기 시작", s == 200, (s, b))
    d = done(admin, D1)
    ok("되돌리기 ok", d and d["rollbackResult"] == "ok" and runs_of(d, "rollback") == [("restore", "ok"), ("after", "ok")], d and d["runs"])
    ok("되돌린 뒤 = 처음 내용(ver=0), 새 파일 지움", read(os.path.join(target, "app.conf")) == "ver=0\n" and read(os.path.join(target, "hello.sh")) is None)
    ok("설치본 이름 = 배포 전(빈 값)", wait(lambda: instance(admin, MAIN).get("installedVersion") == ""), instance(admin, MAIN))
    s, b = call(admin, "POST", f"/admin/updates/deployments/{D1}/rollback", {})
    ok("되돌리기 다시 = 가능(같은 백업)", s == 200, (s, b))
    done(admin, D1)

    # ---- 5. 취소 · 바쁨 ----
    call(admin, "POST", f"/admin/updates/bundles/build/{MAIN}", {"version": "v3", "from": {"owner": MAIN, "version": "v1"}})
    B3 = f"/admin/updates/bundles/build/{MAIN}/v3"
    text(admin, B3, "slow.sh", "#!/bin/sh\necho 시작\nsleep 30\necho 안 나와야 함\n")
    steps3 = dict(steps1)
    steps3["deploy"] = [{"id": "backup", "builtin": "backup"}, {"id": "slow", "script": "slow.sh"}, {"id": "hello", "script": "hello.sh"}]
    text(admin, B3, "steps.json", json.dumps(steps3, indent=2) + "\n")
    s, items = deploy(admin, MAIN, "v3", [SUB])
    D3 = items[0]["deployId"]
    wait(lambda: ("slow", "running") in runs_of(dep(admin, D3), "deploy"))
    s, items2 = deploy(admin, MAIN, "v3", [SUB])
    ok("배포 중 다른 배포 = 그 인스턴스 오류", s == 200 and items2[0].get("error"), items2)
    s, b = call(admin, "POST", f"/admin/updates/devices/{MAIN}/instances/{SUB}/runs", {"version": "v3", "stepId": "hello"})
    ok("배포 중 개별 실행 = 409", s == 409, (s, b))
    s, b = call(admin, "POST", f"/admin/updates/deployments/{D1}/rollback", {})
    ok("배포 중 다른 배포 되돌리기 = 409", s == 409, (s, b))
    s, b = call(admin, "POST", f"/admin/updates/deployments/{D3}/cancel")
    ok("취소 요청", s == 200, (s, b))
    d = done(admin, D3)
    ok("취소 = cancel, 남은 단계 skip", d and d["result"] == "cancel" and runs_of(d, "deploy") == [("backup", "ok"), ("slow", "cancel"), ("hello", "skip")], d and d["runs"])

    # 시작 알림·출력·끝 순서(HandleOrdered) — 짧은 단계를 여러 번 돌려도 실행 중 표시가 남지 않고 로그가 빠지지 않음
    stuck, empty = [], []
    for _ in range(5):
        s, items = deploy(admin, MAIN, "v1", [SUB])
        dd = done(admin, items[0]["deployId"])
        i = instance(admin, MAIN)
        if i.get("running") or i.get("deploying"):
            stuck.append(i)
        for r in dd["runs"]:
            if r["stepId"] == "hello" and "hello v1" not in call(admin, "GET", f"/admin/updates/runs/{r['id']}")[1]["log"]:
                empty.append(r["id"])
    ok("반복 배포 5회: 끝난 뒤 실행 중 표시 없음", not stuck, stuck)
    ok("반복 배포 5회: 짧은 단계 로그 남음", not empty, empty)
    # 백업은 최근 3회만 — D1 백업이 지워져 되돌리기는 이유와 함께 실패
    s, b = call(admin, "POST", f"/admin/updates/deployments/{D1}/rollback", {})
    d = done(admin, D1)
    ok("백업이 지워진 배포 되돌리기 = fail + 이유", s == 200 and d and d["rollbackResult"] == "fail" and "백업이 없습니다" in d["msg"], d and (d["rollbackResult"], d["msg"]))

    # ---- 6. 배포 중 updater 재시작 = 중단 ----
    s, items = deploy(admin, MAIN, "v3", [SUB])
    D4 = items[0]["deployId"]
    wait(lambda: ("slow", "running") in runs_of(dep(admin, D4), "deploy"))
    p.kill(); p.wait()
    subprocess.run(["pkill", "-f", os.path.join(work, "u1")], check=False)  # 남은 slow.sh(staging 안)
    wait(lambda: not instance(admin, MAIN).get("online"))
    p = start_updater(cfg_dir)
    wait(lambda: instance(admin, MAIN).get("online"))
    d = wait(lambda: (lambda x: x if x["result"] != "running" else None)(dep(admin, D4)))
    ok("updater 재시작 = 배포 cancel + 이유", d and d["result"] == "cancel" and "모름" in d["msg"], d and (d["result"], d["msg"]))
    ok("실행 중이던 단계 cancel · 남은 단계 skip", d and runs_of(d, "deploy") == [("backup", "ok"), ("slow", "cancel"), ("hello", "skip")], d and d["runs"])
    ok("다시 바쁘지 않음", d and not d["busy"], d)
finally:
    for p in procs:
        if p.poll() is None:
            p.send_signal(signal.SIGTERM)
            p.wait()
    subprocess.run(["pkill", "-f", work], check=False)
    shutil.rmtree(work, ignore_errors=True)

print(f"\n{'OK' if fails == 0 else f'{fails} FAIL'}")
sys.exit(1 if fails else 0)
