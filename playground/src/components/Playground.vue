<template>
  <div>
    <div class="playground">
      <div class="head">
        <div class="playground-text">
          <strong>{{ msg }}</strong>
        </div>
        <div class="head-buttons">
          <button
            id="run-button"
            :disabled="loading"
            class="button"
            @click="onRun"
          >
            Run
            <span class="key-press hidden-sm">Ctrl+↵</span>
          </button>
          <button
            id="about-button"
            class="button"
            @click="showAboutModal = true"
          >
            About
          </button>
          <div class="loader-wrapper">
            <div
              v-show="loading"
              class="loader"
            />
          </div>
        </div>
        <div class="head-gh hidden-sm">
          <a
            href="https://github.com/ozanh/ugo"
            data-size="large"
            class="github-button"
            aria-label="Fork ozanh/ugo on GitHub"
          >
            Fork
          </a>
        </div>
        <div class="copyright">
          Copyright © 2020 Ozan Hacıbekiroğlu
        </div>
      </div>
      <div class="body-container">
        <prism-editor
          v-model="code"
          :highlight="highlighter"
          line-numbers
          class="playground-editor"
          @input="edited = true"
          @click="onEditorClick"
        />
        <div class="result">
          <div
            v-if="result && result.error != ''"
            class="result-error"
          >
            <pre>{{ result.error }}</pre>
          </div>
          <div
            v-if="result && result.stdout != ''"
            class="result-stdout"
          >
            <pre>{{ result.stdout }}</pre>
          </div>
          <div
            v-if="result && result.value!=''"
            class="result-value"
          >
            <strong>Return Value as JSON:</strong>
            <pre v-text="valueToJSON(result.value)" />
          </div>
        </div>
      </div>
      <div class="footer">
        <div
          v-if="result && result.metrics"
          class="metrics"
        >
          <span>
            Compile:{{ result.metrics.compile }}
          </span>
          <span>
            Exec:{{ result.metrics.exec }}
          </span>
          <span>
            Total:{{ result.metrics.elapsed }}
          </span>
        </div>
      </div>
    </div>
    <teleport to="body">
      <modal
        :show-modal="showAboutModal"
        @update:show-modal="showAboutModal = $event"
      >
        <template #header>
          <h1>About</h1>
        </template>

        <template #body>
          <p>
            uGO Playground is a single page application to test uGO scripts in the browser.<br>
            Thanks to Go's <b>experimental</b> WebAssembly support, uGO and stdlib modules are compiled for WebAssembly.<br>
            Note that native performance of uGO is much faster than WebAssembly port.<br><br>
            <a
              :href="'/'+license"
              target="_blank"
            >LICENSE
            </a><br>
            <a
              :href="'/'+thirdParty"
              target="_blank"
              rel="nofollow"
            >Third Party Notices
            </a>
            <br><br>
            Playground Version: {{ playgroundVersion }}<br>
            uGO Version: {{ uGOVersion }}<br>
            Build Time: {{ buildTime }}<br><br>
            <a
              href="https://github.com/ozanh/ugo"
              target="_blank"
            >uGO Script Language</a><br>
            Copyright © 2020 Ozan Hacıbekiroğlu
          </p>
        </template>

        <template #footer>
          <button
            class="button"
            @click="showAboutModal = false"
          >
            Close
          </button>
        </template>
      </modal>
      <modal
        :show-modal="showWASMErrorModal"
        @update:show-modal="showWASMErrorModal = $event"
      >
        <template #header>
          <h1>WebAssembly Error</h1>
        </template>
        <template #body>
          <br>
        </template>
        <template #footer>
          <button
            class="button"
            @click="showWASMErrorModal = false"
          >
            Close
          </button>
        </template>
      </modal>
    </teleport>
  </div>
</template>

<script>
// import required css files
import '../assets/css/button.css'
import '../assets/css/loader.css'

// import modal component
import Modal from './Modal'

// import Prism Editor and vue component
import { PrismEditor } from 'vue-prism-editor'
import 'vue-prism-editor/dist/prismeditor.min.css'

// import highlighting library (you can use any library you want just return html string)
import { highlight, languages } from 'prismjs/components/prism-core'
import 'prismjs/components/prism-clike'
import 'prismjs/components/prism-ugo'
import 'prismjs/themes/prism-okaidia.css' // import syntax highlighting styles

import { debounce } from '../lib/utils'
// FIXME: During testing, due to jest or babel syntax error is thrown
// for mini-toastr. It is imported with require in methods.
// import miniToastr from 'mini-toastr'
import code from '../../cmd/wasm/testdata/sample.ugo'

export default {
  name: 'Playground',
  components: {
    PrismEditor,
    Modal
  },
  props: {
    msg: {
      type: String,
      default: ''
    },
    checkWASM: {
      type: Boolean,
      default: true
    }
  },
  data: () => ({
    playgroundVersion: process.env.VUE_APP_PLAYGROUND_VERSION,
    uGOVersion: process.env.VUE_APP_UGO_VERSION,
    buildTime: process.env.VUE_APP_BUILD_TIME,
    license: process.env.VUE_APP_LICENSE_PATH,
    thirdParty: process.env.VUE_APP_THIRD_PARTY_PATH,
    showAboutModal: false,
    showWASMErrorModal: false,
    code: code,
    linesMsgs: {},
    loading: true,
    result: null,
    edited: false
  }),
  watch: {
    code () {
      !this.loading && this.checkCode()
    }
  },
  mounted () {
    if (global.window) {
      const unwatch = this.$watch('edited', (newVal) => {
        if (newVal) {
          window.addEventListener('beforeunload', function (e) {
            e.preventDefault()
            e.returnValue = 'Stay on Page?'
          })
          unwatch()
        }
      })
      const ln = document.querySelector('.prism-editor__line-numbers')
      if (ln) {
        ln.addEventListener('click', (e) => {
          if (!e.target.classList.contains('line-number-red')) return
          const msgs = this.linesMsgs[e.target.innerText] || []
          if (!Array.isArray(msgs)) return
          require('mini-toastr').default.error(
            `<pre style="white-space: pre-wrap">${msgs.join('\n\n').trim()}</pre>`,
            'Warning',
            5000,
            undefined,
            { allowHtml: true }
          )
        })
      }
    }
    if (global.document) {
      const pg = document.querySelector('.playground-editor')
      if (pg) {
        pg.addEventListener('keyup', (e) => {
          if (e.ctrlKey && e.keyCode === 13) this.onRun()
        })
      }
    }
    let counter = 0
    const f = () => {
      if (global.runUGO == null) {
        if (counter > 9) {
          this.loading = false
          this.showWASMErrorModal = true
          return
        }
      } else {
        this.loading = false

        this.checkCode = debounce(() => {
          if (global.checkUGO == null) return
          global.checkUGO(this, this.code.toString())
        }, 1000)

        return
      }
      counter++
      setTimeout(f, 250)
    }
    if (this.checkWASM) {
      setTimeout(f, 250)
    }
  },
  methods: {
    highlighter (code) {
      return highlight(code, languages.ugo)
    },
    onRun () {
      this.result = null
      this.loading = true
      try {
        global.runUGO(this, this.code)
      } catch (err) {
        console.log(err)
        this.result = { error: err.toString() }
      } finally {
        this.loading = false
      }
    },
    resultCallback (msg) {
      this.loading = false
      this.result = msg
    },
    valueToJSON (value) {
      try {
        return JSON.stringify(JSON.parse(value), null, 2)
      } catch (err) {
        return `JSON Error: ${err.toString()}`
      }
    },
    onEditorClick () {
      const elem = document.querySelector('.prism-editor__textarea')
      if (elem) {
        elem.focus()
        elem.click()
      }
    },
    checkCallback (result) {
      if (typeof result !== 'undefined') {
        if (result.warning) {
          console.log('check warning:', result.warning)
        }
        this.highlightLine(result.lines || {})
        return
      }
      this.highlightLine({})
    },
    highlightLine (linesMsgs) {
      this.linesMsgs = linesMsgs
      const lines = document.querySelectorAll('.prism-editor__line-number')
      if (!lines) return
      lines.forEach((el) => {
        if (el.innerText in linesMsgs) el.classList.add('line-number-red')
        else el.classList.remove('line-number-red')
      })
    }
  }
}
</script>
<style lang="scss">
.line-number-red {
  background-color: red;
  opacity: 0.8;
  color: white !important;
  cursor: pointer;
}

.modal__dialog {
  font-size: small;
}

.playground {
  display: flex;
  flex-flow: column;
  height: 100%;
}

.head {
  width: 100%;
  flex: 0 0 auto;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.playground-text {
  padding-right: 5px;
  overflow: visible;
  font-size: 16pt;
  flex: 0 1 auto;
}

.head-buttons {
  height: 28px;
  flex: 0 0 auto;
  display: flex;
  flex-wrap: nowrap;
  align-items: stretch;
}

.head-gh {
  padding-right: 5px;
}

.key-press {
  opacity: 0.4;
  font-size: 90%;
}

.head-buttons .button {
  margin-right: 5px;
}

.loader-wrapper {
  width: 20px;
  flex: 0 0 20px;
  margin-top: auto;
  margin-bottom: auto;
}

.copyright {
  flex: 0 0 auto;
  color: #ccc;
  font-size: 10pt;
  text-align: right;
}

.body-container {
  flex: 0 0 auto;
  width: 100%;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

// required class for editor
.playground-editor {
  background: #2d2d2d;
  color: #ccc;
  font-family: Fira code, Fira Mono, Consolas, Menlo, Courier, monospace;
  font-size: 11pt;
  line-height: 1.2;
  padding: 0px;
  max-width: 49%;
  margin-right: 1%;
  flex: 1 1 auto;
  height: 85vh;
}

// optional
.prism-editor__textarea:focus {
  outline: none;
}

/***************************************************/
// FIXME
// this is a work around for issue:
// https://github.com/koca/vue-prism-editor/issues/87
//
// Editor shows line numbers incorrectly due to wrapped content.
// If it is forced to disable wrapping behavior then horizontall scroll
// does not show up because of cascaded overflow:hidden style. After setting
// overflow-x to scroll then editing becomes buggy because text does not
// appear at the cursor. Finally setting width to a big number fixes
// this issue temporarily.
.prism-editor__textarea {
  width: 999999px !important;
}

.prism-editor__editor {
  white-space: pre !important;
}

.prism-editor__container {
  overflow-x: scroll !important;
}
/***************************************************/

.result {
  font-size: 11pt;
  background: #2d2d2d;
  max-width: 48%;
  margin-left: 1%;
  padding-left: 1%;
  height: 85vh;
  flex: 1 1 auto;
  overflow: auto;
}

.result-error {
  color: #ff6565;
  font-weight: 200;
  width: 100%;
}

.head,
.result,
.footer {
  color: white;
}

.footer {
  bottom: 0;
  width: 100%;
  text-align: right;
}

.metrics {
  font-size: 10pt;
}

@media (max-width: 600px) {
  .hidden-sm {
    display: none;
  }
  .playground-text {
    font-size: 14pt;
  }
}

@media (min-width: 900px) {
  .copyright {
    flex: 1 0 auto;
  }
}
</style>
