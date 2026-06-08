import request from '@/utils/request'

export function getDashboardStats() {
  return request({
    url: '/stats/dashboard',
    method: 'get'
  })
}

export function getOrderStats(params) {
  return request({
    url: '/stats/orders',
    method: 'get',
    params
  })
}

export function getDishStats() {
  return request({
    url: '/stats/dishes',
    method: 'get'
  })
}
