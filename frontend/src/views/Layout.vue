<template>
  <el-container class="layout-container">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="28" color="#409EFF"><Food /></el-icon>
        <span class="logo-text">智能配餐系统</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :default-openeds="defaultOpeneds"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        class="sidebar-menu"
      >
        <template v-for="menu in filteredMenus" :key="menu.path">
          <el-sub-menu
            v-if="menu.children && menu.children.length > 1"
            :index="menu.path"
          >
            <template #title>
              <el-icon><component :is="menu.icon" /></el-icon>
              <span>{{ menu.title }}</span>
            </template>
            <el-menu-item
              v-for="child in menu.children"
              :key="child.path"
              :index="child.path"
            >
              <el-icon><component :is="child.icon" /></el-icon>
              <span>{{ child.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item
            v-else-if="menu.children && menu.children.length === 1"
            :index="menu.children[0].path"
          >
            <el-icon><component :is="menu.icon" /></el-icon>
            <span>{{ menu.title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-popover
            placement="bottom-end"
            :width="380"
            trigger="click"
            v-model="notificationPopoverVisible"
            @show="fetchNotifications"
          >
            <template #reference>
              <el-badge :value="unreadCount" :max="99" :hidden="unreadCount === 0" class="notification-bell">
                <el-button type="primary" :icon="Bell" circle plain />
              </el-badge>
            </template>
            <div class="notification-panel">
              <div class="notification-header">
                <span class="notification-title">系统通知</span>
                <el-button type="text" size="small" @click="handleMarkAllRead" :disabled="unreadCount === 0">
                  全部已读
                </el-button>
              </div>
              <el-tabs v-model="activeNotificationTab" size="small">
                <el-tab-pane label="全部" name="all" />
                <el-tab-pane label="未读" name="unread" />
              </el-tabs>
              <div class="notification-list" v-loading="notificationLoading">
                <div
                  v-for="item in filteredNotifications"
                  :key="item.id"
                  class="notification-item"
                  :class="{ unread: item.status === 'unread' }"
                  @click="handleNotificationClick(item)"
                >
                  <div class="notification-icon">
                    <el-icon :size="18" :color="getPriorityColor(item.priority)">
                      <component :is="getNotificationIcon(item.type)" />
                    </el-icon>
                  </div>
                  <div class="notification-content">
                    <div class="notification-item-header">
                      <span class="notification-item-title">{{ item.title }}</span>
                      <span class="notification-item-time">{{ formatNotificationTime(item.created_at) }}</span>
                    </div>
                    <div class="notification-item-body">{{ item.content.split('\n')[0] }}</div>
                    <div class="notification-item-meta">
                      <el-tag v-if="item.related_no" size="small" type="info">{{ item.related_no }}</el-tag>
                      <el-tag v-if="item.priority === 'high'" size="small" type="danger">重要</el-tag>
                    </div>
                  </div>
                  <div v-if="item.status === 'unread'" class="unread-dot" />
                </div>
                <el-empty v-if="filteredNotifications.length === 0" description="暂无通知" :image-size="60" />
              </div>
              <div class="notification-footer">
                <el-button type="primary" link @click="goToNotifications">查看全部</el-button>
              </div>
            </div>
          </el-popover>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" :icon="UserFilled" />
              <span class="username">{{ user?.name || user?.username }}</span>
              <span class="role-tag" :class="user?.role === 'admin' ? 'admin' : 'user'">
                {{ user?.role === 'admin' ? '管理员' : '用户' }}
              </span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>个人信息
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Food, User, UserFilled, ArrowDown, SwitchButton, Bell, Warning, ShoppingCart, InfoFilled, CircleCheck } from '@element-plus/icons-vue'
import { getUser, removeToken, removeUser } from '@/utils/auth'
import { logout } from '@/api/auth'
import { getNotifications, markNotificationRead, markAllNotificationsRead, getUnreadNotificationCount } from '@/api/inventory'
import menuConfig from '@/config/menu'

const route = useRoute()
const router = useRouter()
const user = ref(getUser())

const notificationPopoverVisible = ref(false)
const notificationLoading = ref(false)
const notifications = ref([])
const unreadCount = ref(0)
const activeNotificationTab = ref('all')
let notificationTimer = null

const filteredNotifications = computed(() => {
  if (activeNotificationTab.value === 'unread') {
    return notifications.value.filter(n => n.status === 'unread').slice(0, 10)
  }
  return notifications.value.slice(0, 10)
})

const hasPermission = (menu) => {
  if (menu.roles) {
    return menu.roles.includes(user.value?.role)
  }
  return true
}

const filterMenus = (menus) => {
  return menus.filter(menu => {
    if (!hasPermission(menu)) {
      return false
    }
    if (menu.children) {
      menu.children = menu.children.filter(child => hasPermission(child))
      return menu.children.length > 0
    }
    return true
  })
}

const filteredMenus = computed(() => filterMenus(JSON.parse(JSON.stringify(menuConfig))))

const findParentMenu = (path, menus) => {
  for (const menu of menus) {
    if (menu.children?.some(child => child.path === path)) {
      return menu
    }
  }
  return null
}

const defaultOpeneds = computed(() => {
  const parent = findParentMenu(route.path, filteredMenus.value)
  return parent ? [parent.path] : []
})

const activeMenu = computed(() => route.path)

const currentTitle = computed(() => {
  for (const menu of filteredMenus.value) {
    if (menu.children) {
      for (const child of menu.children) {
        if (child.path === route.path) {
          return child.title
        }
      }
    }
  }
  return route.meta?.title || ''
})

const handleCommand = async (command) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(async () => {
        try {
          await logout()
        } catch (e) {
          console.log('Logout api error, but will continue')
        }
        removeToken()
        removeUser()
        stopNotificationPolling()
        router.push('/login')
      }).catch(() => {})
  } else if (command === 'profile') {
    router.push('/dashboard')
  }
}

const fetchNotifications = async () => {
  if (notificationLoading.value) return
  notificationLoading.value = true
  try {
    const res = await getNotifications()
    if (res.success) {
      notifications.value = res.data.list || []
      unreadCount.value = res.data.unread_count || 0
    }
  } catch (e) {
    console.error('获取通知失败', e)
  } finally {
    notificationLoading.value = false
  }
}

const fetchUnreadCount = async () => {
  try {
    const res = await getUnreadNotificationCount()
    if (res.success) {
      unreadCount.value = res.data.unread_count || 0
    }
  } catch (e) {
    console.error('获取未读通知数量失败', e)
  }
}

const handleNotificationClick = async (item) => {
  if (item.status === 'unread') {
    try {
      await markNotificationRead(item.id)
      item.status = 'read'
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    } catch (e) {
      console.error('标记通知已读失败', e)
    }
  }

  if (item.related_type === 'purchase' && item.related_id) {
    notificationPopoverVisible.value = false
    router.push(`/purchases/${item.related_id}`)
  }
}

const handleMarkAllRead = async () => {
  try {
    await markAllNotificationsRead()
    notifications.value.forEach(n => n.status = 'read')
    unreadCount.value = 0
    ElMessage.success('已全部标记为已读')
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const goToNotifications = () => {
  notificationPopoverVisible.value = false
  router.push('/notifications')
}

const getNotificationIcon = (type) => {
  const icons = {
    auto_replenish: ShoppingCart,
    low_stock: Warning,
    system: InfoFilled,
    success: CircleCheck
  }
  return icons[type] || InfoFilled
}

const getPriorityColor = (priority) => {
  const colors = {
    high: '#f56c6c',
    normal: '#409eff',
    low: '#909399'
  }
  return colors[priority] || '#909399'
}

const formatNotificationTime = (time) => {
  if (!time) return ''
  const now = new Date()
  const notifyTime = new Date(time)
  const diff = now - notifyTime
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  return notifyTime.toLocaleDateString('zh-CN')
}

const startNotificationPolling = () => {
  fetchUnreadCount()
  notificationTimer = setInterval(() => {
    fetchUnreadCount()
  }, 30000)
}

const stopNotificationPolling = () => {
  if (notificationTimer) {
    clearInterval(notificationTimer)
    notificationTimer = null
  }
}

onMounted(() => {
  startNotificationPolling()
})

onUnmounted(() => {
  stopNotificationPolling()
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  transition: width 0.3s;
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  background-color: #2b2f3a;
  color: white;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  margin-left: 10px;
  color: white;
}

.sidebar-menu {
  border-right: none;
}

.header {
  background-color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  height: 60px;
}

.header-left {
  flex: 1;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 0 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  margin: 0 8px 0 10px;
  font-size: 14px;
  color: #303133;
}

.role-tag {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  margin-right: 8px;
}

.role-tag.admin {
  background-color: #ecf5ff;
  color: #409eff;
}

.role-tag.user {
  background-color: #f0f9eb;
  color: #67c23a;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.notification-bell {
  margin-right: 16px;
}

.notification-panel {
  max-height: 500px;
  display: flex;
  flex-direction: column;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
}

.notification-title {
  font-weight: 600;
  font-size: 14px;
  color: #303133;
}

.notification-list {
  flex: 1;
  overflow-y: auto;
  max-height: 340px;
  margin: 8px 0;
}

.notification-item {
  display: flex;
  padding: 12px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  position: relative;
}

.notification-item:hover {
  background-color: #f5f7fa;
}

.notification-item.unread {
  background-color: #ecf5ff;
}

.notification-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: #ecf5ff;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  flex-shrink: 0;
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-item-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 4px;
}

.notification-item-title {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-item-time {
  font-size: 12px;
  color: #909399;
  flex-shrink: 0;
  margin-left: 8px;
}

.notification-item-body {
  font-size: 12px;
  color: #606266;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  margin-bottom: 6px;
}

.notification-item-meta {
  display: flex;
  gap: 6px;
}

.unread-dot {
  position: absolute;
  top: 12px;
  right: 8px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #f56c6c;
}

.notification-footer {
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
  text-align: center;
}
</style>
