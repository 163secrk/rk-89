import request from '@/utils/request'

export function verifyOrder(data) {
  return request({
    url: '/api/verification',
    method: 'post',
    data
  })
}

export function getCurrentMealSession() {
  return request({
    url: '/api/verification/session',
    method: 'get'
  })
}

export function getMealSessions() {
  return request({
    url: '/api/verification/sessions',
    method: 'get'
  })
}

export function getVerificationRecords(params) {
  return request({
    url: '/api/verification/records',
    method: 'get',
    params
  })
}
