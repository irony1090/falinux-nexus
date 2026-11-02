// memberDropOrder 전수 확인: 같은 좌표에서 순서가 왕복하지 않고, 한 번 정한 답이 그대로인지. 실행 = run.sh
import { memberDropOrder, type MemberFrame } from '../../apps/frontend/src/feature/member/util/memberDrop.util'
import { newTile, type MemberTile } from '../../apps/frontend/src/feature/member/util/memberGrid.util'

const SIZES = [[1, 1], [1, 2], [2, 1], [2, 2]] as const
const names = (o: readonly MemberTile[]) => o.map(t => t.key).join()
let runs = 0, bad = 0

for (const [cols, rows] of [[1, 2], [2, 2], [3, 2], [3, 1]] as const) {
    const w = 600, h = 500, gap = 8
    const f: MemberFrame = { gap, cols, rows, grids: [0, 1, 2].map(i => ({ x: i * (w + 16), y: 0, w, h })) }
    for (let m = 0; m < SIZES.length ** 4; m++) {
        let x = m
        const tiles = Array.from({ length: 4 }, (_, i) => {
            const s = SIZES[x % 4]!
            x >>= 2
            return { ...newTile(`m${i}`), w: s[0], h: s[1] }
        })
        for (let px = 5; px < 3 * (w + 16); px += 67) for (let py = 5; py < h; py += 53) for (const t of tiles) {
            runs++
            const once = memberDropOrder(tiles, t.key, { x: px, y: py }, f)
            const twice = memberDropOrder(once, t.key, { x: px, y: py }, f)
            if (names(once) !== names(twice) && bad++ < 3) console.log('왕복', { cols, rows, key: t.key, px, py, once: names(once), twice: names(twice) })
        }
    }
}
console.log(`member-drop: ${runs}회 중 왕복 ${bad}회`)
process.exit(bad ? 1 : 0)
