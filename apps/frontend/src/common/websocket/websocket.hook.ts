import { inject, onUnmounted, provide, ref, watch, type Ref } from 'vue';
import type { Connect } from '../util/index.type';
import { getBaseURL } from '../api/api.util';
import { readTabId } from '../util/tabId.util';

/* ────────────────────────────────────────────────────────────────────────
 * Protocol 계층 (frame 형태 가변)
 *   코어(call/emit/on)는 프레임 필드를 모른다. "의도 → 와이어"와
 *   "와이어 → 분류(Inbound)"만 Protocol에 위임한다.
 *   Go internal/transport.Conn 과 1:1 대응:
 *     call  ↔ Conn.Call  (REQ → RES, id 상관)
 *     emit  ↔ Conn.Emit  (EVENT 단방향)
 *     on    ↔ Conn.On    (EVENT 수신)
 * ──────────────────────────────────────────────────────────────────────── */

export type WireData = string | Blob | BufferSource;

/** 코어가 이해하는 정규화된 수신 분류 (와이어 형태와 무관). */
export type Inbound<D = unknown> =
    | { kind: 'response'; id: number; data?: D; err?: string } // → pending 매칭
    | { kind: 'event'; type: string; data?: D }                // → on 디스패치
    | { kind: 'request'; id: number; type: string; data?: D }  // 브라우저 미처리(분류만)
    | { kind: 'unknown' };                                      // 깨진/모르는 프레임 → 버림

export interface Protocol {
    /** REQ 프레임 생성. id/type/data 는 코어가 정하고 와이어 형태는 프로토콜이 정한다. */
    request(id: number, type: string, data?: unknown): WireData;
    /** EVENT 프레임 생성 (상관 id 없음). */
    event(type: string, data?: unknown): WireData;
    /** 와이어 메시지를 코어 분류로 디코드. 실패 시 { kind: 'unknown' }. */
    decode(raw: string | ArrayBuffer): Inbound;
}

/**
 * 기본 프로토콜 — Go internal/protocol.Frame 미러.
 *   { k: Kind, id, t: type, e?: err, d?: data }  (Kind: REQ=0 / RES=1 / EVENT=2)
 * Go 측이 BinaryMessage 로 보내므로 수신은 ArrayBuffer 를 TextDecoder 로 푼다.
 * 송신은 string(text frame) 으로 보내도 Go Serve 가 바이트만 Decode 하므로 무방.
 */
export const jsonFrame: Protocol = {
    request: (id, type, data) => JSON.stringify({ k: 0, id, t: type, d: data }),
    event: (type, data) => JSON.stringify({ k: 2, id: 0, t: type, d: data }),
    decode: (raw) => {
        let o: any;
        try {
            o = JSON.parse(typeof raw === 'string' ? raw : new TextDecoder().decode(raw));
        } catch {
            return { kind: 'unknown' };
        }
        switch (o?.k) {
            case 1: return { kind: 'response', id: o.id, data: o.d, err: o.e || undefined };
            case 2: return { kind: 'event', type: o.t, data: o.d };
            case 0: return { kind: 'request', id: o.id, type: o.t, data: o.d };
            default: return { kind: 'unknown' };
        }
    },
};

/* ────────────────────────────────────────────────────────────────────────
 * Hook
 * ──────────────────────────────────────────────────────────────────────── */

export type WebsocketOption = Partial<{
    /** 자동 재연결 최대 횟수. 미지정 시 무제한. */
    maxRetries: number;
    /** 첫 재연결 간격(ms). 실패할 때마다 두 배, retryMax까지. 기본 2000. */
    retryInterval: number;
    /** 재연결 간격 상한(ms). 기본 30000. */
    retryMax: number;
    /** 프레임 전략 주입. 기본 jsonFrame. frame 형태가 바뀌면 이것만 교체. */
    protocol: Protocol;
    /** call 기본 타임아웃(ms). 0 이면 무제한. 기본 10000. */
    callTimeout: number;
}>;

export type CallOption = Partial<{
    /** 이 호출만의 타임아웃(ms). 0 이면 무제한. 미지정 시 callTimeout. */
    timeout: number;
    /** 외부 취소 신호. abort 시 reject. */
    signal: AbortSignal;
}>;

type Pending = {
    resolve: (v: unknown) => void;
    reject: (e: unknown) => void;
};

export type SocketContext = {
    status: Ref<Connect['Status']>;
    /** 마지막 error 이벤트(onerror). */
    event: Ref<Event | null>;
    connect: () => void;
    disconnect: () => void;
    /** 화면 복귀·네트워크 복구 때 부른다: 끊겼으면 바로 연결, 연결됨이면 살아 있는지 확인. 연결 중이 아니면(로그아웃) 아무것도 안 함. */
    wake: () => void;
    /** REQ 송신 후 RES 까지 대기. 응답 err 가 있으면 reject. */
    call: <Res = unknown>(type: string, data?: unknown, opt?: CallOption) => Promise<Res>;
    /** EVENT 단방향 송신(fire-and-forget). 미연결이면 조용히 무시. */
    emit: (type: string, data?: unknown) => void;
    /** EVENT 수신 핸들러 등록. 해제 함수를 반환(같은 type 다중 구독 가능). */
    on: <D = unknown>(type: string, handler: (data: D) => void) => () => void;
};

const isConnected = (s: Connect['Status']) => s === 'CONNECTED';
const STALE_MS = 5000;      // 연결 중인 채 이보다 오래면 멈춘 시도로 본다(wake)
const PROBE_MS = 3000;      // 화면 복귀 때 PING 응답을 기다리는 시간

// url이 함수면 연결(재연결 포함)할 때마다 다시 만든다 — 쿼리에 그때의 값(탭 id 등)을 싣기 위해
export const createWebsocketHook = (url: string | (() => string), option: WebsocketOption = {}) => {
    const {
        maxRetries,
        retryInterval = 2000,
        retryMax = 30000,
        protocol = jsonFrame,
        callTimeout = 10000,
    } = option;

    // url 당 단일 컨텍스트(소켓·correlator·핸들러)를 모든 호출부가 공유.
    let ctx: SocketContext | null = null;
    const resolveUrl = () => typeof url === 'function' ? url() : url;
    const label = typeof url === 'function' ? resolveUrl().split('?')[0] : url;
    // 인스턴스별 고유 주입 키(여러 url로 여러 번 호출될 수 있어 문자열 키 충돌을 피함).
    const key = Symbol(`WebsocketContext:${label}`);

    const build = (): SocketContext => {
        const retry = ref(0);
        const event = ref<Event | null>(null);
        const ws = ref<WebSocket | null>(null);
        const status = ref<Connect['Status']>('PENDING');
        let active = false;     // connect() ~ disconnect() 사이 = 연결을 유지해야 함(로그인 중)
        let timer: ReturnType<typeof setTimeout> | undefined;     // 예약된 재연결
        let connectingAt = 0;

        // correlator (Conn.Correlator 대응)
        let seq = 0;
        const pending = new Map<number, Pending>();
        // event 디스패치 테이블 (Conn.on 대응, type 당 다중 핸들러)
        const handlers = new Map<string, Set<(data: any) => void>>();

        const rejectAllPending = (err: Error) => {
            pending.forEach((p) => p.reject(err));
            pending.clear();
        };

        // 소켓을 핸들러째 떼고 닫는다 — 그 소켓의 onclose가 재연결을 또 예약하지 않게
        const drop = () => {
            const s = ws.value;
            if (!s) return;
            s.onclose = null;
            s.onopen = null;
            s.onerror = null;
            s.onmessage = null;
            s.close();
            ws.value = null;
            rejectAllPending(new Error('websocket: connection dropped'));
        };

        const connect = () => {
            clearTimeout(timer);
            timer = undefined;
            // 연결 시도가 이미 진행 중이면 중복 소켓 생성 방지(#3).
            if (ws.value && ws.value.readyState === WebSocket.CONNECTING) return;
            drop();     // 살아 있는 연결이면 조용히 교체, 닫힌 옛 소켓은 정리
            active = true;
            retry.value += 1;
            const sock = new WebSocket(resolveUrl());
            sock.binaryType = 'arraybuffer'; // Go BinaryMessage 수신용
            connectingAt = Date.now();
            ws.value = sock;
            status.value = 'CONNECTING';
        };

        const disconnect = () => {
            active = false;
            clearTimeout(timer);
            timer = undefined;
            rejectAllPending(new Error('websocket: disconnected'));
            drop();
            retry.value = 0;
            status.value = 'PENDING';
        };

        const call = <Res = unknown>(type: string, data?: unknown, opt: CallOption = {}) =>
            new Promise<Res>((resolve, reject) => {
                if (!ws.value || !isConnected(status.value)) {
                    reject(new Error('websocket: not connected'));
                    return;
                }
                if (opt.signal?.aborted) {
                    reject(new Error('websocket: aborted'));
                    return;
                }

                const id = ++seq;
                let timer: ReturnType<typeof setTimeout> | undefined;

                // 단 한 번만 정리(타이머·abort 리스너·pending 제거) 후 결말.
                const settle = (fin: () => void) => {
                    if (timer) clearTimeout(timer);
                    opt.signal?.removeEventListener('abort', onAbort);
                    pending.delete(id);
                    fin();
                };
                const onAbort = () => settle(() => reject(new Error('websocket: aborted')));

                pending.set(id, {
                    resolve: (v) => settle(() => resolve(v as Res)),
                    reject: (e) => settle(() => reject(e)),
                });

                const ms = opt.timeout ?? callTimeout;
                if (ms > 0) {
                    timer = setTimeout(
                        () => settle(() => reject(new Error(`websocket: call timeout (${type})`))),
                        ms,
                    );
                }
                opt.signal?.addEventListener('abort', onAbort);

                ws.value.send(protocol.request(id, type, data));
            });

        const emit = (type: string, data?: unknown) => {
            if (!ws.value || !isConnected(status.value)) return;
            ws.value.send(protocol.event(type, data));
        };

        // 재연결-2(돌아오면 바로)·재연결-3(죽은 연결 확인) — REF-socket-reconnect.md
        // 백그라운드에서 연 소켓은 연결 중인 채 멈춰 있을 수 있어 STALE_MS가 지났으면 버리고 새로 연다
        const wake = () => {
            if (!active) return;
            if (isConnected(status.value)) return probe();
            if (ws.value?.readyState === WebSocket.CONNECTING && Date.now() - connectingAt < STALE_MS) return;
            drop();
            retry.value = 0;
            connect();
        };

        let probing = false;
        const probe = () => {
            if (probing) return;
            probing = true;
            const sock = ws.value;
            call('PING', undefined, { timeout: PROBE_MS })
                .catch((err) => {
                    // 시간 초과만 = 죽은 연결. 닫혀서 실패한 건 onclose가 이미 재연결을 예약했고, 오류 응답은 살아 있다는 뜻(옛 서버엔 PING이 없음)
                    if (!active || ws.value !== sock || !String(err?.message).includes('timeout')) return;
                    drop();
                    retry.value = 0;
                    connect();
                })
                .finally(() => { probing = false; });
        };

        const on = <D = unknown>(type: string, handler: (data: D) => void) => {
            let set = handlers.get(type);
            if (!set) {
                set = new Set();
                handlers.set(type, set);
            }
            set.add(handler as (data: any) => void);
            return () => {
                const s = handlers.get(type);
                if (!s) return;
                s.delete(handler as (data: any) => void);
                if (s.size === 0) handlers.delete(type);
            };
        };

        // 수신 디스패치 (Conn.Serve + dispatch 대응)
        const dispatch = (raw: string | ArrayBuffer) => {
            const m = protocol.decode(raw);
            switch (m.kind) {
                case 'response': {
                    const p = pending.get(m.id);
                    if (!p) return; // 늦게 온/모르는 응답
                    if (m.err) p.reject(new Error(m.err));
                    else p.resolve(m.data);
                    break;
                }
                case 'event': {
                    const set = handlers.get(m.type);
                    if (set) set.forEach((h) => h(m.data));
                    break;
                }
                // 'request' | 'unknown' → 브라우저는 처리하지 않음
            }
        };

        // ws ref 가 바뀔 때마다 수신/수명 핸들러를 (재)설치 = Go Serve 루프 대용.
        watch(
            ws,
            (sock) => {
                if (!sock) return;

                sock.onmessage = (ev) => dispatch(ev.data);

                sock.onopen = () => {
                    status.value = 'CONNECTED';
                    retry.value = 0;
                };

                sock.onerror = (ev) => {
                    event.value = ev;
                    status.value = 'FAILED';
                };

                sock.onclose = () => {
                    // 끊김 → 매달린 call 전부 실패(Go corr.Close 대응).
                    rejectAllPending(new Error('websocket: connection closed'));
                    if (!active || ws.value !== sock) return;
                    if (maxRetries && retry.value >= maxRetries) {
                        retry.value = 0;
                        status.value = 'PENDING';
                        return;
                    }
                    // 재연결-1(포기 없음): 실패할수록 간격을 두 배로(2초, 4초 ... retryMax). 화면 복귀·네트워크 복구 땐 wake가 바로 연결
                    status.value = 'CONNECTING';
                    clearTimeout(timer);
                    timer = setTimeout(connect, Math.min(retryInterval * 2 ** retry.value, retryMax));
                };
            },
            { immediate: true },
        );

        return { status, event, connect, disconnect, wake, call, emit, on };
    };

    // root(App.vue 등) setup에서 한 번 호출 — build()+provide(key, ctx)로 생명주기 초기화 지점을 명시.
    const provideSocket = (): SocketContext => {
        if (!ctx) ctx = build();
        provide(key, ctx);
        return ctx;
    };

    // 하위 트리 어디서나 호출 — provideSocket()이 상위에서 안 불렸으면 즉시 throw.
    const useSocket = (): SocketContext => {
        const shared = inject<SocketContext>(key)!;
        if (!shared) throw new Error(`websocket: provideSocket()이 상위에서 호출되지 않았습니다 (url=${label})`);

        // 이 호출부(컴포넌트)에서 등록한 on 구독은 unmount 시 자동 해제.
        const localUnsubs: Array<() => void> = [];
        const wrapped: SocketContext = {
            ...shared,
            on: (type, handler) => {
                const unsub = shared.on(type, handler);
                localUnsubs.push(unsub);
                return unsub;
            },
        };

        onUnmounted(() => {
            localUnsubs.forEach((u) => u());
            localUnsubs.length = 0;
        });

        return wrapped;
    };

    return [provideSocket, useSocket] as const;
};

// 서버 subscribe.go 의 TEST 핸들러 왕복 검증용 (root에서 provideTestSocket() 먼저 호출):
//   provideTestSocket();
//   const s = useTestSocket(); s.connect();
//   s.call('TEST', 'hello').then(console.log) // → 'RES'
export const [provideTestSocket, useTestSocket] = createWebsocketHook(
    () => {
        // API 주소에서 만든다(http -> ws, https -> wss) — https 화면에서 ws://는 혼합 콘텐츠로 막힘
        const url = new URL(`${getBaseURL()}/subscribe?tabId=${encodeURIComponent(readTabId())}`);
        url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
        return url.toString();
    },
    {},     // 재연결 횟수 제한 없음(재연결-1(포기 없음)) — 옛 maxRetries: 2는 모바일 복귀 때 두 번 다 실패하면 영영 끊긴 채였음
);
