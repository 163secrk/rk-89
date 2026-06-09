import request from '@/utils/request'

export function rechargeMealAllowance(data) {
  return request({
    url: '/meal-allowance/recharge',
    method: 'post',
    data
  })
}

export function getMealAllowanceRecords(params) {
  return request({
    url: '/meal-allowance/records',
    method: 'get',
    params
  })
}

export function getUserMealAllowanceRecords(userId, params) {
  return request({
    url: '/meal-allowance/records/user/' + userId,
    method: 'get',
    params
  })
}

export function getMealAllowanceStats() {
  return request({
    url: '/meal-allowance/stats',
    method: 'get'
  })
}

export function getConsumptionRecords(params) {
  return request({
    url: '/meal-allowance/consumptions',
    method: 'get',
    params
  })
}
