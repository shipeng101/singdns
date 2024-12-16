import request from '@/utils/request'

// 获取规则列表
export function getRules() {
  return request({
    url: '/rules/list',
    method: 'get'
  })
}

// 添加规则
export function addRule(data) {
  return request({
    url: '/rules/add',
    method: 'post',
    data
  })
}

// 更新规则
export function updateRule(data) {
  return request({
    url: '/rules/update',
    method: 'put',
    data
  })
}

// 删除规则
export function deleteRule(id) {
  return request({
    url: `/rules/delete/${id}`,
    method: 'delete'
  })
}

// 更新规则集
export function updateRuleSet(id) {
  return request({
    url: `/rules/update/${id}`,
    method: 'post'
  })
}

// 更新所有规则集
export function updateAllRuleSets() {
  return request({
    url: '/rules/update/all',
    method: 'post'
  })
}

// 获取预设规则集
export function getPresetRuleSets() {
  return request({
    url: '/rules/presets',
    method: 'get'
  })
}

// 导入规则集
export function importRuleSet(data) {
  return request({
    url: '/rules/import',
    method: 'post',
    data
  })
}

// 导出规则集
export function exportRuleSet(id) {
  return request({
    url: `/rules/export/${id}`,
    method: 'get',
    responseType: 'blob'
  })
} 