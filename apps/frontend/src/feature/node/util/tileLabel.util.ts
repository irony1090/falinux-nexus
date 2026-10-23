// 노드 타일 이름 — 이름을 받기 전에는 node id로 표시, nodeId null = 루트 목록
export const nodeLabel = (nodeId: number | null, loaded?: string) => nodeId === null ? '기본' : loaded ?? `node ${nodeId}`

// 장비 파일 실행 터미널 이름 — 파일 이름(폴더-4(실행))
export const fileLabel = (path: string) => path.split(/[\\/]/).pop() || path

// 장비 실행 경로 표시 — 끝 두 단계만(전체는 title로). 길면 파일 이름이 가려져서
export const shortPath = (p: string) => {
    const sep = p.includes('\\') && !p.includes('/') ? '\\' : '/'
    const segs = p.split(/[\\/]/).filter(s => !!s)
    return segs.length > 2 ? `…${sep}${segs.slice(-2).join(sep)}` : p
}

// 장비 타일 이름 — 경로 끝 이름, 최상위면 장비 키(main_key)
export const deviceLabel = (instance: string, path?: string) => {
    const last = path?.split(/[\\/]/).filter(s => !!s).pop()
    return last ?? instance.slice(0, instance.lastIndexOf('#'))
}
