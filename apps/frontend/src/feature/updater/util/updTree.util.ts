import type { UpdFile } from '../api/updater.api'

// Files 타일 폴더 트리 — 서버엔 폴더가 없고 경로뿐이라 `/`로 쪼개 만든다(결정-3(폴더 트리)). path = 'web/assets'(끝 / 없음), 루트 = ''
export type TreeDir = { path: string; name: string; dirs: TreeDir[]; files: UpdFile[]; count: number }
export type TreeRow =
    | { kind: 'dir'; path: string; name: string; depth: number; dir: TreeDir }
    | { kind: 'file'; path: string; name: string; depth: number; file: UpdFile }

export const baseName = (path: string) => path.slice(path.lastIndexOf('/') + 1)
export const parentDir = (path: string) => path.slice(0, Math.max(0, path.lastIndexOf('/')))

export const buildTree = (files: UpdFile[]): TreeDir => {
    const root: TreeDir = { path: '', name: '', dirs: [], files: [], count: 0 }
    const index = new Map<string, TreeDir>([['', root]])
    const dirOf = (path: string): TreeDir => {
        let d = index.get(path)
        if (d) return d
        const parent = dirOf(parentDir(path))
        d = { path, name: baseName(path), dirs: [], files: [], count: 0 }
        parent.dirs.push(d)
        index.set(path, d)
        return d
    }
    for (const f of files) {
        dirOf(parentDir(f.path)).files.push(f)
        for (let p = parentDir(f.path); ; p = parentDir(p)) {
            index.get(p)!.count++
            if (!p) break
        }
    }
    return root
}

// 화면 행 — 폴더마다 파일 먼저, 그다음 하위 폴더. closed에 든 폴더는 안을 건너뛴다
export const treeRows = (root: TreeDir, closed: ReadonlySet<string>): TreeRow[] => {
    const out: TreeRow[] = []
    const walk = (d: TreeDir, depth: number) => {
        for (const f of d.files) out.push({ kind: 'file', path: f.path, name: baseName(f.path), depth, file: f })
        for (const c of d.dirs) {
            out.push({ kind: 'dir', path: c.path, name: c.name, depth, dir: c })
            if (!closed.has(c.path)) walk(c, depth + 1)
        }
    }
    walk(root, 0)
    return out
}

export const filesUnder = (d: TreeDir): string[] => [...d.files.map(f => f.path), ...d.dirs.flatMap(filesUnder)]
