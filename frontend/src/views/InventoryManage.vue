<template>
  <div class="inventory-manage">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>库存管理</span>
        </div>
      </template>

      <div class="filter-bar">
        <el-radio-group v-model="activeZone" @change="fetchData">
          <el-radio-button value="">全部库区</el-radio-button>
          <el-radio-button value="dry">干货区</el-radio-button>
          <el-radio-button value="refrigerated">冷藏区</el-radio-button>
          <el-radio-button value="frozen">冷冻区</el-radio-button>
        </el-radio-group>

        <div class="filter-right">
          <el-select v-model="filterCategory" placeholder="分类" clearable style="width: 120px; margin-right: 12px;" @change="fetchData">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
          <el-input v-model="keyword" placeholder="搜索食材名称" clearable style="width: 200px; margin-right: 12px;" @keyup.enter="fetchData" />
          <el-button type="primary" @click="fetchData">
            <el-icon><Search /></el-icon>搜索
          </el-button>
        </div>
      </div>

      <el-tabs v-model="activeZoneTab" type="card" @tab-change="handleTabChange">
        <el-tab-pane label="干货区" name="dry">
          <InventoryTable :data="zoneData.dry" :zones="zones" @update="handleUpdate" @transfer="handleTransfer" />
        </el-tab-pane>
        <el-tab-pane label="冷藏区" name="refrigerated">
          <InventoryTable :data="zoneData.refrigerated" :zones="zones" @update="handleUpdate" @transfer="handleTransfer" />
        </el-tab-pane>
        <el-tab-pane label="冷冻区" name="frozen">
          <InventoryTable :data="zoneData.frozen" :zones="zones" @update="handleUpdate" @transfer="handleTransfer" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="editDialogVisible" title="编辑食材库存" width="500px">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="食材名称">
          <span>{{ editForm.name }}</span>
        </el-form-item>
        <el-form-item label="当前库存">
          <el-input-number v-model="editForm.stock" :min="0" :precision="2" />
          <span style="margin-left: 8px;">{{ editForm.unit }}</span>
        </el-form-item>
        <el-form-item label="安全库存">
          <el-input-number v-model="editForm.safety_stock" :min="0" :precision="2" />
          <span style="margin-left: 8px;">{{ editForm.unit }}</span>
        </el-form-item>
        <el-form-item label="库区">
          <el-select v-model="editForm.warehouse_zone">
            <el-option v-for="zone in zones" :key="zone.code" :label="zone.name" :value="zone.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="供应商">
          <el-input v-model="editForm.supplier" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="transferDialogVisible" title="库区调拨" width="400px">
      <el-form :model="transferForm" label-width="100px">
        <el-form-item label="食材名称">
          <span>{{ transferForm.name }}</span>
        </el-form-item>
        <el-form-item label="当前库区">
          <el-tag>{{ getZoneName(transferForm.current_zone) }}</el-tag>
        </el-form-item>
        <el-form-item label="调拨数量">
          <el-input-number v-model="transferForm.qty" :min="1" :max="transferForm.max_qty" :precision="2" />
          <span style="margin-left: 8px;">{{ transferForm.unit }}</span>
        </el-form-item>
        <el-form-item label="目标库区">
          <el-select v-model="transferForm.target_zone">
            <el-option v-for="zone in zones.filter(z => z.code !== transferForm.current_zone)" :key="zone.code" :label="zone.name" :value="zone.code" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="transferDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleTransferSave">确认调拨</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, h } from 'vue'
import { Search, Edit, Switch, Warning } from '@element-plus/icons-vue'
import { getInventoryByZone, getWarehouseZones, getIngredientCategories, updateIngredient, updateIngredientZone } from '@/api/inventory'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUser } from '@/utils/auth'

const activeZone = ref('')
const activeZoneTab = ref('dry')
const filterCategory = ref('')
const keyword = ref('')
const categories = ref([])
const zones = ref([])
const inventoryList = ref([])

const editDialogVisible = ref(false)
const editForm = reactive({
  id: 0,
  name: '',
  stock: 0,
  safety_stock: 0,
  unit: '',
  warehouse_zone: '',
  supplier: '',
  remark: ''
})

const transferDialogVisible = ref(false)
const transferForm = reactive({
  id: 0,
  name: '',
  current_zone: '',
  target_zone: '',
  qty: 0,
  max_qty: 0,
  unit: ''
})

const user = getUser()

const zoneData = computed(() => {
  const data = {
    dry: [],
    refrigerated: [],
    frozen: []
  }
  inventoryList.value.forEach(item => {
    if (data[item.warehouse_zone]) {
      data[item.warehouse_zone].push(item)
    }
  })
  return data
})

const getZoneName = (zone) => {
  const names = { dry: '干货区', refrigerated: '冷藏区', frozen: '冷冻区' }
  return names[zone] || zone
}

const handleTabChange = (tab) => {
  activeZone.value = tab
  fetchData()
}

const fetchData = async () => {
  try {
    const params = {}
    if (activeZone.value) params.zone = activeZone.value
    if (filterCategory.value) params.category = filterCategory.value
    if (keyword.value) params.keyword = keyword.value

    const res = await getInventoryByZone(params)
    if (res.success) {
      inventoryList.value = res.data
    }
  } catch (e) {
    ElMessage.error('获取库存数据失败')
  }
}

const fetchZones = async () => {
  try {
    const res = await getWarehouseZones()
    if (res.success) {
      zones.value = res.data
    }
  } catch (e) {
    console.log('Get zones error')
  }
}

const fetchCategories = async () => {
  try {
    const res = await getIngredientCategories()
    if (res.success) {
      categories.value = res.data
    }
  } catch (e) {
    console.log('Get categories error')
  }
}

const handleUpdate = (row) => {
  Object.assign(editForm, {
    id: row.id,
    name: row.name,
    stock: row.stock,
    safety_stock: row.safety_stock,
    unit: row.unit,
    warehouse_zone: row.warehouse_zone,
    supplier: row.supplier,
    remark: row.remark || ''
  })
  editDialogVisible.value = true
}

const handleSave = async () => {
  try {
    const res = await updateIngredient(editForm.id, {
      stock: editForm.stock,
      safety_stock: editForm.safety_stock,
      warehouse_zone: editForm.warehouse_zone,
      supplier: editForm.supplier,
      remark: editForm.remark
    })
    if (res.success) {
      ElMessage.success('保存成功')
      editDialogVisible.value = false
      fetchData()
    }
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

const handleTransfer = (row) => {
  Object.assign(transferForm, {
    id: row.id,
    name: row.name,
    current_zone: row.warehouse_zone,
    target_zone: '',
    qty: row.stock,
    max_qty: row.stock,
    unit: row.unit
  })
  transferDialogVisible.value = true
}

const handleTransferSave = async () => {
  if (!transferForm.target_zone) {
    ElMessage.warning('请选择目标库区')
    return
  }
  try {
    const res = await updateIngredientZone(transferForm.id, {
      warehouse_zone: transferForm.target_zone,
      operator_id: user?.id,
      operator_name: user?.name || user?.username
    })
    if (res.success) {
      ElMessage.success('调拨成功')
      transferDialogVisible.value = false
      fetchData()
    }
  } catch (e) {
    ElMessage.error('调拨失败')
  }
}

const InventoryTable = {
  props: ['data', 'zones'],
  emits: ['update', 'transfer'],
  setup(props, { emit }) {
    const getAlertType = (level) => {
      const types = { normal: 'info', warning: 'warning', critical: 'danger' }
      return types[level] || 'info'
    }

    const getAlertText = (level) => {
      const texts = { normal: '正常', warning: '预警', critical: '严重' }
      return texts[level] || level
    }

    const getZoneName = (zone) => {
      const names = { dry: '干货区', refrigerated: '冷藏区', frozen: '冷冻区' }
      return names[zone] || zone
    }

    return () => h('div', [
      h('el-table', {
        data: props.data,
        style: { width: '100%' },
        stripe: true
      }, [
        h('el-table-column', { prop: 'name', label: '食材名称', minWidth: 120 }),
        h('el-table-column', { prop: 'category', label: '分类', width: 100 }),
        h('el-table-column', { label: '当前库存', width: 120 }, {
          default: ({ row }) => h('span', {
            class: row.alert_level !== 'normal' ? 'text-danger' : ''
          }, `${row.stock} ${row.unit}`)
        }),
        h('el-table-column', { label: '安全库存', width: 100 }, {
          default: ({ row }) => `${row.safety_stock} ${row.unit}`
        }),
        h('el-table-column', { label: '库存状态', width: 100 }, {
          default: ({ row }) => h('el-tag', {
            type: getAlertType(row.alert_level),
            size: 'small'
          }, () => getAlertText(row.alert_level))
        }),
        h('el-table-column', { label: '库存价值', width: 120 }, {
          default: ({ row }) => `¥${Number(row.stock_value).toFixed(2)}`
        }),
        h('el-table-column', { label: '供应商', width: 120, prop: 'supplier' }),
        h('el-table-column', { label: '操作', width: 180, fixed: 'right' }, {
          default: ({ row }) => h('div', { style: { display: 'flex', gap: '8px' } }, [
            h('el-button', {
              type: 'primary',
              size: 'small',
              link: true,
              onClick: () => emit('update', row)
            }, () => [h('el-icon', () => h(Edit)), ' 编辑']),
            h('el-button', {
              type: 'success',
              size: 'small',
              link: true,
              onClick: () => emit('transfer', row)
            }, () => [h('el-icon', () => h(Switch)), ' 调拨'])
          ])
        })
      ]),
      props.data.length === 0 ? h('el-empty', { description: '暂无库存数据' }) : null
    ])
  }
}

onMounted(() => {
  fetchData()
  fetchZones()
  fetchCategories()
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
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.filter-right {
  display: flex;
  align-items: center;
}

.text-danger {
  color: #f56c6c;
  font-weight: 500;
}
</style>
