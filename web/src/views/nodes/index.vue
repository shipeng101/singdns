<template>
  <div class="nodes">
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" @click="showAddDialog">添加节点</el-button>
        <el-button @click="refreshNodes">刷新</el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchQuery"
          placeholder="搜索节点"
          clearable
          style="width: 200px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <el-table :data="filteredNodes" style="width: 100%">
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.type.toUpperCase() }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="server" label="服务器" />
      <el-table-column prop="port" label="端口" width="100" />
      <el-table-column prop="latency" label="延迟" width="120">
        <template #default="{ row }">
          <el-tag :type="getLatencyType(row.latency)">
            {{ row.latency }}ms
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" @click="testNode(row)">测试</el-button>
            <el-button size="small" type="primary" @click="editNode(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteNode(row)">删除</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加/编辑节点对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑节点' : '添加节点'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="节点名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入节点名称" />
        </el-form-item>
        <el-form-item label="节点类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择节点类型" style="width: 100%">
            <el-option label="Shadowsocks" value="ss" />
            <el-option label="VMess" value="vmess" />
            <el-option label="Trojan" value="trojan" />
            <el-option label="VLESS" value="vless" />
            <el-option label="Hysteria" value="hysteria" />
            <el-option label="TUIC" value="tuic" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务器" prop="server">
          <el-input v-model="form.server" placeholder="请输入服务器地址" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="needsPassword">
          <el-input v-model="form.password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="UUID" prop="uuid" v-if="needsUUID">
          <el-input v-model="form.uuid" placeholder="请输入UUID" />
        </el-form-item>
        <el-form-item label="加密方式" prop="cipher" v-if="form.type === 'ss'">
          <el-select v-model="form.cipher" placeholder="请选择加密方式" style="width: 100%">
            <el-option label="aes-128-gcm" value="aes-128-gcm" />
            <el-option label="aes-256-gcm" value="aes-256-gcm" />
            <el-option label="chacha20-poly1305" value="chacha20-poly1305" />
          </el-select>
        </el-form-item>
        <el-form-item label="传输协议" prop="network">
          <el-select v-model="form.network" placeholder="请选择传输协议" style="width: 100%">
            <el-option label="TCP" value="tcp" />
            <el-option label="WebSocket" value="ws" />
            <el-option label="HTTP/2" value="http" />
            <el-option label="gRPC" value="grpc" />
          </el-select>
        </el-form-item>
        <el-form-item label="分组" prop="group">
          <el-input v-model="form.group" placeholder="请输入分组名称" />
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
import { getNodes, addNode, updateNode, deleteNode as removeNode, testNode as testNodeApi } from '@/api/subscription'

// 节点列表
const nodes = ref([])
const searchQuery = ref('')

// 过滤后的节点列表
const filteredNodes = computed(() => {
  if (!searchQuery.value) return nodes.value
  const query = searchQuery.value.toLowerCase()
  return nodes.value.filter(node => 
    node.name.toLowerCase().includes(query) ||
    node.server.toLowerCase().includes(query) ||
    node.type.toLowerCase().includes(query)
  )
})

// 表单相关
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const form = ref({
  name: '',
  type: '',
  server: '',
  port: 443,
  password: '',
  uuid: '',
  cipher: '',
  network: 'tcp',
  group: ''
})

// 表单校验规则
const rules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择节点类型', trigger: 'change' }],
  server: [{ required: true, message: '请输入服务器地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  uuid: [{ required: true, message: '请输入UUID', trigger: 'blur' }],
  cipher: [{ required: true, message: '请选择加密方式', trigger: 'change' }]
}

// 是否需要密码
const needsPassword = computed(() => ['ss', 'trojan'].includes(form.value.type))

// 是否需要UUID
const needsUUID = computed(() => ['vmess', 'vless', 'tuic'].includes(form.value.type))

// 获取延迟标签类型
function getLatencyType(latency) {
  if (latency < 100) return 'success'
  if (latency < 200) return 'warning'
  return 'danger'
}

// 刷新节点列表
async function refreshNodes() {
  try {
    const data = await getNodes()
    nodes.value = data
  } catch (error) {
    ElMessage.error('获取节点列表失败')
  }
}

// 显示添加对话框
function showAddDialog() {
  isEdit.value = false
  form.value = {
    name: '',
    type: '',
    server: '',
    port: 443,
    password: '',
    uuid: '',
    cipher: '',
    network: 'tcp',
    group: ''
  }
  dialogVisible.value = true
}

// 编辑节点
function editNode(node) {
  isEdit.value = true
  form.value = { ...node }
  dialogVisible.value = true
}

// 删除节点
async function deleteNode(node) {
  try {
    await ElMessageBox.confirm('确定要删除该节点吗？', '提示', {
      type: 'warning'
    })
    await removeNode(node.id)
    ElMessage.success('删除成功')
    refreshNodes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 测试节点
async function testNode(node) {
  try {
    const latency = await testNodeApi(node.id)
    ElMessage.success(`测试成功：${latency}ms`)
    refreshNodes()
  } catch (error) {
    ElMessage.error('测试失败')
  }
}

// 提交表单
async function submitForm() {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    if (isEdit.value) {
      await updateNode(form.value)
      ElMessage.success('更新成功')
    } else {
      await addNode(form.value)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    refreshNodes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(isEdit.value ? '更新失败' : '添加失败')
    }
  }
}

// 初始化
refreshNodes()
</script>

<style lang="scss" scoped>
.nodes {
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