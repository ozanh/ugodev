const { defineConfig } = require('@vue/cli-service')
const CopyPlugin = require('copy-webpack-plugin')
const LicenseCheckerWebpackPlugin = require('license-checker-webpack-plugin')
const NodePolyfillPlugin = require("node-polyfill-webpack-plugin");
const { resolve } = require('path')

const env = process.env
env.VUE_APP_PLAYGROUND_VERSION = require('./package.json').version
env.VUE_APP_LICENSE_PATH = 'static/LICENSE.txt'
env.VUE_APP_THIRD_PARTY_PATH = 'static/ThirdPartyNotices.txt'

const outputDir = 'dist'

if (env.NODE_ENV === 'production' &&
  env.VUE_APP_GO_LICENSE &&
  env.VUE_APP_TENGO_LICENSE &&
  env.VUE_APP_THIRD_PARTY_PATH) {
  // append license file contents of Go and other Go libraries to ThirdPartyNotices
  process.on('exit', (code) => {
    if (code !== 0) {
      return
    }
    const fs = require('fs')
    for (const file of [
      env.VUE_APP_GO_LICENSE,
      env.VUE_APP_TENGO_LICENSE
    ]) {
      const s = fs.readFileSync(file, { encoding: 'utf8' })
      const target = resolve(outputDir, env.VUE_APP_THIRD_PARTY_PATH)
      try {
        fs.appendFileSync(target, `\n${'-'.repeat(80)}\n\n`, 'utf8')
        fs.appendFileSync(target, s, 'utf8')
      } catch (err) {
        console.warn(err)
      }
    }
  })
}

let copyFiles = []
if (env.VUE_APP_LICENSE && env.VUE_APP_LICENSE_PATH &&
  env.VUE_APP_WASM_EXEC_FILE) {
  copyFiles = [
    {
      from: resolve(__dirname, env.VUE_APP_LICENSE),
      to: env.VUE_APP_LICENSE_PATH
    },
    {
      from: resolve(__dirname, env.VUE_APP_WASM_EXEC_FILE),
      to: 'static/js'
    }
  ]
}

const plugins = [
  new LicenseCheckerWebpackPlugin({
    allow: '(Apache-2.0 OR BSD-2-Clause OR BSD-3-Clause OR MIT OR ISC)',
    outputFilename: env.VUE_APP_THIRD_PARTY_PATH,
    emitError: true,
    filter: /(^.*[/\\]node_modules[/\\]((?:@[^/\\]+[/\\])?(?:[^@/\\][^/\\]*)))/
  }),
  new NodePolyfillPlugin()
]

if (copyFiles.length > 0) {
  plugins.push(new CopyPlugin({ patterns: copyFiles }))
}

module.exports = defineConfig({
  transpileDependencies: true,
  parallel: false,
  productionSourceMap: false,
  outputDir: outputDir,
  assetsDir: 'static',
  css: {
    extract: false
  },
  configureWebpack: {
    optimization: {
      splitChunks: false
    },
    plugins: plugins
  },
  chainWebpack: config => {
    config.module
      .rule('ugo file')
      .test(/\.ugo$/)
      .use('raw-loader')
      .loader('raw-loader')
      .end()
    config.module
      .rule('worker')
      .test(/\.worker\.js$/)
      .use('worker-loader')
      .loader('worker-loader')
      .tap(options => {
        options = options || {}
        options.inline = 'fallback'
        options.filename = 'static/js/[name].js'
        if (env.NODE_ENV !== 'production') {
          // FIXME: Setting publicPath is required for worker-loader to work
          // properly until a solution is found. Following Open PR may fix this
          // bug: https://github.com/webpack-contrib/worker-loader/pull/291
          // Browser throws invalid URL error.
          options.publicPath = 'http://localhost:8080/'
        } else {
          options.publicPath = 'https://play.verigraf.com/'
        }
        return options
      })
      .end()
    config.module
      .rule('wasm')
      .test(/\.wasm$/)
      .type('javascript/auto')
      .use('file-loader')
      .loader('file-loader')
      .tap(options => {
        options = options || {}
        options.name = '[name].[hash:8].[ext]'
        options.outputPath = 'static'
        return options
      })
      .end()
  }
})
