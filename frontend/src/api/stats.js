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

export function getCleanPlateStats(params) {
  return request({
    url: '/stats/clean-plate',
    method: 'get',
    params
  })
}

export function getCleanPlateDepartmentRanking(params) {
  return request({
    url: '/stats/clean-plate/department-ranking',
    method: 'get',
    params
  })
}

export function getCleanPlateUserRanking(params) {
  return request({
    url: '/stats/clean-plate/user-ranking',
    method: 'get',
    params
  })
}

export function getCleanPlateTrend(params) {
  return request({
    url: '/stats/clean-plate/trend',
    method: 'get',
    params
  })
}

export function getDepartments() {
  return request({
    url: '/stats/departments',
    method: 'get'
  })
}
