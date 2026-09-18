import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { SEO } from './src/app/seo'

// ponytail: per-route HTML for crawlers. Emits dist/<route>/index.html with that
// route's title/description swapped in; Caddy's try_files picks it up. SSR when
// pages need real content, not before.
const seoPages = () => ({
  name: 'seo-pages',
  closeBundle() {
    const base = readFileSync('dist/index.html', 'utf8')
    for (const [path, { title, description }] of Object.entries(SEO)) {
      if (path === '/') continue
      const html = base
        .replace(/<title>.*?<\/title>/, `<title>${title}</title>`)
        .replace(/(name="description" content=")[^"]*/, `$1${description}`)
        .replace(/(property="og:title" content=")[^"]*/, `$1${title}`)
        .replace(/(property="og:description" content=")[^"]*/, `$1${description}`)
      mkdirSync(`dist${path}`, { recursive: true })
      writeFileSync(`dist${path}/index.html`, html)
    }
  },
})

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueDevTools(), seoPages()],
  server: {
    proxy: { '/api': 'http://localhost:8090' },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
