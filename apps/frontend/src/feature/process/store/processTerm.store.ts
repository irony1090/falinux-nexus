import { useTestSocket } from '@/common/websocket/websocket.hook'
import {
    execProcess, killProcess, listProcesses, resizeProcess, toProcessResponse,
    type ProcessResponse, type ProcessResponseDto, type ProcessStatus,
} from '@/feature/process/api/process.api'
import { FitAddon } from '@xterm/addon-fit'
import { Terminal } from '@xterm/xterm'
import { inject, onBeforeUnmount, provide, readonly, ref, watch } from 'vue'
import { useTheme } from 'vuetify'

const PROCESS_TERM_STORE_KEY = Symbol('ProcessTermStore')

// 백엔드 미러: protocol.DataEvent(data = base64) / protocol.StatusEvent(status = execute.CommandStatus 숫자) / protocol.SizeOwnerEvent
type DataEvent = { uid: string; data: string }
type StatusEvent = { uid: string; status: number; pid?: number; exitCode: number }
type SizeOwnerEvent = { uid: string }

const STATUS_BY_CODE: ProcessStatus[] = ['PENDING', 'PROCESS', 'COMPLETED', 'FAILED']

export type TermHandle = { term: Terminal; fit: FitAddon }

const base64ToBytes = (b64: string) => {
    const binary = atob(b64)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
    return bytes
}

export const isRunning = (status?: ProcessStatus) => status === 'PROCESS' || status === 'PENDING'

// 타일 화면의 process: uid -> 상태·크기 소유 여부·xterm. 타일 트리와 수명이 같다(pages/index.vue에서 provide).
// 재발 방지: xterm은 컴포넌트가 아니라 여기서 가진다 — 타일이 다른 Grid로 가면 재마운트되므로
// 컴포넌트가 가지면 화면이 사라진다. 타일은 attach/detach로 DOM만 붙였다 뗀다 (REF-node-ui-terminal.md J)
export const provideProcessTerm = () => {
    const { on, status: socketStatus } = useTestSocket()
    const theme = useTheme()

    const procs_ = ref<Record<string, ProcessResponse>>({})
    const owners_ = ref<Record<string, boolean>>({})
    const handles = new Map<string, TermHandle>()   // 반응형에 넣지 않는다(xterm 객체를 proxy로 감싸면 안 됨)

    const offs: (() => void)[] = []     // 소켓은 App 수명이라 페이지가 내려갈 때 핸들러를 뗀다

    // exec 응답보다 출력·상태가 먼저 올 수 있어(계정 소켓은 relay 기동 전에 구독됨) 요청 중에만 모르는 uid 것을 모아 둔다
    let execInFlight = 0
    const early = new Map<string, Uint8Array[]>()
    const earlyStatus = new Map<string, StatusEvent>()

    const createHandle = () => {
        const color = (name: string) => {
            const v = theme.current.value.colors[name]
            return typeof v === 'string' ? v : undefined
        }
        const term = new Terminal({
            convertEol: true,
            disableStdin: true,     // 입력 배선 전
            fontSize: 12,
            theme: { background: color('terminal'), foreground: color('on-terminal') },
        })
        const fit = new FitAddon()
        term.loadAddon(fit)
        return { term, fit }
    }

    const register = (proc: ProcessResponse) => {
        procs_.value[proc.uid] = proc
        owners_.value[proc.uid] = !!proc.sizeOwner
        if (!handles.has(proc.uid)) handles.set(proc.uid, createHandle())
        const h = handles.get(proc.uid)!
        early.get(proc.uid)?.forEach(chunk => h.term.write(chunk))
        early.delete(proc.uid)
        const st = earlyStatus.get(proc.uid)
        if (st) applyStatus(st)
        earlyStatus.delete(proc.uid)
    }

    offs.push(on<DataEvent>('DATA', ev => {
        const bytes = base64ToBytes(ev.data)
        const h = handles.get(ev.uid)
        if (h) {
            h.term.write(bytes)
        } else if (execInFlight > 0) {
            early.set(ev.uid, [...(early.get(ev.uid) ?? []), bytes])
        }
    }))

    const applyStatus = (ev: StatusEvent) => {
        const proc = procs_.value[ev.uid]
        const status = STATUS_BY_CODE[ev.status]
        if (!proc || !status) return
        const done = status === 'COMPLETED' || status === 'FAILED'
        procs_.value[ev.uid] = {
            ...proc,
            status,
            pid: ev.pid ? ev.pid : proc.pid,
            exitCode: done ? ev.exitCode : proc.exitCode,
        }
        if (done) handles.get(ev.uid)?.term.write(`\r\n\x1b[2m[process exited: ${ev.exitCode}]\x1b[0m\r\n`)
    }

    offs.push(on<StatusEvent>('STATUS', ev => {
        if (procs_.value[ev.uid]) applyStatus(ev)
        else if (execInFlight > 0) earlyStatus.set(ev.uid, ev)     // 최신 것 하나면 충분
    }))

    offs.push(on<ProcessResponseDto>('PROCESS:UPDATE', dto => {
        if (!procs_.value[dto.uid]) return
        procs_.value[dto.uid] = toProcessResponse(dto)
    }))

    offs.push(on<SizeOwnerEvent>('PROCESS:SIZE_OWNER', ev => {
        if (procs_.value[ev.uid]) owners_.value[ev.uid] = true
    }))

    // 연결이 끊긴 사이 소유권이 넘어갔을 수 있다(원래 소유자는 되찾지 않음) — 다시 붙으면 서버 기준으로 맞춘다
    const syncOwners = () => listProcesses()
        .then(list => list.forEach(p => {
            if (procs_.value[p.uid]) owners_.value[p.uid] = !!p.sizeOwner
        }))
        .catch(err => console.log('[PROCESS_TERM] syncOwners', err))

    watch(socketStatus, s => {
        if (s === 'CONNECTED' && Object.keys(procs_.value).length) syncOwners()
    })

    const exec = (nodeId: number, authKey: string) => {
        execInFlight++
        return execProcess({ nodeId, authKey })
            .then(proc => {
                register(proc)
                return proc.uid
            })
            .finally(() => {
                if (--execInFlight === 0) {
                    early.clear()
                    earlyStatus.clear()
                }
            })
    }

    const kill = (uid: string) => killProcess(uid)

    // 소유자만 보낸다. 실패하면(403 포함 — 상태 코드가 버려져 구분 불가) 서버 기준으로 다시 맞춘다
    const syncSize = (uid: string, cols: number, rows: number) => {
        const proc = procs_.value[uid]
        if (!proc || !owners_.value[uid] || !isRunning(proc.status)) return
        if (proc.cols === cols && proc.rows === rows) return
        resizeProcess(uid, { cols, rows })
            .then(res => { procs_.value[uid] = res })
            .catch(err => {
                console.log('[PROCESS_TERM] resize', err)
                syncOwners()
            })
    }

    const handleOf = (uid: string) => handles.get(uid)

    // 첫 attach에서 open(한 번만 가능), 이후엔 xterm이 만든 element를 옮겨 붙인다
    const attach = (uid: string, el: HTMLElement) => {
        const h = handles.get(uid)
        if (!h) return
        if (!h.term.element) h.term.open(el)
        else if (h.term.element.parentElement !== el) el.appendChild(h.term.element)
    }

    const detach = (uid: string, el: HTMLElement) => {
        const node = handles.get(uid)?.term.element
        if (node?.parentElement === el) el.removeChild(node)
    }

    // 타일을 닫을 때 (실행 중엔 닫기가 막혀 있음)
    const dispose = (uid: string) => {
        handles.get(uid)?.term.dispose()
        handles.delete(uid)
        delete procs_.value[uid]
        delete owners_.value[uid]
    }

    onBeforeUnmount(() => {
        offs.forEach(off => off())
        handles.forEach(h => h.term.dispose())
        handles.clear()
    })

    const ctx = {
        procs: readonly(procs_),
        owners: readonly(owners_),
        exec,
        kill,
        syncSize,
        handleOf,
        attach,
        detach,
        dispose,
    }

    provide(PROCESS_TERM_STORE_KEY, ctx)

    return ctx
}

export const useProcessTerm = () => {
    const context = inject<ReturnType<typeof provideProcessTerm>>(PROCESS_TERM_STORE_KEY)
    if (!context) throw new Error('ProcessTermStore is not provided')
    return context
}
