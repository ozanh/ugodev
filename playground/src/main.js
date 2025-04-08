import 'vue-prism-editor/dist/prismeditor.min.css'
import 'prismjs/themes/prism-okaidia.css'
import '@/assets/css/button.css'
import '@/assets/css/loader.css'

import { createApp } from 'vue'
import App from './App.vue'
import miniToastr from 'mini-toastr'

miniToastr.init()

createApp(App).mount('#app')
