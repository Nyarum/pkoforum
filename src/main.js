import './app.css'
import { start } from '@sveltejs/kit/src/runtime/client/start'

start({
  target: document.body,
  paths: {
    base: '',
    assets: ''
  }
}) 