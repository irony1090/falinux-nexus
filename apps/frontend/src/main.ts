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

const app = createApp(App)

registerPlugins(app)

app.mount('#app')
