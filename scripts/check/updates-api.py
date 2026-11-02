#!/usr/bin/env python3
# 업데이트 app P1(보관·화면) API 확인 — 임시 DB·5051에서만(run.sh가 띄움). REF-update-deploy.md "P1 착수 계획" 단위-8(확인)
# 실행: BASE=http://127.0.0.1:5051 CHECK_DB=<임시 DB> python3 scripts/check/updates-api.py
import hashlib, http.cookiejar, io, json, os, subprocess, sys, time, urllib.error, urllib.request, zipfile

BASE = os.environ.get("BASE", "http://127.0.0.1:5051") + "/api"
DB = os.environ.get("CHECK_DB", "")
fails = 0


def client():
    return urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))


def call(op, method, path, body=None, raw=None, ctype="application/json"):
    data = raw if raw is not None else (json.dumps(body).encode() if body is not None else None)
    req = urllib.request.Request(BASE + path, data=data, method=method)
    if data is not None:
        req.add_header("Content-Type", ctype if raw is None else "application/octet-stream")
    try:
        with op.open(req) as r:
            b = r.read()
            return r.status, b, r.headers
    except urllib.error.HTTPError as e:
        return e.code, e.read(), e.headers


def js(b):
    try:
        return json.loads(b)
    except Exception:
        return None


def ok(name, cond, extra=""):
    global fails
    print(("PASS " if cond else "FAIL ") + name + (f"  -- {extra}" if extra and not cond else ""))
    if not cond:
        fails += 1


def psql(sql):
    subprocess.run(["docker", "exec", "postgres15", "psql", "-U", "irony", "-d", DB, "-v", "ON_ERROR_STOP=1", "-qAtc", sql], check=True)


def login(nick, admin=False):
    op = client()
    ident = f"{nick}-{int(time.time()*1000)}"
    call(op, "POST", "/users", {"identification": ident, "nickname": ident, "password": "pw-1234-check"})
    psql(f"UPDATE users SET perms = '{{{'admin' if admin else 'worker_register'}}}' WHERE nickname = '{ident}'")
    s, b, _ = call(op, "POST", "/users/session", {"identification": ident, "password": "pw-1234-check"})
    assert s == 200, (s, b)
    return op


def upload(op, bundle, rel, data, overwrite=False, chunk=512 << 10):
    s, b, _ = call(op, "POST", f"{bundle}/uploads", {"path": rel, "size": len(data), "overwrite": overwrite})
    if s != 200:
        return s, js(b)
    uid = js(b)["id"]
    off = 0
    while True:
        part = data[off:off + chunk]
        s, b, _ = call(op, "PUT", f"/admin/updates/uploads/{uid}?offset={off}", raw=part)
        r = js(b)
        if s != 200 or r.get("done"):
            return s, r
        off = r["received"]


admin = login("upd-admin", admin=True)
other = login("upd-other")

s, _, _ = call(other, "GET", "/admin/updates/devices")
ok("관리자 아니면 403", s == 403, s)

# ---- 장비 루트 ----
s, _, _ = call(admin, "POST", "/admin/updates/devices", {"mainKey": "fa24-PID", "memo": "보드"})
ok("장비 루트 만들기", s == 200, s)
s, _, _ = call(admin, "POST", "/admin/updates/devices", {"mainKey": "fa24-PID"})
ok("같은 메인키 409", s == 409, s)
s, _, _ = call(admin, "POST", "/admin/updates/devices", {"mainKey": "bad name"})
ok("이름 규칙 400", s == 400, s)

# ---- 공용 묶음 ----
C = "/admin/updates/bundles/common/nexus-worker/v1"
s, _, _ = call(admin, "POST", "/admin/updates/bundles/common/nexus-worker", {"version": "v1"})
ok("공용 만들기", s == 200, s)
blob = os.urandom(1_300_000)
s, r = upload(admin, C, "worker", blob)
ok("조각 업로드(3조각)", s == 200 and r["file"]["sha"] == hashlib.sha256(blob).hexdigest() and not r["file"]["text"], (s, r))
s, r = upload(admin, C, "worker", b"x")
ok("같은 이름 업로드 = 409(덮어쓰기 확인)", s == 409 and r.get("exists"), (s, r))
s, b, _ = call(admin, "POST", f"{C}/uploads", {"path": "gap.bin", "size": 10})
uid = js(b)["id"]
s, b, _ = call(admin, "PUT", f"/admin/updates/uploads/{uid}?offset=5", raw=b"12345")
ok("건너뛴 offset = 409 + 받은 양", s == 409 and js(b)["received"] == 0, (s, b))
call(admin, "DELETE", f"/admin/updates/uploads/{uid}")
s, _, _ = call(admin, "PUT", f"{C}/text", {"path": "update.conf", "content": "a=1\n", "create": True})
ok("공용 텍스트 새 파일", s == 200, s)
s, _, _ = call(admin, "PUT", f"{C}/text", {"path": "update.conf", "content": "a=2\n", "create": True})
ok("새 파일인데 이름 겹침 409", s == 409, s)
s, _, _ = call(admin, "PUT", f"{C}/text", {"path": "../x", "content": "x"})
ok("번들 밖 경로 400", s == 400, s)

# ---- 빌드 ----
B = "/admin/updates/bundles/build/fa24-PID/b1"
s, _, _ = call(admin, "POST", "/admin/updates/bundles/build/nope", {"version": "b1"})
ok("없는 장비에 빌드 404", s == 404, s)
s, _, _ = call(admin, "POST", "/admin/updates/bundles/build/fa24-PID", {"version": "b1", "note": "첫 빌드"})
ok("빌드 만들기", s == 200, s)
s, b, _ = call(admin, "PUT", f"{B}/text", {"path": "00-stop.sh", "content": "#!/bin/sh\r\necho stop\r\n"})
ok(".sh 저장", s == 200, (s, b))
s, b, _ = call(admin, "GET", f"{B}/file?path=00-stop.sh")
ok(".sh는 LF로 저장", s == 200 and b == b"#!/bin/sh\necho stop\n", b)

bad = '{\n  "uses": ["nexus-worker@v1"],\n  "deploy": [\n    { "id": "stop", "script": "00-stop.sh" },\n    { "id": "copy", "builtin": "cpoy" }\n  ]\n}\n'
s, b, _ = call(admin, "PUT", f"{B}/text", {"path": "steps.json", "content": bad})
r = js(b)
ok("steps.json 오류 = 422 + 5행", s == 422 and any(i["line"] == 5 and i["level"] == "error" for i in r["issues"]), (s, r))
s, b, _ = call(admin, "POST", f"{B}/check", {"content": bad.replace("@v1", "@v9")})
r = js(b)
ok("검증만: 없는 공용 묶음 = 2행 오류", s == 200 and any(i["line"] == 2 for i in r["issues"]), r)
good = bad.replace("cpoy", "copy").replace('"copy" }', '"copy" },\n    { "id": "start", "script": "03-start.sh" }')
s, b, _ = call(admin, "PUT", f"{B}/text", {"path": "steps.json", "content": good})
r = js(b)
ok("steps.json 저장 + 없는 스크립트는 경고", s == 200 and any(i["level"] == "warn" and "03-start.sh" in i["msg"] for i in r["issues"]), (s, r))

s, b, _ = call(admin, "GET", B)
d = js(b)
ok("빌드 상세: 파일 2 · uses 1(공용 파일 2) · sha", s == 200 and len(d["files"]) == 2 and len(d["uses"]) == 1 and len(d["uses"][0]["files"]) == 2 and len(d["sha"]) == 64, d)
ok("빌드 상세: 텍스트 판정", all(f["text"] for f in d["files"]), d["files"])
ok("올린 사람 닉네임", d["createdByName"].startswith("upd-admin"), d["createdByName"])
sha1 = d["sha"]

s, _, _ = call(admin, "DELETE", C)
ok("쓰는 빌드가 있는 공용 삭제 409", s == 409, s)

# ---- 복사로 새 버전 ----
B2 = "/admin/updates/bundles/build/fa24-PID/b2"
s, _, _ = call(admin, "POST", "/admin/updates/bundles/build/fa24-PID", {"version": "b2", "from": {"owner": "fa24-PID", "version": "b1"}})
ok("빌드 복사", s == 200, s)
s, b, _ = call(admin, "GET", B2)
d2 = js(b)
ok("복사본 = 같은 sha · uses · 설명", d2["sha"] == sha1 and len(d2["uses"]) == 1 and d2["note"] == "첫 빌드", d2)
call(admin, "PUT", f"{B2}/text", {"path": "00-stop.sh", "content": "#!/bin/sh\necho b2\n"})
s, b, _ = call(admin, "GET", f"{B}/file?path=00-stop.sh")
ok("복사본을 고쳐도 원본 그대로(하드링크 분리)", b == b"#!/bin/sh\necho stop\n", b)
s, b, _ = call(admin, "GET", B2)
ok("고치면 sha 바뀜", js(b)["sha"] != sha1)

# ---- zip ----
s, b, h = call(admin, "GET", f"{B}/zip")
z = zipfile.ZipFile(io.BytesIO(b))
names = sorted(z.namelist())
ok("zip 범위별", names == ["common/nexus-worker@v1/update.conf", "common/nexus-worker@v1/worker", "main/00-stop.sh", "main/steps.json"], names)
ok("zip 내용 = 원본", z.read("common/nexus-worker@v1/worker") == blob)
call(admin, "PUT", f"{B}/text", {"path": "update.conf", "content": "a=main\n"})
s, b, _ = call(admin, "GET", f"{B}/zip?merged=1")
z = zipfile.ZipFile(io.BytesIO(b))
ok("zip 합친 결과: 메인키가 공용을 덮음", sorted(z.namelist()) == ["00-stop.sh", "steps.json", "update.conf", "worker"] and z.read("update.conf") == b"a=main\n", z.namelist())

# ---- 잠금 ----
psql("INSERT INTO update_locks (bundle_id, sha) SELECT id, sha FROM update_bundles WHERE owner='fa24-PID' AND version='b1'")
s, _, _ = call(admin, "PUT", f"{B}/text", {"path": "00-stop.sh", "content": "x"})
ok("잠긴 빌드 쓰기 409", s == 409, s)
s, _, _ = call(admin, "DELETE", f"{B}/file?path=00-stop.sh")
ok("잠긴 빌드 파일 삭제 409", s == 409, s)
s, _, _ = call(admin, "PATCH", B, {"note": "잠긴 뒤 설명"})
ok("잠긴 뒤 설명 고치기", s == 200, s)
s, b, _ = call(admin, "GET", "/admin/updates/devices/fa24-PID")
d = js(b)
ok("장비 상세: 빌드 2 · 잠김 표시", len(d["builds"]) == 2 and any(x["locked"] for x in d["builds"]), d)

# ---- 삭제 ----
s, _, _ = call(admin, "DELETE", f"{B2}/file?path=steps.json")
s, b, _ = call(admin, "GET", B2)
ok("steps.json 지우면 uses 비움", s == 200 and js(b)["uses"] == [], js(b))
s, _, _ = call(admin, "DELETE", "/admin/updates/devices/fa24-PID")
ok("빌드가 남은 장비 삭제 409", s == 409, s)
s, _, _ = call(admin, "DELETE", B)
ok("잠긴 빌드도 삭제 가능", s == 200, s)
call(admin, "DELETE", B2)
s, _, _ = call(admin, "DELETE", C)
ok("쓰는 빌드가 없어진 공용 삭제", s == 200, s)
s, _, _ = call(admin, "DELETE", "/admin/updates/devices/fa24-PID")
ok("빈 장비 삭제", s == 200, s)
s, b, _ = call(admin, "GET", "/admin/updates/devices")
ok("목록 비움", js(b)["total"] == 0, b)

print(f"\n{'OK' if fails == 0 else f'{fails} FAIL'}")
sys.exit(1 if fails else 0)
