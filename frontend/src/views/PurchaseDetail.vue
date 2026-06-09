<template>
  <div class="purchase-detail">
    <el-card shadow="hover" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>补货单详情</span>
          <div class="header-actions">
            <el-button :icon="Refresh" @click="fetchData">刷新</el-button>
            <el-button @click="goBack">返回</el-button>
          </div>
        </div>
      </template>

      <template v-if="purchase">
        <el-descriptions :column="3" border class="detail-desc">
          <el-descriptions-item label="补货单号">
            <span class="text-bold">{{ purchase.purchase_no }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="日期">{{ purchase.date }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(purchase.status)" size="small">
              {{ getStatusText(purchase.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="关联预订">
            <span v-if="purchase.booking">{{ purchase.booking.booking_no }}</span>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="总金额">
            <span class="text-danger">¥{{ purchase.total_price.toFixed(2) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="备注">{{ purchase.remark }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(purchase.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="更新时间">
            {{ formatTime(purchase.updated_at) }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="action-bar" v-if="purchase.status === 'draft'">
          <el-button type="primary" :icon="Check" @click="handleApprove">审核通过</el-button>
          <el-button type="danger" :icon="Delete" @click="handleDelete">删除补货单</el-button>
        </div>

        <el-divider>补货明细</el-divider>

        <el-table :data="purchase.items" style="width: 100%" stripe>
          <el-table-column label="序号" width="60" type="index" />
          <el-table-column label="食材名称" min-width="120">
            <template #default="{ row }">{{ row.ingredient?.name || '-' }}</template>
          </el-table-column>
          <el-table-column label="分类" width="100">
            <template #default="{ row }">{{ row.ingredient?.category || '-' }}</template>
          </el-table-column>
          <el-table-column label="单位" width="80">
            <template #default="{ row }">{{ row.ingredient?.unit || '-' }}</template>
          </el-table-column>
          <el-table-column label="需求数量" width="100" align="right">
            <template #default="{ row }">{{ row.required_qty.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="库存数量" width="100" align="right">
            <template #default="{ row }">{{ row.stock_qty.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="采购数量" width="100" align="right">
            <template #default="{ row }">
              <span :class="{ 'text-danger': row.purchase_qty > 0 }">
                {{ row.purchase_qty.toFixed(2) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="单价" width="100" align="right">
            <template #default="{ row }">¥{{ row.unit_price.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="小计" width="120" align="right">
            <template #default="{ row }">
              <span :class="{ 'text-danger': row.subtotal > 0 }">
                ¥{{ row.subtotal.toFixed(2) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="库区" width="100">
            <template #default="{ row }">
              {{ getZoneName(row.ingredient?.warehouse_zone) }}
            </template>
          </el-table-column>
          <el-table-column label="供应商" min-width="120">
            <template #default="{ row }">{{ row.ingredient?.supplier || '-' }}</template>
          </el-table-column>
          <el-table-column label="备注" min-width="150">
            <template #default="{ row }">{{ row.remark || '-' }}</template>
          </el-table-column>
        </el-table>

        <div class="total-bar">
          <span class="total-label">合计金额：</span>
          <span class="total-value">¥{{ purchase.total_price.toFixed(2) }}</span>
        </div>
      </template>

      <el-empty v-else description="补货单不存在" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Check } from '@element-plus/icons-vue'
import { getPurchaseList, updatePurchaseStatus } from '@/api/inventory'

const route = useRoute()
const router = useRouter()

const purchase = ref(null)
const loading = ref(false)

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getStatusText = (status) => {
  const texts = {
    draft: '待审核',
    approved: '已审核',
    partial: '部分入库',
    completed: '已完成'
  }
  return texts[status] || status
}

const getStatusType = (status) => {
  const types = {
    draft: 'warning',
    approved: 'primary',
    partial: 'info',
    completed: 'success'
  }
  return types[status] || 'info'
}

const getZoneName = (zone) => {
  const names = { dry: '干货区', refrigerated: '冷藏区', frozen: '冷冻区' }
  return names[zone] || zone
}

const fetchData = async () => {
  const id = route.params.id
  if (!id) return

  loading.value = true
  try {
    const res = await getPurchaseList(id)
    if (res.success) {
      purchase.value = res.data
    } else {
      ElMessage.error(res.message || '获取补货单失败')
    }
  } catch (e) {
    ElMessage.error('获取补货单失败')
  } finally {
    loading.value = false
  }
}

const handleApprove = async () => {
  ElMessageBox.confirm('确定审核通过该补货单吗？', '审核确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await updatePurchaseStatus(purchase.value.id, { status: 'approved' })
      if (res.success) {
        ElMessage.success('审核通过')
        fetchData()
      } else {
        ElMessage.error(res.message || '审核失败')
      }
    } catch (e) {
      ElMessage.error('审核失败')
    }
  }).catch(() => {})
}

const handleDelete = async () => {
  ElMessageBox.confirm('确定删除该补货单吗？', '删除确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const { deletePurchaseList } = await import('@/api/inventory')
      const res = await deletePurchaseList(purchase.value.id)
      if (res.success) {
        ElMessage.success('删除成功')
        router.push('/stock-inbound')
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (e) {
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const goBack = () => {
  router.back()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.purchase-detail {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.detail-desc {
  margin-bottom: 20px;
}

.action-bar {
  margin-top: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
  display: flex;
  gap: 12px;
}

.total-bar {
  margin-top: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
  text-align: right;
}

.total-label {
  font-size: 16px;
  color: #606266;
}

.total-value {
  font-size: 24px;
  font-weight: 600;
  color: #f56c6c;
  margin-left: 12px;
}

.text-bold {
  font-weight: 600;
  color: #303133;
}

.text-danger {
  color: #f56c6c;
  font-weight: 500;
}
</style>
