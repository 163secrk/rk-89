<template>
  <div class="meal-allowance-records">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">餐补明细记录</span>
          <div class="header-actions">
            <el-radio-group v-model="activeTab" @change="handleTabChange">
              <el-radio-button label="records">充值/消费记录</el-radio-button>
              <el-radio-button label="consumptions">消费订单明细</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <div v-if="activeTab === 'records'">
        <div class="filter-bar">
          <el-select v-model="filter.userId" placeholder="选择用户" style="width: 180px; margin-right: 10px;" clearable @change="fetchRecords">
            <el-option
              v-for="user in users"
              :key="user.id"
              :label="user.name"
              :value="user.id"
            />
          </el-select>
          <el-select v-model="filter.type" placeholder="全部类型" style="width: 120px; margin-right: 10px;" clearable @change="fetchRecords">
            <el-option label="充值" value="recharge" />
            <el-option label="消费" value="consume" />
            <el-option label="退款" value="refund" />
          </el-select>
          <el-date-picker
            v-model="filter.startDate"
            type="date"
            placeholder="开始日期"
            value-format="YYYY-MM-DD"
            style="width: 140px; margin-right: 10px;"
            @change="fetchRecords"
          />
          <el-date-picker
            v-model="filter.endDate"
            type="date"
            placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 140px; margin-right: 10px;"
            @change="fetchRecords"
          />
          <el-button type="primary" :icon="Refresh" @click="fetchRecords">查询</el-button>
          <el-button :icon="Download" @click="exportRecords" style="margin-left: 10px;">导出</el-button>
        </div>

        <el-table :data="records" style="width: 100%" v-loading="loading">
          <el-table-column prop="created_at" label="时间" width="180">
            <template #default="{ row }">
              <span>{{ formatTime(row.created_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="user.name" label="用户" width="120">
            <template #default="{ row }">
              <span>{{ row.user?.name || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="getTypeTagType(row.type)" size="small">
                {{ getTypeText(row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="amount" label="金额" width="120">
            <template #default="{ row }">
              <span :class="row.amount >= 0 ? 'text-success' : 'text-danger'">
                {{ row.amount >= 0 ? '+' : '' }}{{ row.amount.toFixed(2) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="balance_before" label="变动前" width="120">
            <template #default="{ row }">
              <span>¥{{ row.balance_before.toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="balance_after" label="变动后" width="120">
            <template #default="{ row }">
              <span>¥{{ row.balance_after.toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="related_no" label="关联单号" width="180">
            <template #default="{ row }">
              <span v-if="row.related_no">{{ row.related_no }}</span>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="operator_name" label="操作人" width="120">
            <template #default="{ row }">
              <span v-if="row.operator_name">{{ row.operator_name }}</span>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="150">
            <template #default="{ row }">
              <span v-if="row.remark">{{ row.remark }}</span>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
        </el-table>

        <div class="summary-bar">
          <span>合计：{{ records.length }} 条记录</span>
          <span style="margin-left: 30px;">
            充值总额：<span class="text-success">¥{{ totalRecharge.toFixed(2) }}</span>
          </span>
          <span style="margin-left: 30px;">
            消费总额：<span class="text-danger">¥{{ totalConsume.toFixed(2) }}</span>
          </span>
          <span style="margin-left: 30px;">
            退款总额：<span class="text-warning">¥{{ totalRefund.toFixed(2) }}</span>
          </span>
        </div>
      </div>

      <div v-if="activeTab === 'consumptions'">
        <div class="filter-bar">
          <el-select v-model="consumptionFilter.userId" placeholder="选择用户" style="width: 180px; margin-right: 10px;" clearable @change="fetchConsumptions">
            <el-option
              v-for="user in users"
              :key="user.id"
              :label="user.name"
              :value="user.id"
            />
          </el-select>
          <el-date-picker
            v-model="consumptionFilter.startDate"
            type="date"
            placeholder="开始日期"
            value-format="YYYY-MM-DD"
            style="width: 140px; margin-right: 10px;"
            @change="fetchConsumptions"
          />
          <el-date-picker
            v-model="consumptionFilter.endDate"
            type="date"
            placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 140px; margin-right: 10px;"
            @change="fetchConsumptions"
          />
          <el-button type="primary" :icon="Refresh" @click="fetchConsumptions">查询</el-button>
        </div>

        <el-table :data="consumptions" style="width: 100%" v-loading="consumptionsLoading">
          <el-table-column prop="order_no" label="订单号" width="180" />
          <el-table-column prop="user.name" label="用户" width="120">
            <template #default="{ row }">
              <span>{{ row.user?.name || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="meal_date" label="用餐日期" width="120" />
          <el-table-column prop="meal_time" label="用餐时段" width="100">
            <template #default="{ row }">
              {{ getMealTimeText(row.meal_time) }}
            </template>
          </el-table-column>
          <el-table-column label="菜品" min-width="200">
            <template #default="{ row }">
              <span v-for="(item, index) in row.items" :key="item.id">
                {{ item.dish?.name }}{{ index < row.items.length - 1 ? '、' : '' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="total_price" label="消费金额" width="120">
            <template #default="{ row }">
              <span class="text-danger">¥{{ row.total_price.toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusTagType(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="下单时间" width="180">
            <template #default="{ row }">
              <span>{{ formatTime(row.created_at) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <div class="summary-bar">
          <span>合计：{{ consumptions.length }} 笔订单</span>
          <span style="margin-left: 30px;">
            消费总金额：<span class="text-danger">¥{{ totalConsumptionAmount.toFixed(2) }}</span>
          </span>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Download } from '@element-plus/icons-vue'
import { getUsers } from '@/api/user'
import { getMealAllowanceRecords, getConsumptionRecords } from '@/api/mealAllowance'

const activeTab = ref('records')
const loading = ref(false)
const consumptionsLoading = ref(false)
const records = ref([])
const consumptions = ref([])
const users = ref([])

const filter = reactive({
  userId: null,
  type: '',
  startDate: '',
  endDate: ''
})

const consumptionFilter = reactive({
  userId: null,
  startDate: '',
  endDate: ''
})

const totalRecharge = computed(() => {
  return records.value
    .filter(r => r.type === 'recharge')
    .reduce((sum, r) => sum + r.amount, 0)
})

const totalConsume = computed(() => {
  return records.value
    .filter(r => r.type === 'consume')
    .reduce((sum, r) => sum + Math.abs(r.amount), 0)
})

const totalRefund = computed(() => {
  return records.value
    .filter(r => r.type === 'refund')
    .reduce((sum, r) => sum + r.amount, 0)
})

const totalConsumptionAmount = computed(() => {
  return consumptions.value.reduce((sum, c) => sum + c.total_price, 0)
})

const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime()) || date.getFullYear() < 1970) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const getTypeText = (type) => {
  const map = {
    recharge: '充值',
    consume: '消费',
    refund: '退款'
  }
  return map[type] || type
}

const getTypeTagType = (type) => {
  const map = {
    recharge: 'success',
    consume: 'danger',
    refund: 'warning'
  }
  return map[type] || 'info'
}

const getStatusText = (status) => {
  const map = {
    pending: '待核销',
    confirmed: '已确认',
    verified: '已核销',
    cancelled: '已取消'
  }
  return map[status] || status
}

const getStatusTagType = (status) => {
  const map = {
    pending: 'warning',
    confirmed: 'primary',
    verified: 'success',
    cancelled: 'info'
  }
  return map[status] || 'info'
}

const getMealTimeText = (time) => {
  const map = {
    breakfast: '早餐',
    lunch: '午餐',
    dinner: '晚餐'
  }
  return map[time] || time
}

const fetchUsers = async () => {
  try {
    const res = await getUsers()
    users.value = res.data || []
  } catch (e) {
    console.log('Get users error')
  }
}

const fetchRecords = async () => {
  loading.value = true
  try {
    const params = {}
    if (filter.userId) params.user_id = filter.userId
    if (filter.type) params.type = filter.type
    if (filter.startDate) params.start_date = filter.startDate
    if (filter.endDate) params.end_date = filter.endDate

    const res = await getMealAllowanceRecords(params)
    records.value = res.data || []
  } catch (e) {
    console.log('Get records error')
  } finally {
    loading.value = false
  }
}

const fetchConsumptions = async () => {
  consumptionsLoading.value = true
  try {
    const params = {}
    if (consumptionFilter.userId) params.user_id = consumptionFilter.userId
    if (consumptionFilter.startDate) params.start_date = consumptionFilter.startDate
    if (consumptionFilter.endDate) params.end_date = consumptionFilter.endDate

    const res = await getConsumptionRecords(params)
    consumptions.value = res.data || []
  } catch (e) {
    console.log('Get consumptions error')
  } finally {
    consumptionsLoading.value = false
  }
}

const handleTabChange = (tab) => {
  if (tab === 'records') {
    fetchRecords()
  } else {
    fetchConsumptions()
  }
}

const exportRecords = () => {
  if (records.value.length === 0) {
    ElMessage.warning('没有可导出的数据')
    return
  }

  const headers = ['时间', '用户', '类型', '金额', '变动前', '变动后', '关联单号', '操作人', '备注']
  const rows = records.value.map(r => [
    formatTime(r.created_at),
    r.user?.name || '',
    getTypeText(r.type),
    (r.amount >= 0 ? '+' : '') + r.amount.toFixed(2),
    r.balance_before.toFixed(2),
    r.balance_after.toFixed(2),
    r.related_no || '',
    r.operator_name || '',
    r.remark || ''
  ])

  const csvContent = [headers, ...rows]
    .map(row => row.map(cell => `"${cell}"`).join(','))
    .join('\n')

  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `餐补明细_${new Date().toISOString().slice(0, 10)}.csv`
  link.click()
  URL.revokeObjectURL(link.href)

  ElMessage.success('导出成功')
}

onMounted(() => {
  fetchUsers()
  fetchRecords()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 16px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
}

.filter-bar {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.summary-bar {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #ebeef5;
  font-size: 14px;
  color: #606266;
}

.text-success {
  color: #67c23a;
  font-weight: 600;
}

.text-danger {
  color: #f56c6c;
  font-weight: 600;
}

.text-warning {
  color: #e6a23c;
  font-weight: 600;
}

.text-muted {
  color: #909399;
}
</style>
