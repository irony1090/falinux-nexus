import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { boardPath, useBoards } from '../store/boards.store'
import { useTileTree } from '../store/tileTree.store'

// 타일을 하위째 새 보드로 분리 / 다른 보드로 이동 — W1(분리 범위)·W3(이동 범위). 능력-1(루트 가능) 타일이 쓴다.
// run = 그 밖의 타일 메뉴 요청도 같은 busy·오류 표시로
export const useTileBoard = (tileId: () => string) => {
    const { hash, detachToBoard, moveToBoard } = useTileTree()
    const { list: boards, add: addBoard } = useBoards()
    const router = useRouter()
    const { openDialog } = useAppDialog()

    const otherBoards = computed(() => boards.value.filter(b => b.hash !== hash.value))
    const busy = ref(false)
    const run = (job: Promise<unknown>) => {
        busy.value = true
        job.catch(err => openDialog({ type: 'error', content: err?.message || '요청에 실패했습니다' }))
        .finally(() => busy.value = false)
    }

    const detachOpen = ref(false)
    const detach = (name: string) => run(detachToBoard(tileId(), name).then(b => {
        detachOpen.value = false
        addBoard(b)
        router.push(boardPath(b))
    }))
    const move = (to: string) => run(moveToBoard(tileId(), to))

    return { otherBoards, busy, run, detachOpen, detach, move }
}
