<template>
  <div class="settings">
    <el-tabs>
      <!-- 基本设置 -->
      <el-tab-pane label="基本设置">
        <el-form
          ref="basicFormRef"
          :model="basicForm"
          label-width="120px"
          class="settings-form"
        >
          <el-form-item label="服务器地址">
            <el-input v-model="basicForm.host" placeholder="请输入服务器地址" />
          </el-form-item>
          <el-form-item label="服务器端口">
            <el-input-number v-model="basicForm.port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="运行模式">
            <el-select v-model="basicForm.mode" placeholder="请选择运行模式">
              <el-option label="调试" value="debug" />
              <el-option label="生产" value="release" />
            </el-select>
          </el-form-item>
          <el-form-item label="日志级别">
            <el-select v-model="basicForm.logLevel" placeholder="请选择日志级别">
              <el-option label="调试" value="debug" />
              <el-option label="信息" value="info" />
              <el-option label="警告" value="warn" />
              <el-option label="错误" value="error" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveBasicSettings">保存设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- DNS设置 -->
      <el-tab-pane label="DNS设置">
        <el-form
          ref="dnsFormRef"
          :model="dnsForm"
          label-width="120px"
          class="settings-form"
        >
          <el-form-item label="监听地址">
            <el-input v-model="dnsForm.listen" placeholder="请输入监听地址" />
          </el-form-item>
          <el-form-item label="上游DNS">
            <div v-for="(dns, index) in dnsForm.upstream" :key="index" class="dns-item">
              <el-input v-model="dnsForm.upstream[index]" placeholder="请输入DNS服务器地址">
                <template #append>
                  <el-button @click="removeDNS(index, 'upstream')">删除</el-button>
                </template>
              </el-input>
            </div>
            <el-button @click="addDNS('upstream')">添加DNS服务器</el-button>
          </el-form-item>
          <el-form-item label="国内DNS">
            <div v-for="(dns, index) in dnsForm.chinaDNS" :key="index" class="dns-item">
              <el-input v-model="dnsForm.chinaDNS[index]" placeholder="请输入DNS服务器地址">
                <template #append>
                  <el-button @click="removeDNS(index, 'chinaDNS')">删除</el-button>
                </template>
              </el-input>
            </div>
            <el-button @click="addDNS('chinaDNS')">添加DNS服务器</el-button>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveDNSSettings">保存设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 代理设置 -->
      <el-tab-pane label="代理设置">
        <el-form
          ref="proxyFormRef"
          :model="proxyForm"
          label-width="120px"
          class="settings-form"
        >
          <el-form-item label="代理模式">
            <el-select v-model="proxyForm.mode" placeholder="请选择代理模式">
              <el-option label="规则模式" value="rule" />
              <el-option label="全局模式" value="global" />
              <el-option label="直连模式" value="direct" />
            </el-select>
          </el-form-item>
          <el-form-item label="监听地址">
            <el-input v-model="proxyForm.listen" placeholder="请输入监听地址" />
          </el-form-item>
          <el-form-item label="混合端口">
            <el-input-number v-model="proxyForm.port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="Socks端口">
            <el-input-number v-model="proxyForm.socksPort" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="HTTP端口">
            <el-input-number v-model="proxyForm.httpPort" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item label="认证设置">
            <el-input v-model="proxyForm.username" placeholder="用户名">
              <template #prepend>用户名</template>
            </el-input>
            <el-input v-model="proxyForm.password" placeholder="密码" show-password>
              <template #prepend>密码</template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveProxySettings">保存设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 界面设置 -->
      <el-tab-pane label="界面设置">
        <el-form
          ref="uiFormRef"
          :model="uiForm"
          label-width="120px"
          class="settings-form"
        >
          <el-form-item label="主题">
            <el-select v-model="uiForm.theme" placeholder="请选择主题">
              <el-option label="浅色" value="light" />
              <el-option label="深色" value="dark" />
            </el-select>
          </el-form-item>
          <el-form-item label="语言">
            <el-select v-model="uiForm.language" placeholder="请选择语言">
              <el-option label="简体中文" value="zh-CN" />
              <el-option label="English" value="en-US" />
            </el-select>
          </el-form-item>
          <el-form-item label="仪表盘">
            <el-select v-model="uiForm.dashboard" placeholder="请选择仪表盘">
              <el-option label="Yacd" value="yacd" />
              <el-option label="MetaCubeXD" value="metacubexd" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveUISettings">保存设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getSystemSettings,
  getDNSSettings,
  getProxySettings,
  getInterfaceSettings,
  updateSystemSettings,
  updateDNSSettings,
  updateProxySettings,
  updateInterfaceSettings
} from '@/api/system'

// 基本设置
const basicForm = ref({
  host: '',
  port: 8080,
  mode: 'debug',
  logLevel: 'info'
})

// DNS设置
const dnsForm = ref({
  listen: '',
  upstream: [],
  chinaDNS: []
})

// 代理设置
const proxyForm = ref({
  mode: 'rule',
  listen: '',
  port: 1080,
  socksPort: 1081,
  httpPort: 1082,
  username: '',
  password: ''
})

// 界面设置
const uiForm = ref({
  theme: 'light',
  language: 'zh-CN',
  dashboard: 'yacd'
})

// 加载设置
async function loadSettings() {
  try {
    const [system, dns, proxy, ui] = await Promise.all([
      getSystemSettings(),
      getDNSSettings(),
      getProxySettings(),
      getInterfaceSettings()
    ])

    basicForm.value = {
      host: system.host,
      port: system.port,
      mode: system.mode,
      logLevel: system.log_level
    }

    dnsForm.value = {
      listen: dns.listen,
      upstream: dns.upstream,
      chinaDNS: dns.china_dns
    }

    proxyForm.value = {
      mode: proxy.mode,
      listen: proxy.listen,
      port: proxy.port,
      socksPort: proxy.socks_port,
      httpPort: proxy.http_port,
      username: proxy.username,
      password: proxy.password
    }

    uiForm.value = {
      theme: ui.theme,
      language: ui.language,
      dashboard: ui.dashboard
    }
  } catch (error) {
    ElMessage.error('加载设置失败')
  }
}

// 保存基本设置
async function saveBasicSettings() {
  try {
    await updateSystemSettings(basicForm.value)
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 保存DNS设置
async function saveDNSSettings() {
  try {
    await updateDNSSettings(dnsForm.value)
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 保存代理设置
async function saveProxySettings() {
  try {
    await updateProxySettings(proxyForm.value)
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 保存界面设置
async function saveUISettings() {
  try {
    await updateInterfaceSettings(uiForm.value)
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// DNS服务器管理
function addDNS(type) {
  dnsForm.value[type].push('')
}

function removeDNS(index, type) {
  dnsForm.value[type].splice(index, 1)
}

// 初始化
loadSettings()
</script>

<style lang="scss" scoped>
.settings {
  .settings-form {
    max-width: 600px;
    margin: 20px auto;
  }

  .dns-item {
    margin-bottom: 10px;

    &:last-child {
      margin-bottom: 20px;
    }
  }

  :deep(.el-input-group__append) {
    padding: 0;
    
    .el-button {
      margin: 0;
      border: none;
      border-radius: 0;
    }
  }
}
</style> 