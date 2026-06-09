import request from '@/utils/request'

export function verifyOrder(data) {
  return request({
    url: '/verification',
    method: 'post',
    data
  })
}

export function getCurrentMealSession() {
  return request({
    url: '/verification/session',
    method: 'get'
  })
}

export function getMealSessions() {
  return request({
    url: '/verification/sessions',
    method: 'get'
  })
}

export function getVerificationRecords(params) {
  return request({
    url: '/verification/records',
    method: 'get',
    params
  })
}
