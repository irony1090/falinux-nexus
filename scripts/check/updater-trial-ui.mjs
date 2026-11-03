// 시험 타일(배포 전 단계 시험) 확인 — 조회 API · 입구 · 자동 선택 · 단계 행 · 실행·로그 · 배포 기록과 구분 · 잠금 조건 · 잠긴 빌드 · 두 인스턴스.
// 실행 = run.sh(격리 supervisor가 BASE, UPDATER = updater 실행 파일). SHOT_DIR이 있으면 화면을 찍는다 — REF-update-trial.md 시험-5(확인)
import { execFileSync, spawn } from 'node:child_process'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
const { chromium } = createRequire(`${process.env.PW_DIR}/`)('playwright-core')
const BASE = process.env.BASE, DB = process.env.CHECK_DB, SHOT = process.env.SHOT_DIR, UPDATER = process.env.UPDATER

let fails = 0
const ok = (name, cond, extra = '') => { console.log(`${cond ? 'OK  ' : 'FAIL'} ${name} ${cond ? '' : extra}`); if (!cond) fails++ }
const sql = q => execFileSync('docker', ['exec', 'postgres15', 'psql', '-U', 'irony', '-d', DB, '-Atc', q]).toString().trim()
if (SHOT) mkdirSync(SHOT, { recursive: true })
const shot = async (p, name) => { if (SHOT) await p.screenshot({ path: join(SHOT, `${name}.png`) }) }
const until = async (fn, ms = 15000) => { const end = Date.now() + ms; for (;;) { const v = await fn(); if (v || Date.now() > end) return v; await new Promise(r => setTimeout(r, 150)) } }

const work = mkdtempSync(join(tmpdir(), 'upd-trial-ui-'))
const upds = []
const browser = await chromium.launch({ executablePath: process.env.CHROME })
try {
    const ctx = await browser.newContext({ viewport: { width: 1600, height: 900 } })
    const api = ctx.request
    const call = async (m, p, data) => {
        const r = await api[m](`${BASE}/api${p}`, data === undefined ? {} : { data })
        if (!r.ok()) throw new Error(`${m} ${p} ${r.status()} ${await r.text()}`)
        return r.json().catch(() => null)
    }
    const id = `tui${Date.now()}`, pw = 'Check1234!'
    await call('post', '/users', { identification: id, nickname: id, password: pw })
    sql(`update users set perms = '{admin}' where nickname = '${id}'`)
    await call('post', '/users/session', { identification: id, password: pw })

    // 장비 · 빌드 v1(script 1 + 내장 2) · v2(steps.json 없음)
    const MAIN = 'trial-jig'
    const { apiKey } = await call('post', '/admin/updates/devices', { mainKey: MAIN })
    const B = `/admin/updates/bundles/build/${MAIN}`
    await call('post', B, { version: 'v1' })
    await call('post', B, { version: 'v2' })
    const text = (path, content) => call('put', `${B}/v1/text`, { path, content })
    await text('app.conf', 'ver=1\n')
    await text('hello.sh', '#!/bin/sh\necho "hello-trial $STEP_ID"\n')
    await text('steps.json', JSON.stringify({
        deploy: [{ id: 'hello', script: 'hello.sh' }, { id: 'backup', builtin: 'backup' }, { id: 'copy', builtin: 'copy' }],
        rollback: [{ id: 'restore', builtin: 'restore' }],
    }, null, 2) + '\n')

    // updater = 폴더마다 하나(state.json에 서브키) — A는 처음부터, B는 두 인스턴스 검사 때
    const startUpdater = name => {
        const dir = join(work, name), target = join(dir, 'target')
        mkdirSync(target, { recursive: true })
        writeFileSync(join(dir, 'updater.json'), JSON.stringify({ supervisor: BASE.replace('http', 'ws'), mainKey: MAIN, apiKey, targetDir: target }))
        const p = spawn(UPDATER, ['-c', join(dir, 'updater.json')], { stdio: 'ignore' })
        upds.push(p)
        return p
    }
    const online = async () => (await call('get', `/admin/updates/devices/${MAIN}`)).instances.filter(i => i.online)
    startUpdater('a')
    const A = (await until(async () => (await online())[0]))?.subKey
    ok('updater A 접속', !!A)
    const trial = (ver, sub) => call('get', `${B}/${ver}/trial${sub ? `?sub=${sub}` : ''}`)

    // 검사-1(조회 API)
    const t0 = await trial('v1')
    ok('검사-1 조회 = 단계 정의 · sub 없으면 기록 없음', t0.stepsIssue === '' && t0.deploy.map(s => s.id).join() === 'hello,backup,copy' && t0.runs.length === 0, JSON.stringify(t0))
    const t2 = await trial('v2', A)
    ok('검사-1 steps.json 없음 = 이유 · 단계 없음', t2.stepsIssue === 'steps.json이 없습니다' && !t2.deploy.length, JSON.stringify(t2))

    const page = await ctx.newPage()
    const errors = []
    page.on('pageerror', e => errors.push(e.message))
    page.on('console', m => { if (m.type() === 'error' && !/Failed to load resource/.test(m.text())) errors.push(m.text()) })
    const tile = key => page.locator(`[data-tile-key="${key}"]`)
    const X = tile(`x${MAIN}~v1`)
    const selected = async () => (await X.locator('.v-select__selection').allTextContents()).join().trim()
    const note = async () => (await X.locator('.mb-foot').textContent()).trim()
    const row = sid => X.locator('.upd-row', { has: page.locator('.main', { hasText: new RegExp(`^${sid}$`) }) })

    // 검사-2(입구): 배포가 없어도 시험 타일이 열림
    await page.goto(`${BASE}/updater?t=b${MAIN}~v1*1x1`)
    const BT = tile(`b${MAIN}~v1`)
    const openRow = await until(async () => { const t = await BT.locator('.upd-set .ct-seg').allTextContents(); return t.length >= 4 && t })
    ok('검사-2 열기 줄 = Files·steps.json·시험·Run · Log', openRow && openRow.join('|') === 'Files|steps.json|시험|Run · Log', openRow)
    ok('검사-2 배포 전 = Run · Log 꺼짐', await BT.getByRole('button', { name: 'Run · Log' }).isDisabled())
    await BT.getByRole('button', { name: '시험' }).click()
    await X.waitFor()
    ok('검사-2 시험 타일 = 빌드 타일 자식 · 2칸', /x[^,]*\*1x2\^b/.test(decodeURIComponent(page.url())), page.url())

    // 검사-3(자동 선택)
    ok('검사-3 하나뿐인 온라인 인스턴스 자동 선택', !!await until(async () => (await selected()).startsWith(A)), await selected())
    ok('검사-3 안내 = 배포 아님', (await note()).includes('배포가 아닙니다'), await note())

    // 검사-4(단계 행)
    await until(async () => await X.locator('.upd-row').count() >= 3)
    ok('검사-4 내장 = 실행 없음 · 배포에서만', await row('backup').getByRole('button', { name: '실행' }).count() === 0 && (await row('copy').textContent()).includes('배포에서만'))
    ok('검사-4 script = 실행 켜짐', await row('hello').getByRole('button', { name: '실행' }).isEnabled())
    await shot(page, 'trial-1-open')

    // 검사-5(실행·로그)
    await row('hello').getByRole('button', { name: '실행' }).click()
    ok('검사-5 로그 = 스크립트 출력', !!await until(async () => (await X.locator('.upd-log .xterm-rows').textContent()).includes('hello-trial hello')))
    ok('검사-5 행 = 완료 · 기록 1줄', !!await until(async () => (await row('hello').textContent()).includes('완료') && (await X.locator('.mb-sec', { hasText: '시험 기록 1' }).count()) === 1))
    await shot(page, 'trial-2-run')

    // 검사-7(잠금 조건): 시험이 다 성공해도 잠그기 조건을 못 채움
    ok('검사-7 시험 성공 뒤에도 lockable 아님', (await call('get', `${B}/v1`)).lockable === false)

    // 검사-6(구분): 배포를 해도 시험 기록은 그대로
    const dep = (await call('post', `${B}/v1/deploy`, { subKeys: [A] })).items?.[0]
    const done = dep?.deployId && await until(async () => { const d = await call('get', `/admin/updates/deployments/${dep.deployId}`); return d.result !== 'running' && !d.busy && d }, 30000)
    ok('검사-6 배포 성공(같은 인스턴스)', done?.result === 'ok', JSON.stringify(done ?? dep))
    const t1 = await trial('v1', A)
    ok('검사-6 시험 기록 = 시험만(배포 단계 기록 안 섞임)', t1.runs.length === 1 && t1.runs[0].stepId === 'hello', JSON.stringify(t1.runs))

    // 검사-8(잠긴 빌드): 잠금 행을 바로 넣고(정식 잠그기 = updater-deploy-ui.mjs) 시험
    sql(`insert into update_locks (bundle_id, sha) select id, sha from update_bundles where kind = 'build' and owner = '${MAIN}' and version = 'v1'`)
    await page.reload()
    await X.waitFor()
    ok('검사-8 잠김 표시', !!await until(async () => (await X.textContent()).includes('잠김(시험 가능)')))
    await until(async () => (await selected()).startsWith(A))
    ok('검사-8 잠겨도 실행 켜짐', !!await until(async () => row('hello').getByRole('button', { name: '실행' }).isEnabled()))
    await row('hello').getByRole('button', { name: '실행' }).click()
    ok('검사-8 잠긴 빌드 시험 = 기록 2줄', !!await until(async () => (await X.locator('.mb-sec', { hasText: '시험 기록 2' }).count()) === 1))

    // 선택-1(두 인스턴스): 둘 다 온라인이면 자동 선택 안 함, 오프라인은 고를 수 없음
    const ub = startUpdater('b')
    const both = await until(async () => { const o = await online(); return o.length === 2 && o })
    const Bsub = both && both.find(i => i.subKey !== A).subKey
    ok('updater B 접속', !!Bsub)
    await page.reload()
    await X.waitFor()
    await until(async () => await X.locator('.upd-row').count() >= 3)
    await page.waitForTimeout(800)
    ok('두 인스턴스 = 자동 선택 안 함', (await selected()) === '' && (await note()).includes('대상 인스턴스를 고르세요'), `${await selected()} / ${await note()}`)
    ok('두 인스턴스 = 고르기 전 실행 꺼짐', await row('hello').getByRole('button', { name: '실행' }).isDisabled())
    ub.kill()
    await until(async () => (await online()).length === 1)
    await page.reload()
    await X.waitFor()
    await X.locator('.v-select').click()
    const item = sub => page.locator('.v-overlay--active .v-list-item', { hasText: sub })
    await item(A).waitFor()
    ok('오프라인 B = 고를 수 없음', await item(Bsub).evaluate(e => e.classList.contains('v-list-item--disabled')))
    await item(A).click()
    ok('A 고르면 실행 켜짐', !!await until(async () => row('hello').getByRole('button', { name: '실행' }).isEnabled()))
    await shot(page, 'trial-3-pick')

    ok('검사-9 브라우저 오류 없음', errors.length === 0, errors.join(' | '))
} catch (e) {
    ok('예외 없이 끝남', false, e.stack)
} finally {
    upds.forEach(p => p.kill())
    await browser.close()
    rmSync(work, { recursive: true, force: true })
}
console.log(fails ? `\n${fails} FAIL` : '\nOK')
process.exit(fails ? 1 : 0)
