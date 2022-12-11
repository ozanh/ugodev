/* eslint-disable import/first */
/* eslint-disable camelcase */
import { expose } from 'comlink'

// pollyfill for wasm_exec.js
globalThis.require = require
// globalThis.fs = require('fs')
globalThis.TextEncoder = require('util').TextEncoder
globalThis.TextDecoder = require('util').TextDecoder

globalThis.performance = {
  now () {
    const [sec, nsec] = process.hrtime()
    return sec * 1000 + nsec / 1000000
  }
}

const crypto = require('crypto')
globalThis.crypto = {
  getRandomValues (b) {
    crypto.randomFillSync(b)
  }
}
// end of pollyfill

console.log('worker importing wasm_exec')
// eslint-disable-next-line no-undef
self.importScripts(`${__webpack_public_path__}static/js/${process.env.VUE_APP_WASM_EXEC_FILE}`)

console.log('worker importing wasm')
if (!WebAssembly.instantiateStreaming) {
  // polyfill
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer()
    return await WebAssembly.instantiate(source, importObject)
  }
}
const go = new self.Go()
import('../ugo.wasm').then(mod => {
  WebAssembly.instantiateStreaming(
    fetch(mod.default),
    go.importObject
  ).then(async (result) => {
    await go.run(result.instance)
  })
})

const ugo = {
  isLoaded () {
    return typeof self.runUGO !== 'undefined'
  },
  runUGO (obj, script) {
    try {
      self.runUGO(obj, script)
    } catch (err) {
      if (typeof obj.resultCallback !== 'undefined') {
        obj.resultCallback({ error: err.toString() })
      }
    }
  },
  checkUGO (obj, script) {
    try {
      self.checkUGO(obj, script)
    } catch (err) {
      if (typeof obj.checkCallback !== 'undefined') {
        obj.checkCallback({ warning: err.toString() })
      }
    }
  }
}

expose(ugo)
console.log('worker exposed object')
