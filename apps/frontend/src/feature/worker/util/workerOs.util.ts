// worker OS 표시 — 한 글자(L/W) + 툴팁 이름. 백엔드는 빈 값을 linux로 정규화해 보낸다 — WW5(OS 표시)
const OS_NAMES: Record<string, string> = { linux: 'Linux', windows: 'Windows', darwin: 'macOS' };

export const osName = (os: string) => OS_NAMES[os] ?? os;
export const osBadge = (os: string) => osName(os).charAt(0).toUpperCase();
export const osTitle = (os: string) => `OS: ${osName(os)}`;
