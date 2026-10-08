// User-Agent 원문에서 "브라우저 · OS · 모바일 여부"만 뽑는다(라이브러리 없음) — SH7(브라우저 이름). 순서가 중요: Edge·Opera는 Chrome 문구도 포함하고, Chrome은 Safari 문구도 포함한다
export const parseUserAgent = (ua: string): string => {
    if (!ua) return '알 수 없는 브라우저'
    const browser = /Edg\//.test(ua) ? 'Edge'
        : /OPR\/|Opera/.test(ua) ? 'Opera'
        : /Firefox\/|FxiOS/.test(ua) ? 'Firefox'
        : /Chrome\/|CriOS/.test(ua) ? 'Chrome'
        : /Safari\//.test(ua) ? 'Safari'
        : '브라우저'
    const os = /Android/.test(ua) ? 'Android'
        : /iPhone|iPad|iPod/.test(ua) ? 'iOS'
        : /Windows/.test(ua) ? 'Windows'
        : /Mac OS X|Macintosh/.test(ua) ? 'macOS'
        : /Linux/.test(ua) ? 'Linux'
        : ''
    const mobile = /Mobi|Android|iPhone|iPod/.test(ua)
    return [browser, os, mobile ? '모바일' : ''].filter(Boolean).join(' · ')
}

// unix sec -> "방금 / N분 전 / N시간 전 / N일 전"
export const agoText = (sec: number, now = Date.now()): string => {
    const d = Math.max(0, Math.floor(now / 1000) - sec)
    if (d < 60) return '방금'
    if (d < 3600) return `${Math.floor(d / 60)}분 전`
    if (d < 86400) return `${Math.floor(d / 3600)}시간 전`
    return `${Math.floor(d / 86400)}일 전`
}

export const lastSeenText = (online: boolean, lastSeen?: number) =>
    online ? '접속 중' : lastSeen !== undefined ? `마지막 접속 ${agoText(lastSeen)}` : '접속 기록 없음'
