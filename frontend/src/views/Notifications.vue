<template>
  <div class="notifications-page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>系统通知中心</span>
          <div class="header-actions">
            <el-radio-group v-model="statusFilter" size="small" @change="fetchData">
              <el-radio-button value="">全部</el-radio-button>
              <el-radio-button value="unread">未读</el-radio-button>
              <el-radio-button value="read">已读</el-radio-button>
            </el-radio-group>
            <el-button size="small" type="primary" @click="handleMarkAllRead" :disabled="unreadCount === 0">
              全部标记已读
            </el-button>
          </div>
        </div>
      </template>

      <div class="stats-bar">
        <el-statistic title="未读通知" :value="unreadCount" value-color="#f56c6c" />
        <el-divider direction="vertical" />
        <el-statistic title="今日通知" :value="todayCount" />
        <el-divider direction="vertical" />
        <el-statistic title="全部通知" :value="notifications.length" />
      </div>

      <el-table :data="filteredNotifications" style="width: 100%" stripe>
        <el-table-column label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="getPriorityTagType(row.priority)" size="small">
              {{ getPriorityText(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ getTypeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="标题" min-width="200">
          <template #default="{ row }">
            <span :class="{ 'font-bold': row.status === 'unread' }">{{ row.title }}</span>
          </template>
        </el-table-column>
        <el-table-column label="内容" min-width="300">
          <template #default="{ row }">
            <div class="content-cell">
              <pre class="content-text">{{ row.content }}</pre>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="关联单据" width="140">
          <template #default="{ row }">
            <el-link v-if="row.related_no" type="primary" @click="goToRelated(row)">
              {{ row.related_no }}
            </el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'unread' ? 'danger' : 'success'" size="small">
              {{ row.status === 'unread' ? '未读' : '已读' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="阅读时间" width="160">
          <template #default="{ row }">{{ row.read_at ? formatTime(row.read_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'unread'">
              <el-button type="primary" size="small" link @click="handleMarkRead(row)">标记已读</el-button>
            </template>
            <template v-else>
              <span class="text-muted">已处理</span>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="filteredNotifications.length === 0" description="暂无通知数据" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getNotifications, markNotificationRead, markAllNotificationsRead } from '@/api/inventory'

const router = useRouter()
const statusFilter = ref('')
const notifications = ref([])
const loading = ref(false)

const unreadCount = computed(() => {
  return notifications.value.filter(n => n.status === 'unread').length
})

const todayCount = computed(() => {
  const today = new Date().toDateString()
  return notifications.value.filter(n => new Date(n.created_at).toDateString() === today).length
})

const filteredNotifications = computed(() => {
  if (statusFilter.value === '') {
    return notifications.value
  }
  return notifications.value.filter(n => n.status === statusFilter.value)
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

const getPriorityText = (priority) => {
  const texts = { high: '高', normal: '中', low: '低' }
  return texts[priority] || priority
}

const getPriorityTagType = (priority) => {
  const types = { high: 'danger', normal: 'primary', low: 'info' }
  return types[priority] || 'info'
}

const getTypeText = (type) => {
  const texts = {
    auto_replenish: '自动补货',
    low_stock: '库存预警',
    system: '系统通知',
    success: '操作成功'
  }
  return texts[type] || type
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = {}
    if (statusFilter.value) params.status = statusFilter.value
    const res = await getNotifications(params)
    if (res.success) {
      notifications.value = res.data.list || []
    }
  } catch (e) {
    ElMessage.error('获取通知数据失败')
  } finally {
    loading.value = false
  }
}

const handleMarkRead = async (row) => {
  try {
    await markNotificationRead(row.id)
    row.status = 'read'
    row.read_at = new Date().toISOString()
    ElMessage.success('已标记为已读')
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const handleMarkAllRead = async () => {
  ElMessageBox.confirm('确定将所有通知标记为已读吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await markAllNotificationsRead()
      notifications.value.forEach(n => {
        if (n.status === 'unread') {
          n.status = 'read'
          n.read_at = new Date().toISOString()
        }
      })
      ElMessage.success('已全部标记为已读')
    } catch (e) {
      ElMessage.error('操作失败')
    }
  }).catch(() => {})
}

const goToRelated = (row) => {
  if (row.related_type === 'purchase' && row.related_id) {
    router.push(`/purchases/${row.related_id}`)
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.notifications-page {
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

.stats-bar {
  display: flex;
  align-items: center;
  padding: 20px 0;
  margin-bottom: 20px;
  background: #f5f7fa;
  border-radius: 8px;
  justify-content: space-around;
}

.content-cell {
  max-height: 80px;
  overflow: hidden;
}

.content-text {
  margin: 0;
  font-family: inherit;
  font-size: 13px;
  color: #606266;
  white-space: pre-wrap;
  line-height: 1.5;
}

.font-bold {
  font-weight: 600;
  color: #303133;
}

.text-muted {
  color: #909399;
}
</style>
