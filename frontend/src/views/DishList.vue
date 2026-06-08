<template>
  <div class="dish-list">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">菜品管理</span>
          <div class="header-actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索菜品名称"
              style="width: 200px; margin-right: 10px;"
              :prefix-icon="Search"
              clearable
              @keyup.enter="fetchDishes"
            />
            <el-select v-model="categoryFilter" placeholder="分类" style="width: 120px; margin-right: 10px;" clearable @change="fetchDishes">
              <el-option label="主食" value="主食" />
              <el-option label="热菜" value="热菜" />
              <el-option label="凉菜" value="凉菜" />
              <el-option label="汤品" value="汤品" />
              <el-option label="饮品" value="饮品" />
            </el-select>
            <el-button type="primary" :icon="Plus" @click="handleAdd">新增菜品</el-button>
          </div>
        </div>
      </template>
      <el-table :data="dishes" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="菜品名称" />
        <el-table-column prop="category" label="分类" width="100">
          <template #default="{ row }">
            <el-tag type="success" size="small">{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="price" label="价格" width="100">
          <template #default="{ row }">¥{{ row.price }}</template>
        </el-table-column>
        <el-table-column prop="calories" label="热量(kcal)" width="120" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="available" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.available ? 'success' : 'danger'" size="small">
              {{ row.available ? '在售' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="success" link size="small" v-if="row.available" @click="toggleStatus(row, false)">下架</el-button>
            <el-button type="warning" link size="small" v-else @click="toggleStatus(row, true)">上架</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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
        @size-change="fetchDishes"
        @current-change="fetchDishes"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑菜品' : '新增菜品'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="菜品名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜品名称" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="form.category" placeholder="请选择分类" style="width: 100%;">
            <el-option label="主食" value="主食" />
            <el-option label="热菜" value="热菜" />
            <el-option label="凉菜" value="凉菜" />
            <el-option label="汤品" value="汤品" />
            <el-option label="饮品" value="饮品" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number v-model="form.price" :min="0" :precision="2" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="热量" prop="calories">
          <el-input-number v-model="form.calories" :min="0" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入菜品描述" />
        </el-form-item>
        <el-form-item label="图片URL" prop="image">
          <el-input v-model="form.image" placeholder="请输入图片URL" />
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
import { Search, Plus } from '@element-plus/icons-vue'
import { getDishes, createDish, updateDish, deleteDish } from '@/api/dish'

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const searchKeyword = ref('')
const categoryFilter = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dishes = ref([])

const form = reactive({
  id: null,
  name: '',
  category: '',
  price: 0,
  calories: 0,
  description: '',
  image: '',
  available: true
})

const rules = {
  name: [{ required: true, message: '请输入菜品名称', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }]
}

const fetchDishes = async () => {
  loading.value = true
  try {
    const res = await getDishes({
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
      category: categoryFilter.value
    })
    dishes.value = res.data?.items || res.data || []
    total.value = res.data?.total || dishes.value.length
  } catch (e) {
    console.log('Get dishes error')
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null,
    name: '',
    category: '',
    price: 0,
    calories: 0,
    description: '',
    image: '',
    available: true
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (isEdit.value) {
          await updateDish(form.id, form)
          ElMessage.success('更新成功')
        } else {
          await createDish(form)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchDishes()
      } catch (e) {
        console.log('Submit error')
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该菜品吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
      try {
        await deleteDish(row.id)
        ElMessage.success('删除成功')
        fetchDishes()
      } catch (e) {
        console.log('Delete error')
      }
    }).catch(() => {})
}

const toggleStatus = async (row, available) => {
  try {
    await updateDish(row.id, { ...row, available })
    ElMessage.success('操作成功')
    fetchDishes()
  } catch (e) {
    console.log('Toggle status error')
  }
}

onMounted(() => {
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
