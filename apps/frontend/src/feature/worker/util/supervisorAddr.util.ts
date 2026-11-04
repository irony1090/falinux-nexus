// 저장된 주소(ws://host:port)를 주소 변경 창의 두 입력칸(방식·호스트)으로 나눈다 — 순위 눌러 입력, REF-supervisor-move.md
export const splitWsAddr = (url: string): { scheme: 'ws' | 'wss'; host: string } | undefined => {
    const m = /^(wss?):\/\/(.+)$/.exec(url.trim());
    return m ? { scheme: m[1] as 'ws' | 'wss', host: m[2] } : undefined;
};
