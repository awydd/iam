import vue from '@vitejs/plugin-vue'
import path from 'node:path'
import { fileURLToPath, URL } from 'node:url'
import AutoImport from 'unplugin-auto-import/vite'
import IconsResolver from 'unplugin-icons/resolver'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import Components from 'unplugin-vue-components/vite'
import { defineConfig, loadEnv } from 'vite'
import vueDevTools from 'vite-plugin-vue-devtools'
import Icons from 'unplugin-icons/vite'

const srcPath = fileURLToPath(new URL('./src', import.meta.url))

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const allowedHosts = env.VITE_ALLOWED_HOSTS
    ? env.VITE_ALLOWED_HOSTS.split(',').map(host => host.trim())
    : undefined

  return {
    plugins: [
      vue(),
      vueDevTools(),
      AutoImport({
        imports: ['vue', 'vue-router'],
        dts: path.resolve(srcPath, 'auto-imports.d.ts'),
        eslintrc: {
          enabled: true,
          filepath: path.resolve(srcPath, '.eslintrc-auto-import.json'),
          globalsPropValue: true,
        },
        resolvers: [ElementPlusResolver(), IconsResolver({ prefix: 'Icon' })],
      }),
      Components({
        resolvers: [
          ElementPlusResolver(),
          IconsResolver({
            prefix: '',
            enabledCollections: ['ep'],
          }),
        ],
        dts: path.resolve(srcPath, 'components.d.ts'),
      }),
      Icons({
        autoInstall: true,
      }),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_PROXY_TARGET,
          changeOrigin: true,
        },
      },
      ...(allowedHosts ? { allowedHosts } : {}),
    },
  }
})
