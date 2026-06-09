<template>
  <div class="inventory-alerts">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>库存预警管理</span>
          <div class="header-actions">
            <el-radio-group v-model="statusFilter" size="small" @change="fetchData">
              <el-radio-button value="">全部</el-radio-button>
              <el-radio-button value="pending">待处理</el-radio-button>
              <el-radio-button value="handled">已处理</el-radio-button>
              <el-radio-button value="ignored">已忽略</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <el-alert
        v-if="pendingCount > 0"
        :title="`当前有 ${pendingCount} 条待处理预警`"
        type="warning"
        show-icon
        :closable="false"
        style="margin-bottom: 20px;"
      />

      <el-table :data="alerts" style="width: 100%" stripe>
        <el-table-column label="预警时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="食材名称" min-width="120">
          <template #default="{ row }">{{ row.ingredient?.name || '-' }}</template>
        </el-table-column>
        <el-table-column label="分类" width="100">
          <template #default="{ row }">{{ row.ingredient?.category || '-' }}</template>
        </el-table-column>
        <el-table-column label="预警级别" width="100">
          <template #default="{ row }">
            <el-tag :type="row.alert_level === 'critical' ? 'danger' : 'warning'" size="small">
              {{ row.alert_level === 'critical' ? '严重' : '预警' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前库存" width="120">
          <template #default="{ row }">
            <span class="text-danger">{{ row.current_stock }} {{ row.ingredient?.unit || '' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="安全库存" width="100">
          <template #default="{ row }">{{ row.safety_stock }} {{ row.ingredient?.unit || '' }}</template>
        </el-table-column>
        <el-table-column label="缺口数量" width="100">
          <template #default="{ row }">
            <span class="text-danger">{{ row.shortage_qty }}</span>
          </template>
        </el-table-column>
        <el-table-column label="库区" width="100">
          <template #default="{ row }">{{ getZoneName(row.ingredient?.warehouse_zone) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处理人" width="100">
          <template #default="{ row }">{{ row.handled_by_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="处理时间" width="160">
          <template #default="{ row }">{{ row.handled_at ? formatTime(row.handled_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button type="success" size="small" link @click="handleAlert(row, 'handled')">已处理</el-button>
              <el-button type="info" size="small" link @click="handleAlert(row, 'ignored')">忽略</el-button>
            </template>
            <template v-else>
              <span class="text-muted">已处理</span>
            </template>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="alerts.length === 0" description="暂无预警数据" />
    </el-card>

    <el-dialog v-model="handleDialogVisible" title="处理预警" width="400px">
      <el-form :model="handleForm" label-width="80px">
        <el-form-item label="处理结果">
          <el-radio-group v-model="handleForm.status">
            <el-radio value="handled">已采购</el-radio>
            <el-radio value="ignored">忽略</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="handleForm.remark" type="textarea" :rows="3" placeholder="请输入处理备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitHandle">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { getStockAlerts, handleStockAlert } from '@/api/inventory'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUser } from '@/utils/auth'

const statusFilter = ref('pending')
const alerts = ref([])
const handleDialogVisible = ref(false)
const currentAlert = ref(null)
const user = getUser()

const handleForm = reactive({
  status: 'handled',
  remark: ''
})

const pendingCount = computed(() => {
  return alerts.value.filter(a => a.status === 'pending').length
})

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

const getZoneName = (zone) => {
  const names = { dry: '干货区', refrigerated: '冷藏区', frozen: '冷冻区' }
  return names[zone] || zone
}

const getStatusType = (status) => {
  const types = { pending: 'warning', handled: 'success', ignored: 'info' }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = { pending: '待处理', handled: '已处理', ignored: '已忽略' }
  return texts[status] || status
}

const fetchData = async () => {
  try {
    const params = {}
    if (statusFilter.value) params.status = statusFilter.value
    const res = await getStockAlerts(params)
    if (res.success) {
      alerts.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取预警数据失败')
  }
}

const handleAlert = (row, status) => {
  currentAlert.value = row
  handleForm.status = status
  handleForm.remark = ''
  handleDialogVisible.value = true
}

const submitHandle = async () => {
  if (!currentAlert.value) return
  try {
    const res = await handleStockAlert(currentAlert.value.id, {
      status: handleForm.status,
      handle_remark: handleForm.remark,
      handled_by: user?.id,
      handled_by_name: user?.name || user?.username
    })
    if (res.success) {
      ElMessage.success('处理成功')
      handleDialogVisible.value = false
      fetchData()
    }
  } catch (e) {
    ElMessage.error('处理失败')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.text-danger {
  color: #f56c6c;
  font-weight: 500;
}

.text-muted {
  color: #909399;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}
</style>
