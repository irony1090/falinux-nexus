import { authState, ensureSession } from '@/feature/user/store/auth.store';
import { watch } from 'vue';
import type { RouteLocationNormalized, RouteLocationRaw, Router } from 'vue-router';

// 비로그인 = `/login?redirect=원래 주소`, 로그인 상태로 `/login` = redirect 주소로 — 로그인-2(라우터 가드)
// meta.public = 로그인 없이 보는 페이지

const loginRoute = (from: string): RouteLocationRaw => ({ path: '/login', query: from === '/' ? {} : { redirect: from } });

// 같은 사이트 경로만 따른다(`//host` 같은 바깥 주소 거부)
const redirectTarget = (to: RouteLocationNormalized): string => {
    const r = to.query.redirect;
    return typeof r === 'string' && r.startsWith('/') && !r.startsWith('//') ? r : '/';
}

export const installAuthGuard = (router: Router) => {
    router.beforeEach(async to => {
        await ensureSession();
        const authed = !!authState.value;
        if (to.meta.public) return authed ? redirectTarget(to) : true;
        return authed ? true : loginRoute(to.fullPath);
    });

    // 보는 중에 로그인 상태가 바뀜(로그인·로그아웃·세션 만료). 첫 이동이 끝나기 전(matched 없음)은 가드가 맡는다
    watch(authState, val => {
        const cur = router.currentRoute.value;
        if (val === undefined || !cur.matched.length) return;
        if (!val && !cur.meta.public) router.replace(loginRoute(cur.fullPath));
        else if (val && cur.meta.public) router.replace(redirectTarget(cur));
    });
}
