/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Composables
import { createApp } from 'vue'

// Plugins
import { registerPlugins } from '@/plugins'

// Components
import App from './App.vue'

// Styles
// import 'unfonts.css'
import "@/styles/noto-sans-korean.scss"
import "@/styles/global.scss"
import "@/styles/tile-buttons.scss"
import "@/styles/dir-grid.scss"

// iOS Safari 핀치 확대 끔(줌-3) — touch-action이 안 먹는 옛 iOS 대비
document.addEventListener('gesturestart', e => e.preventDefault())

const app = createApp(App)

registerPlugins(app)

app.mount('#app')
