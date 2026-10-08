// 서버 superdb.Perm* 와 같은 값 — 단계 비교는 >=
export const PERM_BLOCK = 0
export const PERM_VIEW = 10
export const PERM_OPERATE = 20
export const PERM_EDIT = 30

const PERM_INFO: Array<{ min: number, icon: string, text: string }> = [
    { min: PERM_EDIT, icon: 'mdi-pencil', text: '편집' },
    { min: PERM_OPERATE, icon: 'mdi-keyboard', text: '조작' },
    { min: PERM_VIEW, icon: 'mdi-eye', text: '보기' },
]

// 공유받은 것의 단계 아이콘·이름. 내 것(소유자 40)이거나 단계를 모르면 undefined
export const permInfo = (perm: number | undefined) => perm === undefined || perm > PERM_EDIT ? undefined : PERM_INFO.find(i => perm >= i.min)

// 보기 단계 = 입력이 막힌 터미널
export const isReadOnly = (perm: number | undefined) => perm !== undefined && perm < PERM_OPERATE

export const PERM_OWNER = 40

// 공유 걸기 창의 선택지 — 차단(0)은 상위 공유가 있을 때만 의미가 있어 공유-e-3 목록과 함께
export const SHARE_LEVELS = [
    { value: PERM_VIEW, title: '보기', desc: '목록·내용, 출력 보기' },
    { value: PERM_OPERATE, title: '조작', desc: '+ 실행, 입력, 종료' },
    { value: PERM_EDIT, title: '편집', desc: '+ 추가·수정·삭제·이동' },
]
export const levelTitle = (perm: number | null | undefined) => SHARE_LEVELS.find(l => l.value === perm)?.title
