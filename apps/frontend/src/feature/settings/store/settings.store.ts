import type { SocketContext } from '@/common/websocket/websocket.hook'
import { readUserPref, removeUserPref, userPrefKey, writeUserPref } from '@/common/util/userPrefs.util'
import { authState } from '@/feature/user/store/auth.store'
import type { TileBarAt } from '@/feature/widget/util/tileBar.util'
import { computed, readonly, ref, watch } from 'vue'
import { getAccountSettings, putAccountSetting, type AccountSettingsResponse } from '../api/settings.api'

// 설정 두 겹 — 쓰이는 값 = 이 브라우저에서만 ?? 계정 ?? 기본값(Q2(설정 범위) 개정, REF-settings-footer.md "설정 두 겹")
// 브라우저 겹 = userPrefs `settings.<항목>`(로그아웃하면 지워짐) / 계정 겹 = 서버 + 사본 userPrefs `settingsAccount`(Q25(새로고침 깜빡임))
type Spec<T> = { init: T; ok: (v: unknown) => v is T }

const SPECS = {
    tileBarAt: { init: 'top', ok: (v): v is TileBarAt => v === 'top' || v === 'bottom' } as Spec<TileBarAt>,
}

type Specs = typeof SPECS
export type SettingKey = keyof Specs
export type Settings = { [K in SettingKey]: Specs[K] extends Spec<infer T> ? T : never }
type Layer = Partial<Settings>

const KEYS = Object.keys(SPECS) as SettingKey[]
const browserKey = (k: SettingKey) => `settings.${k}`
const ACCOUNT_COPY = 'settingsAccount'

// 모양이 다른 값(옛 값·손상·다른 버전 프론트가 쓴 값)은 없는 것으로
const pick = (src: Record<string, unknown>): Layer => {
    const out: Record<string, unknown> = {}
    for (const k of KEYS) if ((SPECS[k] as Spec<unknown>).ok(src[k])) out[k] = src[k]
    return out as Layer
}
const readBrowser = () => pick(Object.fromEntries(KEYS.map(k => [k, readUserPref<unknown>(browserKey(k), undefined)])))
const readAccountCopy = () => pick(readUserPref<Record<string, unknown>>(ACCOUNT_COPY, {}))

const browser = ref<Layer>(readBrowser())
const account = ref<Layer>(readAccountCopy())
const dialogOpen = ref(false)

const layered = (...layers: Layer[]) => Object.fromEntries(
    KEYS.map(k => [k, layers.find(l => l[k] !== undefined)?.[k] ?? SPECS[k].init]),
) as Settings

export const settings = computed(() => layered(browser.value, account.value))
export const accountSettings = computed(() => layered(account.value))
export const browserSettings = readonly(browser)    // 항목이 없으면 = 계정 설정 따름

const applyAccount = (src: Record<string, unknown>) => {
    account.value = pick(src)
    writeUserPref(ACCOUNT_COPY, account.value)
}

/*
 * 서버 값 받기 순서 — HTTP 응답과 SETTINGS:UPDATE push는 다른 길이라 도착 순서가 뒤바뀔 수 있다.
 * - rev(서버 마지막 저장 시각)가 지금보다 작은 것은 버린다(옛 push가 새 값을 되돌리지 않게)
 * - 이 탭의 저장 요청이 진행 중이면 화면은 고른 값 그대로 두고 최신 서버 값만 기억, 마지막 요청이 끝날 때 적용
 *   (실패하면 그때 서버 값으로 되돌아감)
 */
let rev = -1
let latest: Record<string, unknown> | null = null
let inflight = 0
let gen = 0     // 로그인 상태가 바뀌면 증가 — 로그아웃 뒤 도착한 응답이 이전 계정 값을 쓰지 않게

const take = (res: AccountSettingsResponse) => {
    if (res.rev < rev) return
    rev = res.rev
    latest = res.settings
    if (!inflight) applyAccount(latest)
}

const loadAccount = async () => {
    const my = gen
    try {
        const res = await getAccountSettings()
        if (my === gen) take(res)
    } catch (err) {
        console.log('[SETTINGS] load', err)
    }
}

export const setAccountSetting = async <K extends SettingKey>(k: K, v: Settings[K]) => {
    const my = gen
    inflight++
    account.value = { ...account.value, [k]: v }
    try {
        const res = await putAccountSetting(k, v)
        if (my === gen) take(res)
    } catch (err) {
        console.log('[SETTINGS] save', err)
    } finally {
        if (my === gen && !--inflight && latest) applyAccount(latest)
    }
}

// undefined = 이 브라우저 값을 지우고 계정 설정을 따름
export const setBrowserSetting = <K extends SettingKey>(k: K, v: Settings[K] | undefined) => {
    const next = { ...browser.value }
    if (v === undefined) {
        delete next[k]
        removeUserPref(browserKey(k))
    } else {
        next[k] = v
        writeUserPref(browserKey(k), v)
    }
    browser.value = next
}

// storage 이벤트 = 다른 탭의 변경만(key null = 전체 삭제)
window.addEventListener('storage', ev => {
    if (ev.key === null || ev.key === userPrefKey(ACCOUNT_COPY)) account.value = readAccountCopy()
    if (ev.key === null || KEYS.some(k => ev.key === userPrefKey(browserKey(k)))) browser.value = readBrowser()
})
// 로그인 = 계정 값 받기 / 로그아웃 = userPrefs가 지워진 뒤라 두 겹 다 비게 됨(nodeClipboard.store와 같은 방식)
watch(authState, auth => {
    gen++
    rev = -1
    latest = null
    inflight = 0
    browser.value = readBrowser()
    account.value = readAccountCopy()
    if (auth) void loadAccount()
    else dialogOpen.value = false
}, { immediate: true })

// 앱 루트에서 한 번 — 다른 기기·탭의 계정 값 변경(Q26(다른 기기 반영)) + 끊긴 동안 놓친 변경
export const bindAccountSettings = (socket: SocketContext) => {
    socket.on<AccountSettingsResponse>('SETTINGS:UPDATE', take)
    let connected = false
    watch(socket.status, s => {
        if (s !== 'CONNECTED') return
        if (connected) void loadAccount()
        connected = true
    })
}

// 설정 창 — App.vue에 하나(SettingsDialog), 여는 곳은 푸터 오른쪽 메뉴
export const settingsDialog = dialogOpen
export const openSettings = () => dialogOpen.value = true
