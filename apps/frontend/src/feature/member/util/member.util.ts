// 회원 관리 화면 공용 — REF-member.md

// http(보안 문맥 아님)에서는 navigator.clipboard가 없어 숨긴 textarea로 복사한다
export const copyText = async (text: string): Promise<void> => {
    if (navigator.clipboard && window.isSecureContext) return navigator.clipboard.writeText(text)
    const ta = document.createElement('textarea')
    ta.value = text
    ta.setAttribute('readonly', '')
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    const ok = document.execCommand('copy')
    ta.remove()
    if (!ok) throw new Error('복사하지 못했습니다')
}

// worker env에 그대로 붙여 넣는 두 줄
export const workerEnvLines = (mainKey: string, apiKey: string) => `NAME="${mainKey}"\nAPI_KEY="${apiKey}"`

// 장비 이름 규칙 — 서버 deviceKey.go mainKeyPattern과 같게
export const MAIN_KEY_PATTERN = /^[A-Za-z0-9._-]{1,64}$/
export const API_KEY_PREFIX = 'nxk_'

export const dateText = (sec: number) => new Date(sec * 1000).toLocaleString('ko-KR', { dateStyle: 'short', timeStyle: 'short' })

// 인스턴스 표시 이름 — 라벨이 있으면 "라벨 (서브키)", 없으면 서브키
export const instanceText = (subKey: string, label?: string) => label ? `${label} (${subKey})` : subKey

export const dayText = (sec: number) => new Date(sec * 1000).toLocaleDateString('sv-SE')

// 장비 상태 — 서버 memberList.go deviceState와 같은 판정(목록 필터 on·off·none)
export type DeviceState = 'on' | 'off' | 'none'
export const deviceState = (instances: { online: boolean }[]): DeviceState =>
    !instances.length ? 'none' : instances.some(i => i.online) ? 'on' : 'off'
export const deviceStateText = (instances: { online: boolean }[]) => {
    const n = instances.filter(i => i.online).length
    return n ? `접속 ${n}` : instances.length ? '꺼짐' : '기록 없음'
}

// 페이지 번호 — 처음·끝 + 지금 앞뒤(넓으면 1칸씩), 사이는 …(null)
export const pageNumbers = (page: number, pages: number, wide: boolean): Array<number | null> => {
    const keep = new Set([1, pages])
    const r = wide ? 1 : 0
    for (let p = page - r; p <= page + r; p++) if (p >= 1 && p <= pages) keep.add(p)
    if (pages <= (wide ? 7 : 5)) for (let p = 1; p <= pages; p++) keep.add(p)
    const out: Array<number | null> = []
    let last = 0
    for (const p of [...keep].sort((a, b) => a - b)) {
        if (p - last > 1) out.push(null)
        out.push(p)
        last = p
    }
    return out
}
