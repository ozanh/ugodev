<template>
  <div>
    <div class="box head">
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
      <div class="head-right-gh">
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
        class="box playground-editor"
      />
      <div class="box result">
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
      <ul
        v-if="result && result.metrics"
        class="button-group metrics hidden-mob"
      >
        <li>
          <button
            class="button"
            disabled
          >
            Compile:{{ result.metrics.compile }}
          </button>
        </li>
        <li>
          <button
            class="button"
            disabled
          >
            Exec:{{ result.metrics.exec }}
          </button>
        </li>
        <li>
          <button
            class="button"
            disabled
          >
            Total:{{ result.metrics.elapsed }}
          </button>
        </li>
      </ul>
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
            Build Time: {{ buildTime }}
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
    loading: true,
    result: null
  }),
  mounted () {
    if (global.window != null) {
      window.addEventListener('beforeunload', function (e) {
        e.preventDefault()
        e.returnValue = 'Stay on Page?'
      })
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
        return
      }
      counter++
      setTimeout(f, 1000)
    }
    if (this.checkWASM) {
      setTimeout(f, 1000)
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
        this.tryScroll()
      }
    },
    resultCallback (msg) {
      this.loading = false
      this.result = msg
      this.$nextTick(() => {
        this.tryScroll()
      })
    },
    valueToJSON (value) {
      try {
        return JSON.stringify(JSON.parse(value), null, 2)
      } catch (err) {
        return `JSON Error: ${err.toString()}`
      }
    },
    tryScroll () {
      if (!this.checkWASM) {
        return
      }
      const resultDiv = document.querySelector('.result')
      if (resultDiv) {
        if (resultDiv.offsetTop > document.querySelector('.body-container').offsetTop + 50) {
          resultDiv.scrollIntoView(true)
        }
      }
    }
  }
}
</script>
<style lang="scss">
.modal__dialog {
  font-size: small;
}

.box {
  width: 98%;
  margin-left: 1%;
  margin-right: 1%;
}

.head {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  flex-basis: 98%;
  align-items: stretch;
}

.playground-text {
  overflow: visible;
  font-size: 16pt;
  flex: 0 0 auto;
}

.head-buttons {
  height: 28px;
  padding-left: 5px;
  flex: 1 0 auto;
  display: flex;
  flex-wrap: nowrap;
}

.head-buttons .button {
  margin-right: 5px;
}

.loader-wrapper {
  width: 20px;
  flex: 0 0 20px;
}

.copyright {
  color: #ccc;
  font-size: 12px;
  text-align: right;
}

.body-container {
  width: 100%;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.head > div {
  padding: 5px;
}

// required class for editor
.playground-editor {
  background: #2d2d2d;
  color: #ccc;
  font-family: Fira code, Fira Mono, Consolas, Menlo, Courier, monospace;
  font-size: 12pt;
  line-height: 1.2;
  padding: 0px;
  max-height: 100vh;
}

// optional
.prism-editor__textarea:focus {
  outline: none;
}

.result {
  font-size: 14px;
  background: #2d2d2d;
}

.result-error {
  color: #ff6565;
  font-weight: 200;
  width: 100%;
}

.head,
.result {
  color: white;
  padding: 5px;
}

.result-value {
  padding-bottom: 25px;
}

pre {
  white-space: pre-wrap; /* Since CSS 2.1 */
  white-space: -moz-pre-wrap; /* Mozilla, since 1999 */
  white-space: -pre-wrap; /* Opera 4-6 */
  white-space: -o-pre-wrap; /* Opera 7 */
  word-wrap: break-word; /* Internet Explorer 5.5+ */
}

.footer {
  position: fixed;
  bottom: 0;
  width: 100%;
  text-align: right;
}

@media (max-width: 600px) {
  .hidden-mob {
    display: none;
  }
  .playground-text {
    justify-content: center;
    text-align: center;
  }
  .head-buttons {
    justify-content: left;
  }
}

@media (min-width: 992px) {
  .playground-editor {
    width: 48%;
    max-height: unset;
  }
  .result {
    width: 48%;
    margin-left: 0;
  }
}
</style>
