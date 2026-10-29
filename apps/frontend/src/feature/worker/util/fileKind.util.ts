// V1(확장자) — 서버 workerFile.go viewTypes와 같은 목록. 더블클릭·칸 메뉴 "보기"가 같이 쓴다(파트(보기 타일), REF-file-view.md)
const IMAGE_EXT = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg'])
const VIDEO_EXT = new Set(['mp4', 'webm', 'mov'])

export type ViewKind = 'image' | 'video'

export const viewKindOf = (path: string): ViewKind | undefined => {
    const name = path.slice(Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\')) + 1)
    const i = name.lastIndexOf('.')
    if (i <= 0) return undefined
    const ext = name.slice(i + 1).toLowerCase()
    return IMAGE_EXT.has(ext) ? 'image' : VIDEO_EXT.has(ext) ? 'video' : undefined
}

// 보기 타일 아이콘 — 종류(TILE_KINDS.viewer)는 하나, 영상만 다른 아이콘
export const viewIcon = (path: string) => viewKindOf(path) === 'video' ? 'mdi-movie-outline' : 'mdi-image-outline'
