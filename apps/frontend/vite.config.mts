import { fileURLToPath, URL } from 'node:url'
import Vue from '@vitejs/plugin-vue'
// import Fonts from 'unplugin-fonts/vite'
import { defineConfig, UserConfig } from 'vite'
import Vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'


const css: UserConfig['css'] = {
	preprocessorOptions: {
		scss: {
			additionalData: `@use '@/styles/_variables.scss' as vars;`
		}
	}
}

// https://vitejs.dev/config/
export default defineConfig({
  preview: { allowedHosts: true },
  plugins: [
    Vue({
      template: { transformAssetUrls },
    }),
    // https://github.com/vuetifyjs/vuetify-loader/tree/master/packages/vite-plugin#readme
    Vuetify({
      autoImport: true,
      styles: {
        configFile: 'src/styles/settings.scss',
      },
    }),
    // Fonts({
    //   fontsource: {
    //     families: [
    //       {
    //         name: 'Roboto',
    //         weights: [100, 300, 400, 500, 700, 900],
    //         styles: ['normal', 'italic'],
    //       },
    //     ],
    //   },
    // }),
  ],
  define: { 'process.env': {} },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('src', import.meta.url)),
    },
    extensions: [
      '.js',
      '.json',
      '.jsx',
      '.mjs',
      '.ts',
      '.tsx',
      '.vue',
    ],
  },
  css,
  server: {
    port: 3000,
    allowedHosts: true,
  },
})
