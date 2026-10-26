import { inject, provide, type Ref } from 'vue'

const PAGER_BOARD_KEY = Symbol('PagerBoard')

// 미니맵 Grid 하나 — x·w = view와 같은 스크롤 좌표, 칸 위치는 cols·rows 기준
export type PeekGrid = {
    x: number
    w: number
    cols: number
    rows: number
    half: boolean                   // 반 폭 Grid(보드 1칼럼 Grid)
    items: Array<{ id: string; c: number; r: number; cw: number; ch: number }>
}

// 타일 판 하나가 알약 끌기(usePagerDrag)·미니맵(TilePeek)·빈 곳 휠(useGapWheel)에 주는 것 — 보드·관리 페이지가 각자 채운다. REF-member-ui.md "판 계약"
// 칼럼 = 스크롤이 멈추는 한 칸(번호 0부터, 폰 관리 페이지는 0 = 목록)
export type PagerBoard = {
    off: Readonly<Ref<boolean>>                     // 끌기 막음(보드 전체보기)
    span: Readonly<Ref<number>>                     // 화면에 들어가는 칼럼 수
    maxCol: Readonly<Ref<number>>                   // 화면 왼쪽 끝이 될 수 있는 마지막 칼럼
    viewCol: Readonly<Ref<number>>                  // 지금 화면 왼쪽 칼럼
    order: Readonly<Ref<string[]>>                  // 타일 순서 = 위치 번호 순
    activeId: Readonly<Ref<string | null>>
    scrollEl: Readonly<Ref<HTMLElement | undefined>>
    scrubbing: Ref<boolean>                         // 끄는 동안 스냅 끔(판이 class로 건다)
    colLeft: (col: number) => number | undefined    // 칼럼의 scrollLeft — 도착 확인용
    goCol: (col: number) => void
    scrubTo: (col: number) => void                  // 실수 칼럼, 애니메이션 없이
    colRange: (id: string) => { l: number; r: number } | undefined
    setActive: (id: string) => void
    focusTile: (id: string) => void
    // 미니맵
    peekGrids: Readonly<Ref<PeekGrid[]>>
    peekView: Readonly<Ref<{ l: number; r: number }>>   // 지금 보이는 범위(peekGrids 좌표)
    aspect: Readonly<Ref<number>>                   // 전체 폭 Grid의 가로/세로(0 = 모름)
}

export const providePagerBoard = (board: PagerBoard) => {
    provide(PAGER_BOARD_KEY, board)
    return board
}

export const usePagerBoard = () => {
    const board = inject<PagerBoard>(PAGER_BOARD_KEY)
    if (!board) throw new Error('PagerBoard is not provided')
    return board
}
