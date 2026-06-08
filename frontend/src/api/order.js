import request from '@/utils/request'

export function getOrders(params) {
  return request({
    url: '/orders',
    method: 'get',
    params
  })
}

export function getOrder(id) {
  return request({
    url: '/orders/' + id,
    method: 'get'
  })
}

export function createOrder(data) {
  return request({
    url: '/orders',
    method: 'post',
    data
  })
}

export function updateOrder(id, data) {
  return request({
    url: '/orders/' + id,
    method: 'put',
    data
  })
}

export function deleteOrder(id) {
  return request({
    url: '/orders/' + id,
    method: 'delete'
  })
}

export function updateOrderStatus(id, status) {
  return request({
    url: '/orders/' + id + '/status',
    method: 'put',
    data: { status }
  })
}

export function getMyOrders(params) {
  const user = JSON.parse(localStorage.getItem('user') || '{}')
  const userId = user.id
  return request({
    url: '/orders/user/' + userId,
    method: 'get',
    params
  })
}
