<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <!-- 系统状态 -->
      <el-col :span="6">
        <el-card class="status-card">
          <template #header>
            <div class="card-header">
              <span>系统状态</span>
            </div>
          </template>
          <div class="status-content">
            <div class="status-item">
              <span class="label">运行时间</span>
              <span class="value">{{ uptime }}</span>
            </div>
            <div class="status-item">
              <span class="label">CPU 使用率</span>
              <span class="value">{{ cpuUsage }}%</span>
            </div>
            <div class="status-item">
              <span class="label">内存使用率</span>
              <span class="value">{{ memoryUsage }}%</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 流量统计 -->
      <el-col :span="6">
        <el-card class="traffic-card">
          <template #header>
            <div class="card-header">
              <span>流量统计</span>
            </div>
          </template>
          <div class="traffic-content">
            <div class="traffic-item">
              <span class="label">上传速度</span>
              <span class="value">{{ uploadSpeed }}</span>
            </div>
            <div class="traffic-item">
              <span class="label">下载速度</span>
              <span class="value">{{ downloadSpeed }}</span>
            </div>
            <div class="traffic-item">
              <span class="label">总上传量</span>
              <span class="value">{{ totalUpload }}</span>
            </div>
            <div class="traffic-item">
              <span class="label">总下载量</span>
              <span class="value">{{ totalDownload }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 节点状态 -->
      <el-col :span="12">
        <el-card class="node-card">
          <template #header>
            <div class="card-header">
              <span>节点状态</span>
            </div>
          </template>
          <el-table :data="nodes" style="width: 100%">
            <el-table-column prop="name" label="节点名称" />
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="latency" label="延迟" width="100">
              <template #default="{ row }">
                {{ row.latency }}ms
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'connected' ? 'success' : 'danger'">
                  {{ row.status === 'connected' ? '已连接' : '未连接' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 系统日志 -->
    <el-row :gutter="20" class="mt-4">
      <el-col :span="24">
        <el-card class="log-card">
          <template #header>
            <div class="card-header">
              <span>系统日志</span>
              <el-button-group>
                <el-button size="small" @click="clearLogs">清空</el-button>
                <el-button size="small" type="primary" @click="refreshLogs">刷新</el-button>
              </el-button-group>
            </div>
          </template>
          <div class="log-content">
            <div v-for="(log, index) in logs" :key="index" class="log-item">
              <span class="time">{{ log.time }}</span>
              <span :class="['level', log.level]">{{ log.level }}</span>
              <span class="message">{{ log.message }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { getSystemStatus, getTrafficStats, getNodeStatus, getSystemLogs } from '@/api/system'

// 系统状态
const uptime = ref('0:00:00')
const cpuUsage = ref(0)
const memoryUsage = ref(0)

// 流量统计
const uploadSpeed = ref('0 B/s')
const downloadSpeed = ref('0 B/s')
const totalUpload = ref('0 B')
const totalDownload = ref('0 B')

// 节点状态
const nodes = ref([])

// 系统日志
const logs = ref([])

// 定时器
let timer = null

// 格式化时间
function formatUptime(seconds) {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

// 格式化流量
function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 更新系统状态
async function updateSystemStatus() {
  try {
    const data = await getSystemStatus()
    uptime.value = formatUptime(data.uptime)
    cpuUsage.value = data.cpu_usage.toFixed(1)
    memoryUsage.value = data.memory_usage.toFixed(1)
  } catch (error) {
    console.error('Failed to get system status:', error)
  }
}

// 更新流量统计
async function updateTrafficStats() {
  try {
    const data = await getTrafficStats()
    uploadSpeed.value = formatBytes(data.up_speed) + '/s'
    downloadSpeed.value = formatBytes(data.down_speed) + '/s'
    totalUpload.value = formatBytes(data.up_total)
    totalDownload.value = formatBytes(data.down_total)
  } catch (error) {
    console.error('Failed to get traffic stats:', error)
  }
}

// 更新节点状态
async function updateNodeStatus() {
  try {
    const data = await getNodeStatus()
    nodes.value = data
  } catch (error) {
    console.error('Failed to get node status:', error)
  }
}

// 更新系统日志
async function updateLogs() {
  try {
    const data = await getSystemLogs()
    logs.value = data
  } catch (error) {
    console.error('Failed to get system logs:', error)
  }
}

// 清空日志
function clearLogs() {
  logs.value = []
}

// 刷新日志
function refreshLogs() {
  updateLogs()
}

// 启动定时更新
function startTimer() {
  timer = setInterval(() => {
    updateSystemStatus()
    updateTrafficStats()
    updateNodeStatus()
  }, 1000)
}

onMounted(() => {
  updateSystemStatus()
  updateTrafficStats()
  updateNodeStatus()
  updateLogs()
  startTimer()
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style lang="scss" scoped>
.dashboard {
  .mt-4 {
    margin-top: 16px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .status-content,
  .traffic-content {
    .status-item,
    .traffic-item {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;

      &:last-child {
        margin-bottom: 0;
      }

      .label {
        color: var(--el-text-color-secondary);
      }

      .value {
        font-weight: bold;
      }
    }
  }

  .log-content {
    height: 300px;
    overflow-y: auto;
    font-family: monospace;

    .log-item {
      padding: 4px 0;
      border-bottom: 1px solid var(--el-border-color-lighter);

      &:last-child {
        border-bottom: none;
      }

      .time {
        color: var(--el-text-color-secondary);
        margin-right: 8px;
      }

      .level {
        padding: 2px 4px;
        border-radius: 4px;
        margin-right: 8px;
        font-size: 12px;

        &.debug {
          background-color: var(--el-color-info-light-9);
          color: var(--el-color-info);
        }

        &.info {
          background-color: var(--el-color-primary-light-9);
          color: var(--el-color-primary);
        }

        &.warn {
          background-color: var(--el-color-warning-light-9);
          color: var(--el-color-warning);
        }

        &.error {
          background-color: var(--el-color-danger-light-9);
          color: var(--el-color-danger);
        }
      }

      .message {
        color: var(--el-text-color-primary);
      }
    }
  }
}
</style> 