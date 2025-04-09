<template>
  <div>
    <div class="playground">
      <div class="head">
        <div class="playground-text">
          <strong>{{ playgroundText }}</strong>
        </div>
        <div class="head-buttons">
          <button id="run-button" :disabled="loading" class="button" @click="onRun">
            Run
            <span class="key-press hidden-sm">Ctrl+↵</span>
          </button>

          <button id="about-button" class="button" @click="showAboutModal = true">About</button>

          <div style="min-width: 120px">
            <button
              title="Cancel running script"
              id="cancel-button"
              v-show="delayedLoading"
              :disabled="cancelInProcess"
              class="button"
              @click="onCancel"
            >
              <span v-if="cancelInProcess">Canceling</span>
              <span v-else>x Cancel</span>
            </button>
          </div>
          <div class="loader-wrapper" style="padding-right: 4px">
            <div v-show="delayedLoading" class="loader" />
          </div>
        </div>
        <div class="copyright">Copyright © 2020-2025 Ozan Hacıbekiroğlu</div>
        <div class="hidden-sm">
          <a
            href="https://github.com/ozanh/ugo"
            title="Fork ozanh/ugo on GitHub"
            aria-label="Fork ozanh/ugo on GitHub"
            target="_blank"
          >
            <button class="button">
              <svg
                fill="none"
                width="16"
                height="16"
                viewBox="0 -4 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  fill="currentColor"
                  d="M8 18a1 1 0 1 0 0-2 1 1 0 0 0 0 2zm1.033-3.817A3.001 3.001 0 1 1 7 14.17v-1.047c0-.074.003-.148.008-.221a1 1 0 0 0-.462-.637L3.46 10.42A3 3 0 0 1 2 7.845V5.829a3.001 3.001 0 1 1 2 0v2.016a1 1 0 0 0 .487.858l3.086 1.846a3 3 0 0 1 .443.324 3 3 0 0 1 .444-.324l3.086-1.846a1 1 0 0 0 .487-.858V5.841A3.001 3.001 0 0 1 13 0a3 3 0 0 1 1.033 5.817v2.028a3 3 0 0 1-1.46 2.575l-3.086 1.846a1 1 0 0 0-.462.637c.005.073.008.147.008.22v1.06zM3 4a1 1 0 1 0 0-2 1 1 0 0 0 0 2zm10 0a1 1 0 1 0 0-2 1 1 0 0 0 0 2z"
                />
              </svg>
              Fork
            </button>
          </a>
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
          <div v-if="result && result.error != ''" class="result-error">
            <pre>{{ result.error }}</pre>
          </div>
          <div v-if="result && result.stdout != ''" class="result-stdout">
            <pre>{{ result.stdout }}</pre>
          </div>
          <div v-if="result && result.value != ''" class="result-value">
            <strong>Return Value as JSON:</strong>
            <pre v-text="valueToJSON(result.value)" />
          </div>
        </div>
      </div>
      <div class="footer">
        <div v-if="result && result.metrics" class="metrics">
          <span> Compile:{{ result.metrics.compile }} </span>
          <span> Exec:{{ result.metrics.exec }} </span>
          <span> Total:{{ result.metrics.elapsed }} </span>
        </div>
      </div>
    </div>
    <teleport to="body">
      <app-modal :show-modal="showAboutModal" @update:show-modal="showAboutModal = $event">
        <template #header>
          <h1>About</h1>
        </template>

        <template #body>
          <p>
            uGO Playground is a single page application to test uGO scripts in the browser.<br />
            Thanks to Go's WebAssembly support, uGO and stdlib modules are compiled for
            WebAssembly.<br />
            Note that native performance of uGO is much faster than WebAssembly port.<br /><br />
            <a :href="license" target="_blank">LICENSE </a><br />
            <a :href="thirdParty" target="_blank">Third Party Notices </a>
            <br /><br />
            Playground Version: {{ playgroundVersion }}<br />
            Go Version: {{ goVersion }}<br />
            uGO Version: {{ uGOVersion }}<br />
            Build Time: {{ buildTime }}<br /><br />
            <a href="https://github.com/ozanh/ugo" target="_blank">uGO Script Language</a><br />
            <a href="https://github.com/ozanh/ugodev/tree/main/playground" target="_blank"
              >uGO Playground</a
            ><br />
            Copyright © 2020-2025 Ozan Hacıbekiroğlu
          </p>
        </template>

        <template #footer>
          <button class="button" @click="showAboutModal = false">Close</button>
        </template>
      </app-modal>
      <app-modal :show-modal="showWASMErrorModal" @update:show-modal="showWASMErrorModal = $event">
        <template #header>
          <h1>WebAssembly Error</h1>
        </template>
        <template #body>
          <br />
        </template>
        <template #footer>
          <button class="button" @click="showWASMErrorModal = false">Close</button>
        </template>
      </app-modal>
    </teleport>
  </div>
</template>

<script>
import { ref } from 'vue'
import { PrismEditor } from 'vue-prism-editor'

import { highlight, languages } from 'prismjs/components/prism-core'
import 'prismjs/components/prism-clike'
import 'prismjs/components/prism-ugo'

import miniToastr from 'mini-toastr'
import * as Comlink from 'comlink'

import ugoSampleCode from '@/../cmd/wasm/testdata/sample.ugo?raw'
import { debounce } from '@/lib/utils'
import Worker from '@/wasm.worker?worker'
import AppModal from './AppModal'

export default {
  name: 'UserPlayground',
  components: {
    PrismEditor,
    AppModal
  },
  props: {
    checkWASM: {
      type: Boolean,
      default: true
    }
  },
  setup() {
    const playgroundText = 'uGO Playground'
    const playgroundVersion = import.meta.env.VITE_PLAYGROUND_VERSION
    const goVersion = import.meta.env.VITE_GO_VERSION
    const uGOVersion = import.meta.env.VITE_UGO_VERSION
    const buildTime = import.meta.env.VITE_BUILD_TIME
    const license = '/assets/LICENSE.txt'
    const thirdParty = '/assets/ThirdPartyNotices.txt'

    const showAboutModal = ref(false)
    const showWASMErrorModal = ref(false)
    const code = ref(ugoSampleCode)
    const linesMsgs = ref({})
    const loading = ref(false)
    const delayedLoading = ref(false)
    const result = ref(null)
    const edited = ref(false)
    const cancelInProcess = ref(false)

    return {
      playgroundText,
      playgroundVersion,
      goVersion,
      uGOVersion,
      buildTime,
      license,
      thirdParty,

      showAboutModal,
      showWASMErrorModal,
      code,
      linesMsgs,
      loading,
      delayedLoading,
      result,
      edited,
      cancelInProcess
    }
  },
  watch: {
    code() {
      if (!this.loading && this.code && this.checkCode) {
        this.checkCode()
      }
    },
    loading(newVal) {
      if (newVal) {
        setTimeout(() => {
          this.delayedLoading = this.loading
        }, 1000)
      } else {
        this.delayedLoading = false
      }
    }
  },
  created() {
    if (this.checkWASM) {
      this.worker = Comlink.wrap(new Worker())
    }
  },
  mounted() {
    if (typeof globalThis.window !== 'undefined') {
      const unwatch = this.$watch('edited', newVal => {
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
        ln.addEventListener('click', e => {
          if (!e.target.classList.contains('line-number-red')) return
          const msgs = this.linesMsgs[e.target.innerText] || []
          if (!Array.isArray(msgs)) return
          miniToastr.error(
            `<pre style="white-space: pre-wrap">${msgs.join('\n\n').trim()}</pre>`,
            'Warning',
            5000,
            undefined,
            { allowHtml: true }
          )
        })
      }
    }
    if (typeof globalThis.document !== 'undefined') {
      const pg = document.querySelector('.playground-editor')
      if (pg) {
        pg.addEventListener('keyup', e => {
          if (e.ctrlKey && e.keyCode === 13) this.onRun()
        })
      }
    }

    if (!this.checkWASM) return

    let counter = 0

    const f = async () => {
      const ok = await this.worker.isLoaded()
      if (ok) {
        this.loading = false
        this.checkCode = debounce(() => {
          this.worker.checkUGO(Comlink.proxy(this), this.code.toString())
        }, 1000)
      } else {
        counter++
        if (counter > 100) {
          counter = 0
          this.loading = false
          if (!this.showWASMErrorModal) this.showWASMErrorModal = true
        }
        setTimeout(f, 250)
      }
    }
    setTimeout(f, 250)
  },
  methods: {
    highlighter(code) {
      return highlight(code, languages.ugo)
    },
    onRun() {
      if (this.loading) return

      this.result = null
      this.loading = true

      try {
        this.worker.runUGO(Comlink.proxy(this), this.code.toString())
      } catch (err) {
        console.log(err)
        this.result = { error: err.toString() }
        this.loading = false
      }
    },
    async onCancel() {
      this.cancelInProcess = true

      try {
        const canceled = await this.worker.cancelUGO()
        console.log('Cancel result:', canceled)
      } catch (err) {
        console.log('Cancel error:', err)
      } finally {
        this.cancelInProcess = false
      }
    },
    resultCallback(msg) {
      this.loading = false
      this.result = msg
    },
    valueToJSON(value) {
      try {
        return JSON.stringify(JSON.parse(value), null, 2)
      } catch (err) {
        return `JSON Error: ${err.toString()}`
      }
    },
    onEditorClick() {
      const elem = document.querySelector('.prism-editor__textarea')
      if (elem) {
        elem.focus()
        elem.click()
      }
    },
    checkCallback(result) {
      if (typeof result !== 'undefined') {
        if (result.warning) {
          console.log('check warning:', result.warning)
        }
        this.highlightLine(result.lines || {})
        return
      }
      this.highlightLine({})
    },
    highlightLine(linesMsgs) {
      this.linesMsgs = linesMsgs
      const lines = document.querySelectorAll('.prism-editor__line-number')
      if (!lines) return
      lines.forEach(el => {
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
  font-family:
    Fira code,
    Fira Mono,
    Consolas,
    Menlo,
    Courier,
    monospace;
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
  padding-top: 8px;
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
    padding-left: 8px;
    padding-right: 8px;
  }
}
</style>
