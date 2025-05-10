import { fileURLToPath, URL } from 'node:url'
import * as fs from 'node:fs'
import process from 'node:process'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import topLevelAwait from 'vite-plugin-top-level-await'
import { viteStaticCopy } from 'vite-plugin-static-copy'
import legacy from '@vitejs/plugin-legacy'

process.env.VITE_PLAYGROUND_VERSION = JSON.parse(
  fs.readFileSync(new URL('./package.json', import.meta.url))
).version

const createThirdPartyNotices = mode => {
  const target = fileURLToPath(new URL('./ThirdPartyNotices.txt', import.meta.url))
  // truncate target file if it exists
  if (fs.existsSync(target)) {
    fs.truncateSync(target, 0)
  }
  const env = loadEnv(mode, process.cwd())
  console.log('env', env)
  for (const file of [env.VITE_GO_LICENSE, env.VITE_TENGO_LICENSE]) {
    const data = fs.readFileSync(file, { encoding: 'utf8' })

    fs.appendFileSync(target, data, 'utf8')
    fs.appendFileSync(target, `\n${'-'.repeat(80)}\n\n`, 'utf8')
  }
}

const checkEnvFile = mode => {
  const envFile = fileURLToPath(new URL(`./.env.${mode}.local`, import.meta.url))
  if (!fs.existsSync(envFile)) {
    const err = new Error(
      `Environment file ${envFile} does not exist. Don't forget to run "make production" or make development" first.`
    )
    console.error(err)
    throw err
  }
}

export default defineConfig(({ mode }) => {
  checkEnvFile(mode)
  createThirdPartyNotices(mode)

  return {
    esbuild: {
      target: 'es2015'
    },
    build: {
      target: 'es2015'
    },
    plugins: [
      vue(),
      topLevelAwait(),
      viteStaticCopy({
        targets: [
          {
            src: '../LICENSE',
            dest: 'assets',
            rename: 'LICENSE.txt'
          },
          {
            src: './ThirdPartyNotices.txt',
            dest: 'assets'
          }
        ]
      }),

      // https://github.com/vitejs/vite/discussions/10519
      legacy({
        targets: ['chrome >= 64', 'edge >= 79', 'safari >= 11.1', 'firefox >= 67'],
        renderLegacyChunks: false,
        modernPolyfills: true
      })
    ],
    resolve: {
      extensions: ['.js', '.vue', '.json', '.ts', '.mjs', '.jsx', '.tsx'],
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    define: {
      global: 'globalThis'
    },
    worker: {
      format: 'es'
    }
  }
})
