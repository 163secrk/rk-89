<template>
  <div class="meal-plan">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">配餐计划管理</span>
          <div class="header-actions">
            <el-date-picker
              v-model="dateFilter"
              type="date"
              placeholder="选择日期"
              style="width: 180px; margin-right: 10px;"
              @change="fetchMealPlans"
            />
            <el-button type="primary" :icon="Plus" @click="handleAdd">新增计划</el-button>
          </div>
        </div>
      </template>
      <el-table :data="mealPlans" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column prop="meal_type" label="餐次" width="100">
          <template #default="{ row }">
            <el-tag type="success" size="small">{{ row.meal_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="菜品" show-overflow-tooltip>
          <template #default="{ row }">
            {{ (row.dishes || []).map(d => d.name).join('、') }}
          </template>
        </el-table-column>
        <el-table-column label="总热量(kcal)" width="140">
          <template #default="{ row }">
            {{ (row.dishes || []).reduce((sum, d) => sum + (d.calories || 0), 0) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑配餐计划' : '新增配餐计划'" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="日期" prop="date">
          <el-date-picker v-model="form.date" type="date" placeholder="选择日期" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="餐次" prop="mealType">
          <el-select v-model="form.mealType" placeholder="请选择餐次" style="width: 100%;">
            <el-option label="早餐" value="breakfast" />
            <el-option label="午餐" value="lunch" />
            <el-option label="晚餐" value="dinner" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择菜品" prop="dishIds">
          <el-select
            v-model="form.dishIds"
            multiple
            filterable
            placeholder="请选择菜品"
            style="width: 100%;"
            @change="calculateCalories"
          >
            <el-option
              v-for="dish in dishOptions"
              :key="dish.id"
              :label="`${dish.name} (¥${dish.price}) - ${dish.calories}kcal`"
              :value="dish.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="总热量">
          <el-input v-model="form.totalCalories" disabled />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getMealPlans, createMealPlan, updateMealPlan, deleteMealPlan } from '@/api/mealplan'
import { getDishes } from '@/api/dish'

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const dateFilter = ref('')
const mealPlans = ref([])
const dishOptions = ref([])

const form = reactive({
  id: null,
  date: '',
  mealType: '',
  dishIds: [],
  totalCalories: 0
})

const rules = {
  date: [{ required: true, message: '请选择日期', trigger: 'change' }],
  mealType: [{ required: true, message: '请选择餐次', trigger: 'change' }],
  dishIds: [{ required: true, message: '请选择菜品', trigger: 'change' }]
}

const fetchMealPlans = async () => {
  loading.value = true
  try {
    const params = {}
    if (dateFilter.value) {
      params.date = dateFilter.value
    }
    const res = await getMealPlans(params)
    mealPlans.value = res.data?.items || res.data || []
  } catch (e) {
    console.log('Get meal plans error')
  } finally {
    loading.value = false
  }
}

const fetchDishes = async () => {
  try {
    const res = await getDishes({ available: true })
    dishOptions.value = res.data?.items || res.data || []
  } catch (e) {
    console.log('Get dishes error')
  }
}

const calculateCalories = () => {
  let total = 0
  form.dishIds.forEach(id => {
    const dish = dishOptions.value.find(d => d.id === id)
    if (dish) {
      total += dish.calories || 0
    }
  })
  form.totalCalories = total
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null,
    date: '',
    mealType: '',
    dishIds: [],
    totalCalories: 0
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  const dishIds = row.dish_ids ? row.dish_ids.split(',').map(Number) : []
  Object.assign(form, {
    id: row.id,
    date: row.date,
    mealType: row.meal_type,
    dishIds: dishIds,
    totalCalories: (row.dishes || []).reduce((sum, d) => sum + (d.calories || 0), 0)
  })
  calculateCalories()
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const data = {
          date: form.date,
          meal_type: form.mealType,
          dish_ids: form.dishIds.join(',')
        }
        if (isEdit.value) {
          await updateMealPlan(form.id, data)
          ElMessage.success('更新成功')
        } else {
          await createMealPlan(data)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchMealPlans()
      } catch (e) {
        console.log('Submit error')
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该配餐计划吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
      try {
        await deleteMealPlan(row.id)
        ElMessage.success('删除成功')
        fetchMealPlans()
      } catch (e) {
        console.log('Delete error')
      }
    }).catch(() => {})
}

onMounted(() => {
  fetchMealPlans()
  fetchDishes()
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
