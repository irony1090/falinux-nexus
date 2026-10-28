// 노드 타일 이름 — 이름을 받기 전에는 node id로 표시, nodeId null = 루트 목록
export const nodeLabel = (nodeId: number | null, loaded?: string) => nodeId === null ? '기본' : loaded ?? `node ${nodeId}`

// 장비 타일 "여기서 셸" cmd — 서버 protocol.PlaceholderWorkerShell과 같게(worker가 OS 셸로 바꿈)
export const SHELL_CMD = '{WORKER_SHELL}'

// 장비 파일 편집 cmd — 서버 protocol.PlaceholderWorkerEditor와 같게(worker가 편집기로 바꿈), 파일 = args[0]. 파트(파일 편집기)
export const EDITOR_CMD = '{WORKER_EDITOR}'

// 장비 파일 실행 터미널 이름 — 파일 이름(폴더-4(실행)), 셸 = "셸", 편집 = "편집 <파일 이름>"
const baseName = (p: string) => p.split(/[\\/]/).pop() || p
export const fileLabel = (path: string, args: readonly string[] = []) =>
    path === SHELL_CMD ? '셸' : path === EDITOR_CMD ? `편집 ${baseName(args[0] ?? '')}`.trimEnd() : baseName(path)

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
