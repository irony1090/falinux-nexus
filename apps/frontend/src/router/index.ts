/**
 * router/index.ts
 *
 * Manual routes for ./src/pages/*.vue
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router'
import { installAuthGuard } from './authGuard'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        // 보드(작업 공간): 기본 = `/`, 그 외 = `/w/{hash}` — REF-workspace.md "URL". meta.board = 헤더에 보드 메뉴 자리
        {
            path: '/',
            component: () => import('@/pages/index.vue'),
            meta: { board: true },
        },
        {
            path: '/w/:hash',
            component: () => import('@/pages/index.vue'),
            meta: { board: true },
        },
        {
            path: '/login',
            component: () => import('@/pages/Login.vue'),
            meta: { public: true },
        }
    ],
})

installAuthGuard(router)

export default router
