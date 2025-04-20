import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('../views/Home.vue'),
    },
    {
      path: '/car-dealer', // 修改路径
      component: () => import('../views/CarDealer.vue'), // 修改组件路径
    },
    {
      path: '/trading-platform',
      component: () => import('../views/TradingPlatform.vue'),
    },
    {
      path: '/bank',
      component: () => import('../views/Bank.vue'),
    },
  ],
})

export default router
