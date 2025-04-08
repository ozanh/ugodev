import * as Comlink from 'comlink'

console.log('importing wasm worker')

if (!WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer()
    return await WebAssembly.instantiate(source, importObject)
  }
}

await import(`../wasm_exec.js`)

const go = new self.Go()

const mod = await import('../ugo.wasm?url')

WebAssembly.instantiateStreaming(fetch(mod.default), go.importObject).then(async result => {
  await go.run(result.instance)
})

export const ugo = {
  isLoaded() {
    return typeof self.runUGO !== 'undefined'
  },
  runUGO(obj, script) {
    try {
      self.runUGO(obj, script)
    } catch (err) {
      if (typeof obj.resultCallback !== 'undefined') {
        obj.resultCallback({ error: err.toString() })
      }
    }
  },
  checkUGO(obj, script) {
    try {
      self.checkUGO(obj, script)
    } catch (err) {
      if (typeof obj.checkCallback !== 'undefined') {
        obj.checkCallback({ warning: err.toString() })
      }
    }
  }
}

Comlink.expose(ugo)
