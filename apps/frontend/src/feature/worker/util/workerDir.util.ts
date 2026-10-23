import type { DirEntry } from '../api/worker.api'

// 장비 타일 목록 표시 — 폴더-2(장비 타일)

export const sizeText = (n: number) => {
    if (n < 1024) return `${n} B`
    const units = ['K', 'M', 'G', 'T']
    let v = n / 1024, i = 0
    while (v >= 1024 && i < units.length - 1) {
        v /= 1024
        i++
    }
    return `${v < 10 ? v.toFixed(1) : Math.round(v)} ${units[i]}`
}

const pad = (v: number) => String(v).padStart(2, '0')

export const timeText = (sec: number) => {
    if (!sec) return ''
    const d = new Date(sec * 1000)
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export const isDirLike = (e: DirEntry) => e.kind === 'dir' || (e.kind === 'link' && !!e.targetDir)
// 파일 하나로 옮길 수 있는 것(파일 링크는 원본 기준, 끊긴 링크 제외) — 전송-10(worker 간 중계)
export const isFileLike = (e: DirEntry) => e.kind === 'file' || (e.kind === 'link' && !!e.targetPath && !e.targetDir)

// 경로의 상위 폴더(구분자 `/`·`\` 둘 다). `/x` -> `/`, `C:\x` -> `C:\`
export const parentPathOf = (p: string) => {
    const i = Math.max(p.lastIndexOf('/'), p.lastIndexOf('\\'))
    const dir = p.slice(0, Math.max(i, 0))
    if (dir === '') return p.startsWith('/') ? '/' : p
    return /^[A-Za-z]:$/.test(dir) ? `${dir}\\` : dir
}

// 링크의 "원본으로" 갈 곳 — 원본이 폴더면 그 폴더, 파일이면 든 폴더(Q22(링크 원본이 파일))
export const linkDest = (e: DirEntry) => e.targetPath ? (e.targetDir ? e.targetPath : parentPathOf(e.targetPath)) : undefined

export const entryIcon = (e: DirEntry) => {
    if (e.kind === 'dir') return 'mdi-folder'
    if (e.kind === 'link') return e.targetPath ? (e.targetDir ? 'mdi-folder-arrow-right' : 'mdi-file-link') : 'mdi-link-off'
    if (e.kind === 'file') return e.exec ? 'mdi-file-cog' : 'mdi-file-outline'
    return 'mdi-file-question-outline'
}
