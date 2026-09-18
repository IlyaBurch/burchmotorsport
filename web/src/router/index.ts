import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/pages/HomePage.vue'),
    },
    {
      path: '/design-system',
      name: 'design-system',
      component: () => import('@/pages/DesignSystemPage.vue'),
    },
  ],
})

export default router
