#!/usr/bin/env python3
# 업데이트 app P2(updater Linux) 확인 — 로컬 updater를 임시 DB·5051에 붙여 단계 실행(builtin·script·입력·시간 제한·취소·복원)·설치 경로 변경을 API로. REF-update-p2.md 단위-8(확인)
# 실행: BASE=http://127.0.0.1:5051 CHECK_DB=<임시 DB> UPDATER=<updater 실행 파일> python3 scripts/check/updater-run.py
import hashlib, http.cookiejar, json, os, shutil, signal, stat, subprocess, sys, tempfile, time, urllib.error, urllib.request

BASE = os.environ.get("BASE", "http://127.0.0.1:5051")
API = BASE + "/api"
DB = os.environ.get("CHECK_DB", "")
UPDATER = os.environ["UPDATER"]
fails = 0
work = tempfile.mkdtemp(prefix="upd-run-")
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
    ident = f"run-admin-{int(time.time()*1000)}"
    call(op, "POST", "/users", {"identification": ident, "nickname": ident, "password": "pw-1234-check"})
    psql(f"UPDATE users SET perms = '{{admin}}' WHERE nickname = '{ident}'")
    s, b = call(op, "POST", "/users/session", {"identification": ident, "password": "pw-1234-check"})
    assert s == 200, (s, b)
    return op


def text(op, ref, path, content):
    s, b = call(op, "PUT", ref + "/text", {"path": path, "content": content})
    assert s == 200, (path, s, b)


def start_updater(cfg_dir):
    log = open(os.path.join(cfg_dir, "out.log"), "ab")
    p = subprocess.Popen([UPDATER, "-c", os.path.join(cfg_dir, "updater.json")], stdout=log, stderr=log)
    procs.append(p)
    return p


def wait(cond, timeout=10.0, step=0.1):
    end = time.time() + timeout
    while time.time() < end:
        v = cond()
        if v:
            return v
        time.sleep(step)
    return None


def instances(op, main):
    s, b = call(op, "GET", f"/admin/updates/devices/{main}")
    return b["instances"] if s == 200 else []


def run(op, main, sub, step, lst="deploy", version="v1", **kw):
    s, b = call(op, "POST", f"/admin/updates/devices/{main}/instances/{sub}/runs", {"version": version, "list": lst, "stepId": step, **kw})
    if s != 200:
        return s, b
    rid = b["runId"]
    done = wait(lambda: (lambda r: r if r[1]["result"] != "running" else None)(call(op, "GET", f"/admin/updates/runs/{rid}")), timeout=kw.pop("_t", 20))
    return (200, done[1]) if done else (0, call(op, "GET", f"/admin/updates/runs/{rid}")[1])


def start_run(op, main, sub, step):
    s, b = call(op, "POST", f"/admin/updates/devices/{main}/instances/{sub}/runs", {"version": "v1", "stepId": step})
    return s, (b or {}).get("runId")


def get_run(op, rid):
    return call(op, "GET", f"/admin/updates/runs/{rid}")[1]


try:
    admin = login()
    MAIN = "run-jig"
    s, b = call(admin, "POST", "/admin/updates/devices", {"mainKey": MAIN})
    key = (b or {}).get("apiKey", "")
    ok("장비 루트 만들기 = 키 발급(nxu_)", s == 200 and key.startswith("nxu_"), (s, b))
    s, b = call(admin, "GET", f"/admin/updates/devices/{MAIN}/key")
    ok("키 다시 보기 = 같은 키", s == 200 and b["apiKey"] == key, b)

    # 빌드: 실행 파일(권한 700으로 바꿈) · 스크립트 · steps.json
    B = f"/admin/updates/bundles/build/{MAIN}/v1"
    call(admin, "POST", f"/admin/updates/bundles/build/{MAIN}", {"version": "v1"})
    text(admin, B, "app.conf", "ver=1\n")
    text(admin, B, "bin/tool", "#!/bin/sh\necho tool\n")
    s, b = call(admin, "PATCH", B + "/mode", {"path": "bin/tool", "mode": 0o700})
    ok("권한 바꾸기 700", s == 200, (s, b))
    text(admin, B, "env.sh", "#!/bin/sh\necho \"T=$TARGET_DIR V=$VERSION S=$STEP_ID B=$BACKUP_DIR\"\nls -l \"$STAGE_DIR/bin/tool\" | cut -c1-10\n")
    text(admin, B, "fail.sh", "#!/bin/sh\necho 실패합니다 >&2\nexit 3\n")
    text(admin, B, "slow.sh", "#!/bin/sh\necho 시작\nsleep 30\necho 안 나와야 함\n")
    text(admin, B, "ask.sh", "#!/bin/sh\nprintf '이름? '\nread name\necho \"안녕 $name\"\n")
    text(admin, B, "noexec.sh", "#!/bin/sh\necho x\n")
    call(admin, "PATCH", B + "/mode", {"path": "noexec.sh", "mode": 0o644})
    steps = {"deploy": [
        {"id": "backup", "builtin": "backup"},
        {"id": "copy", "builtin": "copy"},
        {"id": "env", "script": "env.sh"},
        {"id": "fail", "script": "fail.sh"},
        {"id": "slow", "script": "slow.sh", "timeout": "2s"},
        {"id": "ask", "script": "ask.sh", "interactive": True},
        {"id": "noexec", "script": "noexec.sh"},
        {"id": "hang", "script": "slow.sh"},
    ], "rollback": [{"id": "restore", "builtin": "restore"}]}
    text(admin, B, "steps.json", json.dumps(steps, indent=2) + "\n")
    s, b = call(admin, "GET", B)
    modes = {f["path"]: f["mode"] for f in b["files"]}
    ok("파일 권한: .sh 755 · 새 파일 644 · 바꾼 것 700", modes["env.sh"] == 0o755 and modes["app.conf"] == 0o644 and modes["bin/tool"] == 0o700 and modes["noexec.sh"] == 0o644, modes)

    # 설치 경로: 기존 파일 1개(덮어쓸 것)
    target = os.path.join(work, "target")
    os.makedirs(target)
    with open(os.path.join(target, "app.conf"), "w") as f:
        f.write("ver=0\n")
    cfg_dir = os.path.join(work, "u1")
    os.makedirs(cfg_dir)
    cfg = {"supervisor": BASE.replace("http", "ws"), "mainKey": MAIN, "apiKey": key, "targetDir": target}

    # 틀린 키 = 접속 거절
    bad = dict(cfg, apiKey="nxu_wrong")
    with open(os.path.join(cfg_dir, "updater.json"), "w") as f:
        json.dump(bad, f)
    p = start_updater(cfg_dir)
    time.sleep(1.5)
    ok("틀린 키 = 인스턴스 안 생김", instances(admin, MAIN) == [])
    p.terminate(); p.wait()

    with open(os.path.join(cfg_dir, "updater.json"), "w") as f:
        json.dump(cfg, f)
    p = start_updater(cfg_dir)
    ins = wait(lambda: [i for i in instances(admin, MAIN) if i["online"]])
    ok("updater 접속 = 인스턴스 online", bool(ins), instances(admin, MAIN))
    SUB = ins[0]["subKey"] if ins else ""
    ok("인스턴스 보고값(OS·설치 경로·태그)", ins and ins[0]["os"] == "linux" and ins[0]["installDir"] == target and ins[0]["version"].startswith("UPDATER"), ins)
    st = json.load(open(os.path.join(cfg_dir, "state.json")))
    ok("state.json에 서브키 저장", st.get("subKey") == SUB, st)

    s, r = run(admin, MAIN, SUB, "backup")
    ok("backup", s == 200 and r["result"] == "ok" and "있던 파일 1개" in r["log"], r)
    s, r = run(admin, MAIN, SUB, "copy")
    ok("copy", s == 200 and r["result"] == "ok", r)
    tool = os.path.join(target, "bin/tool")
    ok("copy 결과: 덮어씀 · 권한 700 · .sh 755 · 644 유지", open(os.path.join(target, "app.conf")).read() == "ver=1\n"
       and stat.S_IMODE(os.stat(tool).st_mode) == 0o700 and stat.S_IMODE(os.stat(os.path.join(target, "env.sh")).st_mode) == 0o755
       and stat.S_IMODE(os.stat(os.path.join(target, "noexec.sh")).st_mode) == 0o644)
    s, r = run(admin, MAIN, SUB, "env")
    ok("script: 환경변수·staging 권한", s == 200 and r["result"] == "ok" and r["exitCode"] == 0 and f"T={target} V=v1 S=env B=" in r["log"] and "-rwx------" in r["log"], r)
    s, r = run(admin, MAIN, SUB, "fail")
    ok("script 실패 = fail + 종료 코드 3 + stderr", s == 200 and r["result"] == "fail" and r["exitCode"] == 3 and "실패합니다" in r["log"], r)
    s, r = run(admin, MAIN, SUB, "slow")
    ok("시간 제한 2s = timeout", s == 200 and r["result"] == "timeout" and "안 나와야 함" not in r["log"], r)
    s, r = run(admin, MAIN, SUB, "noexec")
    ok("권한 644 스크립트 = 실행 실패", s == 200 and r["result"] == "fail", r)

    # interactive: 입력 중계
    s, rid = start_run(admin, MAIN, SUB, "ask")
    ok("interactive 시작", s == 200 and rid)
    wait(lambda: "이름?" in get_run(admin, rid)["log"])
    s2, b2 = call(admin, "POST", f"/admin/updates/runs/{rid}/input", {"data": "nexus\r"})
    r = wait(lambda: (lambda x: x if x["result"] != "running" else None)(get_run(admin, rid)))
    ok("interactive: 입력 -> 출력", s2 == 200 and r and r["result"] == "ok" and "안녕 nexus" in r["log"], (s2, b2, r))
    s, b = call(admin, "POST", f"/admin/updates/runs/{rid}/input", {"data": "x"})
    ok("끝난 실행에 입력 = 409", s == 409, s)

    # 취소 + 실행 중 다른 단계·경로 변경 거절
    s, rid = start_run(admin, MAIN, SUB, "hang")
    wait(lambda: "시작" in get_run(admin, rid)["log"])
    s2, _ = start_run(admin, MAIN, SUB, "env")
    ok("실행 중 다른 단계 = 409", s2 == 409, s2)
    s3, _ = call(admin, "PATCH", f"/admin/updates/devices/{MAIN}/instances/{SUB}", {"targetDir": "/tmp/x"})
    ok("실행 중 설치 경로 변경 = 409", s3 == 409, s3)
    s4, _ = call(admin, "POST", f"/admin/updates/runs/{rid}/input", {"data": "x"})
    ok("비대화형 단계에 입력 = 409", s4 == 409, s4)
    call(admin, "POST", f"/admin/updates/runs/{rid}/cancel")
    r = wait(lambda: (lambda x: x if x["result"] != "running" else None)(get_run(admin, rid)))
    ok("취소 = cancel", r and r["result"] == "cancel", r)

    # 설치 경로 변경 -> 설정 파일 고침 -> restore는 백업 때 경로로
    t2 = os.path.join(work, "target2")
    s, b = call(admin, "PATCH", f"/admin/updates/devices/{MAIN}/instances/{SUB}", {"targetDir": t2, "label": "지그1"})
    c2 = json.load(open(os.path.join(cfg_dir, "updater.json")))
    ok("설치 경로 변경 = updater.json 고침(키 유지)", s == 200 and c2["targetDir"] == t2 and c2["apiKey"] == key, (s, b, c2))
    s, b = call(admin, "PATCH", f"/admin/updates/devices/{MAIN}/instances/{SUB}", {"targetDir": "rel/path"})
    ok("상대 경로 = 거절", s == 409, (s, b))
    s, r = run(admin, MAIN, SUB, "restore", lst="rollback")
    ok("restore = 백업 때 경로", s == 200 and r["result"] == "ok" and open(os.path.join(target, "app.conf")).read() == "ver=0\n"
       and not os.path.exists(tool) and not os.path.exists(t2), r)
    ins = instances(admin, MAIN)
    ok("인스턴스 별칭·경로 반영", ins[0]["label"] == "지그1" and ins[0]["installDir"] == t2, ins)

    # 재시작해도 같은 서브키, 실행 기록 목록
    p.terminate(); p.wait()
    ok("끊김 = offline", wait(lambda: not instances(admin, MAIN)[0]["online"]))
    s, b = call(admin, "DELETE", f"/admin/updates/devices/{MAIN}/instances/{SUB}")
    ok("offline 인스턴스 지우기", s == 200, (s, b))
    p = start_updater(cfg_dir)
    ins = wait(lambda: [i for i in instances(admin, MAIN) if i["online"]])
    ok("다시 접속 = 같은 서브키(행 새로)", ins and ins[0]["subKey"] == SUB and ins[0]["label"] == "", ins)
    s, b = call(admin, "DELETE", f"/admin/updates/devices/{MAIN}/instances/{SUB}")
    ok("online 인스턴스 지우기 = 409", s == 409, s)
    s, b = call(admin, "GET", f"/admin/updates/devices/{MAIN}/instances/{SUB}/runs")
    ok("실행 기록 목록", s == 200 and len(b["items"]) >= 9, b)

    # 같은 state.json 두 개 = 두 번째 거절
    cfg_dir2 = os.path.join(work, "u2")
    shutil.copytree(cfg_dir, cfg_dir2, ignore=shutil.ignore_patterns("staging", "backups", "logs", "out.log"))
    start_updater(cfg_dir2)
    time.sleep(1.5)
    ok("같은 서브키 두 번째 updater = 거절", b"\xec\x9d\xb4\xeb\xaf\xb8 \xec\xa0\x91\xec\x86\x8d" in open(os.path.join(cfg_dir2, "out.log"), "rb").read() and len(instances(admin, MAIN)) == 1)

    # 키 재발급 = 옛 키 다음 접속부터 거절
    s, b = call(admin, "POST", f"/admin/updates/devices/{MAIN}/key")
    ok("키 재발급", s == 200 and b["apiKey"] != key)
    req = urllib.request.Request(BASE + "/updater/file?b=1&path=app.conf", headers={"X-Nexus-Main": MAIN, "X-Nexus-Key": key})
    try:
        urllib.request.urlopen(req)
        code = 200
    except urllib.error.HTTPError as e:
        code = e.code
    ok("옛 키로 파일 받기 = 401", code == 401, code)
finally:
    for p in procs:
        if p.poll() is None:
            p.send_signal(signal.SIGTERM)
            p.wait()
    shutil.rmtree(work, ignore_errors=True)

print(f"\n{'OK' if fails == 0 else f'{fails} FAIL'}")
sys.exit(1 if fails else 0)
