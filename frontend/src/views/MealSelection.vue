<template>
  <div class="meal-selection">
    <el-row :gutter="20">
      <el-col :span="18">
        <el-card shadow="hover" class="dishes-card">
          <template #header>
            <div class="card-header">
              <span class="title">智能配餐</span>
              <div class="header-actions">
                <el-input
                  v-model="searchKeyword"
                  placeholder="搜索菜品"
                  style="width: 200px; margin-right: 10px;"
                  :prefix-icon="Search"
                  clearable
                  @keyup.enter="fetchDishes"
                />
                <el-button type="primary" :icon="Refresh" @click="fetchDishes">刷新</el-button>
              </div>
            </div>
          </template>

          <el-tabs v-model="activeCategory" @tab-change="handleCategoryChange">
            <el-tab-pane label="全部" name="all" />
            <el-tab-pane label="主食" name="主食" />
            <el-tab-pane label="热菜" name="热菜" />
            <el-tab-pane label="凉菜" name="凉菜" />
            <el-tab-pane label="汤品" name="汤品" />
            <el-tab-pane label="饮品" name="饮品" />
          </el-tabs>

          <el-row :gutter="20" v-loading="loading">
            <el-col :span="8" v-for="dish in filteredDishes" :key="dish.id">
              <el-card class="dish-card" shadow="hover">
                <div class="dish-image">
                  <img :src="dish.image || 'https://via.placeholder.com/200x150?text=' + encodeURIComponent(dish.name)" :alt="dish.name" />
                  <el-tag v-if="dish.available" type="success" class="status-tag" size="small">在售</el-tag>
                  <el-tag v-else type="danger" class="status-tag" size="small">售罄</el-tag>
                </div>
                <div class="dish-info">
                  <h3 class="dish-name">{{ dish.name }}</h3>
                  <p class="dish-desc">{{ dish.description }}</p>
                  <div class="dish-meta">
                    <span class="dish-calories">{{ dish.calories }} kcal</span>
                    <span class="dish-category">{{ dish.category }}</span>
                  </div>
                  <div class="dish-footer">
                    <span class="dish-price">¥{{ dish.price }}</span>
                    <div class="dish-actions">
                      <el-button
                        type="primary"
                        :icon="Plus"
                        size="small"
                        circle
                        :disabled="!dish.available"
                        @click="addToCart(dish)"
                      />
                    </div>
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>
          <el-empty v-if="filteredDishes.length === 0 && !loading" description="暂无菜品" />
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="hover" class="cart-card">
          <template #header>
            <div class="cart-header">
              <span class="title"><el-icon><ShoppingCart /></el-icon> 我的订单</span>
              <el-badge :value="cartTotalCount" class="cart-badge" />
            </div>
          </template>

          <div class="cart-list" v-if="cart.length > 0">
            <div class="cart-item" v-for="item in cart" :key="item.id">
              <div class="item-info">
                <span class="item-name">{{ item.name }}</span>
                <span class="item-price">¥{{ item.price }}</span>
              </div>
              <div class="item-actions">
                <el-button type="danger" :icon="Minus" size="small" circle @click="decreaseQuantity(item)" />
                <span class="item-quantity">{{ item.quantity }}</span>
                <el-button type="primary" :icon="Plus" size="small" circle @click="increaseQuantity(item)" />
              </div>
              <div class="item-subtotal">小计: ¥{{ (item.price * item.quantity).toFixed(2) }}</div>
            </div>
          </div>
          <el-empty v-else description="购物车为空，快去选购吧" />

          <el-divider v-if="cart.length > 0" />

          <div class="cart-summary" v-if="cart.length > 0">
            <el-row>
              <el-col :span="12">菜品数量:</el-col>
              <el-col :span="12" class="text-right">{{ cartTotalCount }} 份</el-col>
            </el-row>
            <el-row>
              <el-col :span="12">总热量:</el-col>
              <el-col :span="12" class="text-right">{{ cartTotalCalories }} kcal</el-col>
            </el-row>
            <el-row class="total-row">
              <el-col :span="12">合计金额:</el-col>
              <el-col :span="12" class="text-right total-price">¥{{ cartTotalPrice.toFixed(2) }}</el-col>
            </el-row>
          </div>

          <el-form v-if="cart.length > 0" label-position="top" style="margin-top: 15px;">
            <el-form-item label="备注">
              <el-input v-model="remark" type="textarea" :rows="2" placeholder="请输入备注信息" />
            </el-form-item>
          </el-form>

          <div class="cart-footer">
            <el-button type="danger" :icon="Delete" @click="clearCart" :disabled="cart.length === 0">清空</el-button>
            <el-button type="primary" :icon="Check" @click="submitOrder" :loading="submitting" :disabled="cart.length === 0">提交订单</el-button>
          </div>
        </el-card>

        <el-card shadow="hover" style="margin-top: 20px;">
          <template #header>
            <span class="title">今日推荐</span>
          </template>
          <el-table :data="todayMealPlan" size="small">
            <el-table-column prop="meal_type" label="餐次" width="80">
              <template #default="{ row }">
                <el-tag type="success" size="small">{{ row.meal_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="菜品" show-overflow-tooltip>
              <template #default="{ row }">
                {{ (row.dishes || []).map(d => d.name).join('、') }}
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="todayMealPlan.length === 0" description="暂无推荐" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Plus, Minus, ShoppingCart, Delete, Check } from '@element-plus/icons-vue'
import { getDishes } from '@/api/dish'
import { getTodayMealPlan } from '@/api/mealplan'
import { createOrder } from '@/api/order'

const loading = ref(false)
const submitting = ref(false)
const searchKeyword = ref('')
const activeCategory = ref('all')
const dishes = ref([])
const cart = ref([])
const remark = ref('')
const todayMealPlan = ref([])

const filteredDishes = computed(() => {
  let result = dishes.value
  if (activeCategory.value !== 'all') {
    result = result.filter(d => d.category === activeCategory.value)
  }
  if (searchKeyword.value) {
    result = result.filter(d => d.name.includes(searchKeyword.value))
  }
  return result
})

const cartTotalCount = computed(() => {
  return cart.value.reduce((sum, item) => sum + item.quantity, 0)
})

const cartTotalPrice = computed(() => {
  return cart.value.reduce((sum, item) => sum + item.price * item.quantity, 0)
})

const cartTotalCalories = computed(() => {
  return cart.value.reduce((sum, item) => sum + (item.calories || 0) * item.quantity, 0)
})

const fetchDishes = async () => {
  loading.value = true
  try {
    const res = await getDishes({ available: true, keyword: searchKeyword.value })
    dishes.value = res.data?.items || res.data || []
  } catch (e) {
    console.log('Get dishes error')
  } finally {
    loading.value = false
  }
}

const fetchTodayMealPlan = async () => {
  try {
    const res = await getTodayMealPlan()
    todayMealPlan.value = res.data || []
  } catch (e) {
    console.log('Get meal plan error')
  }
}

const handleCategoryChange = () => {
}

const addToCart = (dish) => {
  const existingItem = cart.value.find(item => item.id === dish.id)
  if (existingItem) {
    existingItem.quantity++
  } else {
    cart.value.push({
      id: dish.id,
      name: dish.name,
      price: dish.price,
      calories: dish.calories,
      quantity: 1
    })
  }
  ElMessage.success(`已添加 ${dish.name}`)
}

const increaseQuantity = (item) => {
  item.quantity++
}

const decreaseQuantity = (item) => {
  if (item.quantity > 1) {
    item.quantity--
  } else {
    const index = cart.value.findIndex(i => i.id === item.id)
    if (index > -1) {
      cart.value.splice(index, 1)
    }
  }
}

const clearCart = () => {
  cart.value = []
  remark.value = ''
  ElMessage.success('购物车已清空')
}

const submitOrder = async () => {
  submitting.value = true
  try {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    const today = new Date().toISOString().split('T')[0]
    const orderData = {
      user_id: user.id,
      meal_time: 'lunch',
      meal_date: today,
      items: cart.value.map(item => ({
        dish_id: item.id,
        quantity: item.quantity
      })),
      remark: remark.value
    }
    await createOrder(orderData)
    ElMessage.success('订单提交成功！')
    clearCart()
  } catch (e) {
    console.log('Submit order error')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchDishes()
  fetchTodayMealPlan()
})
</script>

<style scoped>
.meal-selection {
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

.header-actions {
  display: flex;
  align-items: center;
}

.dishes-card {
  min-height: 600px;
}

.dish-card {
  margin-bottom: 20px;
  overflow: hidden;
}

.dish-image {
  position: relative;
  height: 150px;
  overflow: hidden;
  border-radius: 8px;
  margin-bottom: 12px;
}

.dish-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.status-tag {
  position: absolute;
  top: 10px;
  right: 10px;
}

.dish-name {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #303133;
}

.dish-desc {
  font-size: 12px;
  color: #909399;
  margin: 0 0 8px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.dish-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
  margin-bottom: 12px;
}

.dish-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dish-price {
  font-size: 20px;
  font-weight: 600;
  color: #f56c6c;
}

.cart-card {
  position: sticky;
  top: 20px;
}

.cart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.cart-list {
  max-height: 300px;
  overflow-y: auto;
}

.cart-item {
  padding: 12px 0;
  border-bottom: 1px solid #f0f2f5;
}

.item-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.item-name {
  font-weight: 500;
  color: #303133;
}

.item-price {
  color: #f56c6c;
  font-weight: 500;
}

.item-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 6px;
}

.item-quantity {
  min-width: 30px;
  text-align: center;
  font-weight: 500;
}

.item-subtotal {
  text-align: right;
  font-size: 12px;
  color: #909399;
}

.cart-summary {
  padding: 0 10px;
  font-size: 14px;
}

.cart-summary .el-row {
  margin-bottom: 8px;
}

.text-right {
  text-align: right;
}

.total-row {
  font-size: 16px;
  font-weight: 600;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #dcdfe6;
}

.total-price {
  color: #f56c6c;
  font-size: 20px;
}

.cart-footer {
  display: flex;
  justify-content: space-between;
  margin-top: 15px;
  gap: 10px;
}

.cart-footer .el-button {
  flex: 1;
}
</style>
