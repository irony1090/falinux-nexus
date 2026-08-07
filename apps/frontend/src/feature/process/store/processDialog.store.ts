import type { ProcessResponse } from '@/feature/process/api/process.api'
import { computed, inject, provide, readonly, ref } from 'vue'

const PROCESS_DIALOG_STORE_KEY = Symbol('ProcessDialogStore')

// 현재 다이얼로그가 보여주는 대상 process 1개만 관리한다(동시에 여러 개 띄우는 건 범위 밖).
// 소켓 STATUS 이벤트 구독·필터링은 이 스토어 바깥(호출부)에서 구현 — 여기선 patchStatus로
// 반영할 진입점만 제공한다.
export const provideProcessDialog = () => {
    const process_ = ref<ProcessResponse | null>(null)
    const process = readonly(process_)

    // const open = ref(false);
    const open = computed({
        get: () => !!process.value,
        set: val => {
            if (!val) process_.value = null
        }
    })

    const openProcessDialog = (target: ProcessResponse) => {
        process_.value = target
    }

    const closeProcessDialog = () => {
        process_.value = null
    }

    // 소켓 STATUS 이벤트 등 외부에서 받은 최신 status를 반영한다. uid가 다르면(이미 다른
    // process로 교체됐거나 닫힌 경우) 무시.
    const patchStatus = (uid: string, status: ProcessResponse['status']) => {
        if (process_.value?.uid !== uid) return
        process_.value = { ...process_.value, status }
    }

    const ctx = {
        open,
        process,
        openProcessDialog,
        closeProcessDialog,
        patchStatus,
    }

    provide(PROCESS_DIALOG_STORE_KEY, ctx)

    return ctx
}

export const useProcessDialog = () => {
    const context = inject<ReturnType<typeof provideProcessDialog>>(PROCESS_DIALOG_STORE_KEY)
    if (!context) throw new Error('ProcessDialogStore is not provided')
    return context
}
