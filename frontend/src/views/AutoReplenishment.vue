<template>
  <div class="auto-replenishment">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>自动补货管理</span>
              <div class="header-actions">
                <el-button type="primary" :icon="Refresh" @click="fetchData">刷新</el-button>
              </div>
            </div>
          </template>

          <el-row :gutter="16" class="stats-row">
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-icon primary">
                  <el-icon :size="24"><ShoppingCart /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ stats.totalCount }}</div>
                  <div class="stat-label">补货单总数</div>
                </div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-icon success">
                  <el-icon :size="24"><CircleCheck /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ stats.todayCount }}</div>
                  <div class="stat-label">今日补货</div>
                </div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-icon warning">
                  <el-icon :size="24"><Warning /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ stats.totalShortageCount }}</div>
                  <div class="stat-label">累计缺货项</div>
                </div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-icon danger">
                  <el-icon :size="24"><TrendCharts /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">¥{{ stats.totalAmount.toFixed(2) }}</div>
                  <div class="stat-label">补货总金额</div>
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
              <span>手动触发补货</span>
            </div>
          </template>

          <el-form :model="replenishForm" label-width="100px" class="replenish-form">
            <el-form-item label="选择预订">
              <el-select
                v-model="replenishForm.booking_id"
                placeholder="请选择预报餐预订"
                clearable
                style="width: 100%;"
                @change="handleBookingChange"
              >
                <el-option
                  v-for="booking in bookingList"
                  :key="booking.id"
                  :label="`${booking.date} ${getMealTypeText(booking.meal_type)} - ${booking.people_num}人`"
                  :value="booking.id"
                />
              </el-select>
            </el-form-item>

            <template v-if="!replenishForm.booking_id">
              <el-form-item label="用餐日期">
                <el-date-picker
                  v-model="replenishForm.date"
                  type="date"
                  placeholder="选择日期"
                  value-format="YYYY-MM-DD"
                  style="width: 100%;"
                />
              </el-form-item>

              <el-form-item label="用餐类型">
                <el-select v-model="replenishForm.meal_type" placeholder="请选择" style="width: 100%;">
                  <el-option label="早餐" value="breakfast" />
                  <el-option label="午餐" value="lunch" />
                  <el-option label="晚餐" value="dinner" />
                </el-select>
              </el-form-item>

              <el-form-item label="就餐人数">
                <el-input-number
                  v-model="replenishForm.people_num"
                  :min="1"
                  :max="1000"
                  style="width: 100%;"
                />
              </el-form-item>

              <el-form-item label="菜品清单">
                <el-select
                  v-model="selectedDishIds"
                  multiple
                  filterable
                  placeholder="请选择菜品"
                  style="width: 100%;"
                >
                  <el-option
                    v-for="dish in dishList"
                    :key="dish.id"
                    :label="dish.name"
                    :value="dish.id"
                  />
                </el-select>
              </el-form-item>
            </template>

            <el-form-item label="损耗率">
              <el-slider
                v-model="replenishForm.wastage_rate"
                :min="0"
                :max="0.3"
                :step="0.05"
                :format-tooltip="val => `${(val * 100).toFixed(0)}%`"
                show-input
                style="width: 80%;"
              />
            </el-form-item>

            <el-form-item label="自动审核">
              <el-switch v-model="replenishForm.auto_approve" />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :icon="Search" @click="analyzeDemand">分析需求</el-button>
              <el-button type="success" :icon="ShoppingCart" :disabled="!analysisResult" @click="executeReplenish">
                执行补货
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="hover" v-loading="analyzing">
          <template #header>
            <div class="card-header">
              <span>库区需求分析</span>
            </div>
          </template>

          <template v-if="analysisResult">
            <div class="analysis-summary">
              <el-descriptions :column="3" size="small" border>
                <el-descriptions-item label="日期">
                  {{ analysisResult.date }}
                </el-descriptions-item>
                <el-descriptions-item label="餐次">
                  {{ getMealTypeText(analysisResult.meal_type) }}
                </el-descriptions-item>
                <el-descriptions-item label="人数">
                  {{ analysisResult.people_num }}人
                </el-descriptions-item>
                <el-descriptions-item label="缺货项数">
                  <span class="text-danger">{{ analysisResult.shortage_count }}</span>
                </el-descriptions-item>
                <el-descriptions-item label="总缺口">
                  <span class="text-danger">{{ analysisResult.total_shortage.toFixed(2) }}</span>
                </el-descriptions-item>
              </el-descriptions>
            </div>

            <el-collapse v-model="activeZoneNames" class="zone-collapse">
              <el-collapse-item
                v-for="zone in analysisResult.zone_demands"
                :key="zone.zone"
                :name="zone.zone"
              >
                <template #title>
                  <div class="zone-title">
                    <span>{{ zone.zone_name }}</span>
                    <el-tag size="small" type="info">
                      需求: {{ zone.total_demand.toFixed(2) }} / 库存: {{ zone.total_stock.toFixed(2) }}
                    </el-tag>
                    <el-tag v-if="zone.total_shortage > 0" size="small" type="danger">
                      缺口: {{ zone.total_shortage.toFixed(2) }}
                    </el-tag>
                  </div>
                </template>
                <el-table :data="zone.ingredients" size="small">
                  <el-table-column prop="name" label="食材" min-width="100" />
                  <el-table-column prop="required_qty" label="需求" width="80">
                    <template #default="{ row }">{{ row.required_qty.toFixed(2) }} {{ row.unit }}</template>
                  </el-table-column>
                  <el-table-column prop="stock_qty" label="库存" width="80">
                    <template #default="{ row }">{{ row.stock_qty.toFixed(2) }} {{ row.unit }}</template>
                  </el-table-column>
                  <el-table-column prop="shortage_qty" label="缺口" width="80">
                    <template #default="{ row }">
                      <span :class="{ 'text-danger': row.shortage_qty > 0 }">
                        {{ row.shortage_qty.toFixed(2) }} {{ row.unit }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column label="安全库存" width="80">
                    <template #default="{ row }">
                      <span :class="{ 'text-warning': row.below_safety }">
                        {{ row.safety_stock }} {{ row.unit }}
                      </span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="supplier" label="供应商" min-width="100" />
                </el-table>
              </el-collapse-item>
            </el-collapse>
          </template>

          <el-empty v-else description="请先分析需求" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>自动补货记录</span>
              <div class="header-actions">
                <el-date-picker
                  v-model="dateFilter"
                  type="date"
                  placeholder="选择日期"
                  value-format="YYYY-MM-DD"
                  size="small"
                  @change="fetchRecords"
                />
              </div>
            </div>
          </template>

          <el-table :data="records" style="width: 100%" stripe v-loading="recordsLoading">
            <el-table-column label="日期" width="120">
              <template #default="{ row }">{{ row.date }}</template>
            </el-table-column>
            <el-table-column label="餐次" width="80">
              <template #default="{ row }">{{ getMealTypeText(row.meal_type) }}</template>
            </el-table-column>
            <el-table-column label="人数" width="80">
              <template #default="{ row }">{{ row.people_num }}人</template>
            </el-table-column>
            <el-table-column label="缺货项数" width="100">
              <template #default="{ row }">
                <span class="text-danger">{{ row.shortage_count }}</span>
              </template>
            </el-table-column>
            <el-table-column label="总缺口" width="120">
              <template #default="{ row }">{{ row.total_shortage.toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="关联补货单" width="160">
              <template #default="{ row }">
                <el-link v-if="row.purchase_list" type="primary" @click="goToPurchase(row.purchase_list_id)">
                  {{ row.purchase_list.purchase_no }}
                </el-link>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="补货单状态" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.purchase_list" :type="getPurchaseStatusType(row.purchase_list.status)" size="small">
                  {{ getPurchaseStatusText(row.purchase_list.status) }}
                </el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="备注" min-width="150">
              <template #default="{ row }">{{ row.remark }}</template>
            </el-table-column>
            <el-table-column label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>

          <el-empty v-if="records.length === 0" description="暂无补货记录" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh,
  Search,
  ShoppingCart,
  CircleCheck,
  Warning,
  TrendCharts
} from '@element-plus/icons-vue'
import {
  analyzeZoneInventoryDemand,
  autoReplenish,
  getAutoReplenishmentRecords,
  getPurchaseLists
} from '@/api/inventory'
import { getBookings } from '@/api/booking'
import { getDishes } from '@/api/dish'

const router = useRouter()

const replenishForm = reactive({
  booking_id: null,
  date: '',
  meal_type: '',
  people_num: 1,
  dish_ids: '',
  wastage_rate: 0.1,
  auto_approve: false
})

const selectedDishIds = ref([])
const bookingList = ref([])
const dishList = ref([])
const analysisResult = ref(null)
const analyzing = ref(false)
const activeZoneNames = ref([])
const records = ref([])
const recordsLoading = ref(false)
const dateFilter = ref('')

const stats = computed(() => {
  let totalCount = records.value.length
  let todayCount = 0
  let totalShortageCount = 0
  let totalAmount = 0

  const today = new Date().toDateString()

  records.value.forEach(r => {
    if (new Date(r.created_at).toDateString() === today) {
      todayCount++
    }
    totalShortageCount += r.shortage_count
    if (r.purchase_list) {
      totalAmount += r.purchase_list.total_price
    }
  })

  return {
    totalCount,
    todayCount,
    totalShortageCount,
    totalAmount
  }
})

const getMealTypeText = (type) => {
  const texts = { breakfast: '早餐', lunch: '午餐', dinner: '晚餐' }
  return texts[type] || type
}

const getPurchaseStatusText = (status) => {
  const texts = { draft: '待审核', approved: '已审核', partial: '部分入库', completed: '已完成' }
  return texts[status] || status
}

const getPurchaseStatusType = (status) => {
  const types = { draft: 'warning', approved: 'primary', partial: 'info', completed: 'success' }
  return types[status] || 'info'
}

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

const handleBookingChange = (bookingId) => {
  if (bookingId) {
    const booking = bookingList.value.find(b => b.id === bookingId)
    if (booking) {
      replenishForm.date = booking.date
      replenishForm.meal_type = booking.meal_type
      replenishForm.people_num = booking.people_num
      selectedDishIds.value = booking.dish_ids ? booking.dish_ids.split(',').map(Number) : []
    }
  } else {
    replenishForm.date = ''
    replenishForm.meal_type = ''
    replenishForm.people_num = 1
    selectedDishIds.value = []
  }
}

const analyzeDemand = async () => {
  const data = {}

  if (replenishForm.booking_id) {
    data.booking_id = replenishForm.booking_id
  } else {
    if (!replenishForm.date || !replenishForm.meal_type || selectedDishIds.value.length === 0) {
      ElMessage.warning('请填写完整的补货信息')
      return
    }
    data.date = replenishForm.date
    data.meal_type = replenishForm.meal_type
    data.people_num = replenishForm.people_num
    data.dish_ids = selectedDishIds.value.join(',')
  }

  analyzing.value = true
  try {
    const res = await analyzeZoneInventoryDemand(data)
    if (res.success) {
      analysisResult.value = res.data
      activeZoneNames.value = res.data.zone_demands.map(z => z.zone)
    }
  } catch (e) {
    ElMessage.error('分析需求失败')
  } finally {
    analyzing.value = false
  }
}

const executeReplenish = async () => {
  if (!analysisResult.value) {
    ElMessage.warning('请先分析需求')
    return
  }

  ElMessageBox.confirm(
    `确定执行自动补货吗？\n将生成 ${analysisResult.value.shortage_count} 项缺货食材的补货单。`,
    '确认补货',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      const data = {
        wastage_rate: replenishForm.wastage_rate,
        auto_approve: replenishForm.auto_approve
      }

      if (replenishForm.booking_id) {
        data.booking_id = replenishForm.booking_id
      } else {
        data.date = replenishForm.date
        data.meal_type = replenishForm.meal_type
        data.people_num = replenishForm.people_num
        data.dish_ids = selectedDishIds.value.join(',')
      }

      const res = await autoReplenish(data)
      if (res.success) {
        ElMessage.success(res.message || '自动补货成功')
        if (res.data.purchase_list_id) {
          ElMessage.info(`补货单号: ${res.data.purchase_no}`)
        }
        analysisResult.value = null
        fetchRecords()
      } else {
        ElMessage.error(res.message || '自动补货失败')
      }
    } catch (e) {
      ElMessage.error('自动补货失败')
    }
  }).catch(() => {})
}

const fetchBookings = async () => {
  try {
    const res = await getBookings({ status: 'confirmed' })
    if (res.success) {
      bookingList.value = res.data
    }
  } catch (e) {
    console.error('获取预订列表失败', e)
  }
}

const fetchDishes = async () => {
  try {
    const res = await getDishes()
    if (res.success) {
      dishList.value = res.data
    }
  } catch (e) {
    console.error('获取菜品列表失败', e)
  }
}

const fetchRecords = async () => {
  recordsLoading.value = true
  try {
    const params = {}
    if (dateFilter.value) params.date = dateFilter.value
    const res = await getAutoReplenishmentRecords(params)
    if (res.success) {
      records.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取补货记录失败')
  } finally {
    recordsLoading.value = false
  }
}

const goToPurchase = (id) => {
  router.push(`/purchases/${id}`)
}

const fetchData = () => {
  fetchBookings()
  fetchDishes()
  fetchRecords()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.auto-replenishment {
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
  align-items: center;
}

.stats-row {
  margin-bottom: 0;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-right: 16px;
}

.stat-icon.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.success {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.stat-icon.warning {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.danger {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.replenish-form {
  margin-top: 12px;
}

.analysis-summary {
  margin-bottom: 16px;
}

.zone-collapse {
  margin-top: 16px;
}

.zone-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 500;
}

.text-danger {
  color: #f56c6c;
  font-weight: 500;
}

.text-warning {
  color: #e6a23c;
  font-weight: 500;
}
</style>
