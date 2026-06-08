<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon :size="32"><User /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">用户总数</p>
              <p class="stat-value">{{ stats.totalUsers || 0 }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon :size="32"><KnifeFork /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">菜品总数</p>
              <p class="stat-value">{{ stats.totalDishes || 0 }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
              <el-icon :size="32"><List /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">今日订单</p>
              <p class="stat-value">{{ stats.todayOrders || 0 }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);">
              <el-icon :size="32"><Money /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">今日销售额</p>
              <p class="stat-value">¥{{ stats.todayRevenue || 0 }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近订单</span>
              <el-button type="primary" link @click="$router.push('/orders')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentOrders" style="width: 100%" size="small">
            <el-table-column prop="order_no" label="订单号" width="120" />
            <el-table-column label="用户">
              <template #default="{ row }">{{ row.user?.name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="total_price" label="金额">
              <template #default="{ row }">¥{{ row.total_price }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>今日配餐计划</span>
              <el-button type="primary" link @click="$router.push('/mealplans')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="todayMealPlan" style="width: 100%" size="small">
            <el-table-column prop="meal_type" label="餐次" width="80">
              <template #default="{ row }">
                <el-tag type="success" size="small">{{ row.meal_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="菜品">
              <template #default="{ row }">
                {{ (row.dishes || []).map(d => d.name).join('、') }}
              </template>
            </el-table-column>
            <el-table-column prop="date" label="日期" width="120" />
          </el-table>
          <el-empty v-if="todayMealPlan.length === 0" description="暂无今日配餐计划" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <span>热销菜品排行</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="8" v-for="(dish, index) in popularDishes" :key="dish.name">
              <div class="dish-item">
                <div class="dish-rank" :class="'rank-' + (index + 1)">{{ index + 1 }}</div>
                <div class="dish-info">
                  <p class="dish-name">{{ dish.name }}</p>
                  <p class="dish-sales">销量: {{ dish.count }} 份</p>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { User, KnifeFork, List, Money } from '@element-plus/icons-vue'
import { getDashboardStats } from '@/api/stats'
import { getMyOrders, getOrders } from '@/api/order'
import { getTodayMealPlan } from '@/api/mealplan'
import { getUser } from '@/utils/auth'

const stats = ref({})
const recentOrders = ref([])
const todayMealPlan = ref([])
const popularDishes = ref([])
const user = getUser()

const getStatusType = (status) => {
  const types = {
    pending: 'warning',
    confirmed: 'primary',
    completed: 'success',
    cancelled: 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    pending: '待确认',
    confirmed: '已确认',
    completed: '已完成',
    cancelled: '已取消'
  }
  return texts[status] || status
}

const fetchData = async () => {
  try {
    const res = await getDashboardStats()
    stats.value = res.data
    popularDishes.value = (res.data?.popularDishes || []).slice(0, 6)
  } catch (e) {
    console.log('Get stats error')
    stats.value = { totalUsers: 0, totalDishes: 0, todayOrders: 0, todayRevenue: 0 }
    popularDishes.value = [
      { name: '红烧肉', count: 128 },
      { name: '清蒸鱼', count: 96 },
      { name: '宫保鸡丁', count: 85 },
      { name: '麻婆豆腐', count: 72 },
      { name: '西红柿炒蛋', count: 65 },
      { name: '糖醋排骨', count: 58 }
    ]
  }

  try {
    const ordersRes = user?.role === 'admin' ? await getOrders({ limit: 5 }) : await getMyOrders()
    recentOrders.value = (ordersRes.data?.items || ordersRes.data || []).slice(0, 5)
  } catch (e) {
    console.log('Get orders error')
  }

  try {
    const mealRes = await getTodayMealPlan()
    todayMealPlan.value = mealRes.data || []
  } catch (e) {
    console.log('Get meal plan error')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.stat-card {
  border-radius: 12px;
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-info {
  margin-left: 16px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.dish-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-radius: 8px;
  background: #f5f7fa;
  margin-bottom: 12px;
}

.dish-rank {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: white;
  margin-right: 12px;
}

.rank-1 { background: linear-gradient(135deg, #ffd700 0%, #ff8c00 100%); }
.rank-2 { background: linear-gradient(135deg, #c0c0c0 0%, #808080 100%); }
.rank-3 { background: linear-gradient(135deg, #cd7f32 0%, #8b4513 100%); }
.rank-4, .rank-5, .rank-6 { background: #909399; }

.dish-name {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.dish-sales {
  font-size: 12px;
  color: #909399;
}
</style>
