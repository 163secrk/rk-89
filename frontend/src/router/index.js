import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/utils/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'DataBoard' }
      },
      {
        path: 'dishes',
        name: 'DishList',
        component: () => import('@/views/DishList.vue'),
        meta: { title: '菜品管理', icon: 'KnifeFork', roles: ['admin'] }
      },
      {
        path: 'orders',
        name: 'OrderList',
        component: () => import('@/views/OrderList.vue'),
        meta: { title: '订单管理', icon: 'List' }
      },
      {
        path: 'mealplans',
        name: 'MealPlan',
        component: () => import('@/views/MealPlan.vue'),
        meta: { title: '配餐计划', icon: 'Calendar', roles: ['admin'] }
      },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('@/views/UserList.vue'),
        meta: { title: '用户管理', icon: 'User', roles: ['admin'] }
      },
      {
        path: 'selection',
        name: 'MealSelection',
        component: () => import('@/views/MealSelection.vue'),
        meta: { title: '智能配餐', icon: 'Shop' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = getToken()
  const user = JSON.parse(localStorage.getItem('user') || 'null')
  
  if (to.path === '/login') {
    if (token) {
      next('/')
    } else {
      next()
    }
  } else {
    if (!token) {
      next('/login')
    } else {
      if (to.meta.roles && user && !to.meta.roles.includes(user.role)) {
        next('/dashboard')
      } else {
        next()
      }
    }
  }
})

export default router
