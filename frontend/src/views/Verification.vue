<template>
  <div class="verification-page">
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card shadow="hover" class="main-card">
          <template #header>
            <div class="card-header">
              <span class="title">食堂窗口核销</span>
              <div class="session-info" v-if="currentSession">
                <el-tag type="success" size="large" effect="dark">
                  <el-icon><Clock /></el-icon>
                  {{ currentSession.name }}
                  <span class="time-range">{{ currentSession.start_time }} - {{ currentSession.end_time }}</span>
                </el-tag>
              </div>
              <div class="session-info" v-else>
                <el-tag type="warning" size="large" effect="dark">
                  <el-icon><Warning /></el-icon>
                  当前非用餐时段
                </el-tag>
              </div>
            </div>
          </template>

          <div class="verify-content">
            <el-tabs v-model="activeTab" class="verify-tabs">
              <el-tab-pane label="扫码核销" name="scan">
                <div class="scan-container">
                  <div id="qr-reader" class="qr-reader" v-if="showScanner"></div>
                  <div class="scan-placeholder" v-else @click="startScanner">
                    <el-icon :size="80" color="#409EFF"><Camera /></el-icon>
                    <p class="placeholder-text">点击开启摄像头扫码</p>
                    <p class="placeholder-hint">支持手机、扫码枪等设备</p>
                  </div>
                  <div class="scan-actions">
                    <el-button type="primary" :icon="Camera" @click="startScanner" v-if="!showScanner">开启扫码</el-button>
                    <el-button type="danger" :icon="Close" @click="stopScanner" v-else>关闭扫码</el-button>
                    <el-button :icon="Refresh" @click="restartScanner" v-if="showScanner">重新扫码</el-button>
                  </div>
                </div>
              </el-tab-pane>
              <el-tab-pane label="手动输入" name="manual">
                <div class="manual-input-container">
                  <el-input
                    v-model="orderNoInput"
                    placeholder="请输入取餐码/订单号"
                    size="large"
                    :prefix-icon="Tickets"
                    clearable
                    @keyup.enter="handleVerify"
                    class="order-input"
                  />
                  <el-button type="primary" size="large" :icon="Check" @click="handleVerify" :loading="verifying">
                    确认核销
                  </el-button>
                </div>
              </el-tab-pane>
            </el-tabs>

            <div class="recent-scan-result" v-if="lastResult">
              <el-alert
                :title="lastResult.title"
                :type="lastResult.type"
                :description="lastResult.message"
                :closable="false"
                show-icon
              >
                <template #default v-if="lastResult.order">
                  <el-descriptions :column="2" size="small" class="result-desc">
                    <el-descriptions-item label="取餐码">{{ lastResult.order.order_no }}</el-descriptions-item>
                    <el-descriptions-item label="用餐人">{{ lastResult.order.user?.name || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="餐次">{{ getMealTypeName(lastResult.order.meal_time) }}</el-descriptions-item>
                    <el-descriptions-item label="日期">{{ lastResult.order.meal_date }}</el-descriptions-item>
                    <el-descriptions-item label="菜品" :span="2">
                      {{ (lastResult.order.items || []).map(i => i.dish?.name || '').join('、') }}
                    </el-descriptions-item>
                  </el-descriptions>
                </template>
              </el-alert>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover" class="stats-card">
          <template #header>
            <div class="card-header">
              <span class="title">今日核销统计</span>
              <el-button type="text" :icon="Refresh" @click="loadAllData">刷新</el-button>
            </div>
          </template>
          <el-row :gutter="12" class="stats-row">
            <el-col :span="12">
              <div class="stat-item success">
                <div class="stat-value">{{ todayStats.success }}</div>
                <div class="stat-label">成功核销</div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="stat-item total">
                <div class="stat-value">{{ todayStats.total }}</div>
                <div class="stat-label">今日订单</div>
              </div>
            </el-col>
          </el-row>
          <el-divider />
          <div class="upcoming-sessions" v-if="nextSessions && nextSessions.length > 0">
            <h4 class="section-title">
              <el-icon><Clock /></el-icon>
              下一时段
            </h4>
            <div class="session-list">
              <div class="session-item" v-for="session in nextSessions" :key="session.id">
                <span class="session-name">{{ session.name }}</span>
                <span class="session-time">{{ session.start_time }} - {{ session.end_time }}</span>
              </div>
            </div>
          </div>
          <div class="all-sessions">
            <h4 class="section-title">
              <el-icon><Calendar /></el-icon>
              餐次时段配置
            </h4>
            <div class="session-list">
              <div class="session-item" v-for="session in allSessions" :key="session.id">
                <span class="session-name">{{ session.name }}</span>
                <span class="session-time">{{ session.start_time }} - {{ session.end_time }}</span>
              </div>
            </div>
          </div>
        </el-card>

        <el-card shadow="hover" class="recent-card" style="margin-top: 20px;">
          <template #header>
            <div class="card-header">
              <span class="title">最近核销记录</span>
              <el-button type="text" @click="goToRecords">查看全部</el-button>
            </div>
          </template>
          <div class="recent-list" v-loading="recordsLoading">
            <el-empty v-if="recentRecords.length === 0" description="暂无核销记录" :image-size="80" />
            <div class="recent-item" v-for="record in recentRecords" :key="record.id">
              <div class="item-left">
                <el-avatar :size="36" :icon="User" />
              </div>
              <div class="item-center">
                <div class="item-name">{{ record.user?.name || '未知用户' }}</div>
                <div class="item-info">
                  <el-tag size="small" type="success">{{ getMealTypeName(record.meal_type) }}</el-tag>
                  <span class="item-time">{{ formatTime(record.verified_at) }}</span>
                </div>
              </div>
              <div class="item-right">
                <el-icon color="#67c23a" :size="20"><CircleCheck /></el-icon>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Camera, Close, Refresh, Check, Clock, Warning, Calendar,
  Tickets, User, CircleCheck
} from '@element-plus/icons-vue'
import { Html5Qrcode } from 'html5-qrcode'
import { verifyOrder, getCurrentMealSession, getMealSessions, getVerificationRecords } from '@/api/verification'
import { getOrders } from '@/api/order'
import { getUser } from '@/utils/auth'

const router = useRouter()
const user = getUser()

const activeTab = ref('scan')
const showScanner = ref(false)
const orderNoInput = ref('')
const verifying = ref(false)
const recordsLoading = ref(false)

const currentSession = ref(null)
const nextSessions = ref([])
const allSessions = ref([])
const recentRecords = ref([])
const lastResult = ref(null)
const todayStats = ref({ success: 0, total: 0 })

let html5QrCode = null
let audioContext = null

const getMealTypeName = (type) => {
  const names = {
    breakfast: '早餐',
    lunch: '午餐',
    dinner: '晚餐'
  }
  return names[type] || type
}

const formatTime = (time) => {
  if (!time) return ''
  return time.substring(11, 19)
}

const initAudio = () => {
  if (!audioContext) {
    audioContext = new (window.AudioContext || window.webkitAudioContext)()
  }
}

const playSuccessSound = () => {
  initAudio()
  const oscillator = audioContext.createOscillator()
  const gainNode = audioContext.createGain()
  
  oscillator.connect(gainNode)
  gainNode.connect(audioContext.destination)
  
  oscillator.frequency.value = 880
  oscillator.type = 'sine'
  gainNode.gain.setValueAtTime(0.3, audioContext.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.3)
  
  oscillator.start(audioContext.currentTime)
  oscillator.stop(audioContext.currentTime + 0.3)
}

const playErrorSound = () => {
  initAudio()
  const oscillator = audioContext.createOscillator()
  const gainNode = audioContext.createGain()
  
  oscillator.connect(gainNode)
  gainNode.connect(audioContext.destination)
  
  oscillator.frequency.value = 300
  oscillator.type = 'square'
  gainNode.gain.setValueAtTime(0.3, audioContext.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.3)
  
  oscillator.start(audioContext.currentTime)
  oscillator.stop(audioContext.currentTime + 0.3)
}

const loadSessionInfo = async () => {
  try {
    const [sessionRes, sessionsRes] = await Promise.all([
      getCurrentMealSession(),
      getMealSessions()
    ])
    currentSession.value = sessionRes.data?.current || null
    nextSessions.value = sessionRes.data?.next_sessions || []
    allSessions.value = sessionsRes.data || []
  } catch (e) {
    console.log('Load session error')
  }
}

const loadTodayStats = async () => {
  try {
    const today = new Date().toISOString().split('T')[0]
    const [recordsRes, ordersRes] = await Promise.all([
      getVerificationRecords({ mealDate: today, pageSize: 100 }),
      getOrders({ mealDate: today })
    ])
    todayStats.value.success = recordsRes.data?.total || 0
    todayStats.value.total = ordersRes.data?.length || 0
  } catch (e) {
    console.log('Load stats error')
  }
}

const loadRecentRecords = async () => {
  recordsLoading.value = true
  try {
    const res = await getVerificationRecords({ pageSize: 5 })
    recentRecords.value = res.data?.list || []
  } catch (e) {
    console.log('Load records error')
  } finally {
    recordsLoading.value = false
  }
}

const loadAllData = () => {
  loadSessionInfo()
  loadTodayStats()
  loadRecentRecords()
}

const startScanner = async () => {
  try {
    showScanner.value = true
    await nextTick()
    
    if (!html5QrCode) {
      html5QrCode = new Html5Qrcode('qr-reader')
    }

    const config = {
      fps: 10,
      qrbox: { width: 250, height: 250 },
      aspectRatio: 1.0
    }

    await html5QrCode.start(
      { facingMode: 'environment' },
      config,
      onScanSuccess,
      onScanFailure
    )
  } catch (e) {
    showScanner.value = false
    ElMessage.error('无法启动摄像头，请检查权限')
    console.log('Scanner error:', e)
  }
}

const stopScanner = async () => {
  if (html5QrCode) {
    try {
      await html5QrCode.stop()
    } catch (e) {
      console.log('Stop scanner error')
    }
  }
  showScanner.value = false
}

const restartScanner = async () => {
  await stopScanner()
  await startScanner()
}

const onScanSuccess = (decodedText) => {
  stopScanner()
  orderNoInput.value = decodedText
  handleVerify()
}

const onScanFailure = (error) => {
}

const handleVerify = async () => {
  const orderNo = orderNoInput.value.trim()
  if (!orderNo) {
    ElMessage.warning('请输入取餐码')
    return
  }

  verifying.value = true
  try {
    const res = await verifyOrder({
      order_no: orderNo,
      verified_by: user?.id,
      verifier_name: user?.name
    })

    lastResult.value = {
      type: 'success',
      title: '核销成功',
      message: res.message || '取餐码已核销',
      order: res.data?.order
    }

    ElMessage.success('核销成功')
    playSuccessSound()

    orderNoInput.value = ''
    loadTodayStats()
    loadRecentRecords()

    setTimeout(() => {
      if (activeTab.value === 'scan' && !showScanner.value) {
        startScanner()
      }
    }, 2000)
  } catch (e) {
    const errData = e.response?.data
    lastResult.value = {
      type: 'error',
      title: '核销失败',
      message: errData?.message || '核销失败，请检查取餐码',
      order: errData?.data?.order
    }
    ElMessage.error(errData?.message || '核销失败')
    playErrorSound()
  } finally {
    verifying.value = false
  }
}

const goToRecords = () => {
  router.push('/verification-records')
}

let sessionTimer = null

onMounted(() => {
  loadAllData()
  sessionTimer = setInterval(loadSessionInfo, 60000)
})

onUnmounted(() => {
  stopScanner()
  if (sessionTimer) {
    clearInterval(sessionTimer)
  }
  if (audioContext) {
    audioContext.close()
  }
})
</script>

<style scoped>
.verification-page {
  min-height: 100%;
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

.session-info {
  display: flex;
  align-items: center;
}

.time-range {
  margin-left: 8px;
  opacity: 0.9;
}

.verify-content {
  padding: 20px 0;
}

.verify-tabs {
  margin-bottom: 24px;
}

.scan-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}

.qr-reader {
  width: 100%;
  max-width: 400px;
  margin-bottom: 20px;
}

.scan-placeholder {
  width: 100%;
  max-width: 400px;
  height: 300px;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
  margin-bottom: 20px;
}

.scan-placeholder:hover {
  border-color: #409eff;
  background-color: #ecf5ff;
}

.placeholder-text {
  margin: 16px 0 8px;
  font-size: 16px;
  color: #606266;
}

.placeholder-hint {
  font-size: 14px;
  color: #909399;
}

.scan-actions {
  display: flex;
  gap: 12px;
}

.manual-input-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  padding: 40px 20px;
}

.order-input {
  width: 100%;
  max-width: 400px;
}

.recent-scan-result {
  margin-top: 20px;
}

.result-desc {
  margin-top: 12px;
}

.stats-row {
  margin-bottom: 10px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
}

.stat-item.success {
  background: linear-gradient(135deg, #f0f9eb 0%, #e1f3d8 100%);
}

.stat-item.total {
  background: linear-gradient(135deg, #ecf5ff 0%, #d9ecff 100%);
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #606266;
  margin-top: 8px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 12px;
}

.session-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.session-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background-color: #f5f7fa;
  border-radius: 6px;
  font-size: 14px;
}

.session-name {
  font-weight: 500;
  color: #303133;
}

.session-time {
  color: #606266;
  font-family: monospace;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recent-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.recent-item:hover {
  background-color: #ecf5ff;
}

.item-center {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.item-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-time {
  font-size: 12px;
  color: #909399;
  font-family: monospace;
}
</style>
