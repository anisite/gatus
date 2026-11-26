import { createApp } from 'vue'
import App from './App.vue'
import './index.css'
import './utd-theme.css'
import router from './router'
import { applyUiTheme } from './utils/uiTheme'
import { t } from './i18n'

applyUiTheme()
const app = createApp(App)
app.config.globalProperties.$t = t
app.use(router).mount('#app')
