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
  },
})
