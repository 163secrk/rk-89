import {
  DataBoard,
  KnifeFork,
  List,
  Calendar,
  User,
  Shop,
  Camera,
  DataAnalysis,
  Grid,
  Warning,
  Document,
  Bottom,
  Top,
  Goods,
  Tickets,
  Stamp,
  Setting,
  Refresh,
  Bell,
  Food
} from '@element-plus/icons-vue'

export const menuConfig = [
  {
    path: '/dashboard',
    title: '数据概览',
    icon: DataBoard,
    children: [
      {
        path: '/dashboard',
        title: '仪表盘',
        icon: DataBoard
      },
      {
        path: '/clean-plate',
        title: '光盘行动统计',
        icon: Food,
        roles: ['admin']
      }
    ]
  },
  {
    path: '/dishes',
    title: '菜品管理',
    icon: KnifeFork,
    roles: ['admin'],
    children: [
      {
        path: '/dishes',
        title: '菜品列表',
        icon: KnifeFork,
        roles: ['admin']
      }
    ]
  },
  {
    path: '/orders',
    title: '订单管理',
    icon: Tickets,
    children: [
      {
        path: '/orders',
        title: '订单列表',
        icon: List
      }
    ]
  },
  {
    path: '/mealplans',
    title: '配餐管理',
    icon: Calendar,
    children: [
      {
        path: '/mealplans',
        title: '配餐计划',
        icon: Calendar,
        roles: ['admin']
      },
      {
        path: '/selection',
        title: '智能配餐',
        icon: Shop
      }
    ]
  },
  {
    path: '/inventory-dashboard',
    title: '库存管理',
    icon: Goods,
    roles: ['admin'],
    children: [
      {
        path: '/inventory-dashboard',
        title: '库存看板',
        icon: DataAnalysis,
        roles: ['admin']
      },
      {
        path: '/inventory-manage',
        title: '库存管理',
        icon: Grid,
        roles: ['admin']
      },
      {
        path: '/inventory-alerts',
        title: '库存预警',
        icon: Warning,
        roles: ['admin']
      },
      {
        path: '/inventory-logs',
        title: '操作日志',
        icon: Document,
        roles: ['admin']
      },
      {
        path: '/stock-inbound',
        title: '采购入库',
        icon: Bottom,
        roles: ['admin']
      },
      {
        path: '/stock-outbound',
        title: '备菜出库',
        icon: Top,
        roles: ['admin']
      },
      {
        path: '/auto-replenishment',
        title: '自动补货',
        icon: Refresh,
        roles: ['admin']
      }
    ]
  },
  {
    path: '/verification',
    title: '核销管理',
    icon: Stamp,
    children: [
      {
        path: '/verification',
        title: '窗口核销',
        icon: Camera
      },
      {
        path: '/verification-records',
        title: '核销记录',
        icon: List
      }
    ]
  },
  {
    path: '/users',
    title: '系统管理',
    icon: Setting,
    roles: ['admin'],
    children: [
      {
        path: '/users',
        title: '用户管理',
        icon: User,
        roles: ['admin']
      },
      {
        path: '/notifications',
        title: '通知中心',
        icon: Bell
      }
    ]
  }
]

export default menuConfig
