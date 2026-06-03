<template>
  <div class="explorer">
    <!-- 头部操作区 -->
    <div class="explorer-header">
      <el-button type="primary" size="small" @click="showAddDialog = true" class="add-btn">
        <el-icon><Plus /></el-icon>
        <span>新建连接</span>
      </el-button>
    </div>

    <!-- 树形结构 -->
    <div class="tree-container">
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="defaultProps"
        @node-click="handleNodeClick"
        @node-dblclick="handleNodeDblClick"
        @node-contextmenu="handleContextMenu"
        highlight-current
        lazy
        :load="loadNode"
        class="custom-tree"
      >
        <template #default="{ node, data }">
          <div class="tree-node">
            <div class="node-icon" :class="data.type">
              <el-icon v-if="data.type === 'connection'">
                <Connection />
              </el-icon>
              <el-icon v-else-if="data.type === 'database'">
                <Coin />
              </el-icon>
              <el-icon v-else-if="data.type === 'table'">
                <Grid />
              </el-icon>
              <el-icon v-else>
                <Document />
              </el-icon>
            </div>
            <span class="node-label">{{ node.label }}</span>
            <span v-if="data.type === 'connection'" class="node-badge" :style="{ background: data.color || 'var(--accent-cyan)' }">
              {{ data.connType }}
            </span>
          </div>
        </template>
      </el-tree>

      <!-- 空状态 -->
      <div v-if="treeData.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 48 48" fill="none">
            <rect x="6" y="10" width="36" height="28" rx="3" stroke="currentColor" stroke-width="2" stroke-dasharray="4 2"/>
            <path d="M18 24h12M24 18v12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <p class="empty-text">暂无数据库连接</p>
        <p class="empty-hint">点击上方按钮创建</p>
      </div>
    </div>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <div
        v-show="contextMenu.visible"
        class="context-menu"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      >
        <div class="menu-header">操作菜单</div>
        <div class="menu-item" @click="handleMenuAction('query')">
          <el-icon><Edit /></el-icon>
          <span>新建查询</span>
          <span class="menu-shortcut">Ctrl+Q</span>
        </div>
        <div class="menu-item" @click="handleMenuAction('data')">
          <el-icon><Grid /></el-icon>
          <span>查看数据</span>
          <span class="menu-shortcut">Enter</span>
        </div>
        <div class="menu-item" @click="handleMenuAction('structure')">
          <el-icon><Document /></el-icon>
          <span>查看结构</span>
        </div>
        <div class="menu-divider"></div>
        <div class="menu-item" @click="handleMenuAction('export')">
          <el-icon><Download /></el-icon>
          <span>导出数据</span>
        </div>
        <div class="menu-item" @click="handleMenuAction('copy')">
          <el-icon><CopyDocument /></el-icon>
          <span>复制名称</span>
          <span class="menu-shortcut">Ctrl+C</span>
        </div>
        <div class="menu-divider"></div>
        <div class="menu-item danger" @click="handleMenuAction('drop')">
          <el-icon><Delete /></el-icon>
          <span>删除表</span>
        </div>
      </div>
    </Teleport>

    <!-- 新建连接对话框 -->
    <el-dialog v-model="showAddDialog" title="新建数据库连接" width="520px">
      <el-form :model="connForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="connForm.name" placeholder="MySQL 生产环境" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="connForm.type" style="width: 100%">
            <el-option label="MySQL" value="mysql">
              <span style="color: #00d4ff;">●</span> MySQL
            </el-option>
            <el-option label="PostgreSQL" value="postgresql">
              <span style="color: #a855f7;">●</span> PostgreSQL
            </el-option>
            <el-option label="SQLite" value="sqlite">
              <span style="color: #10b981;">●</span> SQLite
            </el-option>
            <el-option label="SQL Server" value="sqlserver">
              <span style="color: #f59e0b;">●</span> SQL Server
            </el-option>
            <el-option label="Oracle" value="oracle">
              <span style="color: #ef4444;">●</span> Oracle
            </el-option>
          </el-select>
        </el-form-item>
        <template v-if="connForm.type !== 'sqlite'">
          <el-form-item label="主机">
            <el-input v-model="connForm.host" placeholder="localhost" />
          </el-form-item>
          <el-form-item label="端口">
            <el-input-number v-model="connForm.port" :min="1" :max="65535" style="width: 100%" />
          </el-form-item>
          <el-form-item label="用户名">
            <el-input v-model="connForm.username" placeholder="root" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="connForm.password" type="password" show-password placeholder="请输入密码" />
          </el-form-item>
          <el-form-item label="数据库">
            <el-input v-model="connForm.database" placeholder="可选，默认连接的数据库" />
          </el-form-item>
        </template>
        <el-form-item v-else label="文件路径">
          <el-input v-model="connForm.fileName" placeholder="/path/to/database.db" />
        </el-form-item>
        <el-form-item label="颜色">
          <div class="color-options">
            <div
              v-for="color in colorOptions"
              :key="color"
              class="color-option"
              :class="{ active: connForm.color === color }"
              :style="{ background: color }"
              @click="connForm.color = color"
            ></div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testConn">
          <el-icon><Connection /></el-icon> 测试连接
        </el-button>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConn">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useConnectionStore } from '../stores'
import { api } from '../stores'
import { ElMessage, ElMessageBox } from 'element-plus'

const emit = defineEmits(['navigate'])
const connStore = useConnectionStore()
const treeRef = ref(null)
const showAddDialog = ref(false)
const contextMenu = reactive({ visible: false, x: 0, y: 0, data: null })

const colorOptions = [
  '#00d4ff', '#a855f7', '#10b981', '#f59e0b', '#ef4444',
  '#ec4899', '#06b6d4', '#8b5cf6', '#14b8a6', '#f97316'
]

const defaultProps = {
  children: 'children',
  label: 'label',
  isLeaf: 'isLeaf'
}

const defaultPortMap = {
  mysql: 3306,
  postgresql: 5432,
  sqlserver: 1433,
  oracle: 1521
}

const connForm = reactive({
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

const treeData = ref([])

onMounted(() => {
  loadTree()
  document.addEventListener('click', hideContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', hideContextMenu)
})

async function loadTree() {
  await connStore.fetchConnections()
  treeData.value = connStore.connections.map(c => ({
    id: c.id,
    label: c.name,
    type: 'connection',
    color: c.color,
    connId: c.id,
    connType: c.type,
    isLeaf: false
  }))
}

async function loadNode(node, resolve) {
  const data = node.data
  if (data.type === 'connection') {
    try {
      const dbs = await connStore.fetchDatabases(data.connId)
      resolve(dbs.map(db => ({
        id: `${data.connId}-${db}`,
        label: db,
        type: 'database',
        connId: data.connId,
        dbName: db,
        connType: data.connType,
        isLeaf: false
      })))
    } catch (e) {
      resolve([])
      ElMessage.error('加载数据库列表失败')
    }
  } else if (data.type === 'database') {
    try {
      const tables = await connStore.fetchTables(data.connId, data.dbName)
      resolve(tables.map(t => ({
        id: `${data.connId}-${data.dbName}-${t}`,
        label: t,
        type: 'table',
        connId: data.connId,
        dbName: data.dbName,
        tableName: t,
        connType: data.connType,
        isLeaf: true
      })))
    } catch (e) {
      resolve([])
      ElMessage.error('加载表列表失败')
    }
  } else {
    resolve([])
  }
}

function setContext(data) {
  if (data.connId) {
    connStore.setCurrentConn(connStore.connections.find(c => c.id === data.connId))
  }
  if (data.dbName) {
    connStore.setCurrentDb(data.dbName)
  }
  if (data.tableName) {
    connStore.setCurrentTable(data.tableName)
  }
}

function handleNodeClick(data) {
  setContext(data)
}

function handleNodeDblClick(data) {
  setContext(data)
  if (data.type === 'table') {
    emit('navigate', 'table-data')
  } else if (data.type === 'database') {
    emit('navigate', 'query')
  }
}

function handleContextMenu(event, data) {
  event.preventDefault()
  contextMenu.visible = true
  contextMenu.x = event.clientX
  contextMenu.y = event.clientY
  contextMenu.data = data
  setContext(data)
}

function hideContextMenu() {
  contextMenu.visible = false
}

function handleMenuAction(action) {
  const data = contextMenu.data
  if (!data) return
  hideContextMenu()

  switch (action) {
    case 'query':
      emit('navigate', 'query')
      break
    case 'data':
      emit('navigate', 'table-data')
      break
    case 'structure':
      emit('navigate', 'table-structure')
      break
    case 'export':
      emit('navigate', 'import-export')
      break
    case 'copy':
      navigator.clipboard?.writeText(data.tableName || data.dbName || data.label)
      ElMessage.success('已复制到剪贴板')
      break
    case 'drop':
      dropTable(data)
      break
  }
}

async function dropTable(data) {
  if (!data.tableName) return
  try {
    await ElMessageBox.confirm(
      `确定要删除表 "${data.tableName}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await api.post(`/api/conn/${data.connId}/query`, {
      sql: `DROP TABLE \`${data.tableName}\``,
      database: data.dbName
    })
    ElMessage.success('表已删除')
    loadTree()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.error || '删除失败')
    }
  }
}

function testConn() {
  ElMessage.info('测试连接功能开发中...')
}

async function saveConn() {
  try {
    const payload = { ...connForm }
    if (payload.type === 'sqlite') {
      payload.host = ''
      payload.port = 0
      payload.username = ''
      payload.password = ''
      payload.database = ''
    }
    await connStore.addConnection(payload)
    ElMessage.success('连接已保存')
    showAddDialog.value = false
    loadTree()
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}
</script>

<style scoped>
.explorer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.explorer-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.add-btn {
  width: 100%;
  height: 36px;
  font-size: 13px;
  letter-spacing: 0.05em;
}

/* 树形容器 */
.tree-container {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

/* 自定义树 */
.custom-tree {
  --el-tree-node-hover-bg-color: var(--bg-hover);
  background: transparent !important;
}

.custom-tree :deep(.el-tree-node__content) {
  height: 36px;
  padding: 0 8px;
  border-radius: var(--radius-sm);
  margin: 1px 0;
}

.custom-tree :deep(.el-tree-node__content:hover) {
  background: var(--bg-hover);
}

.custom-tree :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background: var(--accent-cyan-dim);
}

/* 树节点 */
.tree-node {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.node-icon {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  font-size: 14px;
  flex-shrink: 0;
}

.node-icon.connection {
  color: var(--accent-cyan);
  background: var(--accent-cyan-dim);
}

.node-icon.database {
  color: var(--accent-green);
  background: var(--accent-green-dim);
}

.node-icon.table {
  color: var(--accent-amber);
  background: var(--accent-amber-dim);
}

.node-label {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-badge {
  font-size: 9px;
  font-weight: 600;
  color: white;
  padding: 2px 6px;
  border-radius: 3px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  opacity: 0.9;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--text-muted);
}

.empty-icon {
  width: 48px;
  height: 48px;
  margin-bottom: 12px;
  opacity: 0.3;
}

.empty-text {
  font-size: 14px;
  margin-bottom: 4px;
}

.empty-hint {
  font-size: 12px;
  opacity: 0.6;
}

/* 右键菜单 */
.context-menu {
  position: fixed;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  z-index: 9999;
  min-width: 200px;
  padding: 4px;
  animation: fadeIn 0.15s ease-out;
}

.menu-header {
  padding: 8px 12px 6px;
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--text-primary);
  transition: all var(--transition-fast);
}

.menu-item:hover {
  background: var(--bg-hover);
}

.menu-item.danger {
  color: var(--accent-red);
}

.menu-item.danger:hover {
  background: var(--accent-red-dim);
}

.menu-shortcut {
  margin-left: auto;
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.menu-divider {
  height: 1px;
  background: var(--border-subtle);
  margin: 4px 8px;
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
