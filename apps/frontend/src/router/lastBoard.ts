import { readUserPref, writeUserPref } from '@/common/util/userPrefs.util';
import type { Router } from 'vue-router';

// 관리 페이지 '보드로' = 마지막으로 보던 보드 주소(focus·view 포함) — REF-member-ui.md 주소-7(보드로)
const KEY = 'board.last';

// 보드 URL은 보는 중에도 replace로 바뀌므로 이동이 끝날 때마다 덮어쓴다
export const installLastBoard = (router: Router) => {
    router.afterEach((to, _from, failure) => {
        if (!failure && to.meta.board) writeUserPref(KEY, to.fullPath);
    });
}

// 직전 기록이 보드면 back(관리 페이지가 기록에 남지 않게), 아니면(새 탭으로 바로 연 경우) 마지막 보드로 push
export const goLastBoard = (router: Router) => {
    const back = router.options.history.state.back;
    if (typeof back === 'string' && router.resolve(back).meta.board) {
        router.back();
        return;
    }
    void router.push(readUserPref<string>(KEY, '/'));
}
