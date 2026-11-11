import { readUserPref, writeUserPref } from '@/common/util/userPrefs.util'
import { watch } from 'vue'
import { useRoute, useRouter, type LocationQuery, type LocationQueryRaw } from 'vue-router'
import { PAGE_TABS, type MemberTab, type provideMemberBoard } from '../store/memberBoard.store'
import { decodeTiles, encodeTiles } from '../util/memberGrid.util'

const SIZES = [10, 20, 50]

type Query = Record<string, string>
const str = (v: LocationQuery[string]) => typeof v === 'string' && v ? v : undefined
const norm = (q: LocationQuery): Query => Object.fromEntries(Object.entries(q).flatMap(([k, v]) => { const s = str(v); return s ? [[k, s]] : [] }))
const same = (a: Query, b: Query) => JSON.stringify(Object.entries(a).sort()) === JSON.stringify(Object.entries(b).sort())

// 관리 페이지 URL 저장 — 새로고침·뒤로 가기·보드 메뉴로 다시 들어와도 같은 화면. REF-member-ui.md 주소-1~주소-6(보드 tileUrl.hook과 같은 틀)
// ?focus = active 타일(들어올 때 이 타일로 이동). 2026-11-11 ?view(주소-4(보던 위치)) 폐지 — REF-column-flow.md 포커스-4(관리 페이지)
export const useMemberUrl = (ctx: ReturnType<typeof provideMemberBoard>) => {
    const { page, keys, tabs, tab, q, filter, pageNo, size, tiles, activeKey, plan, restore } = ctx
    const route = useRoute()
    const router = useRouter()
    const lastKey = `member.last.${page}`

    const current = (): Query => {
        const o: Query = {}
        if (tiles.value.length) o.t = encodeTiles(tiles.value)
        if (tab.value !== tabs[0]) o.tab = tab.value
        if (filter.value !== 'all') o.f = filter.value
        if (pageNo.value !== 1) o.p = String(pageNo.value)
        if (size.value !== 10) o.n = String(size.value)
        if (q.value) o.q = q.value
        if (activeKey.value && tiles.value.some(t => t.key === activeKey.value)) o.focus = activeKey.value
        return o
    }

    // 이 훅이 마지막으로 쓴 값 — route가 이것과 같으면 내 쓰기가 돌아온 것
    let written: Query | null = null
    // 들어온 값을 적용하는 동안엔 쓰지 않는다(덮어쓰면 복원할 값이 사라짐)
    let pending = false

    // 주소-1(진입만 push): 페이지 안에서 바뀌는 값은 전부 replace — 새로고침은 유지, 뒤로 가기는 바로 나감
    const write = () => {
        if (pending) return
        const next = current()
        const now = norm(route.query)
        written = next
        writeUserPref(lastKey, next)   // 주소-6(마지막 구성)
        if (!same(next, now)) void router.replace({ query: next as LocationQueryRaw })
    }

    const apply = (raw: Query) => {
        const t: MemberTab = tabs.includes(raw.tab as MemberTab) ? raw.tab as MemberTab : tabs[0]!
        tab.value = t
        filter.value = raw.f && PAGE_TABS[page][t]?.includes(raw.f) ? raw.f : 'all'
        const p = Number(raw.p)
        pageNo.value = Number.isInteger(p) && p > 0 ? p : 1
        const n = Number(raw.n)
        size.value = SIZES.includes(n) ? n : 10
        q.value = raw.q ?? ''
        // 회원 타일은 관리자 페이지에만
        const all = decodeTiles(raw.t, keys)
        tiles.value = page !== 'mine' ? all : all.filter(x => !x.key.startsWith('m')).map(x => x.parent ? { ...x, parent: null } : x)
    }

    // 주소-6: 쿼리 없이 들어오면(보드 메뉴) 마지막 구성으로
    const saved = readUserPref<Query | null>(lastKey, null)
    if (!Object.keys(norm(route.query)).length && saved && Object.keys(saved).length) void router.replace({ query: saved })

    // 판 측정 전이라 focus 이동을 미룬 URL이 있음 — 측정되면 그때 마저 적용
    let waitPlan = false
    const sync = async () => {
        const raw = norm(route.query)
        if (written && same(raw, written)) return
        pending = true
        apply(raw)
        waitPlan = !plan.value && !!raw.focus
        if (waitPlan) return     // 판 측정 전 — 스크롤할 자리가 없다
        // 들어오는 경로(새로고침·재진입·마지막 구성)뿐이라 늘 애니메이션 없이. 없는 타일이면 restore가 건너뛰고 아래 write가 지금 상태로 덮는다
        if (raw.focus) await restore(raw.focus)
        pending = false
        write()
    }
    // 원천은 route 글자열만 — Grid 유무를 원천에 넣으면 화면에서 연 첫 타일(push가 아직 안 돌아온 때)에
    // 옛 URL을 다시 적용해 타일을 지운다(폰에서 첫 타일로 안 넘어가던 원인, 2026-10-26)
    watch(() => JSON.stringify(norm(route.query)), sync, { immediate: true, flush: 'post' })
    watch(() => !!plan.value, measured => { if (measured && waitPlan) void sync() }, { flush: 'post' })

    watch(() => [tiles.value, activeKey.value, tab.value, filter.value, pageNo.value, size.value, q.value], () => write(), { flush: 'post' })
}
