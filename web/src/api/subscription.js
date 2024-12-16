import request from '@/utils/request'

// 获取订阅列表
export function getSubscriptions() {
  return request({
    url: '/subscription/list',
    method: 'get'
  })
}

// 添加订阅
export function addSubscription(data) {
  return request({
    url: '/subscription/add',
    method: 'post',
    data
  })
}

// 更新订阅
export function updateSubscription(data) {
  return request({
    url: '/subscription/update',
    method: 'put',
    data
  })
}

// 删除订阅
export function deleteSubscription(id) {
  return request({
    url: `/subscription/delete/${id}`,
    method: 'delete'
  })
}

// 获取节点列表
export function getNodes() {
  return request({
    url: '/subscription/nodes',
    method: 'get'
  })
}

// 更新所有订阅
export function updateAll() {
  return request({
    url: '/subscription/update/all',
    method: 'post'
  })
} 