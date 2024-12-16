import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    // 在请求发送之前做一些处理
    return config
  },
  error => {
    // 处理请求错误
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    // 处理响应数据
    const res = response.data

    // 如果响应成功
    if (res.code === 0) {
      return res.data
    }

    // 如果响应失败
    ElMessage.error(res.message || '请求失败')
    return Promise.reject(new Error(res.message || '请求失败'))
  },
  error => {
    // 处理响应错误
    console.error('Response error:', error)
    ElMessage.error(error.message || '请求失败')
    return Promise.reject(error)
  }
)

export default request 