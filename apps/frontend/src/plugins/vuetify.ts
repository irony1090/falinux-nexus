/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Composables
import { createVuetify, type VuetifyOptions } from 'vuetify'
// Styles
import '@mdi/font/css/materialdesignicons.css'
import { ko } from 'vuetify/locale'
import 'vuetify/styles'

const locale: VuetifyOptions['locale'] = {
	locale: 'ko',
	messages: { ko }
}

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  locale,
  theme: {
    defaultTheme: 'system',
    // 테마에 없는 색은 여기 등록 후 이름으로 사용 (REF-frontend "색 규칙") — 터미널은 light/dark 모두 어두운 배경 / share = 공유 관련 표시 전용(청록)
    // acct1~6 = 다른 계정 active 배지(user id % 6) — primary·share·success와 겹치지 않게
    themes: {
      light: { colors: {
        'terminal': '#10161e', 'on-terminal': '#c8d3df', 'share': '#1f8f86', 'on-share': '#ffffff',
        'acct1': '#d84315', 'acct2': '#7b1fa2', 'acct3': '#283593', 'acct4': '#ad1457', 'acct5': '#455a64', 'acct6': '#5d4037',
      } },
      dark: { colors: {
        'terminal': '#10161e', 'on-terminal': '#c8d3df', 'share': '#4fd1c5', 'on-share': '#0b2a28',
        'acct1': '#ff8a65', 'acct2': '#ce93d8', 'acct3': '#9fa8da', 'acct4': '#f48fb1', 'acct5': '#b0bec5', 'acct6': '#bcaaa4',
      } },
    },
  },
})
