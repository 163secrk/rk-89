import request from '@/utils/request'

export function getDishes(params) {
  return request({
    url: '/dishes',
    method: 'get',
    params
  })
}

export function getDish(id) {
  return request({
    url: '/dishes/' + id,
    method: 'get'
  })
}

export function createDish(data) {
  return request({
    url: '/dishes',
    method: 'post',
    data
  })
}

export function updateDish(id, data) {
  return request({
    url: '/dishes/' + id,
    method: 'put',
    data
  })
}

export function deleteDish(id) {
  return request({
    url: '/dishes/' + id,
    method: 'delete'
  })
}

export function getDishesByCategory(category) {
  return request({
    url: '/dishes/category/' + category,
    method: 'get'
  })
}
