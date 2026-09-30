// 사용자별 화면 설정(기기마다 다른 값) — localStorage, 로그인이 풀리면(auth = null) 전부 삭제
// 저장소를 못 쓰는 환경(시크릿 창 등)이면 읽기는 기본값, 쓰기·삭제는 무시

const PREFIX = 'nexus.user.';

export const readUserPref = <T>(key: string, fallback: T): T => {
	try {
		const raw = localStorage.getItem(PREFIX + key);
		return raw === null ? fallback : JSON.parse(raw) as T;
	} catch {
		return fallback;
	}
}

export const writeUserPref = <T>(key: string, value: T) => {
	try {
		localStorage.setItem(PREFIX + key, JSON.stringify(value));
	} catch { /* 저장 불가 */ }
}

export const clearUserPrefs = () => {
	try {
		Object.keys(localStorage)
			.filter(k => k.startsWith(PREFIX))
			.forEach(k => localStorage.removeItem(k));
	} catch { /* 저장 불가 */ }
}
