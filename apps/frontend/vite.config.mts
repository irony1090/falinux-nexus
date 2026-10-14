import { fileURLToPath, URL } from 'node:url'
import Vue from '@vitejs/plugin-vue'
// import Fonts from 'unplugin-fonts/vite'
import { defineConfig, loadEnv, type Plugin, UserConfig } from 'vite'
import Vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
// 폰 사진 -> user vault 올리기(/__upload?t=DEV_UPLOAD_TOKEN). 쓸 때 이 줄과 plugins의 devUpload 줄 주석을 푼다 — REF-test-env.md "폰 사진 올리기"
// import { devUpload } from './dev-upload.plugin.mts'


const css: UserConfig['css'] = {
	preprocessorOptions: {
		scss: {
			additionalData: `@use '@/styles/_variables.scss' as vars;`
		}
	}
}

// 재발 방지: TLS 프록시(nginx)가 30초 무통신 연결을 끊으면 vite 클라이언트가 "서버 재시작"으로 보고 페이지를 새로고침한다.
// 프록시 뒤(HMR_CLIENT_PORT 있음)에서만 HMR 소켓에 주기적으로 빈 사용자 이벤트를 보내 연결을 유지 — REF-external-access.md
const hmrKeepAlive = (ms: number): Plugin => ({
  name: 'hmr-keepalive',
  apply: 'serve',
  configureServer(server) {
    const timer = setInterval(() => server.ws.send('nexus:keepalive'), ms)
    server.httpServer?.on('close', () => clearInterval(timer))
  },
})

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
// 개발 서버 설정은 .env의 DEV_PORT·DEV_HOST·HMR_CLIENT_PORT(접두사 없음 = 클라이언트로 안 나감). 없으면 3000·localhost
// HMR_CLIENT_PORT = TLS 프록시 뒤(https://test1.falinux.com -> 8011)에서 브라우저가 붙을 포트(443)
const env = loadEnv(mode, fileURLToPath(new URL('.', import.meta.url)), '')
return {
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
    ...(env.HMR_CLIENT_PORT ? [hmrKeepAlive(20_000)] : []),
    // devUpload(fileURLToPath(new URL('../../obsidian-vault/user', import.meta.url)), env.DEV_UPLOAD_TOKEN),
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
    port: Number(env.DEV_PORT) || 3000,
    host: env.DEV_HOST || undefined,
    strictPort: !!env.DEV_PORT,
    allowedHosts: true,
    hmr: env.HMR_CLIENT_PORT ? { clientPort: Number(env.HMR_CLIENT_PORT) } : undefined,
  },
} satisfies UserConfig
})
