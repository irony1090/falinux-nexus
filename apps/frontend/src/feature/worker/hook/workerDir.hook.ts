import { ref, watch, type Ref } from 'vue'
import { listWorkerDir, type DirCrumb, type DirEntry } from '../api/worker.api'
import { parentPathOf } from '../util/workerDir.util'

type DirError = { message: string; status?: number }

// 장비 타일 목록 — 첫 페이지 + 스크롤로 다음 페이지(안 다(worker 캐시), 500개씩). 폴더-2(장비 타일)
// 대상(인스턴스·폴더·경로)이 바뀌면 처음부터. 늦게 온 옛 응답은 gen으로 버린다
export const useWorkerDir = (instance: Ref<string | undefined>, folderId: Ref<number | undefined>, path: Ref<string>) => {
    const entries = ref<DirEntry[]>([])
    const crumbs = ref<DirCrumb[]>([])
    const parent = ref<string | null>(null)
    const total = ref(0)
    const loading = ref(false)
    const error = ref<DirError | null>(null)
    let cursor = ''
    let gen = 0

    const reload = async () => {
        const my = ++gen
        const inst = instance.value, fid = folderId.value
        if (!inst || fid === undefined) return
        loading.value = true
        error.value = null
        try {
            const res = await listWorkerDir(inst, fid, { path: path.value })
            if (my !== gen) return
            cursor = res.cursor
            entries.value = res.entries
            crumbs.value = res.crumbs
            parent.value = res.parent
            total.value = res.total
        } catch (err) {
            if (my !== gen) return
            // 실패하면 직전 폴더의 경로 줄·상위가 남지 않게 — 요청한 경로 기준으로
            const p = path.value, up = parentPathOf(p)
            crumbs.value = p ? [{ name: p, path: p }] : []
            parent.value = p && up !== p ? up : null
            entries.value = []
            total.value = 0
            error.value = err as DirError
        } finally {
            if (my === gen) loading.value = false
        }
    }

    // 다음 페이지. 410(보관 만료) = 처음부터 다시
    const loadMore = async () => {
        const inst = instance.value, fid = folderId.value
        if (loading.value || !inst || fid === undefined || !cursor || entries.value.length >= total.value) return
        const my = gen
        loading.value = true
        try {
            const res = await listWorkerDir(inst, fid, { cursor, offset: entries.value.length })
            if (my !== gen) return
            entries.value = [...entries.value, ...res.entries]
        } catch (err) {
            if (my !== gen) return
            if ((err as DirError).status === 410) {
                loading.value = false
                return reload()
            }
            error.value = err as DirError
        } finally {
            if (my === gen) loading.value = false
        }
    }

    watch([instance, folderId, path], () => void reload(), { immediate: true })

    return { entries, crumbs, parent, total, loading, error, reload, loadMore }
}
