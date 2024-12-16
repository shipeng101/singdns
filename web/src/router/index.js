import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/views/layout/index.vue'),
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/dashboard/index.vue'),
          meta: { title: '仪表盘' }
        },
        {
          path: 'nodes',
          name: 'nodes',
          component: () => import('@/views/nodes/index.vue'),
          meta: { title: '节点管理' }
        },
        {
          path: 'rules',
          name: 'rules',
          component: () => import('@/views/rules/index.vue'),
          meta: { title: '规则管理' }
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '系统设置' }
        }
      ]
    }
  ]
})

export default router 