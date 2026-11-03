import { computed, ref, watch, type ComputedRef } from 'vue'
import { readText, type Ref } from '../api/updater.api'

// 열린 편집 타일(steps.json · Script) 내용 맞추기 — 서버 파일 sha가 바뀌면 고치던 것이 없을 때 다시 읽고, 있으면 stale만 켠다(덮지 않음)
// sha = 빌드 상세의 그 파일 sha(같은 탭에서 바꾸면 invalidate로 바로, 다른 곳은 다음 조회 때)
export const useUpdLiveText = (bundle: Ref, path: string, sha: ComputedRef<string | undefined>, onError: (msg: string) => void, onLoad?: () => void) => {
    const original = ref<string>()
    const text = ref('')
    const loadedSha = ref<string>()
    const stale = ref(false)
    const dirty = computed(() => original.value !== undefined && text.value !== original.value)

    const load = async (s: string | undefined) => {
        try {
            const t = await readText(bundle, path)
            original.value = t
            text.value = t
            loadedSha.value = s
            stale.value = false
            onLoad?.()
        } catch (e: any) {
            onError(e?.message || '불러오지 못했습니다')
        }
    }
    // 처음 = 파일이 없을 때의 내용(steps.json 틀) — 고친 것으로 치지 않는다
    const init = (t: string) => { original.value = t; text.value = t }
    // 저장 뒤 = 지금 글이 서버 내용, s = 저장 응답의 sha(이걸 안 맞추면 invalidate 뒤 한 번 더 읽음)
    const saved = (s?: string) => { original.value = text.value; loadedSha.value = s; stale.value = false }

    watch(sha, s => {
        if (s === undefined || original.value === undefined || s === loadedSha.value) return
        if (dirty.value) stale.value = true
        else void load(s)
    })
    return { original, text, dirty, stale, load, init, saved }
}
