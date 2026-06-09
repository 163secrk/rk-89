<template>
  <div class="order-list">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">订单管理</span>
          <div class="header-actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索订单号/用户名"
              style="width: 200px; margin-right: 10px;"
              :prefix-icon="Search"
              clearable
              @keyup.enter="fetchOrders"
            />
            <el-select v-model="statusFilter" placeholder="订单状态" style="width: 120px; margin-right: 10px;" clearable @change="fetchOrders">
              <el-option label="待确认" value="pending" />
              <el-option label="已确认" value="confirmed" />
              <el-option label="已完成" value="completed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="fetchOrders">刷新</el-button>
          </div>
        </div>
      </template>
      <el-table :data="orders" style="width: 100%" v-loading="loading">
        <el-table-column prop="order_no" label="订单号" width="120" />
        <el-table-column label="用户" width="100">
          <template #default="{ row }">{{ row.user?.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="用餐" width="120">
          <template #default="{ row }">
            <div>{{ row.meal_date }}</div>
            <div style="font-size: 12px; color: #909399;">{{ getMealTypeText(row.meal_time) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="菜品" show-overflow-tooltip>
          <template #default="{ row }">
            {{ (row.items || []).map(i => i.dish?.name || '').join('、') }}
          </template>
        </el-table-column>
        <el-table-column prop="total_price" label="金额" width="100">
          <template #default="{ row }">¥{{ row.total_price }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="取消时限" width="160">
          <template #default="{ row }">
            <template v-if="(row.status === 'pending' || row.status === 'confirmed') && cancelInfoMap[row.id]">
              <div v-if="canCancelOrder(row)" style="color: #67c23a;">
                <el-icon><Warning /></el-icon>
                剩余 {{ formatRemainingTime(cancelInfoMap[row.id].remaining_time) }}
              </div>
              <div v-else style="color: #f56c6c; font-size: 12px;">
                {{ getCancelReason(row) }}
              </div>
              <div style="font-size: 11px; color: #909399; margin-top: 4px;">
                截止: {{ cancelInfoMap[row.id].cancel_deadline }}
              </div>
            </template>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="下单时间" width="170" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">查看</el-button>
            <el-button type="success" link size="small" v-if="row.status === 'pending' && user?.role === 'admin'" @click="updateStatus(row, 'confirmed')">确认</el-button>
            <el-button type="success" link size="small" v-if="row.status === 'confirmed' && user?.role === 'admin'" @click="updateStatus(row, 'completed')">完成</el-button>
            <el-button 
              type="danger" 
              link 
              size="small" 
              v-if="(row.status === 'pending' || row.status === 'confirmed')"
              :disabled="!canCancelOrder(row)"
              @click="handleCancel(row)"
            >
              取消
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        style="margin-top: 20px; justify-content: flex-end; display: flex;"
        @size-change="fetchOrders"
        @current-change="fetchOrders"
      />
    </el-card>

    <el-dialog v-model="detailVisible" title="订单详情" width="600px">
      <el-descriptions :column="2" border v-if="currentOrder">
        <el-descriptions-item label="订单号">{{ currentOrder.order_no }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ currentOrder.user?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentOrder.status)">{{ getStatusText(currentOrder.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="总金额">¥{{ currentOrder.total_price }}</el-descriptions-item>
        <el-descriptions-item label="下单时间" :span="2">{{ currentOrder.created_at }}</el-descriptions-item>
        <el-descriptions-item label="订单备注" :span="2">{{ currentOrder.remark || '无' }}</el-descriptions-item>
      </el-descriptions>
      <el-divider content-position="left">菜品明细</el-divider>
      <el-table :data="currentOrder?.items || []" size="small">
        <el-table-column label="菜品名称">
          <template #default="{ row }">{{ row.dish?.name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="100" />
        <el-table-column prop="price" label="单价" width="100">
          <template #default="{ row }">¥{{ row.price }}</template>
        </el-table-column>
        <el-table-column label="小计" width="100">
          <template #default="{ row }">¥{{ (row.price * row.quantity).toFixed(2) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Warning } from '@element-plus/icons-vue'
import { getOrders, getMyOrders, updateOrderStatus, cancelOrder, getOrderCancelInfo } from '@/api/order'
import { getUser, setUser } from '@/utils/auth'

const loading = ref(false)
const detailVisible = ref(false)
const searchKeyword = ref('')
const statusFilter = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const orders = ref([])
const currentOrder = ref(null)
const cancelInfoMap = ref({})
const user = ref(getUser())

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

const getMealTypeText = (mealType) => {
  const texts = {
    breakfast: '早餐',
    lunch: '午餐',
    dinner: '晚餐'
  }
  return texts[mealType] || mealType
}

const formatRemainingTime = (seconds) => {
  if (seconds <= 0) return '已超时'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  }
  return `${minutes}分钟`
}

const fetchOrders = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
      status: statusFilter.value
    }
    const res = user.value?.role === 'admin' ? await getOrders(params) : await getMyOrders(params)
    orders.value = res.data?.items || res.data || []
    total.value = res.data?.total || orders.value.length
    
    for (const order of orders.value) {
      if (order.status === 'pending' || order.status === 'confirmed') {
        loadCancelInfo(order.id)
      }
    }
  } catch (e) {
    console.log('Get orders error')
  } finally {
    loading.value = false
  }
}

const loadCancelInfo = async (orderId) => {
  try {
    const res = await getOrderCancelInfo(orderId)
    cancelInfoMap.value[orderId] = res.data
  } catch (e) {
    console.log('Get cancel info error')
  }
}

const canCancelOrder = (order) => {
  const info = cancelInfoMap.value[order.id]
  return info?.can_cancel || false
}

const getCancelReason = (order) => {
  const info = cancelInfoMap.value[order.id]
  return info?.cancel_reason || ''
}

const handleView = (row) => {
  currentOrder.value = row
  detailVisible.value = true
}

const updateStatus = async (row, status) => {
  ElMessageBox.confirm(`确定要将订单状态更新为"${getStatusText(status)}"吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
      try {
        await updateOrderStatus(row.id, status)
        ElMessage.success('状态更新成功')
        fetchOrders()
      } catch (e) {
        console.log('Update status error')
      }
    }).catch(() => {})
}

const handleCancel = async (row) => {
  const info = cancelInfoMap.value[row.id]
  const refundAmount = info?.refund_amount || row.total_price
  const remainingTime = info?.remaining_time || 0
  
  let content = `确定要取消该订单吗？`
  if (remainingTime > 0) {
    content += `<br/><br/>
      <div style="padding: 12px; background: #f0f9eb; border-radius: 4px; margin-top: 10px;">
        <div style="color: #67c23a; font-weight: 500; margin-bottom: 8px;"><el-icon><Warning /></el-icon> 取消须知</div>
        <div style="font-size: 13px; color: #606266; line-height: 1.8;">
          • 距离取消截止还剩：<strong style="color: #e6a23c;">${formatRemainingTime(remainingTime)}</strong><br/>
          • 取消截止时间：${info?.cancel_deadline || '-'}<br/>
          • 将返还餐补金额：<strong style="color: #67c23a;">¥${refundAmount.toFixed(2)}</strong><br/>
          • 对应菜品库存将自动恢复
        </div>
      </div>`
  }
  
  ElMessageBox.confirm(content, '取消订单', {
    confirmButtonText: '确认取消',
    cancelButtonText: '再想想',
    type: 'warning',
    dangerouslyUseHTMLString: true
  }).then(async () => {
      try {
        const res = await cancelOrder(row.id)
        if (res.data && res.data.user) {
          user.value = res.data.user
          setUser(res.data.user)
        }
        ElMessage.success(res.message || '订单取消成功')
        fetchOrders()
      } catch (e) {
        console.log('Cancel order error')
      }
    }).catch(() => {})
}

onMounted(() => {
  fetchOrders()
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
</style>
