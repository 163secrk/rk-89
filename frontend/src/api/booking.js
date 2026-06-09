import request from '@/utils/request'

export function getBookings(params) {
  return request({
    url: '/bookings',
    method: 'get',
    params
  })
}

export function getBooking(id) {
  return request({
    url: '/bookings/' + id,
    method: 'get'
  })
}

export function createBooking(data) {
  return request({
    url: '/bookings',
    method: 'post',
    data
  })
}

export function updateBooking(id, data) {
  return request({
    url: '/bookings/' + id,
    method: 'put',
    data
  })
}

export function deleteBooking(id) {
  return request({
    url: '/bookings/' + id,
    method: 'delete'
  })
}

export function updateBookingStatus(id, data) {
  return request({
    url: '/bookings/' + id + '/status',
    method: 'put',
    data
  })
}

export function calculateBookingIngredients(data) {
  return request({
    url: '/bookings/calculate',
    method: 'post',
    data
  })
}
