import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('../views/Layout.vue'),
    redirect: '/connections',
    children: [
      {
        path: 'connections',
        name: 'Connections',
        component: () => import('../views/ConnectionManager.vue')
      },
      {
        path: 'query',
        name: 'Query',
        component: () => import('../views/QueryEditor.vue')
      },
      {
        path: 'table-data',
        name: 'TableData',
        component: () => import('../views/TableData.vue')
      },
      {
        path: 'table-structure',
        name: 'TableStructure',
        component: () => import('../views/TableStructure.vue')
      },
      {
        path: 'import-export',
        name: 'ImportExport',
        component: () => import('../views/ImportExport.vue')
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// navigation guard
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
