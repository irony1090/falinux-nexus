import router from '../router';
import vueQuery, { vueQueryPluginOption } from './vueQuery'
/**
 * plugins/index.ts
 *
 * Automatically included in `./src/main.ts`
 */

// Types
import type { App } from 'vue'

// Plugins
import vuetify from './vuetify'

export function registerPlugins (app: App) {
    app.use(vuetify)
    .use(router)
    .use(vueQuery, vueQueryPluginOption)
}