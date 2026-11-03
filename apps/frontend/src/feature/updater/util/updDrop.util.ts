// Files 타일 끌어 놓기(단위-나) — PC 탐색기에서 놓은 파일·폴더 -> 올릴 목록. 폴더는 이름을 유지한다(결정-5(끌어 놓기 규칙))
export type UpItem = { file: File; path: string }

type Taken = { entry: FileSystemEntry | null; file: File | null }

// drop 처리 안에서 await 전에 동기로 불러야 한다 — 이벤트가 끝나면 items가 비어 entry를 못 꺼낸다
export const takeDropped = (dt: DataTransfer | null): Taken[] =>
    Array.from(dt?.items ?? []).filter(it => it.kind === 'file').map(it => ({ entry: it.webkitGetAsEntry?.() ?? null, file: it.getAsFile() }))

const fileOf = (e: FileSystemFileEntry) => new Promise<File>((ok, ng) => e.file(ok, ng))
// readEntries는 한 번에 일부(Chrome 100개)만 주므로 빈 배열이 나올 때까지 반복
const childrenOf = async (d: FileSystemDirectoryEntry) => {
    const reader = d.createReader(), out: FileSystemEntry[] = []
    for (;;) {
        const part = await new Promise<FileSystemEntry[]>((ok, ng) => reader.readEntries(ok, ng))
        if (!part.length) return out
        out.push(...part)
    }
}
const walk = async (e: FileSystemEntry, out: UpItem[]) => {
    if (e.isFile) out.push({ file: await fileOf(e as FileSystemFileEntry), path: e.fullPath.replace(/^\/+/, '') })
    else if (e.isDirectory) for (const c of await childrenOf(e as FileSystemDirectoryEntry)) await walk(c, out)
}

// dir = 놓을 폴더('' = 루트). 빈 폴더는 파일이 없어 목록에 안 남는다(끌기-2(빈 폴더))
export const readDropped = async (taken: Taken[], dir: string): Promise<UpItem[]> => {
    const out: UpItem[] = []
    for (const t of taken) {
        if (t.entry) await walk(t.entry, out)
        else if (t.file) out.push({ file: t.file, path: t.file.name })
    }
    return dir ? out.map(it => ({ ...it, path: `${dir}/${it.path}` })) : out
}

// 폴더 업로드 버튼 — 고른 폴더 이름부터 유지(결정-6(폴더 업로드 버튼)), 파일 고르기 = 이름만
export const pickedItems = (list: FileList | File[]): UpItem[] =>
    Array.from(list).map(f => ({ file: f, path: f.webkitRelativePath || f.name }))
