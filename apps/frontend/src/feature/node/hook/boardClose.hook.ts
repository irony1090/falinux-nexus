import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { computed, readonly, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { BoardCloseMode, BoardItem } from '../api/board.api'
import { boardLabel, useBoards } from '../store/boards.store'

// 닫는 중 — 루트 타일 ×와 보드 메뉴가 같은 상태를 보도록 모듈에 둔다
const closing = ref(false)

const NOTE = [
    '기본 보드로 보내기: 안의 폴더·터미널 타일을 기본 보드 끝으로 옮깁니다. 실행 중인 터미널은 그대로 둡니다.',
    '모두 종료하고 닫기: 실행 중인 터미널에 종료를 요청하고 보드를 지웁니다. 10초 안에 끝나지 않으면 강제 종료됩니다(조작할 수 있는 다른 계정이 열어 두지 않았을 때).',
    '보기 권한만 있는 공유 터미널은 종료되지 않고 이 보드에서만 빠집니다.',
]

// 보드 닫기 확인 — 루트 타일 ×·보드 메뉴 "보드 삭제" 공용. 기본 보드는 닫을 수 없다 — REF-board-close.md
export const useBoardClose = () => {
    const route = useRoute()
    const router = useRouter()
    const { byHash, remove } = useBoards()
    const { open: dialogOpen, openDialog } = useAppDialog()

    const current = computed(() => typeof route.params.hash === 'string' ? byHash(route.params.hash) : undefined)
    const closable = computed(() => !!current.value && !current.value.isDefault)

    const run = (b: BoardItem, mode: BoardCloseMode) => {
        dialogOpen.value = false
        closing.value = true
        remove(b.hash, mode)
        .then(() => router.replace('/'))
        .catch(err => openDialog({ type: 'error', content: err?.message || '보드를 닫지 못했습니다' }))
        .finally(() => closing.value = false)
    }

    // 이름은 title로만 넘긴다 — AppDialog의 content는 v-html로 그려진다
    const confirmClose = () => {
        const b = current.value
        if (!b || b.isDefault || closing.value) return
        openDialog({
            type: 'warning',
            title: `'${boardLabel(b)}' 닫기`,
            content: NOTE,
            buttons: [
                { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
                // Enter를 묶지 않는다 — 묶지 않은 Enter·Space는 AppDialog가 닫기(취소)로 처리
                { view: '기본 보드로 보내기', type: 'info', action: () => run(b, 'move') },
                { view: '모두 종료하고 닫기', type: 'error', action: () => run(b, 'kill') },
            ],
        })
    }

    return { closable, closing: readonly(closing), confirmClose }
}
