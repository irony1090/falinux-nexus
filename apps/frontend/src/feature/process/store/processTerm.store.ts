import { useTestSocket } from '@/common/websocket/websocket.hook'
import {
    execProcess, killProcess, listProcesses, resizeProcess, toProcessResponse,
    type ProcessResponse, type ProcessResponseDto, type ProcessStatus,
} from '@/feature/process/api/process.api'
import type { TileSize } from '@/feature/widget/util/tile.type'
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

export type ExecPlace = { parentTileId: string; size: TileSize }

const base64ToBytes = (b64: string) => {
    const binary = atob(b64)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
    return bytes
}

const bytesToBase64 = (bytes: Uint8Array) => {
    let binary = ''
    for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]!)
    return btoa(binary)
}

const utf8 = new TextEncoder()

export const isRunning = (status?: ProcessStatus) => status === 'PROCESS' || status === 'PENDING'

// 타일 화면의 process: uid -> 상태·크기 소유 여부·xterm. 타일 트리와 수명이 같다(pages/index.vue에서 provide).
// 재발 방지: xterm은 컴포넌트가 아니라 여기서 가진다 — 타일이 다른 Grid로 가면 재마운트되므로
// 컴포넌트가 가지면 화면이 사라진다. 타일은 attach/detach로 DOM만 붙였다 뗀다 (REF-node-ui-terminal.md J)
export const provideProcessTerm = () => {
    const { on, emit, status: socketStatus } = useTestSocket()
    const theme = useTheme()

    const procs_ = ref<Record<string, ProcessResponse>>({})
    const owners_ = ref<Record<string, boolean>>({})
    const restored_ = ref(false)        // 복원 전엔 모르는 uid의 터미널 타일을 닫지 못하게 — 3-d②(복원 전 닫기)
    const handles = new Map<string, TermHandle>()   // 반응형에 넣지 않는다(xterm 객체를 proxy로 감싸면 안 됨)

    const offs: (() => void)[] = []     // 소켓은 App 수명이라 페이지가 내려갈 때 핸들러를 뗀다

    // exec·복원 응답보다 출력·상태가 먼저 올 수 있어(계정 소켓은 relay 기동 전에 구독됨) 요청 중에만 모르는 uid 것을 모아 둔다
    let fetchInFlight = 0
    const early = new Map<string, Uint8Array[]>()
    const earlyStatus = new Map<string, StatusEvent>()

    // PROCESS일 때만 보낸다(서버도 버리지만 불필요한 전송 방지) — REF-process-input.md I3(PENDING 중 입력)
    const sendInput = (uid: string, bytes: Uint8Array) => {
        if (procs_.value[uid]?.status !== 'PROCESS') return
        emit('PROCESS:INPUT', { uid, data: bytesToBase64(bytes) })
    }

    // I4(Ctrl+C 복사): 선택 영역이 있으면 복사하고 0x03을 보내지 않는다. 클립보드를 못 쓰는 환경(비보안 http)에서도
    // 막는다 — 복사하려던 Ctrl+C가 프로세스를 중단시키면 안 되므로
    // I5(Ctrl+V 붙여넣기): xterm이 ^V(0x16)로 바꾸며 기본 동작을 막으므로 넘겨서 브라우저 paste -> xterm onData로 가게 한다
    const clipboardKeys = (term: Terminal) => (ev: KeyboardEvent) => {
        const ctrlOnly = ev.ctrlKey && !ev.shiftKey && !ev.altKey && !ev.metaKey
        if (ctrlOnly && ev.code === 'KeyV') return false
        if (!ctrlOnly || ev.code !== 'KeyC' || !term.hasSelection()) return true
        if (ev.type === 'keydown') {
            navigator.clipboard?.writeText(term.getSelection()).catch(err => console.log('[PROCESS_TERM] copy', err))
            term.clearSelection()
        }
        return false
    }

    const createHandle = (uid: string) => {
        const color = (name: string) => {
            const v = theme.current.value.colors[name]
            return typeof v === 'string' ? v : undefined
        }
        const term = new Terminal({
            convertEol: true,
            fontSize: 12,
            theme: { background: color('terminal'), foreground: color('on-terminal') },
        })
        const fit = new FitAddon()
        term.loadAddon(fit)
        term.attachCustomKeyEventHandler(clipboardKeys(term))
        term.onData(data => sendInput(uid, utf8.encode(data)))
        // onBinary = 문자 하나가 바이트 하나(마우스 보고 등) — UTF-8로 다시 인코딩하면 안 됨
        term.onBinary(data => sendInput(uid, Uint8Array.from(data, ch => ch.charCodeAt(0) & 0xff)))
        return { term, fit }
    }

    const endFetch = () => {
        if (--fetchInFlight === 0) {
            early.clear()
            earlyStatus.clear()
        }
    }

    // 소유권 이벤트(PROCESS:SIZE_OWNER)가 응답보다 먼저 왔으면 응답의 false로 덮지 않는다
    const writeNotice = (uid: string, text: string) =>
        handles.get(uid)?.term.write(`\r\n\x1b[2m[${text}]\x1b[0m\r\n`)

    const register = (proc: ProcessResponse) => {
        procs_.value[proc.uid] = proc
        owners_.value[proc.uid] = !!proc.sizeOwner || !!owners_.value[proc.uid]
        if (!handles.has(proc.uid)) handles.set(proc.uid, createHandle(proc.uid))
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
        } else if (fetchInFlight > 0) {
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
        // 실행 중 -> 끝남일 때만 — 복원 때 이미 쓴 종료 줄을 같은 상태 이벤트로 또 쓰지 않게
        if (done && isRunning(proc.status)) writeNotice(ev.uid, `process exited: ${ev.exitCode}`)
    }

    offs.push(on<StatusEvent>('STATUS', ev => {
        if (procs_.value[ev.uid]) applyStatus(ev)
        else if (fetchInFlight > 0) earlyStatus.set(ev.uid, ev)     // 최신 것 하나면 충분
    }))

    offs.push(on<ProcessResponseDto>('PROCESS:UPDATE', dto => {
        if (!procs_.value[dto.uid]) return
        procs_.value[dto.uid] = toProcessResponse(dto)
    }))

    // 등록 전(복원 중)에 와도 기록한다 — register가 이 값을 유지
    offs.push(on<SizeOwnerEvent>('PROCESS:SIZE_OWNER', ev => {
        owners_.value[ev.uid] = true
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

    // place = 서버가 터미널 타일을 넣을 자리. 응답의 tile·tileVersion은 호출부가 타일 트리에 반영
    const exec = (nodeId: number, authKey: string, place: ExecPlace) => {
        fetchInFlight++
        return execProcess({ nodeId, authKey, ...place })
            .then(res => {
                register(res.proc)
                return res
            })
            .finally(endFetch)
    }

    // 새로고침 뒤 터미널 타일의 process를 다시 등록한다(⑪ 3-d(터미널 복원)). 이전 출력은 없다 — SNAPSHOT 전까지 ⑪-3 감수.
    // 이미 등록된 uid(복원 중 이 탭에서 exec)는 건드리지 않는다
    const restore = (fetch: () => Promise<ProcessResponse[]>) => {
        fetchInFlight++
        return fetch()
            .then(list => {
                list.filter(p => !handles.has(p.uid)).forEach(p => {
                    handles.set(p.uid, createHandle(p.uid))
                    // 3-d③(복원 안내 줄): 빈 화면이 고장처럼 보이지 않게. 모아 둔 출력보다 먼저 쓴다
                    writeNotice(p.uid, isRunning(p.status) ? 'reconnected — earlier output not shown' : `process exited: ${p.exitCode ?? '?'}`)
                    register(p)
                })
                restored_.value = true
                return list
            })
            .finally(endFetch)
    }

    // 3-d④(계정 바뀜): 이전 계정의 xterm·상태를 비운다
    const reset = () => {
        handles.forEach(h => h.term.dispose())
        handles.clear()
        procs_.value = {}
        owners_.value = {}
        restored_.value = false
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
        restored: readonly(restored_),
        exec,
        restore,
        reset,
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
