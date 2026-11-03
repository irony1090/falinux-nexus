import type { RunResult, StepDef, UpdRun } from '../api/updater.api'
import { durText } from './updGrid.util'

// 단계 실행 행 표시 — Run · Log 타일과 시험 타일 공용(REF-update-trial.md 공용-1(행 표시))

export const RUN_ICON: Record<RunResult | 'none', string> = { ok: '✓', fail: '✕', timeout: '✕', cancel: '–', skip: '–', running: '●', wait: '○', none: '○' }
export const RUN_TEXT: Record<RunResult, string> = { ok: '완료', fail: '실패', timeout: '시간 초과', cancel: '취소', skip: '건너뜀', running: '진행 중', wait: '대기' }

export const runTone = (res?: RunResult) => !res ? 'idle' : res === 'ok' ? 'ok' : res === 'running' ? 'run' : ['fail', 'timeout'].includes(res) ? 'bad' : 'idle'

// now = 초 단위 현재 시각(useUpdNow) — 진행 중이면 지난 시간을 센다
export const runText = (now: number, r?: UpdRun, def?: StepDef) => {
    if (!r) return def?.manual ? '수동 단계' : '기록 없음'
    const t = RUN_TEXT[r.result]
    if (r.result === 'running') return `${t} ${durText(Math.max(0, Math.floor(now - r.startedAt)))}${def?.interactive ? ' · 입력 받음' : def?.timeoutSec ? ` / ${durText(def.timeoutSec)}` : ''}`
    if (r.result === 'wait' || r.result === 'skip' || !r.endedAt) return t
    return `${t} ${durText(Math.max(0, r.endedAt - r.startedAt))}${r.exitCode != null && r.exitCode !== 0 ? ` · 종료 코드 ${r.exitCode}` : ''}`
}
