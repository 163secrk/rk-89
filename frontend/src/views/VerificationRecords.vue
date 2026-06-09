<template>
  <div class="verification-records">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">核销记录</span>
          <div class="header-actions">
            <el-date-picker
              v-model="searchDate"
              type="date"
              placeholder="选择日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              style="width: 180px; margin-right: 10px;"
              :prefix-icon="Calendar"
              @change="fetchRecords"
            />
            <el-select v-model="mealTypeFilter" placeholder="餐次" style="width: 120px; margin-right: 10px;" clearable @change="fetchRecords">
              <el-option label="早餐" value="breakfast" />
              <el-option label="午餐" value="lunch" />
              <el-option label="晚餐" value="dinner" />
            </el-select>
            <el-button type="primary" :icon="Refresh" @click="fetchRecords">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table :data="records" style="width: 100%" v-loading="loading">
        <el-table-column prop="order_no" label="取餐码" width="140" />
        <el-table-column label="用餐人" width="120">
          <template #default="{ row }">
            {{ row.user?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="餐次" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getMealTypeTag(row.meal_type)">
              {{ getMealTypeName(row.meal_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="meal_date" label="用餐日期" width="120" />
        <el-table-column label="菜品" show-overflow-tooltip>
          <template #default="{ row }">
            {{ (row.order?.items || []).map(i => i.dish?.name || '').join('、') }}
          </template>
        </el-table-column>
        <el-table-column label="核销人" width="100">
          <template #default="{ row }">
            {{ row.verifier_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="verified_at" label="核销时间" width="160" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'success' ? 'success' : 'danger'">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">详情</el-button>
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
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <el-dialog v-model="detailVisible" title="核销详情" width="600px">
      <el-descriptions :column="2" border v-if="currentRecord">
        <el-descriptions-item label="取餐码">{{ currentRecord.order_no }}</el-descriptions-item>
        <el-descriptions-item label="用餐人">{{ currentRecord.user?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="餐次">
          <el-tag size="small" :type="getMealTypeTag(currentRecord.meal_type)">
            {{ getMealTypeName(currentRecord.meal_type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="用餐日期">{{ currentRecord.meal_date }}</el-descriptions-item>
        <el-descriptions-item label="核销人">{{ currentRecord.verifier_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="核销时间">{{ currentRecord.verified_at }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag size="small" :type="currentRecord.status === 'success' ? 'success' : 'danger'">
            {{ currentRecord.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ currentRecord.remark || '无' }}</el-descriptions-item>
      </el-descriptions>
      <el-divider content-position="left">菜品明细</el-divider>
      <el-table :data="currentRecord?.order?.items || []" size="small">
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
import { ElMessage } from 'element-plus'
import { Refresh, Calendar } from '@element-plus/icons-vue'
import { getVerificationRecords } from '@/api/verification'

const loading = ref(false)
const detailVisible = ref(false)
const searchDate = ref(new Date().toISOString().split('T')[0])
const mealTypeFilter = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const records = ref([])
const currentRecord = ref(null)

const getMealTypeName = (type) => {
  const names = {
    breakfast: '早餐',
    lunch: '午餐',
    dinner: '晚餐'
  }
  return names[type] || type
}

const getMealTypeTag = (type) => {
  const types = {
    breakfast: 'warning',
    lunch: 'success',
    dinner: 'primary'
  }
  return types[type] || 'info'
}

const fetchRecords = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      mealDate: searchDate.value,
      mealType: mealTypeFilter.value
    }
    const res = await getVerificationRecords(params)
    records.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('获取核销记录失败')
  } finally {
    loading.value = false
  }
}

const handleView = (row) => {
  currentRecord.value = row
  detailVisible.value = true
}

const handleSizeChange = () => {
  page.value = 1
  fetchRecords()
}

const handleCurrentChange = () => {
  fetchRecords()
}

onMounted(() => {
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
</style>
