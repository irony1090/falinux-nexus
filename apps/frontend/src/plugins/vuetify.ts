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
    themes: {
      light: { colors: { 'terminal': '#10161e', 'on-terminal': '#c8d3df', 'share': '#1f8f86', 'on-share': '#ffffff' } },
      dark: { colors: { 'terminal': '#10161e', 'on-terminal': '#c8d3df', 'share': '#4fd1c5', 'on-share': '#0b2a28' } },
    },
  },
})
