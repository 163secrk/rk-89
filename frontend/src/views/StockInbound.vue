<template>
  <div class="stock-inbound">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>采购入库管理</span>
        </div>
      </template>

      <div class="filter-bar">
        <el-radio-group v-model="statusFilter" size="small" @change="fetchPurchaseLists">
          <el-radio-button value="approved">待入库</el-radio-button>
          <el-radio-button value="partial">部分入库</el-radio-button>
          <el-radio-button value="completed">已完成</el-radio-button>
          <el-radio-button value="">全部</el-radio-button>
        </el-radio-group>
        <el-date-picker
          v-model="dateFilter"
          type="date"
          placeholder="选择日期"
          style="width: 180px;"
          @change="fetchPurchaseLists"
        />
      </div>

      <el-table :data="purchaseLists" style="width: 100%" stripe>
        <el-table-column prop="purchase_no" label="采购单号" width="180" />
        <el-table-column label="日期" width="120" prop="date" />
        <el-table-column label="关联预订" width="100">
          <template #default="{ row }">{{ row.booking?.booking_no || '-' }}</template>
        </el-table-column>
        <el-table-column label="总金额" width="120">
          <template #default="{ row }">¥{{ row.total_price.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="食材种类" width="100">
          <template #default="{ row }">{{ row.items?.length || 0 }} 种</template>
        </el-table-column>
        <el-table-column label="备注" prop="remark" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="viewDetail(row)">查看明细</el-button>
            <el-button
              v-if="row.status === 'approved' || row.status === 'partial'"
              type="success"
              size="small"
              link
              @click="handleInbound(row)"
            >
              确认入库
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="purchaseLists.length === 0" description="暂无采购单" />
    </el-card>

    <el-dialog v-model="detailVisible" title="采购单明细" width="800px">
      <div v-if="currentPurchase">
        <el-descriptions :column="3" border size="small" style="margin-bottom: 20px;">
          <el-descriptions-item label="采购单号">{{ currentPurchase.purchase_no }}</el-descriptions-item>
          <el-descriptions-item label="日期">{{ currentPurchase.date }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentPurchase.status)" size="small">
              {{ getStatusText(currentPurchase.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="关联预订">{{ currentPurchase.booking?.booking_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="总金额">¥{{ currentPurchase.total_price.toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="备注">{{ currentPurchase.remark || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-table :data="currentPurchase.items || []" style="width: 100%" size="small" border>
          <el-table-column label="食材名称" min-width="120">
            <template #default="{ row }">{{ row.ingredient?.name || '-' }}</template>
          </el-table-column>
          <el-table-column label="分类" width="100">
            <template #default="{ row }">{{ row.ingredient?.category || '-' }}</template>
          </el-table-column>
          <el-table-column label="需求数量" width="100">
            <template #default="{ row }">{{ row.required_qty }} {{ row.ingredient?.unit || '' }}</template>
          </el-table-column>
          <el-table-column label="采购数量" width="100">
            <template #default="{ row }">
              <span class="text-primary">{{ row.purchase_qty }} {{ row.ingredient?.unit || '' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="已入库" width="100">
            <template #default="{ row }">{{ row.stock_qty }} {{ row.ingredient?.unit || '' }}</template>
          </el-table-column>
          <el-table-column label="单价" width="100">
            <template #default="{ row }">¥{{ row.unit_price.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="小计" width="100">
            <template #default="{ row }">¥{{ row.subtotal.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="库区" width="100">
            <template #default="{ row }">{{ getZoneName(row.ingredient?.warehouse_zone) }}</template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button
          v-if="currentPurchase?.status === 'approved' || currentPurchase?.status === 'partial'"
          type="primary"
          @click="handleInbound(currentPurchase)"
        >
          确认入库
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="inboundVisible" title="确认入库" width="500px">
      <div v-if="currentPurchase">
        <el-alert
          title="入库确认"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 20px;"
        >
          <p>采购单号: {{ currentPurchase.purchase_no }}</p>
          <p>食材种类: {{ currentPurchase.items?.length || 0 }} 种</p>
          <p>入库后将自动增加对应食材的库存数量</p>
        </el-alert>

        <el-form label-width="80px">
          <el-form-item label="备注">
            <el-input
              v-model="inboundRemark"
              type="textarea"
              :rows="3"
              placeholder="请输入入库备注（可选）"
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="inboundVisible = false">取消</el-button>
        <el-button type="primary" @click="submitInbound" :loading="submitting">确认入库</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getPurchaseLists, stockInbound, updatePurchaseStatus } from '@/api/inventory'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUser } from '@/utils/auth'

const statusFilter = ref('approved')
const dateFilter = ref('')
const purchaseLists = ref([])
const detailVisible = ref(false)
const inboundVisible = ref(false)
const currentPurchase = ref(null)
const inboundRemark = ref('')
const submitting = ref(false)
const user = getUser()

const getStatusType = (status) => {
  const types = {
    draft: 'info',
    approved: 'warning',
    partial: 'warning',
    completed: 'success',
    cancelled: 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    draft: '草稿',
    approved: '待入库',
    partial: '部分入库',
    completed: '已完成',
    cancelled: '已取消'
  }
  return texts[status] || status
}

const getZoneName = (zone) => {
  const names = { dry: '干货区', refrigerated: '冷藏区', frozen: '冷冻区' }
  return names[zone] || zone
}

const fetchPurchaseLists = async () => {
  try {
    const params = {}
    if (statusFilter.value) params.status = statusFilter.value
    if (dateFilter.value) params.date = dateFilter.value
    const res = await getPurchaseLists(params)
    if (res.success) {
      purchaseLists.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取采购单列表失败')
  }
}

const viewDetail = (row) => {
  currentPurchase.value = row
  detailVisible.value = true
}

const handleInbound = async (row) => {
  if (row.status !== 'approved' && row.status !== 'partial') {
    ElMessage.warning('该采购单状态不允许入库')
    return
  }
  currentPurchase.value = row
  inboundRemark.value = ''
  inboundVisible.value = true
}

const submitInbound = async () => {
  if (!currentPurchase.value) return

  try {
    submitting.value = true
    const res = await stockInbound({
      purchase_list_id: currentPurchase.value.id,
      operator_id: user?.id,
      operator_name: user?.name || user?.username,
      remark: inboundRemark.value
    })
    if (res.success) {
      ElMessage.success('入库成功')
      inboundVisible.value = false
      detailVisible.value = false
      fetchPurchaseLists()
    }
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '入库失败')
  } finally {
    submitting.value = false
  }
}

const approvePurchase = async (row) => {
  try {
    await ElMessageBox.confirm('确定要审批该采购单吗？', '确认审批', {
      type: 'warning'
    })
    const res = await updatePurchaseStatus(row.id, { status: 'approved' })
    if (res.success) {
      ElMessage.success('审批成功')
      fetchPurchaseLists()
    }
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('审批失败')
    }
  }
}

onMounted(() => {
  fetchPurchaseLists()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.filter-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.text-primary {
  color: #409eff;
  font-weight: 500;
}
</style>
