import { createApp } from 'vue'
import App from './App.vue'
import miniToastr from 'mini-toastr'

miniToastr.init()

createApp(App).mount('#app')
