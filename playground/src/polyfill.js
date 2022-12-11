// polyfill for wasm_exec.js
if (!globalThis.require) {
  globalThis.require = require
}

// globalThis.fs = require('fs')

if (!globalThis.TextEncoder) {
  globalThis.TextEncoder = require('util').TextEncoder
}
if (!globalThis.TextDecoder) {
  globalThis.TextDecoder = require('util').TextDecoder
}

if (!globalThis.performance || !globalThis.performance.now) {
  globalThis.performance = {
    now () {
      const [sec, nsec] = process.hrtime()
      return sec * 1000 + nsec / 1000000
    }
  }
}

if (!globalThis.crypto || !globalThis.crypto.getRandomValues) {
  const crypto = require('crypto')
  globalThis.crypto = {
    getRandomValues (b) {
      crypto.randomFillSync(b)
    }
  }
}
