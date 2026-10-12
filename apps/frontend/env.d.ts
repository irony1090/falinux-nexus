/// <reference types="vite/client" />
/// <reference types="vite-plugin-vue-layouts-next/client" />

interface ImportMetaEnv {
    readonly VITE_API_URL?: string
    readonly VITE_DEV_LOGIN_ID?: string     // 개발용 로그인 폼 기본값 — `.env.development.local`(git 제외)
    readonly VITE_DEV_LOGIN_PW?: string
}
