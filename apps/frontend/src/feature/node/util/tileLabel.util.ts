// 노드 타일 이름 — 이름을 받기 전에는 node id로 표시, nodeId null = 루트 목록
export const nodeLabel = (nodeId: number | null, loaded?: string) => nodeId === null ? '기본' : loaded ?? `node ${nodeId}`
