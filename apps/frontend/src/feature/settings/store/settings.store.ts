import { readUserPref, userPrefKey, writeUserPref } from '@/common/util/userPrefs.util'
import { authState } from '@/feature/user/store/auth.store'
import type { TileBarAt } from '@/feature/widget/util/tileBar.util'
import { readonly, ref, watch } from 'vue'

// 설정 저장소 — 항목마다 범위를 둔다(Q2(설정 범위)). 지금은 브라우저 범위뿐 = userPrefs(로그아웃하면 지워짐, Q14(로그아웃 때 설정)) — REF-settings-footer.md
type Spec<T> = { scope: 'browser'; init: T; ok: (v: unknown) => v is T }

const SPECS = {
    tileBarAt: { scope: 'browser', init: 'top', ok: (v): v is TileBarAt => v === 'top' || v === 'bottom' } as Spec<TileBarAt>,
}

type Specs = typeof SPECS
export type SettingKey = keyof Specs
export type Settings = { [K in SettingKey]: Specs[K] extends Spec<infer T> ? T : never }

const KEYS = Object.keys(SPECS) as SettingKey[]
const prefKey = (k: SettingKey) => `settings.${k}`

// 저장값이 없거나 모양이 다르면(옛 값·손상) 기본값
const read = <K extends SettingKey>(k: K): Settings[K] => {
    const s = SPECS[k] as Spec<Settings[K]>
    const v = readUserPref<unknown>(prefKey(k), s.init)
    return s.ok(v) ? v : s.init
}
const readAll = () => Object.fromEntries(KEYS.map(k => [k, read(k)])) as Settings

const values = ref<Settings>(readAll())
const dialogOpen = ref(false)

// storage 이벤트 = 다른 탭의 변경만(key null = 전체 삭제)
window.addEventListener('storage', ev => {
    if (ev.key === null || KEYS.some(k => ev.key === userPrefKey(prefKey(k)))) values.value = readAll()
})
// 이 탭의 로그아웃은 storage 이벤트가 없어 로그인 상태로 다시 읽는다(nodeClipboard.store와 같은 방식)
watch(authState, auth => {
    values.value = readAll()
    if (!auth) dialogOpen.value = false
})

export const settings = readonly(values)

export const setSetting = <K extends SettingKey>(k: K, v: Settings[K]) => {
    values.value = { ...values.value, [k]: v }
    writeUserPref(prefKey(k), v)
}

// 설정 창 — App.vue에 하나(SettingsDialog), 여는 곳은 푸터 오른쪽 메뉴
export const settingsDialog = dialogOpen
export const openSettings = () => dialogOpen.value = true
