import { createApp } from 'vue'
import { createPinia } from 'pinia'

import '@/shared/assets/brand/tokens.css'
import '@/shared/assets/brand/bm.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
