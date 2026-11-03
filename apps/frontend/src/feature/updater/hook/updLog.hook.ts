import '@xterm/xterm/css/xterm.css'
import { FitAddon } from '@xterm/addon-fit'
import { Terminal } from '@xterm/xterm'
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { useTheme } from 'vuetify'
import { getRun, runInput, type UpdRun } from '../api/updater.api'
import { onUpdOutput, updLogReload } from './updWatch.hook'

// 단계 실행 로그 상자 — Run · Log 타일과 시험 타일 공용(REF-update-trial.md 결정-8(코드 공용화))

// 1초마다 바뀌는 현재 시각(초) — 진행 중 단계의 지난 시간 표시용
export const useUpdNow = () => {
    const now = ref(Date.now() / 1000)
    let clock: ReturnType<typeof setInterval> | undefined
    onMounted(() => { clock = setInterval(() => now.value = Date.now() / 1000, 1000) })
    onBeforeUnmount(() => clearInterval(clock))
    return now
}

type LogOpts = {
    runs: () => UpdRun[] | undefined             // 따라갈 후보(고르지 않았을 때 진행 중 > 최근)
    interactive: (r: UpdRun) => boolean          // 진행 중인 이 실행에 키 입력을 보낼지
}

export const useUpdLog = (opts: LogOpts) => {
    const selected = ref<number>()
    const picked = ref(false)   // 사람이 행을 골랐으면 따라가지 않음
    const shown = computed(() => opts.runs()?.find(r => r.id === selected.value))

    watch(() => opts.runs()?.filter(r => r.result !== 'wait' && r.result !== 'skip'), runs => {
        if (!runs?.length || picked.value) return
        const pick = runs.find(r => r.result === 'running') ?? [...runs].sort((a, b) => b.startedAt - a.startedAt || b.id - a.id)[0]
        if (pick && pick.id !== selected.value) selected.value = pick.id
    }, { immediate: true })
    const pick = (r?: UpdRun) => { if (!r || r.result === 'wait' || r.result === 'skip') return; picked.value = true; selected.value = r.id }
    // 방금 시작한 실행으로 고정(startStepRun 응답의 runId)
    const follow = (runId: number) => { picked.value = true; selected.value = runId }
    // 다시 따라가기(되돌리기 시작 등)
    const unpin = () => { picked.value = false }

    const box = ref<HTMLElement>()
    const term = shallowRef<Terminal>()
    const fit = new FitAddon()
    const theme = useTheme()
    let ro: ResizeObserver | undefined
    // 상자는 조회가 온 뒤에 그려진다 — 생길 때 만든다
    watch(box, el => {
        if (!el || term.value) return
        const color = (n: string) => { const v = theme.current.value.colors[n]; return typeof v === 'string' ? v : undefined }
        const t = new Terminal({ convertEol: false, fontSize: 12, scrollback: 5000, theme: { background: color('terminal'), foreground: color('on-terminal') } })
        t.loadAddon(fit)
        t.open(el)
        // 입력 받는 단계(interactive)가 진행 중일 때만 보낸다 — 정-6(콘솔)
        t.onData(data => {
            const r = shown.value
            if (r?.result === 'running' && opts.interactive(r)) void runInput(r.id, data).catch(() => undefined)
        })
        term.value = t
        ro = new ResizeObserver(() => { try { fit.fit() } catch { /* 숨김 */ } })
        ro.observe(el)
    })

    let unsub: (() => void) | undefined
    let loadSeq = 0
    // 조회(logEnd)와 실시간 조각(off)의 이음매 — 이미 쓴 끝(last)보다 앞부분은 버린다.
    // 조회 응답 전에 온 조각은 모아 두었다가 응답 뒤에 이음매 기준으로 붙인다(순서를 바꾸면 로그가 겹치거나 빠짐)
    const load = async (id: number) => {
        const seq = ++loadSeq
        unsub?.()
        let last = -1   // -1 = 조회 응답 전(조각은 모아 둠)
        const pending: { b: Uint8Array; off: number }[] = []
        const put = (b: Uint8Array, off: number) => {
            const start = off - b.length
            if (off <= last) return
            term.value?.write(start < last ? b.subarray(last - start) : b)
            last = off
        }
        unsub = onUpdOutput(id, (b, off) => { if (last < 0) pending.push({ b, off }); else put(b, off) })
        try {
            const r = await getRun(id)
            if (seq !== loadSeq) return
            term.value?.reset()
            term.value?.write(r.log ?? '')
            last = r.logEnd ?? 0
            pending.forEach(p => put(p.b, p.off))
        } catch {
            if (seq === loadSeq) { term.value?.reset(); term.value?.write('[로그를 불러오지 못했습니다]\r\n') }
        }
    }
    watch([selected, term], ([id, t]) => { if (id && t) void load(id) }, { immediate: true })
    watch(() => selected.value && updLogReload.value[selected.value], (n, o) => { if (n && n !== o && selected.value) void load(selected.value) })
    onBeforeUnmount(() => { unsub?.(); ro?.disconnect(); term.value?.dispose() })

    // 실행 시작 요청에 실을 터미널 크기
    const size = () => term.value ? { cols: term.value.cols, rows: term.value.rows } : undefined

    // 다른 대상으로 바꿀 때 — 지난 로그를 지우고 다시 따라간다
    const reset = () => { unsub?.(); loadSeq++; term.value?.reset(); selected.value = undefined; picked.value = false }

    return { box, selected, shown, pick, follow, unpin, reset, size }
}
