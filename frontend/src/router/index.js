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
      },
      {
        path: 'verification',
        name: 'Verification',
        component: () => import('@/views/Verification.vue'),
        meta: { title: '窗口核销', icon: 'Camera' }
      },
      {
        path: 'verification-records',
        name: 'VerificationRecords',
        component: () => import('@/views/VerificationRecords.vue'),
        meta: { title: '核销记录', icon: 'List' }
      },
      {
        path: 'inventory-dashboard',
        name: 'InventoryDashboard',
        component: () => import('@/views/InventoryDashboard.vue'),
        meta: { title: '库存看板', icon: 'DataAnalysis', roles: ['admin'] }
      },
      {
        path: 'inventory-manage',
        name: 'InventoryManage',
        component: () => import('@/views/InventoryManage.vue'),
        meta: { title: '库存管理', icon: 'Grid', roles: ['admin'] }
      },
      {
        path: 'inventory-alerts',
        name: 'InventoryAlerts',
        component: () => import('@/views/InventoryAlerts.vue'),
        meta: { title: '库存预警', icon: 'Warning', roles: ['admin'] }
      },
      {
        path: 'inventory-logs',
        name: 'InventoryLogs',
        component: () => import('@/views/InventoryLogs.vue'),
        meta: { title: '操作日志', icon: 'Document', roles: ['admin'] }
      },
      {
        path: 'stock-inbound',
        name: 'StockInbound',
        component: () => import('@/views/StockInbound.vue'),
        meta: { title: '采购入库', icon: 'Bottom', roles: ['admin'] }
      },
      {
        path: 'stock-outbound',
        name: 'StockOutbound',
        component: () => import('@/views/StockOutbound.vue'),
        meta: { title: '备菜出库', icon: 'Top', roles: ['admin'] }
      },
      {
        path: 'auto-replenishment',
        name: 'AutoReplenishment',
        component: () => import('@/views/AutoReplenishment.vue'),
        meta: { title: '自动补货', icon: 'Refresh', roles: ['admin'] }
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('@/views/Notifications.vue'),
        meta: { title: '通知中心', icon: 'Bell' }
      },
      {
        path: 'purchases/:id',
        name: 'PurchaseDetail',
        component: () => import('@/views/PurchaseDetail.vue'),
        meta: { title: '补货单详情', icon: 'Document', roles: ['admin'] }
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
