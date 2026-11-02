-- supervisor(PostgreSQL) — 업데이트 app: 보관(빌드·공용 묶음·파일 메타)·updater 명부·잠금·배포 이력 → REF-update-deploy.md "DB"·"P1 착수 계획"
-- 파일 내용은 UPDATE_STORE_DIR(서버 파일시스템), 여기는 메타만. 관리자 전용이라 장비 주인(devices)과 연결하지 않는다

-- +goose Up
-- 장비 루트(메인키) — updater 첫 접속 때 자동 생성, P1은 화면에서 만든다
CREATE TABLE update_devices (
    main_key   TEXT        PRIMARY KEY,
    memo       TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- updater 인스턴스 명부(관측용, online은 메모리 파생). 서브키 = supervisor 발급
CREATE TABLE updaters (
    main_key          TEXT        NOT NULL REFERENCES update_devices(main_key) ON DELETE CASCADE,
    sub_key           TEXT        NOT NULL,
    label             TEXT        NOT NULL DEFAULT '',
    memo              TEXT        NOT NULL DEFAULT '',
    os                TEXT        NOT NULL DEFAULT '',
    arch              TEXT        NOT NULL DEFAULT '',
    install_dir       TEXT        NOT NULL DEFAULT '',
    installed_version TEXT        NOT NULL DEFAULT '', -- updater 보고값(마지막으로 끝까지 성공한 배포)
    last_seen_at      TIMESTAMPTZ NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (main_key, sub_key)
);

-- 빌드(kind=build, owner=메인키)와 공용 묶음(kind=common, owner=묶음 이름). sha = 폴더 sha(파일 sha에서 계산)
CREATE TABLE update_bundles (
    id         BIGSERIAL   PRIMARY KEY,
    kind       TEXT        NOT NULL CHECK (kind IN ('build', 'common')),
    owner      TEXT        NOT NULL,
    version    TEXT        NOT NULL,
    note       TEXT        NOT NULL DEFAULT '',
    sha        TEXT        NOT NULL DEFAULT '',
    created_by BIGINT      NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (kind, owner, version)
);

-- 파일 메타. scope = main(번들 본체) | sub(그 서브키 인스턴스에만, P1은 자리만)
CREATE TABLE update_files (
    bundle_id  BIGINT      NOT NULL REFERENCES update_bundles(id) ON DELETE CASCADE,
    scope      TEXT        NOT NULL CHECK (scope IN ('main', 'sub')),
    sub_key    TEXT        NOT NULL DEFAULT '',
    path       TEXT        NOT NULL,
    size       BIGINT      NOT NULL,
    sha        TEXT        NOT NULL,
    updated_by BIGINT      NULL REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bundle_id, scope, sub_key, path),
    CHECK ((scope = 'sub') = (sub_key <> ''))
);

-- 빌드가 쓰는 공용 묶음(steps.json uses). RESTRICT = 쓰는 빌드가 남아 있으면 공용 삭제 불가(REF-update-rules.md "삭제"). ord = uses 순서(깔 때 앞부터)
CREATE TABLE update_uses (
    bundle_id BIGINT NOT NULL REFERENCES update_bundles(id) ON DELETE CASCADE,
    common_id BIGINT NOT NULL REFERENCES update_bundles(id) ON DELETE RESTRICT,
    ord       INT    NOT NULL DEFAULT 0,
    PRIMARY KEY (bundle_id, common_id)
);
CREATE INDEX idx_update_uses_common ON update_uses(common_id);

-- 잠금 기록(해제 없음). 빌드를 잠글 때 빌드와 쓰는 공용 묶음마다 한 행, via = 잠그기를 누른 빌드
CREATE TABLE update_locks (
    id           BIGSERIAL   PRIMARY KEY,
    bundle_id    BIGINT      NOT NULL REFERENCES update_bundles(id) ON DELETE CASCADE,
    via_build_id BIGINT      NULL REFERENCES update_bundles(id) ON DELETE SET NULL,
    sha          TEXT        NOT NULL,
    locked_by    BIGINT      NULL REFERENCES users(id) ON DELETE SET NULL,
    locked_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_update_locks_bundle ON update_locks(bundle_id);

-- 배포 이력. 번들이 삭제돼도 남게 bundle_id는 SET NULL + 이름·sha 사본
CREATE TABLE update_deployments (
    id         BIGSERIAL   PRIMARY KEY,
    bundle_id  BIGINT      NULL REFERENCES update_bundles(id) ON DELETE SET NULL,
    main_key   TEXT        NOT NULL,
    sub_key    TEXT        NOT NULL,
    version    TEXT        NOT NULL,
    sha        TEXT        NOT NULL,
    result     TEXT        NOT NULL DEFAULT 'running' CHECK (result IN ('running', 'ok', 'fail', 'cancel')),
    started_by BIGINT      NULL REFERENCES users(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at   TIMESTAMPTZ NULL
);
CREATE INDEX idx_update_deployments_bundle ON update_deployments(bundle_id);
CREATE INDEX idx_update_deployments_instance ON update_deployments(main_key, sub_key);

-- 단계 실행 기록. deployment_id NULL = 개별 실행(잠금 판정에 안 들어감). 로그 = 요약만, 전체는 파일
CREATE TABLE update_step_runs (
    id            BIGSERIAL   PRIMARY KEY,
    deployment_id BIGINT      NULL REFERENCES update_deployments(id) ON DELETE CASCADE,
    bundle_id     BIGINT      NULL REFERENCES update_bundles(id) ON DELETE SET NULL,
    main_key      TEXT        NOT NULL,
    sub_key       TEXT        NOT NULL,
    version       TEXT        NOT NULL,
    list          TEXT        NOT NULL CHECK (list IN ('deploy', 'rollback')),
    step_id       TEXT        NOT NULL,
    step_kind     TEXT        NOT NULL CHECK (step_kind IN ('script', 'builtin')),
    result        TEXT        NOT NULL DEFAULT 'running' CHECK (result IN ('running', 'ok', 'fail', 'timeout', 'cancel')),
    exit_code     INT         NULL,
    log_tail      TEXT        NOT NULL DEFAULT '',
    log_file      TEXT        NOT NULL DEFAULT '',
    started_by    BIGINT      NULL REFERENCES users(id) ON DELETE SET NULL,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at      TIMESTAMPTZ NULL
);
CREATE INDEX idx_update_step_runs_deployment ON update_step_runs(deployment_id);
CREATE INDEX idx_update_step_runs_instance ON update_step_runs(main_key, sub_key);

-- +goose Down
DROP TABLE update_step_runs;
DROP TABLE update_deployments;
DROP TABLE update_locks;
DROP TABLE update_uses;
DROP TABLE update_files;
DROP TABLE update_bundles;
DROP TABLE updaters;
DROP TABLE update_devices;
