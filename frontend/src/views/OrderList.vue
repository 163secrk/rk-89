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
        <el-table-column label="用户" width="120">
          <template #default="{ row }">{{ row.user?.name || '-' }}</template>
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
        <el-table-column prop="created_at" label="下单时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">查看</el-button>
            <el-button type="success" link size="small" v-if="row.status === 'pending'" @click="updateStatus(row, 'confirmed')">确认</el-button>
            <el-button type="success" link size="small" v-if="row.status === 'confirmed'" @click="updateStatus(row, 'completed')">完成</el-button>
            <el-button type="danger" link size="small" v-if="row.status === 'pending' || row.status === 'confirmed'" @click="updateStatus(row, 'cancelled')">取消</el-button>
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
import { Search, Refresh } from '@element-plus/icons-vue'
import { getOrders, getMyOrders, updateOrderStatus } from '@/api/order'
import { getUser } from '@/utils/auth'

const loading = ref(false)
const detailVisible = ref(false)
const searchKeyword = ref('')
const statusFilter = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const orders = ref([])
const currentOrder = ref(null)
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

const fetchOrders = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
      status: statusFilter.value
    }
    const res = user?.role === 'admin' ? await getOrders(params) : await getMyOrders(params)
    orders.value = res.data?.items || res.data || []
    total.value = res.data?.total || orders.value.length
  } catch (e) {
    console.log('Get orders error')
  } finally {
    loading.value = false
  }
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
