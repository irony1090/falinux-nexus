// 업데이트 app P3(배포) 브라우저 확인 — 배포 팝업 · Run · Log(단계 상태·로그·입력) · 잠그기 · 되돌리기. 실행 = run.sh(격리 supervisor가 BASE, UPDATER = updater 실행 파일)
// SHOT_DIR이 있으면 화면을 찍는다 — REF-update-p3.md 단위-7(확인)
import { execFileSync, spawn } from 'node:child_process'
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
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

const work = mkdtempSync(join(tmpdir(), 'upd-dep-ui-'))
let upd
const browser = await chromium.launch({ executablePath: process.env.CHROME })
try {
    const ctx = await browser.newContext({ viewport: { width: 1600, height: 900 } })
    const api = ctx.request
    const call = async (m, p, data) => {
        const r = await api[m](`${BASE}/api${p}`, data === undefined ? {} : { data })
        if (!r.ok()) throw new Error(`${m} ${p} ${r.status()} ${await r.text()}`)
        return r.json().catch(() => null)
    }
    const id = `dui${Date.now()}`, pw = 'Check1234!'
    await call('post', '/users', { identification: id, nickname: id, password: pw })
    sql(`update users set perms = '{admin}' where nickname = '${id}'`)
    await call('post', '/users/session', { identification: id, password: pw })

    // 장비·빌드(API) + updater
    const MAIN = 'ui-jig'
    const { apiKey } = await call('post', '/admin/updates/devices', { mainKey: MAIN })
    const B = `/admin/updates/bundles/build/${MAIN}`
    await call('post', B, { version: 'v1' })
    const text = (path, content) => call('put', `${B}/v1/text`, { path, content })
    await text('app.conf', 'ver=1\n')
    await text('ask.sh', "#!/bin/sh\nprintf 'name? '\nread n\necho \"hi $n\"\n")
    await text('steps.json', JSON.stringify({
        deploy: [{ id: 'backup', builtin: 'backup' }, { id: 'copy', builtin: 'copy' }, { id: 'ask', script: 'ask.sh', interactive: true }],
        rollback: [{ id: 'restore', builtin: 'restore' }],
    }, null, 2) + '\n')
    const target = join(work, 'target')
    mkdirSync(target)
    writeFileSync(join(target, 'app.conf'), 'ver=0\n')
    writeFileSync(join(work, 'updater.json'), JSON.stringify({ supervisor: BASE.replace('http', 'ws'), mainKey: MAIN, apiKey, targetDir: target }))
    upd = spawn(UPDATER, ['-c', join(work, 'updater.json')], { stdio: 'ignore' })
    const inst = await until(async () => (await call('get', `/admin/updates/devices/${MAIN}`)).instances.find(i => i.online))
    ok('updater 접속', !!inst)

    const page = await ctx.newPage()
    const errors = []
    page.on('pageerror', e => errors.push(e.message))
    page.on('console', m => { if (m.type() === 'error' && !/Failed to load resource/.test(m.text())) errors.push(m.text()) })
    const tile = key => page.locator(`[data-tile-key="${key}"]`)
    const dialog = () => page.locator('.v-dialog.v-overlay--active')

    await page.goto(`${BASE}/updater`)
    await page.waitForSelector('.UpdList')
    await page.locator('.UpdList').getByText(MAIN, { exact: true }).first().click()
    const dev = tile(`u${MAIN}`)
    await dev.waitFor()
    ok('배포 전 Run · Log 비활성', await dev.locator('.upd-row.build').getByRole('button', { name: 'Run · Log' }).isDisabled())
    await dev.locator('.upd-row.build').getByRole('button', { name: '열기' }).click()
    const bt = tile(`b${MAIN}~v1`)
    await bt.waitFor()

    // 배포 팝업 — 붙어 있는 인스턴스 기본 체크
    await bt.getByRole('button', { name: '배포' }).click()
    const dd = page.locator('.UpdDeployDialog')
    await dd.waitFor()
    ok('배포 팝업 = 인스턴스 기본 체크', await dd.locator('input[type=checkbox]').first().isChecked())
    await shot(page, 'd01-dialog')
    await dd.getByRole('button', { name: '배포 1대' }).click()
    await dd.waitFor({ state: 'detached' }).catch(() => {})
    const dep = (await call('get', `/admin/updates/devices/${MAIN}`)).builds[0].lastDeployId
    const run = tile(`r${dep}`)
    await run.waitFor()
    const keys = await page.evaluate(() => [...document.querySelectorAll('[data-tile-key]')].map(e => e.dataset.tileKey))
    ok('Run · Log 타일 = 빌드 타일 바로 뒤', keys.indexOf(`r${dep}`) === keys.indexOf(`b${MAIN}~v1`) + 1, keys.join())

    // 입력 받는 단계 — 프롬프트가 로그에 보이고, 터미널에 치면 끝남
    const askRow = run.locator('.upd-row', { hasText: 'ask' })
    ok('ask 진행 중 표시', !!await until(async () => /진행 중.*입력 받음/.test(await askRow.textContent())), await askRow.textContent())
    ok('로그에 프롬프트', !!await until(async () => (await run.locator('.log .xterm-rows').textContent()).includes('name?')))
    await shot(page, 'd02-ask')
    await run.locator('.log .xterm').click()
    await page.keyboard.type('nexus')
    await page.keyboard.press('Enter')
    ok('입력 뒤 배포 성공', !!await until(async () => (await run.locator('.val .mb-tag').first().textContent()).includes('배포 성공')))
    const d1 = await call('get', `/admin/updates/deployments/${dep}`)
    const askRun = d1.runs.find(r => r.stepId === 'ask')
    ok('ask 로그 = 입력·출력', (await call('get', `/admin/updates/runs/${askRun.id}`)).log.includes('hi nexus'))
    ok('설치 경로에 깔림', readFileSync(join(target, 'app.conf'), 'utf8') === 'ver=1\n')
    ok('로그 끝까지 따라감(마지막 단계)', !!await until(async () => (await run.locator('.log .xterm-rows').textContent()).includes('hi nexus')))
    await shot(page, 'd03-done')

    // 잠그기 — 확인 창
    const lockBtn = run.locator('.mb-foot').getByRole('button', { name: '잠그기' })
    ok('잠그기 켜짐(성공 배포)', await lockBtn.isEnabled())
    await lockBtn.click()
    await dialog().waitFor()
    ok('잠그기 확인 창 = sha·공용 안내', (await dialog().textContent()).includes('같이 잠기는 공용 묶음'))
    await dialog().getByRole('button', { name: '잠그기' }).click()
    ok('잠김 표시', !!await until(async () => (await run.locator('.mb-foot').textContent()).includes('잠김')))
    ok('빌드 타일도 잠김', !!await until(async () => (await bt.textContent()).includes('잠김')))

    // 되돌리기 — rollback 목록으로 바뀌고 restore 완료
    await run.locator('.mb-foot').getByRole('button', { name: '되돌리기' }).click()
    await dialog().waitFor()
    await dialog().getByRole('button', { name: '되돌리기' }).click()
    ok('되돌리기 = rollback 목록 + 성공', !!await until(async () => (await run.locator('.val').filter({ hasText: '되돌리기' }).textContent().catch(() => '')).includes('되돌리기 성공')
        && await run.locator('.upd-set button[aria-pressed=true]').textContent() === 'rollback'))
    ok('되돌린 뒤 = 옛 내용', readFileSync(join(target, 'app.conf'), 'utf8') === 'ver=0\n')
    await shot(page, 'd04-rollback')

    // 장비 타일 Run · Log = 마지막 배포
    await run.locator('.mb-head button[title="닫기"]').click()
    await run.waitFor({ state: 'detached' })
    await dev.locator('.upd-row.build').getByRole('button', { name: 'Run · Log' }).click()
    ok('장비 타일 Run · Log = 마지막 배포', await run.waitFor().then(() => true, () => false))
    ok('브라우저 오류 없음', errors.length === 0, errors.join(' | '))
} finally {
    upd?.kill()
    await browser.close()
    rmSync(work, { recursive: true, force: true })
}
console.log(fails ? `\n${fails} FAIL` : '\nOK')
process.exit(fails ? 1 : 0)
