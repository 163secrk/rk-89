<template>
  <div class="clean-plate-dashboard">
    <el-row :gutter="20" class="filter-row">
      <el-col :span="24">
        <el-card shadow="hover">
          <div class="filter-content">
            <div class="filter-item">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              @change="fetchAllData"
            />
            </div>
            <div class="filter-item">
            <el-select v-model="selectedDepartment" placeholder="选择部门" clearable @change="fetchAllData">
              <el-option label="全部部门" value="" />
              <el-option v-for="dept in departments" :key="dept" :label="dept" :value="dept" />
            </el-select>
            </div>
            <div class="filter-item">
            <el-button type="primary" @click="fetchAllData">
              <el-icon><Search /></el-icon>
              查询
            </el-button>
            <el-button @click="resetFilter">
              <el-icon><Refresh /></el-icon>
              重置
            </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon :size="32"><Tickets /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">总报餐数</p>
              <p class="stat-value">{{ stats.total_bookings || 0 }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);">
              <el-icon :size="32"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">实际核销</p>
              <p class="stat-value">{{ stats.total_verified || 0 }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon :size="32"><Delete /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">浪费数量</p>
              <p class="stat-value">{{ stats.total_wasted || 0 }}</p>
              <p class="stat-tip" :class="getWasteRateClass()">{{ stats.waste_rate || 0 }}%</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);">
              <el-icon :size="32"><Medal /></el-icon>
            </div>
            <div class="stat-info">
              <p class="stat-label">报餐诚信度</p>
              <p class="stat-value">{{ stats.integrity_rate || 0 }}%</p>
              <p class="stat-tip" :class="getIntegrityRateClass()">{{ getIntegrityLevel() }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">部门诚信度排名</span>
            </div>
          </template>
          <div class="ranking-list">
            <div v-for="(dept, index) in departmentRankings" :key="dept.department" class="ranking-item">
              <div class="ranking-number" :class="'rank-' + (index + 1)">
                {{ index + 1 }}
              </div>
              <div class="ranking-info">
                <div class="ranking-header">
                  <span class="dept-name">{{ dept.department }}</span>
                  <span class="integrity-badge" :class="getIntegrityBadgeClass(dept.integrity_rate)">
                    {{ dept.integrity_rate }}%
                  </span>
                </div>
                <div class="ranking-progress">
                  <el-progress
                    :percentage="dept.integrity_rate"
                    :color="getProgressColor(dept.integrity_rate)"
                    :stroke-width="8"
                    :show-text="false"
                  />
                </div>
                <div class="ranking-stats">
                  <span>报餐: {{ dept.total_bookings }}</span>
                  <span>核销: {{ dept.total_verified }}</span>
                  <span class="text-warning">浪费: {{ dept.total_wasted }}</span>
                </div>
              </div>
            </div>
          </div>
          <el-empty v-if="departmentRankings.length === 0" description="暂无部门数据" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">个人诚信度排名</span>
            </div>
          </template>
          <div class="ranking-list">
            <div v-for="(user, index) in userRankings" :key="user.user_id" class="ranking-item">
              <div class="ranking-number" :class="'rank-' + (index + 1)">
                {{ index + 1 }}
              </div>
              <div class="ranking-info">
                <div class="ranking-header">
                  <span class="dept-name">{{ user.user_name }}</span>
                  <span class="integrity-badge" :class="getIntegrityBadgeClass(user.integrity_rate)">
                    {{ user.integrity_rate }}%
                  </span>
                </div>
                <div class="ranking-progress">
                  <el-progress
                    :percentage="user.integrity_rate"
                    :color="getProgressColor(user.integrity_rate)"
                    :stroke-width="8"
                    :show-text="false"
                  />
                </div>
                <div class="ranking-stats">
                  <span class="text-muted">{{ user.department }}</span>
                  <span>报餐: {{ user.total_bookings }}</span>
                  <span>核销: {{ user.total_verified }}</span>
                </div>
              </div>
            </div>
          </div>
          <el-empty v-if="userRankings.length === 0" description="暂无个人数据" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">诚信度趋势</span>
            </div>
          </template>
          <div class="trend-chart">
            <div class="chart-container">
              <div class="chart-y-axis">
                <span>100%</span>
                <span>80%</span>
                <span>60%</span>
                <span>40%</span>
                <span>20%</span>
                <span>0%</span>
              </div>
              <div class="chart-content">
                <div class="chart-bars">
                  <div v-for="item in trendData" :key="item.date" class="chart-bar-group">
                    <div class="bar-wrapper">
                      <div
                        class="bar bookings-bar"
                        :style="{ height: (item.total_bookings / maxBookings * 100) + '%' }"
                        :title="'报餐: ' + item.total_bookings"
                      ></div>
                    </div>
                    <div class="bar-wrapper">
                      <div
                        class="bar verified-bar"
                        :style="{ height: (item.total_verified / maxBookings * 100) + '%' }"
                        :title="'核销: ' + item.total_verified"
                      ></div>
                    </div>
                    <div class="bar-wrapper">
                      <div
                        class="bar integrity-bar"
                        :style="{ height: item.integrity_rate + '%' }"
                        :title="'诚信度: ' + item.integrity_rate + '%'"
                      ></div>
                    </div>
                    <div class="bar-label">{{ formatDate(item.date) }}</div>
                  </div>
                </div>
                <div class="chart-legend">
                  <span class="legend-item">
                    <span class="legend-color bookings-color"></span>
                    报餐数
                  </span>
                  <span class="legend-item">
                    <span class="legend-color verified-color"></span>
                    核销数
                  </span>
                  <span class="legend-item">
                    <span class="legend-color integrity-color"></span>
                    诚信度
                  </span>
                </div>
              </div>
            </div>
          </div>
          <el-empty v-if="trendData.length === 0" description="暂无趋势数据" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span class="card-title">浪费情况详细统计</span>
            </div>
          </template>
          <el-table :data="wasteStats" style="width: 100%" size="default">
            <el-table-column label="排名" width="80" align="center">
              <template #default="{ $index }">
                <el-tag :type="$index < 3 ? 'danger' : 'info'" size="small">{{ $index + 1 }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="department" label="部门" />
            <el-table-column label="浪费比例" width="120" align="center">
              <template #default="{ row }">
                <el-tag :type="getWasteTagType(row.waste_rate)">{{ row.waste_rate }}%</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="total_bookings" label="报餐数" width="100" align="center" />
            <el-table-column prop="total_verified" label="核销数" width="100" align="center" />
            <el-table-column prop="total_wasted" label="浪费数" width="100" align="center">
              <template #default="{ row }">
                <span class="text-danger">{{ row.total_wasted }}</span>
              </template>
            </el-table-column>
            <el-table-column label="诚信度" width="120" align="center">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.integrity_rate"
                  :color="getProgressColor(row.integrity_rate)"
                  :width="80"
                />
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="wasteStats.length === 0" description="暂无数据" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Search, Refresh, Tickets, CircleCheck, Delete, Medal } from '@element-plus/icons-vue'
import {
  getCleanPlateStats,
  getCleanPlateDepartmentRanking,
  getCleanPlateUserRanking,
  getCleanPlateTrend,
  getDepartments
} from '@/api/stats'
import { ElMessage } from 'element-plus'

const dateRange = ref([])
const selectedDepartment = ref('')
const departments = ref([])
const stats = ref({})
const departmentRankings = ref([])
const userRankings = ref([])
const trendData = ref([])

const maxBookings = computed(() => {
  let max = 1
  trendData.value.forEach(item => {
    if (item.total_bookings > max) max = item.total_bookings
  })
  return max
})

const wasteStats = computed(() => {
  return [...departmentRankings.value
    .filter(d => d.total_wasted > 0)
    .sort((a, b) => b.waste_rate - a.waste_rate)
})

const getWasteRateClass = () => {
  const rate = stats.value.waste_rate || 0
  if (rate <= 5) return 'text-success'
  if (rate <= 15) return 'text-warning'
  return 'text-danger'
}

const getIntegrityRateClass = () => {
  const rate = stats.value.integrity_rate || 0
  if (rate >= 90) return 'text-success'
  if (rate >= 70) return 'text-warning'
  return 'text-danger'
}

const getIntegrityLevel = () => {
  const rate = stats.value.integrity_rate || 0
  if (rate >= 90) return '优秀'
  if (rate >= 80) return '良好'
  if (rate >= 70) return '一般'
  return '待改进'
}

const getProgressColor = (percentage) => {
  if (percentage >= 90) return '#67c23a'
  if (percentage >= 70) return '#e6a23c'
  return '#f56c6c'
}

const getIntegrityBadgeClass = (rate) => {
  if (rate >= 90) return 'badge-excellent'
  if (rate >= 80) return 'badge-good'
  if (rate >= 70) return 'badge-normal'
  return 'badge-poor'
}

const getWasteTagType = (rate) => {
  if (rate > 20) return 'danger'
  if (rate > 10) return 'warning'
  return 'success'
}

const formatDate = (date) => {
  if (!date) return ''
  const parts = date.split('-')
  return `${parts[1]}/${parts[2]}`
}

const fetchStats = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.startDate = dateRange.value[0]
      params.endDate = dateRange.value[1]
    }
    const res = await getCleanPlateStats(params)
    if (res.success) {
      stats.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取统计数据失败')
  }
}

const fetchDepartmentRanking = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.startDate = dateRange.value[0]
      params.endDate = dateRange.value[1]
    }
    const res = await getCleanPlateDepartmentRanking(params)
    if (res.success) {
      departmentRankings.value = res.data.rankings || []
    }
  } catch (e) {
    ElMessage.error('获取部门排名失败')
  }
}

const fetchUserRanking = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.startDate = dateRange.value[0]
      params.endDate = dateRange.value[1]
    }
    if (selectedDepartment.value) {
      params.department = selectedDepartment.value
    }
    const res = await getCleanPlateUserRanking(params)
    if (res.success) {
      userRankings.value = (res.data.rankings || []).slice(0, 10)
    }
  } catch (e) {
    ElMessage.error('获取个人排名失败')
  }
}

const fetchTrend = async () => {
  try {
    const params = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.startDate = dateRange.value[0]
      params.endDate = dateRange.value[1]
    }
    const res = await getCleanPlateTrend(params)
    if (res.success) {
      trendData.value = res.data.trend || []
    }
  } catch (e) {
    ElMessage.error('获取趋势数据失败')
  }
}

const fetchDepartments = async () => {
  try {
    const res = await getDepartments()
    if (res.success) {
      departments.value = res.data || []
    }
  } catch (e) {
    console.log('获取部门列表失败')
  }
}

const fetchAllData = () => {
  fetchStats()
  fetchDepartmentRanking()
  fetchUserRanking()
  fetchTrend()
}

const resetFilter = () => {
  dateRange.value = []
  selectedDepartment.value = ''
  fetchAllData()
}

onMounted(() => {
  fetchDepartments()
  fetchAllData()
})
</script>

<style scoped>
.filter-row {
  margin-bottom: 0;
}

.filter-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-card {
  border-radius: 12px;
  transition: transform 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-info {
  margin-left: 16px;
  flex: 1;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.stat-tip {
  font-size: 12px;
  margin-top: 4px;
  font-weight: 500;
}

.text-success {
  color: #67c23a;
}

.text-warning {
  color: #e6a23c;
}

.text-danger {
  color: #f56c6c;
}

.text-muted {
  color: #909399;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-weight: 600;
  font-size: 16px;
}

.ranking-list {
  max-height: 400px;
  overflow-y: auto;
}

.ranking-item {
  display: flex;
  align-items: flex-start;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.ranking-item:last-child {
  border-bottom: none;
}

.ranking-number {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: white;
  margin-right: 12px;
  flex-shrink: 0;
}

.rank-1 {
  background: linear-gradient(135deg, #ffd700 0%, #ff8c00 100%);
}

.rank-2 {
  background: linear-gradient(135deg, #c0c0c0 0%, #808080 100%);
}

.rank-3 {
  background: linear-gradient(135deg, #cd7f32 0%, #8b4513 100%);
}

.rank-4,
.rank-5,
.rank-6,
.rank-7,
.rank-8,
.rank-9,
.rank-10 {
  background: #909399;
}

.ranking-info {
  flex: 1;
  min-width: 0;
}

.ranking-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.dept-name {
  font-size: 15px;
  font-weight: 500;
  color: #303133;
}

.integrity-badge {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.badge-excellent {
  background: #f0f9eb;
  color: #67c23a;
}

.badge-good {
  background: #ecf5ff;
  color: #409eff;
}

.badge-normal {
  background: #fdf6ec;
  color: #e6a23c;
}

.badge-poor {
  background: #fef0f0;
  color: #f56c6c;
}

.ranking-progress {
  margin-bottom: 6px;
}

.ranking-stats {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #909399;
}

.trend-chart {
  padding: 16px 0;
}

.chart-container {
  display: flex;
  height: 300px;
}

.chart-y-axis {
  width: 50px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: flex-end;
  padding-right: 10px;
  font-size: 12px;
  color: #909399;
  border-right: 1px solid #f0f0f0;
}

.chart-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chart-bars {
  flex: 1;
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  padding: 0 20px;
  border-bottom: 1px solid #f0f0f0;
}

.chart-bar-group {
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  position: relative;
}

.bar-wrapper {
  height: 100%;
  display: flex;
  align-items: flex-end;
  margin: 0 2px;
}

.bar {
  width: 20px;
  border-radius: 4px 4px 0 0;
  transition: all 0.3s;
  cursor: pointer;
}

.bar:hover {
  opacity: 0.8;
}

.bookings-bar {
  background: linear-gradient(180deg, #667eea 0%, #764ba2 100%);
}

.verified-bar {
  background: linear-gradient(180deg, #43e97b 0%, #38f9d7 100%);
}

.integrity-bar {
  background: linear-gradient(180deg, #fa709a 0%, #fee140 100%);
}

.bar-label {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

.chart-legend {
  display: flex;
  justify-content: center;
  gap: 32px;
  margin-top: 16px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #606266;
}

.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 4px;
}

.bookings-color {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.verified-color {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.integrity-color {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}
</style>
