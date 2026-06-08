import request from '@/utils/request'

export function getMealPlans(params) {
  return request({
    url: '/meal-plans',
    method: 'get',
    params
  })
}

export function getMealPlan(id) {
  return request({
    url: '/meal-plans/' + id,
    method: 'get'
  })
}

export function createMealPlan(data) {
  return request({
    url: '/meal-plans',
    method: 'post',
    data
  })
}

export function updateMealPlan(id, data) {
  return request({
    url: '/meal-plans/' + id,
    method: 'put',
    data
  })
}

export function deleteMealPlan(id) {
  return request({
    url: '/meal-plans/' + id,
    method: 'delete'
  })
}

export function getTodayMealPlan(params) {
  return request({
    url: '/meal-plans/today',
    method: 'get',
    params
  })
}
