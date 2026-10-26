import { readUserPref, writeUserPref } from '@/common/util/userPrefs.util'
import { onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter, type LocationQuery, type LocationQueryRaw } from 'vue-router'
import type { MemberTab, provideMemberBoard } from '../store/memberBoard.store'
import { decodeTiles, encodeTiles } from '../util/memberGrid.util'

// 보던 위치는 스크롤이 멈춘 뒤에 쓴다
const VIEW_DELAY = 300
const SIZES = [10, 20, 50]
const FILTERS: Record<MemberTab, string[]> = { users: ['all', 'admin', 'suspended', 'noreg'], devices: ['all', 'on', 'off', 'none'] }

type Query = Record<string, string>
const str = (v: LocationQuery[string]) => typeof v === 'string' && v ? v : undefined
const norm = (q: LocationQuery): Query => Object.fromEntries(Object.entries(q).flatMap(([k, v]) => { const s = str(v); return s ? [[k, s]] : [] }))
const same = (a: Query, b: Query) => JSON.stringify(Object.entries(a).sort()) === JSON.stringify(Object.entries(b).sort())
// 기록이 쌓이는 값 — 주소-1(push): 타일 구성 t · 목록 종류 tab · 필터 f. 나머지(p·n·q·view)는 replace(주소-2)
const pushPart = (q: Query) => JSON.stringify([q.t, q.tab, q.f])

// 관리 페이지 URL 저장 — 새로고침·뒤로 가기·보드 메뉴로 다시 들어와도 같은 화면. REF-member-ui.md 주소-1~주소-6(보드 tileUrl.hook과 같은 틀)
export const useMemberUrl = (ctx: ReturnType<typeof provideMemberBoard>) => {
    const { page, admin, tab, q, filter, pageNo, size, tiles, grids, anchorKey, restoreView, replaceNext } = ctx
    const route = useRoute()
    const router = useRouter()
    const lastKey = `member.last.${page}`

    const current = (): Query => {
        const o: Query = {}
        if (tiles.value.length) o.t = encodeTiles(tiles.value)
        if (admin && tab.value !== 'users') o.tab = tab.value
        if (filter.value !== 'all') o.f = filter.value
        if (pageNo.value !== 1) o.p = String(pageNo.value)
        if (size.value !== 10) o.n = String(size.value)
        if (q.value) o.q = q.value
        if (anchorKey.value) o.view = anchorKey.value
        return o
    }

    // 이 훅이 마지막으로 쓴 값 — route가 이것과 같으면 내 쓰기가 돌아온 것
    let written: Query | null = null
    // 들어온 값을 적용하는 동안엔 쓰지 않는다(덮어쓰면 복원할 값이 사라짐)
    let pending = false
    let viewTimer: ReturnType<typeof setTimeout> | undefined

    const write = (force?: 'replace') => {
        if (pending) return
        const next = current()
        const now = norm(route.query)
        written = next
        writeUserPref(lastKey, next)   // 주소-6(마지막 구성)
        if (same(next, now)) return
        const push = !force && !replaceNext.value && pushPart(next) !== pushPart(now)
        replaceNext.value = false
        void (push ? router.push({ query: next as LocationQueryRaw }) : router.replace({ query: next as LocationQueryRaw }))
    }

    const apply = (raw: Query) => {
        const t: MemberTab = admin && raw.tab === 'devices' ? 'devices' : admin ? 'users' : 'devices'
        tab.value = t
        filter.value = raw.f && FILTERS[t].includes(raw.f) ? raw.f : 'all'
        const p = Number(raw.p)
        pageNo.value = Number.isInteger(p) && p > 0 ? p : 1
        const n = Number(raw.n)
        size.value = SIZES.includes(n) ? n : 10
        q.value = raw.q ?? ''
        // 회원 타일은 관리자 페이지에만
        tiles.value = decodeTiles(raw.t).filter(x => admin || !x.key.startsWith('m'))
            .map(x => admin || !x.parent ? x : { ...x, parent: null })
    }

    // 주소-6: 쿼리 없이 들어오면(보드 메뉴) 마지막 구성으로
    const saved = readUserPref<Query | null>(lastKey, null)
    if (!Object.keys(norm(route.query)).length && saved && Object.keys(saved).length) void router.replace({ query: saved })

    // Grid 측정 전이라 보던 자리 복원을 미룬 URL이 있음 — 측정되면 그때 마저 적용
    let waitGrid = false
    const sync = async () => {
        const raw = norm(route.query)
        if (written && same(raw, written)) return
        pending = true
        apply(raw)
        waitGrid = !grids.value.length && !!raw.t
        if (waitGrid) return     // Grid 측정 전 — 스크롤할 자리가 없다
        if (raw.view) await restoreView(raw.view)
        pending = false
        write('replace')
    }
    // 원천은 route 글자열만 — Grid 유무를 원천에 넣으면 화면에서 연 첫 타일(push가 아직 안 돌아온 때)에
    // 옛 URL을 다시 적용해 타일을 지운다(폰에서 첫 타일로 안 넘어가던 원인, 2026-10-26)
    watch(() => JSON.stringify(norm(route.query)), sync, { immediate: true, flush: 'post' })
    watch(() => grids.value.length > 0, measured => { if (measured && waitGrid) void sync() }, { flush: 'post' })

    watch(() => [tiles.value, tab.value, filter.value, pageNo.value, size.value, q.value], () => write(), { flush: 'post' })
    watch(anchorKey, () => {
        clearTimeout(viewTimer)
        viewTimer = setTimeout(() => write(), VIEW_DELAY)
    })
    onBeforeUnmount(() => clearTimeout(viewTimer))
}
