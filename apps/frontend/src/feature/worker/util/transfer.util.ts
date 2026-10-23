import type { TransferInfo } from '../api/transfer.api'

// 한 줄 단계 표시 — Q25(진행률 구간)
export const transferLabel = (t: TransferInfo) => {
    switch (t.state) {
        case 'done': return '완료'
        case 'aborted': return '중단됨'
        case 'failed': return CODE_TEXT[t.code ?? ''] ?? '실패'
    }
    return PHASE_TEXT[t.phase]
}

const PHASE_TEXT: Record<TransferInfo['phase'], string> = {
    upload: '올리는 중',
    send: '장비로 보내는 중',
    download: '내려받는 중',
}

export const KIND_ICON: Record<TransferInfo['kind'], string> = {
    upload: 'mdi-upload',
    download: 'mdi-download',
    copy: 'mdi-content-copy',
}

const CODE_TEXT: Record<string, string> = {
    exists: '같은 이름 파일 있음',
    is_dir: '같은 이름 폴더 있음',
    no_parent: '받을 폴더 없음',
    bad_path: '잘못된 경로',
}

// 크기를 모르면(zip) undefined — 진행 막대는 움직이는 막대로
export const transferPercent = (t: TransferInfo) => t.state === 'done' ? 100 : t.size < 0 ? undefined : t.size ? Math.floor(t.done * 100 / t.size) : 0

export const transferColor = (t: TransferInfo) => ({ done: 'success', failed: 'error', aborted: undefined, running: 'primary' } as const)[t.state]
