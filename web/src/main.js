import { createApp } from 'vue'
import App from './App.vue'
import MiniApp from './MiniApp.vue'
import './style.css'
import { isTelegramMiniApp } from './telegramWebApp'

const Root = isTelegramMiniApp() ? MiniApp : App
createApp(Root).mount('#app')
