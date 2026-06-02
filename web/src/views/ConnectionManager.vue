<template>
  <div class="connection-manager">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">连接管理</h2>
        <span class="page-subtitle">管理数据库连接配置</span>
      </div>
      <el-button type="primary" @click="showDialog = true">
        <el-icon><Plus /></el-icon> 新建连接
      </el-button>
    </div>

    <!-- 连接列表 -->
    <div class="connection-grid" v-if="connStore.connections.length > 0">
      <div
        v-for="conn in connStore.connections"
        :key="conn.id"
        class="connection-card"
        :style="{ '--accent': conn.color || 'var(--accent-cyan)' }"
      >
        <div class="card-header">
          <div class="card-icon">
            <svg viewBox="0 0 24 24" fill="none">
              <rect x="2" y="4" width="20" height="16" rx="2" stroke="currentColor" stroke-width="1.5"/>
              <ellipse cx="12" cy="10" rx="6" ry="2.5" stroke="currentColor" stroke-width="1"/>
              <path d="M6 10v4c0 1.2 2.7 2.5 6 2.5s6-1.3 6-2.5v-4" stroke="currentColor" stroke-width="1"/>
            </svg>
          </div>
          <div class="card-title-section">
            <h3 class="card-title">{{ conn.name }}</h3>
            <span class="card-type" :style="{ background: conn.color }">{{ conn.type }}</span>
          </div>
        </div>

        <div class="card-body">
          <div class="info-row">
            <span class="info-label">主机</span>
            <span class="info-value">{{ conn.type === 'sqlite' ? conn.fileName : conn.host }}</span>
          </div>
          <div class="info-row" v-if="conn.type !== 'sqlite'">
            <span class="info-label">端口</span>
            <span class="info-value mono">{{ conn.port }}</span>
          </div>
          <div class="info-row" v-if="conn.type !== 'sqlite'">
            <span class="info-label">用户</span>
            <span class="info-value">{{ conn.username }}</span>
          </div>
          <div class="info-row" v-if="conn.database">
            <span class="info-label">数据库</span>
            <span class="info-value mono">{{ conn.database }}</span>
          </div>
        </div>

        <div class="card-footer">
          <el-button size="small" @click="testConnection(conn)" :loading="testingId === conn.id">
            <el-icon><Connection /></el-icon> 测试
          </el-button>
          <el-button size="small" @click="editConnection(conn)">
            <el-icon><Edit /></el-icon> 编辑
          </el-button>
          <el-button size="small" type="danger" @click="deleteConnection(conn)">
            <el-icon><Delete /></el-icon> 删除
          </el-button>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <div class="empty-illustration">
        <svg viewBox="0 0 120 80" fill="none">
          <rect x="10" y="15" width="100" height="50" rx="5" stroke="currentColor" stroke-width="1.5" stroke-dasharray="4 2"/>
          <circle cx="35" cy="40" r="8" stroke="currentColor" stroke-width="1.5"/>
          <circle cx="60" cy="40" r="8" stroke="currentColor" stroke-width="1.5"/>
          <circle cx="85" cy="40" r="8" stroke="currentColor" stroke-width="1.5"/>
          <path d="M43 40h9M68 40h9" stroke="currentColor" stroke-width="1.5"/>
        </svg>
      </div>
      <h3 class="empty-title">暂无连接配置</h3>
      <p class="empty-desc">创建你的第一个数据库连接开始使用</p>
      <el-button type="primary" @click="showDialog = true">
        <el-icon><Plus /></el-icon> 创建连接
      </el-button>
    </div>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="showDialog"
      :title="editing ? '编辑连接' : '新建连接'"
      width="520px"
      @closed="resetForm"
    >
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="MySQL 生产环境" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" @change="onTypeChange" style="width: 100%">
            <el-option label="MySQL" value="mysql" />
            <el-option label="PostgreSQL" value="postgresql" />
            <el-option label="SQLite" value="sqlite" />
            <el-option label="SQL Server" value="sqlserver" />
            <el-option label="Oracle" value="oracle" />
          </el-select>
        </el-form-item>
        <template v-if="form.type !== 'sqlite'">
          <el-form-item label="主机">
            <el-input v-model="form.host" placeholder="localhost" />
          </el-form-item>
          <el-form-item label="端口">
            <el-input-number v-model="form.port" :min="1" :max="65535" style="width: 100%" />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="form.username" placeholder="root" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" />
          </el-form-item>
          <el-form-item label="数据库">
            <el-input v-model="form.database" placeholder="可选" />
          </el-form-item>
        </template>
        <el-form-item v-else label="文件路径">
          <el-input v-model="form.fileName" placeholder="/path/to/database.db" />
        </el-form-item>
        <el-form-item label="颜色">
          <div class="color-options">
            <div
              v-for="color in colorOptions"
              :key="color"
              class="color-option"
              :class="{ active: form.color === color }"
              :style="{ background: color }"
              @click="form.color = color"
            ></div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testCurrent">测试连接</el-button>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useConnectionStore } from '../stores'
import { ElMessage, ElMessageBox } from 'element-plus'

const connStore = useConnectionStore()
const showDialog = ref(false)
const editing = ref(false)
const editingId = ref('')
const testingId = ref('')

const colorOptions = [
  '#00d4ff', '#a855f7', '#10b981', '#f59e0b', '#ef4444',
  '#ec4899', '#06b6d4', '#8b5cf6', '#14b8a6', '#f97316'
]

const defaultPortMap = {
  mysql: 3306,
  postgresql: 5432,
  sqlserver: 1433,
  oracle: 1521
}

const form = reactive({
  name: '',
  type: 'mysql',
  host: 'localhost',
  port: 3306,
  username: 'root',
  password: '',
  database: '',
  fileName: '',
  color: '#00d4ff'
})

function onTypeChange(type) {
  form.port = defaultPortMap[type] || 3306
}

function resetForm() {
  Object.assign(form, {
    name: '',
    type: 'mysql',
    host: 'localhost',
    port: 3306,
    username: 'root',
    password: '',
    database: '',
    fileName: '',
    color: '#00d4ff'
  })
  editing.value = false
  editingId.value = ''
}

function editConnection(row) {
  editing.value = true
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    type: row.type,
    host: row.host,
    port: row.port,
    username: row.username,
    password: row.password,
    database: row.database,
    fileName: row.fileName || '',
    color: row.color || '#00d4ff'
  })
  showDialog.value = true
}

async function save() {
  try {
    if (editing.value) {
      await connStore.updateConnection({ ...form, id: editingId.value })
      ElMessage.success('连接已更新')
    } else {
      await connStore.addConnection({ ...form })
      ElMessage.success('连接已创建')
    }
    showDialog.value = false
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '操作失败')
  }
}

async function testConnection(row) {
  testingId.value = row.id
  try {
    const res = await connStore.testConnection(row.id)
    if (res.success) {
      ElMessage.success('连接成功')
    } else {
      ElMessage.error(res.error || '连接失败')
    }
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '测试失败')
  } finally {
    testingId.value = ''
  }
}

async function testCurrent() {
  if (!editing.value) {
    ElMessage.warning('请先保存连接后再测试')
    return
  }
  await testConnection({ id: editingId.value })
}

async function deleteConnection(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除连接 "${row.name}" 吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await connStore.deleteConnection(row.id)
    ElMessage.success('连接已删除')
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  connStore.fetchConnections()
})
</script>

<style scoped>
.connection-manager {
  height: 100%;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.page-title {
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.05em;
}

.page-subtitle {
  font-size: 13px;
  color: var(--text-muted);
}

/* 连接网格 */
.connection-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

/* 连接卡片 */
.connection-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-normal);
  position: relative;
}

.connection-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--accent);
  opacity: 0.8;
}

.connection-card:hover {
  border-color: var(--border-hover);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md), 0 0 30px color-mix(in srgb, var(--accent) 10%, transparent);
}

.card-header {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  border-bottom: 1px solid var(--border-subtle);
}

.card-icon {
  width: 40px;
  height: 40px;
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-icon svg {
  width: 22px;
  height: 22px;
}

.card-title-section {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-type {
  font-size: 10px;
  font-weight: 600;
  color: white;
  padding: 2px 8px;
  border-radius: 3px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.card-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-row {
  display: flex;
  align-items: center;
  font-size: 12px;
}

.info-label {
  width: 50px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.info-value {
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-value.mono {
  font-family: var(--font-mono);
}

.card-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  gap: 8px;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: var(--text-muted);
}

.empty-illustration {
  width: 120px;
  height: 80px;
  margin-bottom: 20px;
  color: var(--text-muted);
  opacity: 0.2;
}

.empty-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--text-secondary);
}

.empty-desc {
  font-size: 13px;
  margin-bottom: 24px;
}

/* 颜色选择器 */
.color-options {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.color-option {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all var(--transition-fast);
}

.color-option:hover {
  transform: scale(1.1);
}

.color-option.active {
  border-color: white;
  box-shadow: 0 0 0 2px var(--bg-primary), 0 0 10px currentColor;
}
</style>
