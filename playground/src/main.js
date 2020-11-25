import { createApp } from 'vue'
import App from './App.vue'
import miniToastr from 'mini-toastr'

if (!WebAssembly.instantiateStreaming) { // polyfill
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer()
    return await WebAssembly.instantiate(source, importObject)
  }
}

import('../ugo.wasm').then(module => {
  console.log(module)
  global.go = new global.Go()
  WebAssembly.instantiateStreaming(
    fetch(module.default),
    global.go.importObject
  ).then(async (result) => {
    await global.go.run(result.instance)
  })
})

miniToastr.init()

createApp(App).mount('#app')
