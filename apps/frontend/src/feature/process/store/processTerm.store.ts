import { useTestSocket } from '@/common/websocket/websocket.hook'
import {
    execProcess, getProcessSnapshot, killProcess, listProcesses, resizeProcess, toProcessResponse,
    type ProcessResponse, type ProcessResponseDto, type ProcessSnapshot, type ProcessStatus, type ProcessType,
} from '@/feature/process/api/process.api'
import type { TileSize } from '@/feature/widget/util/tile.type'
import { FitAddon } from '@xterm/addon-fit'
import { Terminal } from '@xterm/xterm'
import { inject, onBeforeUnmount, provide, readonly, ref, watch } from 'vue'
import { useTheme } from 'vuetify'

const PROCESS_TERM_STORE_KEY = Symbol('ProcessTermStore')

// 백엔드 미러: protocol.DataEvent(data = base64, off = 이 묶음까지의 누적 바이트) / protocol.StatusEvent(status = execute.CommandStatus 숫자) / protocol.SizeOwnerEvent
type DataEvent = { uid: string; data: string; off?: number }
type StatusEvent = { uid: string; status: number; pid?: number; exitCode: number }
type SizeOwnerEvent = { uid: string }

const ALT_SCREEN_ENTER = '\x1b[?1049h'

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

    // 모르는 uid의 출력·상태를 잠깐 보관했다가 등록할 때 쓴다 — 4③(다른 탭 실행 터미널 출력).
    // 계정 소켓은 exec 때 relay 기동 전에 구독되므로 다른 탭이 실행한 터미널 출력도 처음부터 오지만, 타일 push로 uid를 알기 전엔
    // 등록돼 있지 않다. exec·복원 응답보다 먼저 온 출력도 같은 길. 넘치면 오래된 것부터 버리고 overflow 표시(안내 줄)
    const HOLD_BYTES = 64 * 1024
    const HOLD_MS = 10_000
    type Chunk = { bytes: Uint8Array; off?: number }
    type Held = { chunks: Chunk[]; size: number; overflow: boolean; at: number; status?: StatusEvent }
    const held = new Map<string, Held>()
    const expired = new Set<string>()      // 보관 시간이 지나 버린 uid — 다시 보관해도 처음부터가 아님

    const holdOf = (uid: string) => {
        let h = held.get(uid)
        if (!h) {
            h = { chunks: [], size: 0, overflow: expired.has(uid), at: Date.now() }
            held.set(uid, h)
        }
        return h
    }

    const purge = setInterval(() => {
        const now = Date.now()
        held.forEach((h, uid) => {
            if (now - h.at <= HOLD_MS) return
            held.delete(uid)
            expired.add(uid)
        })
    }, HOLD_MS / 2)

    // 복원 출력을 쓰는 중인 uid — 스냅샷 속 옛 질의(ESC[6n 등)에 xterm이 답한 응답이 실행 중인 process에 입력으로 가지 않게
    const replaying = new Set<string>()

    // PROCESS일 때만 보낸다(서버도 버리지만 불필요한 전송 방지) — REF-process-input.md I3(PENDING 중 입력)
    const sendInput = (uid: string, bytes: Uint8Array) => {
        if (procs_.value[uid]?.status !== 'PROCESS' || replaying.has(uid)) return
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

    const writeNotice = (uid: string, text: string) =>
        handles.get(uid)?.term.write(`\r\n\x1b[2m[${text}]\x1b[0m\r\n`)

    // 보관분 중 from(스냅샷 off) 뒤만 — S1(이음매 처리). 스냅샷에 이미 든 묶음은 버리고 걸친 묶음은 뒷부분만.
    // from 뒤 첫 묶음이 from보다 뒤에서 시작하면(보관함이 넘쳐 버림) 그 자리에 안내 줄
    const writeHeld = (uid: string, term: Terminal, chunks: Chunk[], from?: number) => {
        let first = true
        chunks.forEach(({ bytes, off }) => {
            if (from === undefined || off === undefined) return term.write(bytes)
            if (off <= from) return
            const start = off - bytes.length
            if (first && start > from) writeNotice(uid, 'earlier output not shown')
            first = false
            term.write(start < from ? bytes.subarray(from - start) : bytes)
        })
    }

    // 소유권 이벤트(PROCESS:SIZE_OWNER)가 응답보다 먼저 왔으면 응답의 false로 덮지 않는다.
    // from = 스냅샷 off(복원) — 없으면(exec·스냅샷 실패) 보관분 전부
    const register = (proc: ProcessResponse, from?: number) => {
        procs_.value[proc.uid] = proc
        owners_.value[proc.uid] = !!proc.sizeOwner || !!owners_.value[proc.uid]
        if (!handles.has(proc.uid)) handles.set(proc.uid, createHandle(proc.uid))
        const h = handles.get(proc.uid)!
        const kept = held.get(proc.uid)
        held.delete(proc.uid)
        writeHeld(proc.uid, h.term, kept?.chunks ?? [], from)
        if (kept?.status) applyStatus(kept.status)
    }

    offs.push(on<DataEvent>('DATA', ev => {
        const bytes = base64ToBytes(ev.data)
        const h = handles.get(ev.uid)
        if (h) {
            h.term.write(bytes)
            return
        }
        const k = holdOf(ev.uid)
        k.chunks.push({ bytes, off: ev.off })
        k.size += bytes.length
        while (k.size > HOLD_BYTES && k.chunks.length > 1) {
            k.size -= k.chunks.shift()!.bytes.length
            k.overflow = true
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
        else holdOf(ev.uid).status = ev     // 최신 것 하나면 충분
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

    // place = 서버가 터미널 타일을 넣을 자리. 응답의 tile·tileVersion은 호출부가 타일 트리에 반영. type EDIT = worker vi 편집
    const exec = (nodeId: number, authKey: string, place: ExecPlace, type: ProcessType = 'EXEC') => execProcess({ nodeId, authKey, type, ...place })
        .then(res => {
            register(res.proc)
            return res
        })

    // 스냅샷을 xterm에 쓴다 — PTY 크기로 맞춘 뒤에(커서 위치 지정 출력이 어긋나지 않게). C(ring + alt screen만 다시 그리기):
    // 안 잘렸으면(off = data 길이) 처음부터라 그대로 — alt screen 이전 일반 화면도 남는다.
    // 잘렸고 alt면 진입만(다시 그리기 출력이 off 뒤 DATA로 옴), 다시 그리기 못 함(worker 끊김)이면 data라도
    const writeSnapshot = (term: Terminal, uid: string, snap: ProcessSnapshot) => {
        if (snap.cols > 0 && snap.rows > 0) term.resize(snap.cols, snap.rows)
        const data = base64ToBytes(snap.data)
        if (snap.off === data.length) return term.write(data)
        if (!snap.alt) writeNotice(uid, 'earlier output truncated')
        else term.write(ALT_SCREEN_ENTER)
        if (!(snap.alt && snap.redraw)) term.write(data)
    }

    // 내가 실행하지 않은 터미널을 등록하는 유일한 길(새로고침 복원·다른 탭 push·409 재적용) — 4-c.
    // 실행 중이면 스냅샷 + 그 off 뒤 보관분(S1(이음매 처리)). 스냅샷을 받은 뒤에 handle을 만든다 — 그 사이 출력이 xterm이 아니라
    // 보관함으로 가야 off로 거를 수 있다. 스냅샷 실패(끝남·supervisor 재시작)면 예전 동작: reload = 3-d③(복원 안내 줄),
    // push는 보관한 출력이 처음부터라 넘쳤을 때만 안내. 이미 등록된 uid(이 탭에서 exec)는 건드리지 않는다
    const restore = (fetch: () => Promise<ProcessResponse[]>, { reload }: { reload: boolean }) => fetch()
        .then(async list => {
            const targets = list.filter(p => !handles.has(p.uid))
            const snaps = await Promise.all(targets.map(p => isRunning(p.status)
                ? getProcessSnapshot(p.uid).catch(err => {
                    console.log('[PROCESS_TERM] snapshot', p.uid, err)
                    return null
                })
                : null))
            targets.forEach((p, i) => {
                if (handles.has(p.uid)) return      // 스냅샷을 받는 사이 이 탭에서 등록됨
                const h = createHandle(p.uid)
                handles.set(p.uid, h)
                const snap = snaps[i]
                replaying.add(p.uid)
                if (snap) writeSnapshot(h.term, p.uid, snap)
                else if (isRunning(p.status) && (reload || held.get(p.uid)?.overflow)) {
                    writeNotice(p.uid, reload ? 'reconnected — earlier output not shown' : 'earlier output not shown')
                }
                register(p, snap?.off)
                // 보관한 출력 뒤에 — 상태 이벤트는 실행 중 -> 끝남 전이 때만 종료 줄을 쓰므로 이미 끝난 건 여기서
                if (!isRunning(p.status)) writeNotice(p.uid, `process exited: ${p.exitCode ?? '?'}`)
                h.term.write('', () => replaying.delete(p.uid))     // 앞서 넣은 쓰기가 모두 해석된 뒤에 불림
            })
            restored_.value = true
            return list
        })

    // 3-d④(계정 바뀜): 이전 계정의 xterm·상태를 비운다
    const reset = () => {
        replaying.clear()
        handles.forEach(h => h.term.dispose())
        handles.clear()
        held.clear()
        expired.clear()
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

    // 트리에서 빠진 끝난 터미널의 xterm 정리 — 4④(다른 탭이 닫은 터미널). 실행 중은 남김(이 탭 exec 직후 타일 반영 전일 수 있음)
    const prune = (keep: ReadonlySet<string>) => {
        Object.values(procs_.value).forEach(p => {
            if (!keep.has(p.uid) && !isRunning(p.status)) dispose(p.uid)
        })
    }

    onBeforeUnmount(() => {
        clearInterval(purge)
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
        prune,
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
