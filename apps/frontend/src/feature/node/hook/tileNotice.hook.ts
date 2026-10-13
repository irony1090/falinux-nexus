import { useTestSocket } from '@/common/websocket/websocket.hook'
import { onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { boardLabel, boardPath, focusPath, useBoards } from '../store/boards.store'
import { useTileNotices, type NoticeItem, type TileNoticeEvent, type TileSpot } from '../store/tileNotice.store'
import { useTileTree } from '../store/tileTree.store'

// TILES:NOTICE -> 문구·이동할 곳·토스트 여부. 토스트 = 출발·도착 보드를 보는 탭만, 위치 없음은 모든 탭(N4(받는 탭)·K2(위치 없음 토스트))
export const useTileNoticeFeed = () => {
    const { on } = useTestSocket()
    const { hash } = useTileTree()
    const { byHash } = useBoards()
    const { push } = useTileNotices()

    const where = (s?: TileSpot) => !s ? '' : s.hash === hash.value ? '이 보드' : boardLabel(byHash(s.hash)) || '다른 보드'
    const here = (s?: TileSpot) => !!s && s.hash === hash.value

    const off = on<TileNoticeEvent>('TILES:NOTICE', ev => {
        const to = ev.to ? { hash: ev.to.hash, tileId: ev.to.tileId } : undefined
        let text = ''
        let toast = here(ev.from) || here(ev.to)
        let go: NoticeItem['go'] = to ?? { tab: 'terminals' }
        switch (ev.kind) {
        case 'BIND':        // M1(내가 바인딩)·M3(다른 곳으로 빠짐)
            text = here(ev.to) ? `${ev.name} — 이 보드에 바인딩됨` : here(ev.from) ? `${ev.name} — ${where(ev.to)}(으)로 이동됨` : `${ev.name} — ${where(ev.to)}에 바인딩됨`
            break
        case 'ATTACH':      // M2(자동으로 붙음)
            text = `${ev.name} — 다른 계정이 실행, ${where(ev.to)}에 열림`
            break
        case 'NO_PLACE':    // M4(위치 없음)
            text = `${ev.name} — 다른 계정이 실행, 위치 없음`
            toast = true
            go = { tab: 'terminals' }
            break
        case 'PIN_MOVED':   // M5(고정 이동)
            text = `${ev.name} — 타일 고정이 ${where(ev.to)}(으)로`
            break
        }
        push({ ...ev, text, go }, toast)
    })
    onUnmounted(off)
}

// 알림 누르기 — 보드 경로 + ?focus / ?tab 은 TileWorkspace가 처리(N3(누르면 이동))
export const useNoticeGo = () => {
    const router = useRouter()
    const { hash } = useTileTree()
    const { byHash, defaultBoard } = useBoards()
    return (item: NoticeItem) => {
        if ('tab' in item.go) {
            const cur = byHash(hash.value) ?? defaultBoard.value
            if (cur) void router.push({ path: boardPath(cur), query: { tab: 'terminals' } })
            return
        }
        const b = byHash(item.go.hash)
        if (b && item.go.tileId) void router.push(focusPath(b, item.go.tileId))
    }
}
