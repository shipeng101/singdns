<template>
  <div class="rules">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" @click="showAddDialog">添加规则</el-button>
        <el-button @click="refreshRules">刷新</el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchQuery"
          placeholder="搜索规则"
          clearable
          style="width: 200px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <el-table :data="filteredRules" style="width: 100%">
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.type.toUpperCase() }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="value" label="规则值" />
      <el-table-column prop="target" label="目标" width="120">
        <template #default="{ row }">
          <el-tag :type="getTargetType(row.target)">
            {{ row.target.toUpperCase() }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" type="primary" @click="editRule(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteRule(row)">删除</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加/编辑规则对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑规则' : '添加规则'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="规则类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择规则类型" style="width: 100%">
            <el-option label="域名" value="domain" />
            <el-option label="域名后缀" value="domain_suffix" />
            <el-option label="域名关键字" value="domain_keyword" />
            <el-option label="域名正则" value="domain_regex" />
            <el-option label="IP CIDR" value="ip_cidr" />
            <el-option label="GeoIP" value="geoip" />
            <el-option label="GeoSite" value="geosite" />
            <el-option label="端口" value="port" />
            <el-option label="端口范围" value="port_range" />
            <el-option label="进程名" value="process_name" />
            <el-option label="进程路径" value="process_path" />
            <el-option label="协议" value="protocol" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则值" prop="value">
          <el-input v-model="form.value" :placeholder="getValuePlaceholder(form.type)" />
        </el-form-item>
        <el-form-item label="目标" prop="target">
          <el-select v-model="form.target" placeholder="请选择目标" style="width: 100%">
            <el-option label="代理" value="proxy" />
            <el-option label="直连" value="direct" />
            <el-option label="拦截" value="block" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitForm">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getRules, addRule, updateRule, deleteRule as removeRule } from '@/api/rules'

// 规则列表
const rules = ref([])
const searchQuery = ref('')

// 过滤后的规则列表
const filteredRules = computed(() => {
  if (!searchQuery.value) return rules.value
  const query = searchQuery.value.toLowerCase()
  return rules.value.filter(rule => 
    rule.type.toLowerCase().includes(query) ||
    rule.value.toLowerCase().includes(query) ||
    rule.target.toLowerCase().includes(query)
  )
})

// 表单相关
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const form = ref({
  type: '',
  value: '',
  target: ''
})

// 表单校验规则
const formRules = {
  type: [{ required: true, message: '请选择规则类型', trigger: 'change' }],
  value: [{ required: true, message: '请输入规则值', trigger: 'blur' }],
  target: [{ required: true, message: '请选择目标', trigger: 'change' }]
}

// 获取目标标签类型
function getTargetType(target) {
  const types = {
    proxy: 'primary',
    direct: 'success',
    block: 'danger'
  }
  return types[target] || ''
}

// 获取规则值占位符
function getValuePlaceholder(type) {
  const placeholders = {
    domain: '例如：example.com',
    domain_suffix: '例如：.example.com',
    domain_keyword: '例如：example',
    domain_regex: '例如：^example\\.com$',
    ip_cidr: '例如：192.168.1.0/24',
    geoip: '例如：cn',
    geosite: '例如：cn',
    port: '例如：80',
    port_range: '例如：1000-2000',
    process_name: '例如：chrome.exe',
    process_path: '例如：/usr/bin/chrome',
    protocol: '例如：tcp'
  }
  return placeholders[type] || '请输入规则值'
}

// 刷新规则列表
async function refreshRules() {
  try {
    const data = await getRules()
    rules.value = data
  } catch (error) {
    ElMessage.error('获取规则列表失败')
  }
}

// 显示添加对话框
function showAddDialog() {
  isEdit.value = false
  form.value = {
    type: '',
    value: '',
    target: ''
  }
  dialogVisible.value = true
}

// 编辑规则
function editRule(rule) {
  isEdit.value = true
  form.value = { ...rule }
  dialogVisible.value = true
}

// 删除规则
async function deleteRule(rule) {
  try {
    await ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
      type: 'warning'
    })
    await removeRule(rule.id)
    ElMessage.success('删除成功')
    refreshRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 提交表单
async function submitForm() {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    if (isEdit.value) {
      await updateRule(form.value)
      ElMessage.success('更新成功')
    } else {
      await addRule(form.value)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    refreshRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(isEdit.value ? '更新失败' : '添加失败')
    }
  }
}

// 初始化
refreshRules()
</script>

<style lang="scss" scoped>
.rules {
  .toolbar {
    margin-bottom: 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;

    .toolbar-left {
      display: flex;
      gap: 10px;
    }
  }

  :deep(.el-tag) {
    text-transform: uppercase;
  }
}
</style> 