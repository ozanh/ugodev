const CopyPlugin = require('copy-webpack-plugin')
const LicenseCheckerWebpackPlugin = require('license-checker-webpack-plugin')
const { resolve } = require('path')

process.env.VUE_APP_PLAYGROUND_VERSION = require('./package.json').version
process.env.VUE_APP_LICENSE_PATH = 'static/LICENSE.txt'
process.env.VUE_APP_THIRD_PARTY_PATH = 'static/ThirdPartyNotices.txt'

const outputDir = 'dist'

if (process.env.VUE_APP_GO_LICENSE &&
  process.env.VUE_APP_TENGO_LICENSE &&
  process.env.VUE_APP_THIRD_PARTY_PATH) {
  // append license file contents of Go and other Go libraries to ThirdPartyNotices
  process.on('exit', (code) => {
    if (code !== 0) {
      return
    }
    const fs = require('fs')
    for (const file of [
      process.env.VUE_APP_GO_LICENSE,
      process.env.VUE_APP_TENGO_LICENSE
    ]) {
      const s = fs.readFileSync(file, { encoding: 'utf8' })
      const target = resolve(outputDir, process.env.VUE_APP_THIRD_PARTY_PATH)
      fs.appendFileSync(target, `\n${'-'.repeat(80)}\n\n`, 'utf8')
      fs.appendFileSync(target, s, 'utf8')
    }
  })
}

let copyFiles = []
if (process.env.VUE_APP_WASM_EXEC_FILE &&
  process.env.VUE_APP_WASM_EXEC_FILE &&
  process.env.VUE_APP_LICENSE &&
  process.env.VUE_APP_LICENSE_PATH) {
  copyFiles = [
    {
      from: resolve(__dirname, process.env.VUE_APP_WASM_EXEC_FILE),
      to: 'static/'
    },
    {
      from: resolve(__dirname, process.env.VUE_APP_WASM_FILE),
      to: 'static/'
    },
    {
      from: resolve(__dirname, process.env.VUE_APP_LICENSE),
      to: process.env.VUE_APP_LICENSE_PATH
    }
  ]
}

module.exports = {
  productionSourceMap: false,
  outputDir: outputDir,
  assetsDir: 'static',
  configureWebpack: {
    plugins: [
      new CopyPlugin(copyFiles),
      new LicenseCheckerWebpackPlugin({
        allow: '(Apache-2.0 OR BSD-2-Clause OR BSD-3-Clause OR MIT)',
        outputFilename: process.env.VUE_APP_THIRD_PARTY_PATH,
        emitError: true
      })
    ]
  },
  chainWebpack: config => {
    config.module
      .rule('raw')
      .test(/\.ugo$/)
      .use('raw-loader')
      .loader('raw-loader')
      .end()
  }
}
