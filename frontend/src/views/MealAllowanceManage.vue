<template>
  <div class="meal-allowance-manage">
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">累计充值</div>
            <div class="stat-value text-primary">¥{{ stats.total_recharge?.toFixed(2) || '0.00' }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">累计消费</div>
            <div class="stat-value text-danger">¥{{ stats.total_consume?.toFixed(2) || '0.00' }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">当前总余额</div>
            <div class="stat-value text-success">¥{{ stats.total_balance?.toFixed(2) || '0.00' }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">用户总数</div>
            <div class="stat-value text-warning">{{ stats.user_count || 0 }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span class="title">餐补余额管理</span>
          <div class="header-actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索姓名/部门"
              style="width: 200px; margin-right: 10px;"
              :prefix-icon="Search"
              clearable
              @keyup.enter="fetchUsers"
            />
            <el-button type="primary" :icon="Tickets" @click="goToRecords">查看明细</el-button>
          </div>
        </div>
      </template>
      <el-table :data="filteredUsers" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="department" label="部门" width="140">
          <template #default="{ row }">
            <span v-if="row.department">{{ row.department }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="meal_allowance" label="当前余额" width="140">
          <template #default="{ row }">
            <span class="balance-text">¥{{ (row.meal_allowance || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleRecharge(row)">充值</el-button>
            <el-button type="success" link size="small" @click="viewRecords(row)">明细</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="rechargeDialogVisible" title="餐补充值" width="500px">
      <el-form ref="rechargeFormRef" :model="rechargeForm" :rules="rechargeRules" label-width="100px">
        <el-form-item label="用户">
          <el-input v-model="rechargeForm.userName" disabled />
        </el-form-item>
        <el-form-item label="当前余额">
          <span class="balance-text">¥{{ (rechargeForm.currentBalance || 0).toFixed(2) }}</span>
        </el-form-item>
        <el-form-item label="充值金额" prop="amount">
          <el-input-number
            v-model="rechargeForm.amount"
            :min="1"
            :precision="2"
            :step="10"
            style="width: 100%;"
            placeholder="请输入充值金额"
          />
        </el-form-item>
        <el-form-item label="充值后余额">
          <span class="balance-text text-success">
            ¥{{ ((rechargeForm.currentBalance || 0) + (rechargeForm.amount || 0)).toFixed(2) }}
          </span>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="rechargeForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注信息（可选）"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRecharge" :loading="recharging">确认充值</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="recordsDialogVisible" :title="recordsUserName + ' - 餐补明细'" width="900px">
      <div class="records-filter">
        <el-select v-model="recordTypeFilter" placeholder="全部类型" style="width: 150px; margin-right: 10px;" clearable @change="fetchUserRecords">
          <el-option label="充值" value="recharge" />
          <el-option label="消费" value="consume" />
          <el-option label="退款" value="refund" />
        </el-select>
      </div>
      <el-table :data="userRecords" style="width: 100%" v-loading="recordsLoading">
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            <span>{{ formatTime(row.created_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)" size="small">
              {{ getTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="金额" width="120">
          <template #default="{ row }">
            <span :class="row.amount >= 0 ? 'text-success' : 'text-danger'">
              {{ row.amount >= 0 ? '+' : '' }}{{ row.amount.toFixed(2) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="balance_before" label="变动前" width="120">
          <template #default="{ row }">
            <span>¥{{ row.balance_before.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="balance_after" label="变动后" width="120">
          <template #default="{ row }">
            <span>¥{{ row.balance_after.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="related_no" label="关联单号" width="180">
          <template #default="{ row }">
            <span v-if="row.related_no">{{ row.related_no }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="operator_name" label="操作人" width="120">
          <template #default="{ row }">
            <span v-if="row.operator_name">{{ row.operator_name }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注">
          <template #default="{ row }">
            <span v-if="row.remark">{{ row.remark }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="recordsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Tickets } from '@element-plus/icons-vue'
import { getUsers } from '@/api/user'
import { rechargeMealAllowance, getMealAllowanceStats, getUserMealAllowanceRecords } from '@/api/mealAllowance'

const router = useRouter()
const loading = ref(false)
const recharging = ref(false)
const recordsLoading = ref(false)
const users = ref([])
const searchKeyword = ref('')
const stats = ref({})

const rechargeDialogVisible = ref(false)
const rechargeFormRef = ref(null)
const rechargeForm = reactive({
  userId: null,
  userName: '',
  currentBalance: 0,
  amount: null,
  remark: ''
})
const rechargeRules = {
  amount: [{ required: true, message: '请输入充值金额', trigger: 'blur' }]
}

const recordsDialogVisible = ref(false)
const recordsUserName = ref('')
const userRecords = ref([])
const recordTypeFilter = ref('')

const filteredUsers = computed(() => {
  if (!searchKeyword.value) return users.value
  const keyword = searchKeyword.value.toLowerCase()
  return users.value.filter(u =>
    u.name?.toLowerCase().includes(keyword) ||
    u.department?.toLowerCase().includes(keyword)
  )
})

const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime()) || date.getFullYear() < 1970) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const getTypeText = (type) => {
  const map = {
    recharge: '充值',
    consume: '消费',
    refund: '退款'
  }
  return map[type] || type
}

const getTypeTagType = (type) => {
  const map = {
    recharge: 'success',
    consume: 'danger',
    refund: 'warning'
  }
  return map[type] || 'info'
}

const fetchStats = async () => {
  try {
    const res = await getMealAllowanceStats()
    stats.value = res.data || {}
  } catch (e) {
    console.log('Get stats error')
  }
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await getUsers()
    users.value = res.data || []
  } catch (e) {
    console.log('Get users error')
  } finally {
    loading.value = false
  }
}

const handleRecharge = (row) => {
  rechargeForm.userId = row.id
  rechargeForm.userName = row.name
  rechargeForm.currentBalance = row.meal_allowance || 0
  rechargeForm.amount = null
  rechargeForm.remark = ''
  rechargeDialogVisible.value = true
}

const submitRecharge = async () => {
  if (!rechargeFormRef.value) return
  await rechargeFormRef.value.validate(async (valid) => {
    if (valid) {
      recharging.value = true
      try {
        const currentUser = JSON.parse(localStorage.getItem('user') || '{}')
        await rechargeMealAllowance({
          user_id: rechargeForm.userId,
          amount: rechargeForm.amount,
          operator_id: currentUser.id,
          operator_name: currentUser.name,
          remark: rechargeForm.remark
        })
        ElMessage.success('充值成功')
        rechargeDialogVisible.value = false
        fetchUsers()
        fetchStats()
      } catch (e) {
        console.log('Recharge error')
      } finally {
        recharging.value = false
      }
    }
  })
}

const viewRecords = async (row) => {
  recordsUserName.value = row.name
  recordTypeFilter.value = ''
  userRecords.value = []
  recordsDialogVisible.value = true
  await fetchUserRecords(row.id)
}

const fetchUserRecords = async (userId) => {
  recordsLoading.value = true
  try {
    const uid = userId || rechargeForm.userId
    const res = await getUserMealAllowanceRecords(uid, {
      type: recordTypeFilter.value || undefined
    })
    userRecords.value = res.data?.records || []
  } catch (e) {
    console.log('Get user records error')
  } finally {
    recordsLoading.value = false
  }
}

const goToRecords = () => {
  router.push('/meal-allowance-records')
}

onMounted(() => {
  fetchStats()
  fetchUsers()
})
</script>

<style scoped>
.stat-item {
  text-align: center;
  padding: 10px 0;
}

.stat-label {
  color: #909399;
  font-size: 14px;
  margin-bottom: 10px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
}

.text-primary {
  color: #409eff;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}

.text-warning {
  color: #e6a23c;
}

.balance-text {
  font-weight: 600;
  color: #303133;
}

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

.text-muted {
  color: #909399;
}

.records-filter {
  margin-bottom: 15px;
}
</style>
