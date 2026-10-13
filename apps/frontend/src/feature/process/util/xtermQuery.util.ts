import type { IDisposable, Terminal } from '@xterm/xterm'

// 창 정보 보고를 요청하는 CSI t 첫 번째 인자(크기·위치·제목 보고) — 나머지(제목 저장·복원 등)는 응답이 없어 그대로 둔다
const WINDOW_REPORTS = new Set([11, 13, 14, 15, 16, 18, 19, 20, 21])

// muteQueries는 allow()가 false인 동안 xterm이 질의 시퀀스에 응답하지 않게 한다.
// 같은 터미널을 연 탭마다 xterm이 응답하면 process가 응답을 탭 수만큼 받으므로 크기 소유 탭만 응답한다(파트(질의 응답 중복)).
// 핸들러가 true를 돌려주면 xterm 기본 처리(응답 생성)가 돌지 않는다. 입력 경로(onData)는 건드리지 않는다.
export const muteQueries = (term: Terminal, allow: () => boolean): IDisposable => {
    const mute = () => !allow()
    const p = term.parser
    const ds: IDisposable[] = [
        p.registerCsiHandler({ final: 'c' }, mute),                                   // DA1(단말 종류)
        p.registerCsiHandler({ prefix: '>', final: 'c' }, mute),                      // DA2
        p.registerCsiHandler({ final: 'n' }, mute),                                   // DSR(상태·커서 위치)
        p.registerCsiHandler({ prefix: '?', final: 'n' }, mute),                      // DEC DSR
        p.registerCsiHandler({ intermediates: '$', final: 'p' }, mute),               // DECRQM(모드 질의)
        p.registerCsiHandler({ prefix: '?', intermediates: '$', final: 'p' }, mute),  // DEC DECRQM
        p.registerDcsHandler({ intermediates: '$', final: 'q' }, mute),               // DECRQSS(설정 질의)
        p.registerCsiHandler({ final: 't' }, params => WINDOW_REPORTS.has(Number(params[0])) && mute()),
        // 색 질의만(데이터에 `?`) — 색 설정은 기본 처리로 보낸다
        ...[4, 10, 11, 12].map(id => p.registerOscHandler(id, data => data.split(';').includes('?') && mute())),
    ]
    return { dispose: () => ds.forEach(d => d.dispose()) }
}
