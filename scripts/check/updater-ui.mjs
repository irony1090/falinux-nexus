// 업데이트 app P1(보관·화면) 브라우저 확인 + 메인 보드 스크립트 편집(CtCode). 실행 = run.sh(격리 supervisor가 BASE에 떠 있어야 함)
// SHOT_DIR이 있으면 화면을 그 폴더에 찍는다(눈으로 확인용) — REF-update-deploy.md "P1 착수 계획" 단위-8(확인)
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
const { chromium } = createRequire(`${process.env.PW_DIR}/`)('playwright-core')
const BASE = process.env.BASE, DB = process.env.CHECK_DB, SHOT = process.env.SHOT_DIR

let fails = 0
const ok = (name, cond, extra = '') => { console.log(`${cond ? 'OK  ' : 'FAIL'} ${name} ${cond ? '' : extra}`); if (!cond) fails++ }
const sql = q => execFileSync('docker', ['exec', 'postgres15', 'psql', '-U', 'irony', '-d', DB, '-Atc', q]).toString().trim()
if (SHOT) mkdirSync(SHOT, { recursive: true })
const shot = async (p, name) => { if (SHOT) await p.screenshot({ path: join(SHOT, `${name}.png`) }) }

const browser = await chromium.launch({ executablePath: process.env.CHROME })
const ctx = await browser.newContext({ viewport: { width: 1600, height: 900 } })
const api = ctx.request
const id = `upd${Date.now()}`, pw = 'Check1234!'
const call = async (m, p, data) => {
    const r = await api[m](`${BASE}/api${p}`, data === undefined ? {} : { data })
    if (!r.ok()) throw new Error(`${m} ${p} ${r.status()} ${await r.text()}`)
    return r.json().catch(() => null)
}
await call('post', '/users', { identification: id, nickname: id, password: pw })
if (sql(`update users set perms = '{worker_register,admin}' where nickname = '${id}' returning id`) === '') throw new Error('권한 넣기 실패')
await call('post', '/users/session', { identification: id, password: pw })

// 공용 묶음은 API로 미리(화면 확인은 빌드 쪽)
await call('post', '/admin/updates/bundles/common/nexus-worker', { version: 'v1' })
await call('put', '/admin/updates/bundles/common/nexus-worker/v1/text', { path: 'update.conf', content: 'a=1\n' })

const page = await ctx.newPage()
const errors = []
page.on('pageerror', e => errors.push(e.message))
page.on('console', m => { if (m.type() === 'error' && !/Failed to load resource/.test(m.text())) errors.push(m.text()) })
const tile = key => page.locator(`[data-tile-key="${key}"]`)
const dialog = () => page.locator('.v-dialog.v-overlay--active')
const fillPrompt = async (vals, ok) => {
    const d = dialog()
    await d.waitFor()
    const inputs = d.locator('input:not([type=hidden]):not([readonly])')
    for (const [i, v] of vals.entries()) if (v !== null) await inputs.nth(i).fill(v)
    await d.getByRole('button', { name: ok }).click()
    await d.waitFor({ state: 'detached' }).catch(() => {})
}
const typeCode = async (root, text) => {
    await root.locator('.cm-content').click()
    await page.keyboard.press('Control+A')
    await page.keyboard.insertText(text)
}

// ---- 장비 -> 빌드 -> Files ----
await page.goto(`${BASE}/updater`)
await page.waitForSelector('.UpdList')
await shot(page, '01-empty')
await page.locator('.UpdList button.add').click()
await fillPrompt(['fa24-PID', '보드 1'], '추가')
await tile('ufa24-PID').waitFor()
ok('장비 추가 = 장비 타일 열림', await tile('ufa24-PID').isVisible())
await tile('ufa24-PID').getByRole('button', { name: '+ 새 빌드' }).click()
await fillPrompt(['b1'], '만들기')
await tile('bfa24-PID~b1').waitFor()
ok('새 빌드 = 빌드 타일이 장비 바로 뒤', (await page.evaluate(() => [...document.querySelectorAll('[data-tile-key]')].map(e => e.dataset.tileKey))).join() === 'ufa24-PID,bfa24-PID~b1')
ok('빌드 열기 줄 = Files·steps.json·시험·Run · Log(Script 없음)', (await tile('bfa24-PID~b1').locator('.upd-set .ct-seg').allTextContents()).join('|') === 'Files|steps.json|시험|Run · Log')
await tile('bfa24-PID~b1').getByRole('button', { name: 'Files' }).click()
const F = tile('ffa24-PID~b1')
await F.waitFor()

// 업로드(파일 둘) + 폴더 업로드(폴더 이름 유지, 결정-6(폴더 업로드 버튼))
const bin = Buffer.alloc(700 * 1024, 7)
await F.locator('input[type=file]:not([webkitdirectory])').setInputFiles([
    { name: 'worker', mimeType: 'application/octet-stream', buffer: bin },
    { name: '00-stop.sh', mimeType: 'text/plain', buffer: Buffer.from('#!/bin/sh\r\necho stop\r\n') },
])
await F.locator('.upd-row', { hasText: '00-stop.sh' }).waitFor()
const dir = mkdtempSync(join(tmpdir(), 'upd-ui-'))
mkdirSync(join(dir, 'release', 'conf'), { recursive: true })
writeFileSync(join(dir, 'release', 'conf', 'app.env'), 'X=1\n')
writeFileSync(join(dir, 'release', '03-start.sh'), '#!/bin/sh\necho start\n')
await F.locator('input[webkitdirectory]').setInputFiles(join(dir, 'release'))
let paths = ''
for (const t0 = Date.now(); Date.now() - t0 < 5000 && paths.split(',').length < 4; await page.waitForTimeout(100))
    paths = (await call('get', '/admin/updates/bundles/build/fa24-PID/b1')).files.map(f => f.path).sort().join()
ok('업로드·폴더 업로드 = 경로 그대로(폴더 이름 유지)', paths === '00-stop.sh,release/03-start.sh,release/conf/app.env,worker', paths)
ok('바이너리 표시', await F.locator('.upd-row', { hasText: 'worker' }).locator('.sub').textContent().then(t => t.includes('바이너리')))
// 같은 이름 다시 = 확인 창
await F.locator('input[type=file]:not([webkitdirectory])').setInputFiles([{ name: 'worker', mimeType: 'application/octet-stream', buffer: Buffer.from('x') }])
await dialog().waitFor()
ok('같은 이름 업로드 = 덮어쓰기 확인', (await dialog().textContent()).includes('이미 있습니다'))
await dialog().getByRole('button', { name: '취소' }).click()
await shot(page, '02-files')

// ---- Script 타일: 목록에서 텍스트 파일 누름 -> 편집·저장 ----
await F.locator('.upd-row', { hasText: '00-stop.sh' }).click()
const T = tile('tbfa24-PID~b1~00-stop.sh')
await T.waitFor()
await T.locator('.cm-content').waitFor()
ok('Script 타일 = 줄 끝 표시', /CRLF · 2줄/.test(await T.locator('.upd-hint').textContent()), await T.locator('.upd-hint').textContent())
await T.locator('.cm-line').nth(1).click()
await page.keyboard.press('End')
await page.keyboard.type('\nexit 0')
ok('고치면 수정됨 표시', (await T.locator('.upd-hint').textContent()).includes('수정됨'))
// 닫기 확인(저장 안 한 편집)
await T.locator('.mb-head button[title="닫기"]').click()
await dialog().waitFor()
ok('저장 안 하고 닫기 = 확인 창', (await dialog().textContent()).includes('저장하지 않은'))
await dialog().getByRole('button', { name: '취소' }).click()
await page.waitForTimeout(200)
ok('취소 = 타일 남음', await T.isVisible())
await T.locator('.cm-content').click()
await page.keyboard.press('Control+S')
await T.locator('.upd-hint:not(:has-text("수정됨"))').waitFor({ timeout: 3000 }).catch(() => {})
const saved = await (await api.get(`${BASE}/api/admin/updates/bundles/build/fa24-PID/b1/file?path=00-stop.sh`)).text()
ok('Ctrl+S 저장 + .sh는 LF', saved === '#!/bin/sh\necho stop\nexit 0\n', JSON.stringify(saved))
await shot(page, '03-script')

// ---- steps.json: 틀 -> 오류(줄) -> 고쳐 저장 ----
await tile('bfa24-PID~b1').getByRole('button', { name: 'steps.json' }).click()
const S = tile('sfa24-PID~b1')
await S.locator('.cm-content').waitFor()
const bad = '{\n  "uses": ["nexus-worker@v1"],\n  "deploy": [\n    { "id": "stop", "script": "00-stop.sh" },\n    { "id": "copy", "builtin": "cpoy" }\n  ]\n}\n'
await typeCode(S, bad)
await S.getByRole('button', { name: '저장' }).click()
await S.locator('.issues li').first().waitFor()
ok('steps.json 오류 = 5행 표시', (await S.locator('.issues').textContent()).includes('5행'))
ok('오류 밑줄', await S.locator('.cm-lintRange-error').count() > 0)
await shot(page, '04-steps-error')
await typeCode(S, bad.replace('cpoy', 'copy'))
await S.getByRole('button', { name: '저장' }).click()
await S.locator('.mb-note', { hasText: '오류 0건' }).waitFor({ timeout: 3000 }).catch(() => {})
ok('고쳐 저장 = 오류 0건', (await S.locator('.mb-note').textContent()).includes('오류 0건'))
await tile('bfa24-PID~b1').locator('.ct-seg', { hasText: 'nexus-worker@v1' }).waitFor({ timeout: 3000 }).catch(() => {})
ok('빌드 타일 uses 갱신', await tile('bfa24-PID~b1').locator('.ct-seg', { hasText: 'nexus-worker@v1' }).count() === 1)

// ---- 공용 탭(읽기 전용) · 공용 묶음 열기 ----
await F.getByRole('button', { name: /^공용/ }).click()
ok('Files 공용 탭 = 공용 파일', await F.locator('.upd-row', { hasText: 'update.conf' }).count() === 1)
await F.getByRole('button', { name: '공용 묶음 열기' }).click()
await tile('cnexus-worker~v1').waitFor()
ok('공용 묶음 타일 = 쓰는 빌드 1·삭제 불가', await tile('cnexus-worker~v1').locator('.mb-foot .mb-btn--danger').isDisabled())

// ---- 전체 받기 메뉴 · 복사로 새 버전 ----
await tile('bfa24-PID~b1').getByRole('button', { name: '전체 받기' }).click()
await page.locator('.UpdZipMenu').waitFor()
ok('전체 받기 = 범위별·합친 결과', (await page.locator('.UpdZipMenu').textContent()).includes('합친 결과'))
await shot(page, '05-zip')
await page.keyboard.press('Escape')
await tile('bfa24-PID~b1').getByRole('button', { name: '새 버전' }).click()
await fillPrompt(['b2'], '만들기')
await tile('bfa24-PID~b2').waitFor()
const b2 = await call('get', '/admin/updates/bundles/build/fa24-PID/b2')
ok('복사로 새 버전 = 파일·uses 그대로', b2.files.length === 5 && b2.uses.length === 1, JSON.stringify(b2.files.map(f => f.path)))

// ---- 새로고침 유지 · 폰 ----
const keys = () => page.evaluate(() => [...document.querySelectorAll('[data-tile-key]')].map(e => e.dataset.tileKey).sort().join())
const before = await keys()
await page.reload()
await page.waitForSelector('[data-tile-key]')
await page.waitForTimeout(500)
ok('새로고침해도 타일 유지', await keys() === before, `${before} / ${await keys()}`)
await shot(page, '06-reload')
await page.setViewportSize({ width: 400, height: 800 })
await page.waitForTimeout(400)
await shot(page, '07-phone')
ok('폰 = 가로 넘침 없음', await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth))
await page.setViewportSize({ width: 1600, height: 900 })

// ---- 잠긴 빌드 = 고칠 수 없음 ----
sql(`insert into update_locks (bundle_id, sha) select id, sha from update_bundles where owner='fa24-PID' and version='b1'`)
await page.reload()
await tile('ffa24-PID~b1').waitFor()
await page.waitForTimeout(500)
ok('잠긴 빌드 = 업로드 꺼짐', await tile('ffa24-PID~b1').getByRole('button', { name: '업로드', exact: true }).isDisabled())

// ---- 메인 보드 스크립트 편집(CtCode) ----
{
    const tree = await call('get', '/tiles')
    const doc = tree.tree
    const f = await call('post', '/nodes', { kind: 'FOLDER', parentId: null, name: 'upd-f' })
    const s = await call('post', '/nodes', { kind: 'SCRIPT', parentId: f.id, name: 'hello.sh' })
    await call('patch', `/nodes/${s.id}`, { content: { valid: true, value: '#!/bin/sh\necho hi\n' } })
    const tid = crypto.randomUUID()
    doc.tiles[tid] = { id: tid, type: 'folder', parent: doc.rootId, size: { w: 0.5, h: 1 }, nodeId: f.id }
    doc.order = [doc.rootId, tid]
    await call('put', '/tiles', { tree: doc, version: tree.version })
    await page.goto(`${BASE}/`)
    await page.waitForSelector(`[data-tile-id="${tid}"]`)
    await page.locator(`[data-tile-id="${tid}"]`).getByText('hello.sh').click({ button: 'right' })
    await page.getByText('내용 편집').click()
    const D = page.locator('.ScriptEditDialog')
    await D.locator('.cm-content').waitFor()
    await page.waitForTimeout(300)
    ok('스크립트 편집 = CtCode(줄 번호)', await D.locator('.cm-lineNumbers').count() === 1)
    ok('불러온 내용', (await D.locator('.cm-content').textContent()).includes('echo hi'))
    ok('다 불러온 뒤 편집기에 포커스', await page.evaluate(() => !!document.activeElement?.closest('.cm-editor')))
    await page.keyboard.press('Control+End')
    await page.keyboard.press('Tab')
    await page.keyboard.type('done')
    await shot(page, '08-script-dialog')
    await page.keyboard.press('Control+S')
    await page.waitForTimeout(500)
    const n = await call('get', `/nodes/${s.id}`)
    ok('Tab = 탭 문자 · Ctrl+S 저장', n.content.endsWith('\tdone'), JSON.stringify(n.content))
}

ok('콘솔·페이지 오류 없음', errors.length === 0, errors.join(' | '))
await browser.close()
console.log(fails ? `updater-ui: ${fails}개 실패` : 'updater-ui: 전부 통과')
process.exit(fails ? 1 : 0)
