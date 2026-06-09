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
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Food, User, UserFilled, ArrowDown, SwitchButton } from '@element-plus/icons-vue'
import { getUser, removeToken, removeUser } from '@/utils/auth'
import { logout } from '@/api/auth'
import menuConfig from '@/config/menu'

const route = useRoute()
const router = useRouter()
const user = ref(getUser())

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
        router.push('/login')
      }).catch(() => {})
  } else if (command === 'profile') {
    router.push('/dashboard')
  }
}
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
</style>
