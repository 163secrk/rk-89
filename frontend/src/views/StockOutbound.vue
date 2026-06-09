<template>
  <div class="stock-outbound">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>备菜出库管理</span>
        </div>
      </template>

      <div class="filter-bar">
        <el-date-picker
          v-model="dateFilter"
          type="date"
          placeholder="选择日期"
          style="width: 180px; margin-right: 16px;"
          @change="fetchMealPlans"
        />
        <el-select v-model="mealTypeFilter" placeholder="餐次" clearable style="width: 120px; margin-right: 16px;" @change="fetchMealPlans">
          <el-option label="早餐" value="breakfast" />
          <el-option label="午餐" value="lunch" />
          <el-option label="晚餐" value="dinner" />
        </el-select>
      </div>

      <el-table :data="mealPlans" style="width: 100%" stripe>
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column label="餐次" width="100">
          <template #default="{ row }">
            <el-tag type="success" size="small">{{ getMealTypeText(row.meal_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="菜品" min-width="200">
          <template #default="{ row }">
            {{ (row.dishes || []).map(d => d.name).join('、') }}
          </template>
        </el-table-column>
        <el-table-column label="就餐人数" width="120">
          <template #default="{ row }">
            <el-input-number
              v-model="row.peopleNum"
              :min="1"
              :max="999"
              size="small"
              @change="() => {}"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="viewDemand(row)">查看需求</el-button>
            <el-button type="danger" size="small" link @click="handleOutbound(row)">确认出库</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="mealPlans.length === 0" description="暂无配餐计划" />
    </el-card>

    <el-dialog v-model="demandVisible" title="食材需求明细" width="900px">
      <div v-if="currentMealPlan && demandData">
        <el-descriptions :column="3" border size="small" style="margin-bottom: 20px;">
          <el-descriptions-item label="日期">{{ currentMealPlan.date }}</el-descriptions-item>
          <el-descriptions-item label="餐次">{{ getMealTypeText(currentMealPlan.meal_type) }}</el-descriptions-item>
          <el-descriptions-item label="就餐人数">{{ currentMealPlan.peopleNum }} 人</el-descriptions-item>
        </el-descriptions>

        <el-alert
          v-if="demandData.warnings?.length > 0"
          title="库存预警"
          type="warning"
          show-icon
          style="margin-bottom: 20px;"
        >
          <ul>
            <li v-for="(warning, idx) in demandData.warnings" :key="idx">{{ warning }}</li>
          </ul>
        </el-alert>

        <el-table :data="demandData.ingredients || []" style="width: 100%" size="small" border>
          <el-table-column label="食材名称" min-width="120" prop="ingredient_name" />
          <el-table-column label="分类" width="100" prop="category" />
          <el-table-column label="单位" width="80" prop="unit" />
          <el-table-column label="需求数量" width="120">
            <template #default="{ row }">
              <span class="text-danger">{{ row.required_qty }}</span>
            </template>
          </el-table-column>
          <el-table-column label="当前库存" width="120">
            <template #default="{ row }">
              <span :class="row.stock_qty < row.required_qty ? 'text-danger' : 'text-success'">
                {{ row.stock_qty }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="库存状态" width="100">
            <template #default="{ row }">
              <el-tag
                :type="row.stock_qty >= row.required_qty ? 'success' : 'danger'"
                size="small"
              >
                {{ row.stock_qty >= row.required_qty ? '充足' : '不足' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="需采购" width="100">
            <template #default="{ row }">
              <span v-if="row.need_purchase > 0" class="text-danger">
                {{ row.need_purchase.toFixed(2) }}
              </span>
              <span v-else class="text-success">-</span>
            </template>
          </el-table-column>
          <el-table-column label="库区" width="100">
            <template #default="{ row }">{{ getZoneName(row.warehouse_zone) }}</template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="demandVisible = false">关闭</el-button>
        <el-button
          v-if="!hasShortage"
          type="primary"
          @click="handleOutbound(currentMealPlan)"
        >
          确认出库
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="outboundVisible" title="确认出库" width="500px">
      <div v-if="currentMealPlan && demandData">
        <el-alert
          v-if="hasShortage"
          title="部分食材库存不足，无法完成出库"
          type="error"
          :closable="false"
          show-icon
          style="margin-bottom: 20px;"
        >
          <ul>
            <li v-for="(item, idx) in shortageItems" :key="idx">
              {{ item.ingredient_name }}: 需求 {{ item.required_qty }}{{ item.unit }}, 库存 {{ item.stock_qty }}{{ item.unit }}, 缺口 {{ (item.required_qty - item.stock_qty).toFixed(2) }}{{ item.unit }}
            </li>
          </ul>
        </el-alert>

        <el-alert
          v-else
          title="出库确认"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 20px;"
        >
          <p>日期: {{ currentMealPlan.date }}</p>
          <p>餐次: {{ getMealTypeText(currentMealPlan.meal_type) }}</p>
          <p>就餐人数: {{ currentMealPlan.peopleNum }} 人</p>
          <p>食材种类: {{ demandData.ingredients?.length || 0 }} 种</p>
          <p>出库后将自动扣减对应食材的库存数量</p>
        </el-alert>

        <el-form v-if="!hasShortage" label-width="80px">
          <el-form-item label="备注">
            <el-input
              v-model="outboundRemark"
              type="textarea"
              :rows="3"
              placeholder="请输入出库备注（可选）"
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="outboundVisible = false">取消</el-button>
        <el-button
          v-if="!hasShortage"
          type="primary"
          @click="submitOutbound"
          :loading="submitting"
        >
          确认出库
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getMealPlans, calculateMealPlanDemand, stockOutbound } from '@/api/inventory'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUser } from '@/utils/auth'

const dateFilter = ref(new Date().toISOString().split('T')[0])
const mealTypeFilter = ref('')
const mealPlans = ref([])
const demandVisible = ref(false)
const outboundVisible = ref(false)
const currentMealPlan = ref(null)
const demandData = ref(null)
const outboundRemark = ref('')
const submitting = ref(false)
const user = getUser()

const hasShortage = computed(() => {
  if (!demandData.value?.ingredients) return false
  return demandData.value.ingredients.some(ing => ing.stock_qty < ing.required_qty)
})

const shortageItems = computed(() => {
  if (!demandData.value?.ingredients) return []
  return demandData.value.ingredients.filter(ing => ing.stock_qty < ing.required_qty)
})

const getMealTypeText = (type) => {
  const texts = { breakfast: '早餐', lunch: '午餐', dinner: '晚餐' }
  return texts[type] || type
}

const getZoneName = (zone) => {
  const names = { dry: '干货区', refrigerated: '冷藏区', frozen: '冷冻区' }
  return names[zone] || zone
}

const fetchMealPlans = async () => {
  try {
    const params = {}
    if (dateFilter.value) params.date = dateFilter.value
    if (mealTypeFilter.value) params.mealType = mealTypeFilter.value
    const res = await getMealPlans(params)
    if (res.success) {
      mealPlans.value = (res.data || []).map(plan => ({
        ...plan,
        peopleNum: plan.peopleNum || 50
      }))
    }
  } catch (e) {
    ElMessage.error('获取配餐计划失败')
  }
}

const viewDemand = async (row) => {
  currentMealPlan.value = row
  try {
    const res = await calculateMealPlanDemand({
      meal_plan_id: row.id,
      people_num: row.peopleNum
    })
    if (res.success) {
      demandData.value = res.data
      demandVisible.value = true
    }
  } catch (e) {
    ElMessage.error('获取需求明细失败')
  }
}

const handleOutbound = async (row) => {
  currentMealPlan.value = row
  outboundRemark.value = ''

  try {
    const res = await calculateMealPlanDemand({
      meal_plan_id: row.id,
      people_num: row.peopleNum
    })
    if (res.success) {
      demandData.value = res.data
      outboundVisible.value = true
    }
  } catch (e) {
    ElMessage.error('计算需求失败')
  }
}

const submitOutbound = async () => {
  if (!currentMealPlan.value) return
  if (hasShortage.value) {
    ElMessage.error('存在库存不足的食材，无法出库')
    return
  }

  try {
    submitting.value = true
    const res = await stockOutbound({
      meal_plan_id: currentMealPlan.value.id,
      people_num: currentMealPlan.value.peopleNum,
      operator_id: user?.id,
      operator_name: user?.name || user?.username,
      remark: outboundRemark.value
    })
    if (res.success) {
      ElMessage.success('出库成功')
      outboundVisible.value = false
      demandVisible.value = false
    }
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '出库失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchMealPlans()
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
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.text-success {
  color: #67c23a;
  font-weight: 500;
}

.text-danger {
  color: #f56c6c;
  font-weight: 500;
}
</style>
