import { useTestSocket } from '@/common/websocket/websocket.hook'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, onUnmounted, ref, type Ref } from 'vue'
import { invalidateNodeRuns, useDeviceRuns, useNodeRuns, type RunResponse } from '../api/node.api'
import { listViewers } from '../api/tile.api'
import { useTileTree } from '../store/tileTree.store'

// 실행·종료·종료 요청·타일 변화·공유 변화가 오면 process 목록을 다시 받는다
const useRunsRefresh = () => {
    const queryClient = useQueryClient()
    const { on } = useTestSocket()
    const invalidate = () => { void invalidateNodeRuns(queryClient) }
    const offs = ['TILES:UPDATE', 'STATUS', 'PROCESS:UPDATE', 'SHARE:CHANGE'].map(type => on(type, invalidate))
    onUnmounted(() => offs.forEach(off => off()))
}

// 노드의 상태 리스트(실시간-6a)·터미널 탭(deep, 개편-6(목록 UI))
export const useScriptRuns = (folderId: Ref<number | null>, deep = false) => {
    const query = useNodeRuns(folderId, deep)
    useRunsRefresh()

    const byNode = computed(() => {
        const map = new Map<number, RunResponse[]>()
        for (const r of query.data.value ?? []) {
            if (r.nodeId !== null) map.set(r.nodeId, [...(map.get(r.nodeId) ?? []), r])   // 직접 자식 목록엔 장비 실행이 없다
        }
        return map
    })

    const list = computed(() => query.data.value ?? [])

    return { byNode, list }
}

// 루트 타일 터미널 탭 목록 — 루트 종류별: 폴더 = 그 경로 하위 전체 / 장비 = 그 장비의 process 전부(Q25(루트 장비 탭 범위)).
// 루트가 바뀌면 다시 조회한다(최소화한 타일도 경로를 옮겨 찾는다)
export const useRootRuns = (enabled: Ref<boolean> = ref(true)) => {
    const { tiles, rootId } = useTileTree()
    const root = computed(() => tiles.value[rootId.value])
    const folderId = computed(() => root.value?.type === 'folder' ? root.value.nodeId : null)
    const device = computed(() => root.value?.type === 'device' ? root.value : undefined)

    const folderRuns = useNodeRuns(folderId, true, computed(() => enabled.value && root.value?.type === 'folder'))
    const deviceRuns = useDeviceRuns(
        computed(() => device.value ? device.value.instance.slice(0, device.value.instance.lastIndexOf('#')) : ''),
        computed(() => device.value?.nodeId ?? NaN),
        computed(() => enabled.value && !!device.value),
    )
    useRunsRefresh()

    return computed(() => (device.value ? deviceRuns.data.value : folderRuns.data.value) ?? [])
}

const VIEWERS_KEY = ['TILES', 'VIEWERS']

// 루트 타일 터미널 탭의 보기 줄 — 범위는 useRootRuns와 같음. 다른 보드·탭 변화 = TILES:UPDATE, 이 보드 = 보기 타일 목록이 바뀌면(키에 포함)
export const useRootViewers = (enabled: Ref<boolean> = ref(true)) => {
    const { tiles, rootId } = useTileTree()
    const root = computed(() => tiles.value[rootId.value])
    const folderId = computed(() => root.value?.type === 'folder' || root.value?.type === 'device' ? root.value.nodeId ?? 0 : 0)
    const mainKey = computed(() => root.value?.type === 'device' ? root.value.instance.slice(0, root.value.instance.lastIndexOf('#')) : '')
    const here = computed(() => Object.values(tiles.value).filter(t => t.type === 'viewer').map(t => t.id).sort().join())
    const query = useQuery({
        enabled: computed(() => enabled.value && !!root.value),
        queryKey: [...VIEWERS_KEY, folderId, mainKey, here],
        queryFn: ({ queryKey }) => listViewers(queryKey[2] as number, queryKey[3] as string),
    })
    const queryClient = useQueryClient()
    const { on } = useTestSocket()
    const off = on('TILES:UPDATE', () => { void queryClient.invalidateQueries({ queryKey: VIEWERS_KEY }) })
    onUnmounted(off)
    return computed(() => query.data.value ?? [])
}
