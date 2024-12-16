import request from '@/utils/request'

// Get system settings
export function getSystemSettings() {
  return request({
    url: '/system/settings',
    method: 'get'
  })
}

// Update system settings
export function updateSystemSettings(data) {
  return request({
    url: '/system/settings',
    method: 'put',
    data
  })
}

// Get system status
export function getSystemStatus() {
  return request({
    url: '/system/status',
    method: 'get'
  })
}

// Get traffic statistics
export function getTrafficStats() {
  return request({
    url: '/system/traffic',
    method: 'get'
  })
}

// Get node status
export function getNodeStatus() {
  return request({
    url: '/system/nodes/status',
    method: 'get'
  })
}

// Get system logs
export function getSystemLogs(params) {
  return request({
    url: '/system/logs',
    method: 'get',
    params
  })
}

// Get DNS settings
export function getDNSSettings() {
  return request({
    url: '/system/dns',
    method: 'get'
  })
}

// Update DNS settings
export function updateDNSSettings(data) {
  return request({
    url: '/system/dns',
    method: 'put',
    data
  })
}

// Get proxy settings
export function getProxySettings() {
  return request({
    url: '/system/proxy',
    method: 'get'
  })
}

// Update proxy settings
export function updateProxySettings(data) {
  return request({
    url: '/system/proxy',
    method: 'put',
    data
  })
}

// Get interface settings
export function getInterfaceSettings() {
  return request({
    url: '/system/interface',
    method: 'get'
  })
}

// Update interface settings
export function updateInterfaceSettings(data) {
  return request({
    url: '/system/interface',
    method: 'put',
    data
  })
} 