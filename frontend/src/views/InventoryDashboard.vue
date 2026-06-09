<template>
  <div class="inventory-dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon :size="32"><Goods /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">食材种类</p>
              <p class="stat-value">{{ dashboard.total_ingredients || 0 }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon :size="32"><Money /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">库存总价值</p>
              <p class="stat-value">¥{{ formatNumber(dashboard.total_stock_value) }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);">
              <el-icon :size="32"><Bottom /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">今日入库</p>
              <p class="stat-value">{{ formatNumber(dashboard.today_inbound) }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
              <el-icon :size="32"><Top /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">今日出库</p>
              <p class="stat-value">{{ formatNumber(dashboard.today_outbound) }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="8">
        <el-card class="stat-card warning-card" shadow="hover" @click="$router.push('/inventory-alerts')">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);">
              <el-icon :size="32"><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">库存预警</p>
              <p class="stat-value">{{ dashboard.low_stock_count || 0 }} 种</p>
              <p class="stat-tip">待处理: {{ dashboard.pending_alerts || 0 }} 条</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>库区库存分布</span>
            </div>
          </template>
          <el-row :gutter="20">
            <el-col :span="8" v-for="zone in dashboard.zone_stats || []" :key="zone.zone">
              <div class="zone-card" :class="'zone-' + zone.zone">
                <div class="zone-icon">
                  <el-icon :size="28">
                    <component :is="getZoneIcon(zone.zone)" />
                  </el-icon>
                </div>
                <div class="zone-info">
                  <p class="zone-name">{{ zone.zone_name }}</p>
                  <p class="zone-count">{{ zone.count }} 种食材</p>
                  <p class="zone-value">¥{{ formatNumber(zone.value) }}</p>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近库存变动</span>
              <el-button type="primary" link @click="$router.push('/inventory-records')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="dashboard.recent_records || []" style="width: 100%" size="small">
            <el-table-column label="时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="食材">
              <template #default="{ row }">{{ row.ingredient?.name || '-' }}</template>
            </el-table-column>
            <el-table-column label="类型" width="80">
              <template #default="{ row }">
                <el-tag :type="row.change_type === 'in' ? 'success' : row.change_type === 'out' ? 'danger' : 'info'" size="small">
                  {{ getChangeTypeText(row.change_type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="数量" width="100">
              <template #default="{ row }">
                <span :class="row.change_type === 'in' ? 'text-success' : 'text-danger'">
                  {{ row.change_type === 'in' ? '+' : '-' }}{{ row.change_qty }}{{ row.ingredient?.unit || '' }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="库区" width="100">
              <template #default="{ row }">{{ getZoneName(row.warehouse_zone) }}</template>
            </el-table-column>
          </el-table>
          <el-empty v-if="(dashboard.recent_records || []).length === 0" description="暂无库存变动记录" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>库存不足预警</span>
              <el-button type="primary" link @click="$router.push('/inventory-alerts')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="dashboard.low_stock_items || []" style="width: 100%" size="small">
            <el-table-column prop="name" label="食材名称" />
            <el-table-column prop="category" label="分类" width="100" />
            <el-table-column label="当前库存" width="120">
              <template #default="{ row }">
                <span class="text-danger">{{ row.stock }} {{ row.unit }}</span>
              </template>
            </el-table-column>
            <el-table-column label="安全库存" width="100">
              <template #default="{ row }">{{ row.safety_stock }} {{ row.unit }}</template>
            </el-table-column>
            <el-table-column label="库区" width="100">
              <template #default="{ row }">{{ getZoneName(row.warehouse_zone) }}</template>
            </el-table-column>
          </el-table>
          <el-empty v-if="(dashboard.low_stock_items || []).length === 0" description="所有食材库存充足" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>快捷操作</span>
            </div>
          </template>
          <el-row :gutter="20">
            <el-col :span="6">
              <div class="quick-action" @click="$router.push('/stock-inbound')">
                <el-icon :size="32" color="#67c23a"><Bottom /></el-icon>
                <p>采购入库</p>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="quick-action" @click="$router.push('/stock-outbound')">
                <el-icon :size="32" color="#f56c6c"><Top /></el-icon>
                <p>备菜出库</p>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="quick-action" @click="$router.push('/inventory-manage')">
                <el-icon :size="32" color="#409eff"><Grid /></el-icon>
                <p>库存管理</p>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="quick-action" @click="$router.push('/inventory-logs')">
                <el-icon :size="32" color="#909399"><Document /></el-icon>
                <p>操作日志</p>
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
import { Goods, Money, Warning, Bottom, Top, Grid, Document, Box, Refrigerator, ColdDrink } from '@element-plus/icons-vue'
import { getInventoryDashboard } from '@/api/inventory'
import { ElMessage } from 'element-plus'

const dashboard = ref({})

const formatNumber = (num) => {
  if (num === undefined || num === null) return '0'
  return Number(num).toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getZoneIcon = (zone) => {
  const icons = {
    dry: Box,
    refrigerated: Refrigerator,
    frozen: ColdDrink
  }
  return icons[zone] || Box
}

const getZoneName = (zone) => {
  const names = {
    dry: '干货区',
    refrigerated: '冷藏区',
    frozen: '冷冻区'
  }
  return names[zone] || zone
}

const getChangeTypeText = (type) => {
  const texts = {
    in: '入库',
    out: '出库',
    transfer: '调拨'
  }
  return texts[type] || type
}

const fetchData = async () => {
  try {
    const res = await getInventoryDashboard()
    if (res.success) {
      dashboard.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取库存看板数据失败')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.stat-card {
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
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
  flex: 1;
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

.stat-tip {
  font-size: 12px;
  color: #f56c6c;
  margin-top: 4px;
}

.warning-card {
  border: 1px solid #f56c6c;
  background: linear-gradient(135deg, #fff5f5 0%, #fff 100%);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.zone-card {
  display: flex;
  align-items: center;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 12px;
  transition: all 0.2s;
}

.zone-card:hover {
  transform: translateX(4px);
}

.zone-dry {
  background: linear-gradient(135deg, #ecf5ff 0%, #fff 100%);
  border-left: 4px solid #409eff;
}

.zone-refrigerated {
  background: linear-gradient(135deg, #f0f9eb 0%, #fff 100%);
  border-left: 4px solid #67c23a;
}

.zone-frozen {
  background: linear-gradient(135deg, #fdf6ec 0%, #fff 100%);
  border-left: 4px solid #e6a23c;
}

.zone-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  color: white;
}

.zone-dry .zone-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.zone-refrigerated .zone-icon {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.zone-frozen .zone-icon {
  background: linear-gradient(135deg, #2193b0 0%, #6dd5ed 100%);
}

.zone-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.zone-count {
  font-size: 13px;
  color: #909399;
  margin-bottom: 2px;
}

.zone-value {
  font-size: 14px;
  font-weight: 500;
  color: #409eff;
}

.text-success {
  color: #67c23a;
  font-weight: 500;
}

.text-danger {
  color: #f56c6c;
  font-weight: 500;
}

.quick-action {
  text-align: center;
  padding: 24px;
  border-radius: 8px;
  background: #f5f7fa;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-action:hover {
  background: #e4e7ed;
  transform: translateY(-2px);
}

.quick-action p {
  margin-top: 8px;
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}
</style>
