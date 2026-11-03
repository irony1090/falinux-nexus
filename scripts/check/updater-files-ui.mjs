// Files 타일 개선 브라우저 확인 — 단위-가(폴더 트리 + 선택 삭제) · 단위-나(끌어 놓기 + 폴더 이름 유지) · 단위-다(갈아치우기 토글). 실행 = run.sh(격리 supervisor가 BASE에 떠 있어야 함) → REF-update-files-tile.md
// SHOT_DIR이 있으면 화면을 그 폴더에 찍는다
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { createRequire } from 'node:module'
import { dirname, join } from 'node:path'
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
const id = `fil${Date.now()}`, pw = 'Check1234!'
const call = async (m, p, data) => {
    const r = await api[m](`${BASE}/api${p}`, data === undefined ? {} : { data })
    if (!r.ok()) throw new Error(`${m} ${p} ${r.status()} ${await r.text()}`)
    return r.json().catch(() => null)
}
await call('post', '/users', { identification: id, nickname: id, password: pw })
if (sql(`update users set perms = '{worker_register,admin}' where nickname = '${id}' returning id`) === '') throw new Error('권한 넣기 실패')
await call('post', '/users/session', { identification: id, password: pw })

const OWNER = 'files-PID', B = `/admin/updates/bundles/build/${OWNER}/f1`
const tree = ['00-stop.sh', 'conf/a.env', 'conf/b.env', 'web/index.html', 'web/favicon.txt', 'web/assets/app.js', 'web/assets/app.css']
const have = async () => (await call('get', B)).files.map(f => f.path).sort()

const page = await ctx.newPage()
const errors = []
page.on('pageerror', e => errors.push(e.message))
page.on('console', m => { if (m.type() === 'error' && !/Failed to load resource/.test(m.text())) errors.push(m.text()) })
const tile = key => page.locator(`[data-tile-key="${key}"]`)
const dialog = () => page.locator('.v-dialog.v-overlay--active')
const F = tile(`f${OWNER}~f1`)
const row = name => F.locator('.tree > .upd-row', { has: page.locator('.main', { hasText: new RegExp(`^${name.replace(/[.]/g, '\\.')}$`) }) })
const names = async () => (await F.locator('.tree > .upd-row .main').allTextContents()).join(',')
const hold = async (loc, ms = 650) => {
    await loc.scrollIntoViewIfNeeded()
    const b = await loc.locator('.txt').boundingBox()
    await page.mouse.move(b.x + 10, b.y + b.height / 2)
    await page.mouse.down()
    await page.waitForTimeout(ms)
    await page.mouse.up()
}

// PC 탐색기 끌어 놓기 흉내 — headless에선 폴더 entry를 못 만들어 파일만(끌기-3(자동 확인 범위)). over만 하고 drop은 따로
const dragOver = async (loc, files) => {
    const dt = await page.evaluateHandle(fs => { const d = new DataTransfer(); for (const n of fs) d.items.add(new File([`${n}\n`], n)); return d }, files)
    await loc.dispatchEvent('dragenter', { dataTransfer: dt })
    await loc.dispatchEvent('dragover', { dataTransfer: dt })
    return dt
}
const waitHave = async (pred, ms = 5000) => {
    for (const t0 = Date.now(); Date.now() - t0 < ms; await page.waitForTimeout(100)) if (pred(await have())) return true
    return false
}

const fillPrompt = async (vals, label) => {
    const d = dialog()
    await d.waitFor()
    const inputs = d.locator('input:not([type=hidden]):not([readonly])')
    for (const [i, v] of vals.entries()) if (v !== null) await inputs.nth(i).fill(v)
    await d.getByRole('button', { name: label }).click()
    await d.waitFor({ state: 'detached' }).catch(() => {})
}

// 장비·빌드 = 화면으로(updater-ui.mjs와 같은 흐름), 파일 = API로 넣은 뒤 Files 타일을 연다
await page.goto(`${BASE}/updater`)
await page.waitForSelector('.UpdList')
await page.locator('.UpdList button.add').click()
await fillPrompt([OWNER, ''], '추가')
await tile(`u${OWNER}`).waitFor()
await tile(`u${OWNER}`).getByRole('button', { name: '+ 새 빌드' }).click()
await fillPrompt(['f1'], '만들기')
await tile(`b${OWNER}~f1`).waitFor()
for (const p of tree) await call('put', `${B}/text`, { path: p, content: `${p}\n`, create: true })
await page.reload()   // 빌드 타일이 먼저 받아 둔 빈 목록 대신 새로 읽게
await tile(`b${OWNER}~f1`).waitFor()
await tile(`b${OWNER}~f1`).getByRole('button', { name: 'Files', exact: true }).click()
await F.waitFor()
await row('web/').waitFor()

// ---- 트리 ----
ok('트리 = 폴더마다 파일 먼저, 처음엔 모두 펼침', await names() === '00-stop.sh,conf/,a.env,b.env,web/,favicon.txt,index.html,assets/,app.css,app.js', await names())
ok('폴더 행 = 하위 파일 개수', (await row('web/').locator('.sub').textContent()) === '4개 파일')
ok('깊이 = 들여쓰기', await row('app.js').evaluate(e => getComputedStyle(e).getPropertyValue('--d').trim()) === '2')
await shot(page, 'f01-tree')
await row('conf/').click()
ok('폴더 누르기 = 접힘', await names() === '00-stop.sh,conf/,web/,favicon.txt,index.html,assets/,app.css,app.js', await names())
await row('conf/').click()
ok('다시 누르기 = 펼침', (await names()).includes('conf/,a.env,b.env'))
await row('index.html').click()
const T = page.locator(`[data-tile-key^="tb${OWNER}~f1~"]`)
await T.first().waitFor({ timeout: 3000 }).catch(() => {})
ok('파일 누르기 = Script 타일', await T.count() === 1 && (await T.textContent()).includes('web/index.html'))

// ---- 선택: 길게 누르기(마우스) -> 짧게 = 토글, 폴더 = 하위 전체 ----
await hold(row('index.html'))
ok('길게 누르기 = 선택 모드 + 그 행 체크', await F.locator('.cb').count() > 0 && await row('index.html').locator('.cb.all').count() === 1)
ok('선택 모드 = 행 버튼 숨김', await F.locator('.tree .ct-grp').count() === 0)
ok('푸터 = 1개 선택 · 삭제 1', (await F.locator('.mb-note').textContent()).includes('1개 선택') && await F.getByRole('button', { name: '삭제 1' }).isVisible())
await row('assets/').click()
ok('폴더 짧게 = 하위 전체 체크', await row('app.js').locator('.cb.all').count() === 1 && await row('app.css').locator('.cb.all').count() === 1)
ok('부모 폴더 = 일부 선택', await row('web/').locator('.cb.some').count() === 1)
ok('3개 선택', (await F.locator('.mb-note').textContent()).includes('3개 선택'))
await shot(page, 'f02-select')
await row('assets/').click()
ok('폴더 다시 = 하위 전체 해제', (await F.locator('.mb-note').textContent()).includes('1개 선택'))
await F.getByRole('button', { name: '전체 선택' }).click()
ok('전체 선택 = 7개', (await F.locator('.mb-note').textContent()).includes('7개 선택'))
await F.getByRole('button', { name: '취소' }).click()
ok('취소 = 선택 모드 끝, 버튼 돌아옴', await F.locator('.cb').count() === 0 && await F.locator('.tree .ct-grp').count() > 0)
ok('취소 뒤 파일 안 지워짐', (await have()).length === 7)

// 선택 삭제 — 확인창 1회, 개수 + 경로
await hold(row('a.env'))
await row('assets/').click()
await F.getByRole('button', { name: '삭제 3' }).click()
await dialog().waitFor()
const dtext = await dialog().textContent()
ok('삭제 확인창 = 개수 + 경로', dtext.includes('3개 파일') && dtext.includes('web/assets/app.js'), dtext)
await dialog().getByRole('button', { name: '삭제 3' }).click()
await row('assets/').waitFor({ state: 'detached', timeout: 5000 }).catch(() => {})
ok('선택 삭제 = 3개 지움', (await have()).join() === '00-stop.sh,conf/b.env,web/favicon.txt,web/index.html', (await have()).join())
ok('지운 뒤 선택 모드 끝', await F.locator('.cb').count() === 0)
ok('빈 폴더 없음(assets/ 사라짐)', await row('assets/').count() === 0)

// 폴더 삭제 버튼 = 하위 전체
await row('web/').locator('button[title^="폴더 삭제"]').click()
await dialog().waitFor()
ok('폴더 삭제 확인 = 2개', (await dialog().textContent()).includes('2개 파일'))
await dialog().getByRole('button', { name: '삭제 2' }).click()
await row('web/').waitFor({ state: 'detached', timeout: 5000 }).catch(() => {})
ok('폴더 삭제 = 하위 전부', (await have()).join() === '00-stop.sh,conf/b.env', (await have()).join())

// 한 개 삭제 = 예전 문구 그대로
await row('b.env').locator('button[title="삭제"]').click()
await dialog().waitFor()
ok('한 개 삭제 확인 = 예전 문구', (await dialog().textContent()).includes('이 빌드에서 파일을 지웁니다'))
await dialog().getByRole('button', { name: '취소' }).click()
await dialog().waitFor({ state: 'detached' }).catch(() => {})

// ---- 끌어 놓기(결정-5(끌어 놓기 규칙)): 폴더 행 = 그 폴더, 파일 행 = 든 폴더, 빈 곳 = 루트 ----
const tip = () => F.locator('.drop-tip')
let dt = await dragOver(row('conf/').locator('.txt'), ['x.env'])
ok('끄는 중 = 몸통 점선', await F.locator('.mb-body.dropping').count() === 1)
ok('끄는 중 = 놓일 폴더 행 강조 + 여기에 놓기', await row('conf/').evaluate(e => e.classList.contains('drop')) && (await row('conf/').locator('.tail').textContent()) === '여기에 놓기')
ok('끄는 중 = 안내 conf/ (이름 없이)', (await tip().textContent()).includes('conf/ 안으로 올라갑니다(폴더 이름 유지)'), await tip().textContent())
await shot(page, 'f05-drop')
await row('conf/').locator('.txt').dispatchEvent('drop', { dataTransfer: dt })
ok('폴더 행에 놓기 = 그 폴더 안', await waitHave(h => h.includes('conf/x.env')), (await have()).join())
ok('놓은 뒤 = 점선·안내 사라짐', await F.locator('.mb-body.dropping').count() === 0 && await tip().count() === 0)
dt = await dragOver(row('b.env').locator('.txt'), ['y.txt'])
ok('파일 행 위 = 든 폴더 행 강조', await row('conf/').evaluate(e => e.classList.contains('drop')))
await row('b.env').locator('.txt').dispatchEvent('drop', { dataTransfer: dt })
ok('파일 행에 놓기 = 든 폴더 안', await waitHave(h => h.includes('conf/y.txt')), (await have()).join())
dt = await dragOver(F.locator('.upd-fields'), ['z.txt'])
ok('빈 곳 = 안내 루트', (await tip().textContent()).includes('맨 위(루트)'))
await F.locator('.upd-fields').dispatchEvent('drop', { dataTransfer: dt })
ok('빈 곳에 놓기 = 루트', await waitHave(h => h.includes('z.txt')), (await have()).join())
dt = await dragOver(row('conf/').locator('.txt'), ['x.env'])
await row('conf/').locator('.txt').dispatchEvent('drop', { dataTransfer: dt })
await dialog().waitFor({ timeout: 3000 }).catch(() => {})
ok('겹치면 덮어쓰기 확인 1회', (await dialog().textContent().catch(() => '')).includes('1개 파일이 이미 있습니다'))
await dialog().getByRole('button', { name: '취소' }).click()
await dialog().waitFor({ state: 'detached' }).catch(() => {})
await F.getByRole('button', { name: /^공용/ }).click()
dt = await dragOver(F.locator('.upd-fields'), ['w.txt'])
ok('다른 탭 = 점선 없음', await F.locator('.mb-body.dropping').count() === 0)
await F.locator('.upd-fields').dispatchEvent('drop', { dataTransfer: dt })
await page.waitForTimeout(500)
ok('다른 탭 = 놓아도 안 올라감', !(await have()).includes('w.txt'))
await F.getByRole('button', { name: /^메인키/ }).click()

// 폴더 업로드 버튼 = 고른 폴더 이름 유지(결정-6(폴더 업로드 버튼))
const tmp = mkdtempSync(join(tmpdir(), 'upd-files-'))
for (const p of ['pkg/a.txt', 'pkg/sub/b.txt']) { mkdirSync(dirname(join(tmp, p)), { recursive: true }); writeFileSync(join(tmp, p), `${p}\n`) }
await F.locator('input[webkitdirectory]').setInputFiles(join(tmp, 'pkg'))
ok('폴더 업로드 = 폴더 이름 유지', await waitHave(h => h.includes('pkg/a.txt') && h.includes('pkg/sub/b.txt')), (await have()).join())
rmSync(tmp, { recursive: true, force: true })

// 열린 Script 타일 = 같은 이름 파일을 덮어쓰면 바로 다시 읽음(고치던 것 없음)
await row('a.txt').click()
const A = T.filter({ hasText: 'pkg/a.txt' })
await A.locator('.cm-content').waitFor()
ok('Script 타일 = 처음 내용', (await A.locator('.cm-content').textContent()).includes('pkg/a.txt'))
dt = await dragOver(row('pkg/').locator('.txt'), ['a.txt'])
await row('pkg/').locator('.txt').dispatchEvent('drop', { dataTransfer: dt })
await dialog().getByRole('button', { name: '덮어쓰기' }).click()
await page.waitForFunction(() => [...document.querySelectorAll('[data-tile-key^="tb"] .cm-content')].some(e => e.textContent === 'a.txt'), null, { timeout: 5000 }).catch(() => {})
ok('덮어쓰면 Script 타일 바로 반영', (await A.locator('.cm-content').textContent()) === 'a.txt', await A.locator('.cm-content').textContent())

// ---- 갈아치우기 토글(결정-15(토글 모양)): steps.json이 없으면 틀부터, 부모를 켜면 자식 항목은 지움, 자식은 켜진 채 흐림 ----
const rep = async () => (await call('get', B)).replace.join()
const seg = (name, label) => row(name).getByRole('button', { name: label, exact: true })
const waitRep = async want => { for (const t0 = Date.now(); Date.now() - t0 < 5000 && await rep() !== want; await page.waitForTimeout(100)); return await rep() === want }
await row('pkg/').waitFor()
ok('처음 = 합치기 켜짐', await seg('pkg/', '합치기').getAttribute('aria-pressed') === 'true' && await seg('pkg/', '갈아치우기').getAttribute('aria-pressed') === 'false')
await seg('conf/', '갈아치우기').click()
await dialog().waitFor()
ok('켤 때 확인 안내', (await dialog().textContent()).includes('conf/ 를 갈아치우기로 바꿀까요?'))
await dialog().getByRole('button', { name: '취소' }).click()
await dialog().waitFor({ state: 'detached' }).catch(() => {})
ok('취소 = 그대로(steps.json도 안 만듦)', await rep() === '' && !(await have()).includes('steps.json'))
await seg('sub/', '갈아치우기').click()
await dialog().getByRole('button', { name: '갈아치우기로 바꾸기' }).click()
ok('steps.json 없음 = 틀 만들고 replace 저장', await waitRep('pkg/sub') && (await have()).includes('steps.json'), await rep())
await row('steps.json').click()
const S = tile(`s${OWNER}~f1`), code = () => S.locator('.cm-content').textContent()
await S.locator('.cm-content').waitFor()
await seg('pkg/', '갈아치우기').click()
await dialog().getByRole('button', { name: '갈아치우기로 바꾸기' }).click()
ok('부모 켜기 = 자식 항목 지움', await waitRep('pkg'), await rep())
await page.waitForFunction(k => document.querySelector(`[data-tile-key="${k}"] .cm-content`)?.textContent.includes('"replace": ["pkg"]'), `s${OWNER}~f1`, { timeout: 5000 }).catch(() => {})
ok('열린 steps.json 타일 = 바로 반영', (await code()).includes('"replace": ["pkg"]'), await code())
await page.waitForFunction(() => document.querySelector('.mb-foot .mb-note')?.textContent.includes('갈아치우기 1곳'), null, { timeout: 3000 }).catch(() => {})
ok('켠 행 = 강조 + 이유 한 줄', await row('pkg/').evaluate(e => e.classList.contains('replace')) && (await row('pkg/').locator('.sub').textContent()).includes('pkg/ 를 비우고 채움'))
ok('자식 = 켜진 채 흐림', await seg('sub/', '갈아치우기').getAttribute('aria-pressed') === 'true' && await seg('sub/', '갈아치우기').isDisabled() && await seg('sub/', '합치기').isDisabled())
ok('푸터 = 갈아치우기 1곳', (await F.locator('.mb-note').textContent()).includes('갈아치우기 1곳'))
await row('pkg/').scrollIntoViewIfNeeded()
await shot(page, 'f06-replace')
// steps.json 타일에서 고치던 중이면 덮지 않고 안내 + 다시 불러오기
await S.locator('.cm-content').click()
await page.keyboard.press('Control+End')
await page.keyboard.type(' ')
await seg('pkg/', '합치기').click()
ok('합치기 = 확인 없이 끔', await waitRep('') && await dialog().count() === 0)
await S.getByRole('button', { name: '다시 불러오기' }).waitFor({ timeout: 5000 }).catch(() => {})
ok('고치던 중 = 덮지 않고 안내', (await code()).includes('"pkg"') && (await S.locator('.mb-note').textContent()).includes('다른 곳에서 steps.json이 바뀌었습니다'))
await S.getByRole('button', { name: '다시 불러오기' }).click()
await dialog().getByRole('button', { name: '버리고 불러오기' }).click()
await page.waitForFunction(k => !document.querySelector(`[data-tile-key="${k}"] .cm-content`)?.textContent.includes('replace'), `s${OWNER}~f1`, { timeout: 5000 }).catch(() => {})
ok('다시 불러오기 = 새 내용, 안내 사라짐', !(await code()).includes('replace') && await S.getByRole('button', { name: '다시 불러오기' }).count() === 0, await code())

// ---- 단위-라(실행 권한 이름): 글자 칸, 누르면 실행 비트만 ----
const modeOf = async p => (await call('get', B)).files.find(f => f.path === p).mode
const ex = row('00-stop.sh').getByRole('button', { name: '실행 권한', exact: true })
ok('실행 권한 칸 = .sh 켜짐', await ex.getAttribute('aria-pressed') === 'true' && await modeOf('00-stop.sh') === 0o755)
await ex.click()
for (const t0 = Date.now(); Date.now() - t0 < 5000 && await ex.getAttribute('aria-pressed') !== 'false'; await page.waitForTimeout(100));
ok('실행 권한 끄기 = 644, 칸 흐림', await modeOf('00-stop.sh') === 0o644 && await ex.getAttribute('aria-pressed') === 'false')
await shot(page, 'f07-exec')

// ---- 폰 폭 ----
await page.setViewportSize({ width: 400, height: 800 })
await page.waitForTimeout(300)
ok('폰 = 가로 넘침 없음', await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth))
await shot(page, 'f03-phone')
await page.setViewportSize({ width: 1600, height: 900 })

// ---- 잠긴 빌드 = 접기/펴기만(결정-7(잠긴 빌드)) ----
sql(`insert into update_locks (bundle_id, sha) select id, sha from update_bundles where owner='${OWNER}' and version='f1'`)
await page.reload()
await F.waitFor()
await row('conf/').waitFor()
await hold(row('00-stop.sh'))
ok('잠김 = 길게 눌러도 선택 모드 안 됨', await F.locator('.cb').count() === 0)
await row('conf/').click()
ok('잠김 = 접기는 됨', !(await names()).includes('b.env'))
dt = await dragOver(F.locator('.upd-fields'), ['v.txt'])
ok('잠김 = 끌어도 점선 없음', await F.locator('.mb-body.dropping').count() === 0)
await F.locator('.upd-fields').dispatchEvent('drop', { dataTransfer: dt })
await page.waitForTimeout(500)
ok('잠김 = 놓아도 안 올라감', !(await have()).includes('v.txt'))
ok('잠김 = 토글 두 칸 꺼짐', await seg('conf/', '합치기').isDisabled() && await seg('conf/', '갈아치우기').isDisabled())
ok('잠김 = 폴더 삭제 꺼짐', await row('conf/').locator('button[title^="폴더 삭제"]').isDisabled())
await page.waitForTimeout(200)   // 접힘 화살표 회전(0.12s)이 끝난 뒤 찍는다
await shot(page, 'f04-locked')

ok('콘솔·페이지 오류 없음', errors.length === 0, errors.join(' | '))
await browser.close()
console.log(fails ? `Files 타일 확인 실패 ${fails}건` : 'Files 타일 확인 통과')
process.exit(fails ? 1 : 0)
