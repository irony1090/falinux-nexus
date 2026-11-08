import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'

// 백엔드: apps/core/cmd/supervisor/router/settings.go (group "/settings") — 계정 겹, 값 모양 검사는 settings.store SPECS

// rev = 가장 최근 저장 시각(µs) — 늦게 온 옛 값을 버리는 기준. SETTINGS:UPDATE push도 같은 모양
export type AccountSettingsResponse = {
    settings: Record<string, unknown>
    rev: number
}

// GET /settings — 계정 값 전체
export const getAccountSettings = () => BaseAxios.get(
    '/settings'
).then(throwThen<AccountSettingsResponse>).catch(throwCatch)

// PUT /settings/:key — 저장 뒤 계정 값 전체
export const putAccountSetting = (key: string, value: unknown) => BaseAxios.put(
    `/settings/${key}`, { value }
).then(throwThen<AccountSettingsResponse>).catch(throwCatch)
