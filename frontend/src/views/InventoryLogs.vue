<template>
  <div class="inventory-logs">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>库存操作日志</span>
          <div class="header-actions">
            <el-select v-model="moduleFilter" placeholder="模块" clearable style="width: 120px; margin-right: 12px;" @change="fetchData">
              <el-option label="库存管理" value="inventory" />
              <el-option label="采购管理" value="purchase" />
            </el-select>
            <el-select v-model="operationFilter" placeholder="操作类型" clearable style="width: 140px; margin-right: 12px;" @change="fetchData">
              <el-option label="入库" value="stock_inbound" />
              <el-option label="出库" value="stock_outbound" />
              <el-option label="库区调拨" value="zone_transfer" />
              <el-option label="处理预警" value="handle_alert" />
              <el-option label="审批采购" value="approve_purchase" />
            </el-select>
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              style="margin-right: 12px;"
              @change="fetchData"
            />
            <el-button type="primary" @click="fetchData">
              <el-icon><Refresh /></el-icon>刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="logs" style="width: 100%" stripe>
        <el-table-column label="操作时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="模块" width="100">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ getModuleText(row.module) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-tag :type="getOperationType(row.operation)" size="small">
              {{ getOperationText(row.operation) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作内容" min-width="300">
          <template #default="{ row }">
            <span v-if="row.content">{{ formatContent(row.content) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作人" width="120" prop="operator_name" />
        <el-table-column label="IP地址" width="140" prop="ip_address" />
      </el-table>
      <el-empty v-if="logs.length === 0" description="暂无操作日志" />
    </el-card>

    <el-card shadow="hover" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>库存变动记录</span>
          <div class="header-actions">
            <el-radio-group v-model="changeTypeFilter" size="small" @change="fetchRecords">
              <el-radio-button value="">全部</el-radio-button>
              <el-radio-button value="in">入库</el-radio-button>
              <el-radio-button value="out">出库</el-radio-button>
              <el-radio-button value="transfer">调拨</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <el-table :data="records" style="width: 100%" stripe>
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="食材" min-width="120">
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
              {{ row.change_type === 'in' ? '+' : row.change_type === 'out' ? '-' : '' }}{{ row.change_qty }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="变动前" width="100">
          <template #default="{ row }">{{ row.stock_before }}</template>
        </el-table-column>
        <el-table-column label="变动后" width="100">
          <template #default="{ row }">{{ row.stock_after }}</template>
        </el-table-column>
        <el-table-column label="库区" width="100">
          <template #default="{ row }">{{ getZoneName(row.warehouse_zone) }}</template>
        </el-table-column>
        <el-table-column label="关联单据" width="140" prop="related_no" />
        <el-table-column label="操作人" width="100" prop="operator_name" />
        <el-table-column label="备注" min-width="120" prop="remark" />
      </el-table>
      <el-empty v-if="records.length === 0" description="暂无变动记录" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { getOperationLogs, getStockRecords } from '@/api/inventory'
import { ElMessage } from 'element-plus'

const moduleFilter = ref('')
const operationFilter = ref('')
const changeTypeFilter = ref('')
const dateRange = ref([])
const logs = ref([])
const records = ref([])

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const getModuleText = (module) => {
  const texts = { inventory: '库存管理', purchase: '采购管理' }
  return texts[module] || module
}

const getOperationType = (op) => {
  const types = {
    stock_inbound: 'success',
    stock_outbound: 'danger',
    zone_transfer: 'warning',
    handle_alert: 'info',
    approve_purchase: 'primary'
  }
  return types[op] || 'info'
}

const getOperationText = (op) => {
  const texts = {
    stock_inbound: '入库',
    stock_outbound: '出库',
    zone_transfer: '库区调拨',
    handle_alert: '处理预警',
    approve_purchase: '审批采购'
  }
  return texts[op] || op
}

const getChangeTypeText = (type) => {
  const texts = { in: '入库', out: '出库', transfer: '调拨' }
  return texts[type] || type
}

const getZoneName = (zone) => {
  const names = { dry: '干货区', refrigerated: '冷藏区', frozen: '冷冻区' }
  return names[zone] || zone
}

const formatContent = (content) => {
  try {
    const obj = JSON.parse(content)
    return Object.entries(obj).map(([k, v]) => `${k}: ${v}`).join(', ')
  } catch {
    return content
  }
}

const fetchData = async () => {
  try {
    const params = {}
    if (moduleFilter.value) params.module = moduleFilter.value
    if (operationFilter.value) params.operation = operationFilter.value
    if (dateRange.value?.length === 2) {
      params.startDate = dateRange.value[0]
      params.endDate = dateRange.value[1]
    }
    const res = await getOperationLogs(params)
    if (res.success) {
      logs.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取操作日志失败')
  }
}

const fetchRecords = async () => {
  try {
    const params = {}
    if (changeTypeFilter.value) params.changeType = changeTypeFilter.value
    const res = await getStockRecords(params)
    if (res.success) {
      records.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取变动记录失败')
  }
}

onMounted(() => {
  fetchData()
  fetchRecords()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.text-success {
  color: #67c23a;
  font-weight: 500;
}

.text-danger {
  color: #f56c6c;
  font-weight: 500;
}

.text-muted {
  color: #909399;
}
</style>
