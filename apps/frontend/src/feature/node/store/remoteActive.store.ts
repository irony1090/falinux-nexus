import { useTestSocket } from '@/common/websocket/websocket.hook'
import { tabReady } from '@/common/util/tabId.util'
import { computed, inject, onBeforeUnmount, provide, ref, watch } from 'vue'
import { listActive, type RemoteActive } from '../api/active.api'

const REMOTE_ACTIVE_STORE_KEY = Symbol('RemoteActiveStore')

export type ActiveUser = { userId: number; nickname: string }

// 다른 계정 탭의 active 타일 — 탭 key -> 대상. 보드와 무관(계정 토픽)이라 pages/index.vue에서 provide. REF-node-ui-active.md "원격 active 구조안"
export const provideRemoteActive = () => {
    const { on } = useTestSocket()
    const byKey = ref<Record<string, RemoteActive>>({})

    const apply = (a: RemoteActive) => {
        if (!a.uid && a.nodeId == null) delete byKey.value[a.key]
        else byKey.value[a.key] = a
    }

    // 불러오는 사이 온 이벤트는 응답을 쓴 뒤 다시 적용 — 응답이 더 오래된 상태일 수 있다
    let pending: RemoteActive[] | null = null
    const load = () => {
        const buf: RemoteActive[] = []
        pending = buf
        listActive()
            .then(list => {
                if (pending !== buf) return     // 더 늦게 시작한 load가 있다
                byKey.value = Object.fromEntries(list.map(a => [a.key, a]))
                buf.forEach(apply)
            })
            .catch(err => console.log('[REMOTE_ACTIVE] load', err))
            .finally(() => { if (pending === buf) pending = null })
    }

    const offs = [
        on<RemoteActive>('ACTIVE:UPDATE', a => {
            pending?.push(a)
            apply(a)
        }),
        on('SHARE:CHANGE', load),     // 볼 수 있는 범위가 바뀜
    ]
    // 소켓이 (다시) 붙을 때마다 — 끊긴 사이 이벤트를 놓쳤을 수 있다
    watch(tabReady, r => {
        if (r) load()
        else byKey.value = {}
    }, { immediate: true })
    onBeforeUnmount(() => offs.forEach(off => off()))

    // 대상별 계정 목록(같은 계정의 탭 여러 개 = 한 번)
    const group = (keyOf: (a: RemoteActive) => string | number | undefined) => computed(() => {
        const m = new Map<string | number, ActiveUser[]>()
        Object.values(byKey.value).forEach(a => {
            const k = keyOf(a)
            if (k === undefined) return
            const users = m.get(k) ?? []
            if (!users.some(u => u.userId === a.userId)) users.push({ userId: a.userId, nickname: a.nickname })
            m.set(k, users)
        })
        return m
    })
    const byUid = group(a => a.uid || undefined)
    const byNodeId = group(a => a.uid ? undefined : a.nodeId)

    const ctx = { byUid, byNodeId }
    provide(REMOTE_ACTIVE_STORE_KEY, ctx)
    return ctx
}

export const useRemoteActive = () => {
    const context = inject<ReturnType<typeof provideRemoteActive>>(REMOTE_ACTIVE_STORE_KEY)
    if (!context) throw new Error('RemoteActiveStore is not provided')
    return context
}
