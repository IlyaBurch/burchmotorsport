import { createRouter, createWebHistory } from 'vue-router'
import { SEO } from '../seo'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/pages/home/ui/HomePage.vue'),
    },
    {
      path: '/live',
      name: 'live',
      component: () => import('@/pages/live/ui/LivePage.vue'),
      meta: { full: true }, // edge-to-edge, no bm-container
    },
    {
      path: '/design-system',
      name: 'design-system',
      component: () => import('@/pages/design-system/ui/DesignSystemPage.vue'),
    },
  ],
})

router.afterEach((to) => {
  const seo = SEO[to.path] ?? SEO['/']!
  document.title = seo.title
  document.querySelector('meta[name="description"]')?.setAttribute('content', seo.description)
})

export default router
