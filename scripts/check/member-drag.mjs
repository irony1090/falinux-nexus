// 내 장비·관리자·보드 끌기 + 색인 금지 브라우저 확인. 실행 = run.sh(격리 supervisor가 BASE에 떠 있어야 함)
import { execFileSync } from 'node:child_process'
import { createRequire } from 'node:module'
const { chromium } = createRequire(`${process.env.PW_DIR}/`)('playwright-core')
const BASE = process.env.BASE, DB = process.env.CHECK_DB

let fails = 0
const ok = (name, cond, extra = '') => { console.log(`${cond ? 'OK  ' : 'FAIL'} ${name} ${extra}`); if (!cond) fails++ }
const sql = q => execFileSync('docker', ['exec', 'postgres15', 'psql', '-U', 'irony', '-d', DB, '-Atc', q]).toString().trim()

const browser = await chromium.launch({ executablePath: process.env.CHROME })
const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } })
const api = ctx.request
const id = `mdrag${Date.now()}`, pw = 'Check1234!'
const call = async (m, p, data) => {
    const r = await api[m](`${BASE}/api${p}`, data === undefined ? {} : { data })
    if (!r.ok()) throw new Error(`${m} ${p} ${r.status()} ${await r.text()}`)
    return r.json().catch(() => null)
}
await call('post', '/users', { identification: id, nickname: id, password: pw })
if (sql(`update users set perms = '{worker_register,admin}', worker_limit = 20 where nickname = '${id}' returning id`) === '') throw new Error('권한 넣기 실패')
await call('post', '/users/session', { identification: id, password: pw })
for (let i = 1; i <= 8; i++) await call('post', '/devices', { mainKey: `dev-${i}` })

const page = await ctx.newPage()
const errors = []
page.on('pageerror', e => errors.push(e.message))
page.on('console', m => { if (m.type() === 'error' && !/Failed to load resource/.test(m.text())) errors.push(m.text()) })

const order = p => p.evaluate(() => [...document.querySelectorAll('[data-tile-key]')]
    .map(e => [+e.querySelector('.pos')?.textContent, e.dataset.tileKey]).sort((a, b) => a[0] - b[0]).map(x => x[1]).join())
const open = async (p, path, t) => {
    await p.goto(`${BASE}${path}?t=${encodeURIComponent(t)}`)
    await p.waitForSelector('[data-tile-key]')
    await p.waitForTimeout(400)
}
const box = (p, key) => p.locator(`[data-tile-key="${key}"]`).boundingBox()
const grab = async (p, key) => {
    const b = await p.locator(`[data-tile-key="${key}"] .pos`).boundingBox()
    await p.mouse.move(b.x + b.width / 2, b.y + b.height / 2)
    await p.mouse.down()
    await p.mouse.move(b.x + b.width / 2 + 10, b.y + b.height / 2 + 10, { steps: 3 })
}
const drop = async p => { await p.mouse.up(); await p.waitForTimeout(150) }

// 세로 2칸 타일 위/아래 절반(사용자 지적 1)
await open(page, '/devices', 'ddev-1*1x1,ddev-2*1x2')
let b = await box(page, 'ddev-2')
await grab(page, 'ddev-1')
await page.mouse.move(b.x + b.width / 2, b.y + b.height * 0.8, { steps: 8 })
ok('세로 2칸 아래 절반 = 뒤로', await order(page) === 'ddev-2,ddev-1', await order(page))
await drop(page)
b = await box(page, 'ddev-2')
await grab(page, 'ddev-1')
await page.mouse.move(b.x + b.width / 2, b.y + b.height * 0.2, { steps: 8 })
await drop(page)
ok('세로 2칸 위 절반 = 앞으로 돌아옴', await order(page) === 'ddev-1,ddev-2', await order(page))

// 반 칸 타일 아래 빈 칸(사용자 지적 2) — 3열이어야 끌 타일이 같은 Grid에 보인다
await page.setViewportSize({ width: 1700, height: 800 })
await open(page, '/devices', 'ddev-1*1x1,ddev-2*1x2,ddev-3*1x1')
b = await box(page, 'ddev-1')
await grab(page, 'ddev-3')
await page.mouse.move(b.x + b.width / 2, b.y + b.height + 60, { steps: 8 })
await drop(page)
ok('반 칸 타일 아래 빈 칸에 들어감', await order(page) === 'ddev-1,ddev-3,ddev-2', await order(page))
ok('URL t = 놓은 순서', new URL(page.url()).searchParams.get('t').split(',').map(x => x.split('*')[0]).join() === 'ddev-1,ddev-3,ddev-2')
await page.reload(); await page.waitForSelector('[data-tile-key]'); await page.waitForTimeout(300)
ok('새로고침 유지', await order(page) === 'ddev-1,ddev-3,ddev-2')
await page.setViewportSize({ width: 1280, height: 800 })

// 코드-4(넓은 타일 위 왕복): 같은 자리에서 조금씩 움직여도 순서가 한 가지
await open(page, '/devices', 'ddev-1*2x1,ddev-2*1x1,ddev-3*1x1')
b = await box(page, 'ddev-1')
await grab(page, 'ddev-2')
const seen = []
for (let i = 0; i < 12; i++) {
    await page.mouse.move(b.x + b.width * 0.75 + (i % 2), b.y + 20)
    await page.waitForTimeout(30)
    seen.push(await order(page))
}
ok('넓은 타일 위 = 순서 고정', new Set(seen.slice(1)).size === 1, [...new Set(seen)].join(' / '))
await page.keyboard.press('Escape')

// Esc 취소 + 크기 팝업 닫기(사용자 지적 3)
await open(page, '/devices', 'ddev-1*1x1,ddev-2*1x1')
await page.locator('[data-tile-key="ddev-1"] [aria-label="크기"]').click()
await page.waitForTimeout(250)
const menus = () => page.locator('.v-overlay--active.v-menu').count()
ok('크기 팝업 열림', await menus() === 1)
await grab(page, 'ddev-2')
await page.waitForTimeout(350)
ok('끌기 시작 = 팝업 닫힘', await menus() === 0)
await page.keyboard.press('Escape')
await page.mouse.up()
ok('Esc = 순서 그대로', await order(page) === 'ddev-1,ddev-2')

// 열기 규칙: 연 곳 없음 = 맨 앞
await open(page, '/devices', 'ddev-1*1x1,ddev-2*1x1')
await page.locator('.MemberList').getByText('dev-5', { exact: true }).first().click()
await page.waitForSelector('[data-tile-key="ddev-5"]')
ok('목록에서 연 타일 = 맨 앞', (await order(page)).startsWith('ddev-5,'), await order(page))

// 끝 띠: 넘김 + 코드-5(왼쪽 끝 띠 자리)
await open(page, '/devices', Array.from({ length: 7 }, (_, i) => `ddev-${i + 1}*1x1`).join(','))
const strip = await page.locator('.strip').boundingBox()
await grab(page, 'ddev-1')
await page.mouse.move(strip.x + strip.width - 15, strip.y + 200, { steps: 8 })
await page.waitForSelector('.PageEdgeBand.next')
ok('관리 페이지 오른쪽 끝 띠', true)
await page.waitForFunction(() => document.querySelector('.strip').scrollLeft > 100, null, { timeout: 4000 })
await page.mouse.move(strip.x + 200, strip.y + 200, { steps: 5 })
await page.waitForSelector('.PageEdgeBand.prev')
const prev = await page.locator('.PageEdgeBand.prev').boundingBox()
ok('왼쪽 끝 띠 = strip 왼쪽', Math.abs(prev.x - strip.x) <= 1, `band ${prev.x} / strip ${strip.x}`)
await page.keyboard.press('Escape'); await page.mouse.up()

// 코드-3(대신 등록 이름)
await open(page, '/admin', 'new*1x1')
await grab(page, 'new')
ok('관리자 칩 = 대신 등록', (await page.locator('.TileDragChip .name').first().textContent()) === '대신 등록')
await page.keyboard.press('Escape'); await page.mouse.up()

// 회원 -> 장비 자식: 연 곳 바로 뒤 + 끌어도 parent 유지
{
    const m = `m${sql(`select id from users where nickname = '${id}'`)}`
    await open(page, '/admin', `${m}*1x1,ddev-2*1x1`)
    await page.locator(`[data-tile-key="${m}"] button.mb-row`).filter({ hasText: 'dev-1' }).first().click()
    await page.waitForSelector('[data-tile-key="ddev-1"]')
    ok('회원에서 연 장비 = 회원 바로 뒤', await order(page) === `${m},ddev-1,ddev-2`, await order(page))
    ok('자식 표시', await page.locator('[data-tile-key="ddev-1"].child').count() === 1)
    ok('URL parent', new URL(page.url()).searchParams.get('t').includes(`ddev-1*1x2^${m}`), new URL(page.url()).searchParams.get('t'))
}

// 폰 터치: 길게 눌러 끌기
{
    const c2 = await browser.newContext({ viewport: { width: 400, height: 800 }, hasTouch: true, isMobile: true, storageState: await ctx.storageState() })
    const p = await c2.newPage()
    await open(p, '/devices', 'ddev-1*1x1,ddev-2*1x1')
    await p.evaluate(() => document.querySelector('.outer').scrollTo({ left: 1e4, behavior: 'instant' }))
    await p.waitForTimeout(300)
    await p.evaluate(() => document.querySelector('[data-tile-key="ddev-1"]').scrollIntoView({ block: 'nearest', inline: 'start', behavior: 'instant' }))
    await p.waitForTimeout(300)
    const a = await box(p, 'ddev-1'), z = await box(p, 'ddev-2')
    const cdp = await c2.newCDPSession(p)
    const touch = (type, x, y) => cdp.send('Input.dispatchTouchEvent', { type, touchPoints: type === 'touchEnd' ? [] : [{ x, y }] })
    const sx = a.x + a.width / 2, sy = a.y + 12
    await touch('touchStart', sx, sy)
    await p.waitForTimeout(650)
    await touch('touchMove', sx, sy + 40)
    await touch('touchMove', z.x + z.width / 2, z.y + z.height * 0.8)
    await p.waitForTimeout(100)
    await touch('touchEnd')
    await p.waitForTimeout(200)
    ok('폰 길게 눌러 끌기', await order(p) === 'ddev-2,ddev-1', `${await order(p)} tiles ${a.y}/${z.y}`)
    await c2.close()
}

// 코드-2(보드 끝 띠 사라짐)
{
    const tree = await call('get', '/tiles')
    const doc = tree.tree, ids = []
    for (let i = 1; i <= 8; i++) {
        const n = await call('post', '/nodes', { kind: 'FOLDER', parentId: null, name: `f${i}` })
        const tid = crypto.randomUUID()
        doc.tiles[tid] = { id: tid, type: 'folder', parent: doc.rootId, size: { w: 0.5, h: 1 }, nodeId: n.id }
        ids.push(tid)
    }
    doc.order = [doc.rootId, ...ids]
    await call('put', '/tiles', { tree: doc, version: tree.version })
    await page.goto(`${BASE}/`)
    await page.waitForSelector(`[data-tile-id="${ids[0]}"]`)
    await page.waitForTimeout(500)
    const h = await page.locator(`[data-tile-id="${ids[0]}"] .TileFrame__head`).boundingBox()
    await page.mouse.move(h.x + h.width * 0.45, h.y + h.height / 2)
    await page.mouse.down()
    await page.mouse.move(h.x + h.width * 0.45 + 12, h.y + 12, { steps: 3 })
    await page.mouse.move(1265, 400, { steps: 8 })
    const shown = await page.waitForSelector('.PageEdgeBand.next', { timeout: 2000 }).then(() => true, () => false)
    ok('보드 오른쪽 끝 띠 표시', shown)
    ok('보드 끌기 칩 표시', await page.locator('.TileDragChip').count() === 1)
    const moved = await page.waitForFunction(() => [...document.querySelectorAll('*')].some(e => e.scrollLeft > 100), null, { timeout: 4000 }).then(() => true, () => false)
    ok('보드 끝 띠 = 페이지 넘김', moved)
    await page.keyboard.press('Escape'); await page.mouse.up()
    await page.goto(`${BASE}/`)
    await page.waitForSelector(`[data-tile-id="${ids[1]}"]`)
    await page.waitForTimeout(400)
    // 보드 크기 팝업·메뉴: 끌기 시작하면 닫힘(77a9426 TileFrame)
    await page.locator(`[data-tile-id="${ids[0]}"] button[title="메뉴"]`).click()
    await page.waitForTimeout(250)
    const opened = await page.locator('.v-overlay--active.v-menu').count()
    await page.keyboard.press('Escape')
    await page.waitForTimeout(200)
    await page.locator(`[data-tile-id="${ids[0]}"] .TileFrame__head .tile-btn .mdi-resize`).click()
    await page.waitForTimeout(250)
    const sized = await page.locator('.v-overlay--active.v-menu').count()
    const h2 = await page.locator(`[data-tile-id="${ids[0]}"] .TileFrame__head`).boundingBox()
    await page.mouse.move(h2.x + h2.width * 0.45, h2.y + h2.height / 2)
    await page.mouse.down()
    await page.mouse.move(h2.x + h2.width * 0.45 + 40, h2.y + 40, { steps: 5 })
    await page.waitForTimeout(350)
    ok('보드 메뉴 열림·크기 팝업 열림', opened === 1 && sized === 1, `${opened}/${sized}`)
    ok('보드 끌기 시작 = 팝업 닫힘', await page.locator('.v-overlay--active.v-menu').count() === 0)
    await page.keyboard.press('Escape'); await page.mouse.up()
}

// 코드-6(색인 금지)·코드-7(robots.txt 상충)
{
    const robots = 'x-robots-tag'
    for (const path of ['/', '/w/none', '/api/nope']) {
        const r = await api.get(`${BASE}${path}`)
        ok(`X-Robots-Tag ${path}`, r.headers()[robots] === 'noindex, nofollow')
    }
    const r = await api.get(`${BASE}/robots.txt`)
    const t = await r.text()
    ok('robots.txt = 크롤 허용·text/plain', r.headers()['content-type'].startsWith('text/plain') && /Allow: \//.test(t) && !/Disallow/.test(t))
    await page.goto(`${BASE}/login`)
    ok('meta robots', await page.locator('meta[name="robots"]').getAttribute('content') === 'noindex, nofollow')
}

ok('콘솔·페이지 오류 없음', errors.length === 0, errors.slice(0, 3).join(' | '))
await browser.close()
console.log(fails ? `member-drag: 실패 ${fails}건` : 'member-drag: 전부 통과')
process.exit(fails ? 1 : 0)
