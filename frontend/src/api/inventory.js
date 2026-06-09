import request from '@/utils/request'

export function getInventoryDashboard() {
  return request({
    url: '/inventory/dashboard',
    method: 'get'
  })
}

export function getInventoryByZone(params) {
  return request({
    url: '/inventory/stock',
    method: 'get',
    params
  })
}

export function getWarehouseZones() {
  return request({
    url: '/inventory/zones',
    method: 'get'
  })
}

export function getIngredientCategories() {
  return request({
    url: '/ingredients/categories',
    method: 'get'
  })
}

export function stockInbound(data) {
  return request({
    url: '/inventory/inbound',
    method: 'post',
    data
  })
}

export function stockOutbound(data) {
  return request({
    url: '/inventory/outbound',
    method: 'post',
    data
  })
}

export function calculateMealPlanDemand(data) {
  return request({
    url: '/inventory/calculate-demand',
    method: 'post',
    data
  })
}

export function getStockRecords(params) {
  return request({
    url: '/inventory/records',
    method: 'get',
    params
  })
}

export function getStockAlerts(params) {
  return request({
    url: '/inventory/alerts',
    method: 'get',
    params
  })
}

export function handleStockAlert(id, data) {
  return request({
    url: `/inventory/alerts/${id}`,
    method: 'put',
    data
  })
}

export function getOperationLogs(params) {
  return request({
    url: '/inventory/logs',
    method: 'get',
    params
  })
}

export function updateIngredientZone(id, data) {
  return request({
    url: `/inventory/ingredients/${id}/zone`,
    method: 'put',
    data
  })
}

export function getPurchaseLists(params) {
  return request({
    url: '/purchases',
    method: 'get',
    params
  })
}

export function getPurchaseList(id) {
  return request({
    url: `/purchases/${id}`,
    method: 'get'
  })
}

export function updatePurchaseStatus(id, data) {
  return request({
    url: `/purchases/${id}/status`,
    method: 'put',
    data
  })
}

export function getMealPlans(params) {
  return request({
    url: '/meal-plans',
    method: 'get',
    params
  })
}

export function getIngredients(params) {
  return request({
    url: '/ingredients',
    method: 'get',
    params
  })
}

export function updateIngredient(id, data) {
  return request({
    url: `/ingredients/${id}`,
    method: 'put',
    data
  })
}

export function analyzeZoneInventoryDemand(data) {
  return request({
    url: '/inventory/analyze-zone-demand',
    method: 'post',
    data
  })
}

export function autoReplenish(data) {
  return request({
    url: '/inventory/auto-replenish',
    method: 'post',
    data
  })
}

export function getAutoReplenishmentRecords(params) {
  return request({
    url: '/inventory/auto-replenish-records',
    method: 'get',
    params
  })
}

export function getNotifications(params) {
  return request({
    url: '/inventory/notifications',
    method: 'get',
    params
  })
}

export function markNotificationRead(id) {
  return request({
    url: `/inventory/notifications/${id}/read`,
    method: 'put'
  })
}

export function markAllNotificationsRead() {
  return request({
    url: '/inventory/notifications/read-all',
    method: 'put'
  })
}

export function getUnreadNotificationCount() {
  return request({
    url: '/inventory/notifications/unread-count',
    method: 'get'
  })
}

export function generatePurchaseList(data) {
  return request({
    url: '/purchases/generate',
    method: 'post',
    data
  })
}
