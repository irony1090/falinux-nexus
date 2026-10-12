import { onUnauthorized, type CatchErrorResponse } from '@/common/api/api.util';
import { clearUserPrefs } from '@/common/util/userPrefs.util';
import { readonly, ref } from 'vue';
import { checkSession, signIn, signOut, type UserResponse } from '../api/user.api';

// 로그인 상태는 모듈 하나에 둔다 — 라우터 가드(컴포넌트 밖)와 화면이 같은 값을 본다 — 로그인-1(상태를 모듈로)
// undefined = 아직 확인 전 / null = 로그인 안 됨 / 값 = 로그인됨
const auth = ref<UserResponse | null | undefined>();
const offline = ref(false);     // 마지막 세션 확인이 서버 연결 실패(401 아님)
const loading = ref(false);
let pending: Promise<void> | null = null;

// 사용자 설정은 로그아웃·세션 만료(401)일 때만 지운다. 서버 연결 실패로는 지우지 않는다 — 로그인-3(401만 로그아웃)
const expire = () => {
    auth.value = null;
    clearUserPrefs();
}

const confirm = (): Promise<void> => {
    if (pending) return pending;
    loading.value = true;
    pending = checkSession()
    .then(res => {
        auth.value = res;
        offline.value = false;
    })
    .catch((err: CatchErrorResponse) => {
        if (err?.status === 401) {
            offline.value = false;
            expire();
            return;
        }
        offline.value = true;
        if (auth.value === undefined) auth.value = null;    // 처음 확인이 실패하면 로그인 화면으로(설정은 유지)
    })
    .finally(() => {
        loading.value = false;
        pending = null;
    });
    return pending;
}

// 라우터 가드용 — 처음 한 번만 서버에 묻고, 이후로는 바로 끝난다
export const ensureSession = () => auth.value === undefined ? confirm() : Promise.resolve();

// 다른 API가 401을 받으면 = 세션 만료
onUnauthorized(() => {
    if (auth.value) expire();
});

const login = (id: string, pw: string) => {
    loading.value = true;
    return signIn({ identification: id, password: pw })
    .then(() => confirm())
    .finally(() => loading.value = false);
}

// 서버 요청이 실패해도 이 화면은 로그아웃 상태로 둔다
const logout = () => {
    loading.value = true;
    return signOut()
    .catch(() => {})
    .finally(() => {
        loading.value = false;
        expire();
    });
}

const ctx = {
    auth: readonly(auth),
    offline: readonly(offline),
    loading: readonly(loading),
    login, logout,
}

export const authState = ctx.auth;

// App.vue에서 한 번 — 앱이 뜨자마자 세션을 확인한다
export const provideAuthStore = () => {
    ensureSession();
    return ctx;
}

export const useAuthStore = () => ctx;
