-- supervisor(PostgreSQL) — 업데이트 app P3(배포): 배포 = 단계 기록을 미리 만들고 updater가 차례로 실행 · 되돌리기 · 잠그기 판정 → REF-update-p3.md 단위-1(DB)

-- +goose Up
-- 되돌리기 = 같은 배포 안 rollback 목록(Run·Log 타일 하나에 두 목록). uses_sig = 배포 때 공용 묶음 sha 서명(잠그기 판정)
ALTER TABLE update_deployments
    ADD COLUMN rollback_result TEXT    NOT NULL DEFAULT '' CHECK (rollback_result IN ('', 'running', 'ok', 'fail', 'cancel')),
    ADD COLUMN auto_rollback   BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN uses_sig        TEXT    NOT NULL DEFAULT '',
    ADD COLUMN msg             TEXT    NOT NULL DEFAULT '';

-- wait = 배포가 미리 만든 아직 안 돈 단계, skip = 앞 단계가 실패·취소돼 돌지 않음
ALTER TABLE update_step_runs DROP CONSTRAINT update_step_runs_result_check;
ALTER TABLE update_step_runs ADD CONSTRAINT update_step_runs_result_check
    CHECK (result IN ('wait', 'running', 'ok', 'fail', 'timeout', 'cancel', 'skip'));

-- 빌드가 쓰는 공용 묶음의 지금 sha 서명(id 순) — 배포 때 저장한 값과 같아야 "같은 내용으로 성공"
-- +goose StatementBegin
CREATE FUNCTION update_uses_sig(p_bundle BIGINT) RETURNS TEXT LANGUAGE sql STABLE AS $$
    SELECT COALESCE(string_agg(c.id::text || ':' || c.sha, ',' ORDER BY c.id), '')
    FROM update_uses u JOIN update_bundles c ON c.id = u.common_id
    WHERE u.bundle_id = p_bundle
$$;
-- +goose StatementEnd

-- 잠그기 판정 한 곳(REF-update-rules.md "잠금"): 지금 빌드 sha·공용 서명과 같은 내용으로 끝까지 성공한 배포가 한 번이라도 있는가. 개별 실행은 안 셈
-- +goose StatementBegin
CREATE FUNCTION update_lockable(p_bundle BIGINT) RETURNS BOOLEAN LANGUAGE sql STABLE AS $$
    SELECT EXISTS (
        SELECT 1 FROM update_deployments d JOIN update_bundles b ON b.id = d.bundle_id
        WHERE d.bundle_id = p_bundle AND d.result = 'ok' AND d.sha = b.sha AND d.uses_sig = update_uses_sig(p_bundle)
    )
$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION update_lockable(BIGINT);
DROP FUNCTION update_uses_sig(BIGINT);
DELETE FROM update_step_runs WHERE result IN ('wait', 'skip');
ALTER TABLE update_step_runs DROP CONSTRAINT update_step_runs_result_check;
ALTER TABLE update_step_runs ADD CONSTRAINT update_step_runs_result_check
    CHECK (result IN ('running', 'ok', 'fail', 'timeout', 'cancel'));
ALTER TABLE update_deployments DROP COLUMN msg, DROP COLUMN uses_sig, DROP COLUMN auto_rollback, DROP COLUMN rollback_result;
