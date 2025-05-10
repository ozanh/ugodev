import * as Comlink from 'comlink'

console.log('init wasm worker')

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

const ugo = {
  isLoaded() {
    return Boolean(self.runUGO)
  },
  runUGO(obj, script) {
    try {
      const ret = self.runUGO(obj, script)
      if (ret && typeof ret === 'object') {
        if (ret.error) {
          throw new Error(`Internal error: ${ret.error}`)
        } else {
          throw new Error(`Unexpected result from runUGO wasm: ${ret}`)
        }
      }
    } catch (err) {
      if (obj.resultCallback) {
        obj.resultCallback({ error: err.toString() })
      } else {
        console.error(`runUGO error: ${err}`)
      }
    }
  },
  checkUGO(obj, script) {
    try {
      const ret = self.checkUGO(obj, script)
      if (ret && typeof ret === 'object') {
        if (ret.error) {
          throw new Error(`Internal error: ${ret.error}`)
        } else {
          throw new Error(`Unexpected result from checkUGO wasm: ${ret}`)
        }
      }
    } catch (err) {
      if (obj.checkCallback) {
        obj.checkCallback({ warning: err.toString() })
      } else {
        console.error(`checkUGO error: ${err}`)
      }
    }
  },
  cancelUGO() {
    try {
      return self.cancelUGO()
    } catch (err) {
      console.error(`cancelUGO error: ${err}`)
      return false
    }
  }
}

Comlink.expose(ugo)
