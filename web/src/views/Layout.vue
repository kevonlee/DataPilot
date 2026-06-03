<template>
  <div class="layout">
    <!-- 侧边栏 -->
    <aside class="sidebar" :style="{ width: sidebarWidth + 'px' }">
      <!-- 侧边栏头部 -->
      <div class="sidebar-header">
        <div class="logo-mini">
          <svg viewBox="0 0 24 24" fill="none" class="logo-svg">
            <rect x="2" y="5" width="20" height="14" rx="2" stroke="currentColor" stroke-width="1.5"/>
            <ellipse cx="12" cy="10" rx="6" ry="2.5" stroke="currentColor" stroke-width="1"/>
            <path d="M6 10v4c0 1.2 2.7 2.5 6 2.5s6-1.3 6-2.5v-4" stroke="currentColor" stroke-width="1"/>
          </svg>
        </div>
        <div class="logo-text">
          <span class="brand">DataPilot</span>
          <span class="version">v1.0</span>
        </div>
      </div>

      <!-- 数据库浏览器 -->
      <div class="sidebar-content">
        <DatabaseExplorer @navigate="addTab" />
      </div>

      <!-- 侧边栏底部 -->
      <div class="sidebar-footer">
        <div class="status-indicator">
          <span class="status-dot"></span>
          <span class="status-text">已连接</span>
        </div>
        <el-dropdown @command="handleCommand" trigger="click">
          <div class="user-btn">
            <div class="user-avatar">
              <el-icon><User /></el-icon>
            </div>
            <span class="user-name">{{ authStore.username }}</span>
            <el-icon class="user-arrow"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="changePassword">
                <el-icon><Key /></el-icon> 修改密码
              </el-dropdown-item>
              <el-dropdown-item command="logout" divided>
                <el-icon><SwitchButton /></el-icon> 退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </aside>

    <!-- 拖拽调整条 -->
    <div class="resize-handle" @mousedown="startResize">
      <div class="resize-line"></div>
    </div>

    <!-- 主内容区 -->
    <main class="main">
      <!-- 顶部标签栏 -->
      <div class="tab-bar">
        <div class="tab-list">
          <div
            v-for="tab in openTabs"
            :key="tab.name"
            class="tab-item"
            :class="{ active: activeTab === tab.name }"
            @click="switchTab(tab.name)"
          >
            <el-icon class="tab-icon"><component :is="tab.icon" /></el-icon>
            <span class="tab-label">{{ tab.title }}</span>
            <el-icon
              v-if="tab.closable"
              class="tab-close"
              @click.stop="removeTab(tab.name)"
            >
              <Close />
            </el-icon>
          </div>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="content-area">
        <router-view />
      </div>
    </main>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="showPasswordDialog" title="修改密码" width="420px">
      <el-form :model="passwordForm" label-width="80px">
        <el-form-item label="旧密码">
          <el-input v-model="passwordForm.oldPassword" type="password" show-password placeholder="请输入旧密码" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="passwordForm.newPassword" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" @click="changePassword">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore, useConnectionStore } from '../stores'
import { api } from '../stores'
import { ElMessage } from 'element-plus'
import DatabaseExplorer from './DatabaseExplorer.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const connStore = useConnectionStore()

// Tab definitions
const tabDefs = {
  connections: { title: '连接管理', icon: 'Connection', closable: false, path: '/connections' },
  query: { title: 'SQL 查询', icon: 'Edit', closable: true, path: '/query' },
  'table-data': { title: '表数据', icon: 'Grid', closable: true, path: '/table-data' },
  'table-structure': { title: '表结构', icon: 'Document', closable: true, path: '/table-structure' },
  'import-export': { title: '导入导出', icon: 'Download', closable: true, path: '/import-export' },
  dashboard: { title: '数据概览', icon: 'DataAnalysis', closable: true, path: '/dashboard' }
}

// Open tabs list - always starts with connections
const openTabs = ref([
  { name: 'connections', ...tabDefs.connections }
])

// Active tab name - derived from current route
const activeTab = computed({
  get() {
    // Find tab matching current route
    const name = Object.keys(tabDefs).find(k => route.path === tabDefs[k].path)
    return name || 'connections'
  },
  set(val) {
    // When set, navigate to that route
    const def = tabDefs[val]
    if (def) router.push(def.path)
  }
})

// Sync route changes to tabs - auto-add tab when route changes
watch(() => route.path, (path) => {
  const name = Object.keys(tabDefs).find(k => path === tabDefs[k].path)
  if (name && !openTabs.value.find(t => t.name === name)) {
    openTabs.value.push({ name, ...tabDefs[name] })
  }
}, { immediate: true })

function switchTab(name) {
  const def = tabDefs[name]
  if (def) router.push(def.path)
}

function addTab(name) {
  if (!tabDefs[name]) return
  if (!openTabs.value.find(t => t.name === name)) {
    openTabs.value.push({ name, ...tabDefs[name] })
  }
  router.push(tabDefs[name].path)
}

function removeTab(name) {
  const idx = openTabs.value.findIndex(t => t.name === name)
  if (idx >= 0 && openTabs.value[idx].closable) {
    openTabs.value.splice(idx, 1)
    // If removing active tab, switch to nearest
    if (activeTab.value === name) {
      const next = openTabs.value[Math.min(idx, openTabs.value.length - 1)]
      if (next) router.push(tabDefs[next.name].path)
    }
  }
}

const sidebarWidth = ref(280)
const showPasswordDialog = ref(false)
const passwordForm = ref({ oldPassword: '', newPassword: '' })

function handleCommand(cmd) {
  if (cmd === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (cmd === 'changePassword') {
    showPasswordDialog.value = true
  }
}

async function changePassword() {
  try {
    await api.post('/api/auth/change-password', passwordForm.value)
    ElMessage.success('密码已修改')
    showPasswordDialog.value = false
    passwordForm.value = { oldPassword: '', newPassword: '' }
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '修改失败')
  }
}

let resizing = false
function startResize(e) {
  resizing = true
  const startX = e.clientX
  const startWidth = sidebarWidth.value
  const onMove = (e) => {
    if (!resizing) return
    sidebarWidth.value = Math.max(220, Math.min(500, startWidth + e.clientX - startX))
  }
  const onUp = () => {
    resizing = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

onMounted(() => {
  connStore.fetchConnections()
})
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--bg-primary);
}

/* 侧边栏 */
.sidebar {
  background: var(--bg-secondary);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  border-right: 1px solid var(--border-subtle);
  position: relative;
}

.sidebar::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 1px;
  height: 100%;
  background: linear-gradient(to bottom, var(--accent-cyan), transparent 30%, transparent 70%, var(--accent-cyan));
  opacity: 0.2;
}

.sidebar-header {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--border-subtle);
  background: linear-gradient(to bottom, rgba(0, 212, 255, 0.02), transparent);
}

.logo-mini {
  width: 32px;
  height: 32px;
  color: var(--accent-cyan);
  filter: drop-shadow(0 0 6px var(--accent-cyan-glow));
}

.logo-svg {
  width: 100%;
  height: 100%;
}

.logo-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.brand {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: 0.1em;
}

.version {
  font-size: 10px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.sidebar-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: var(--text-muted);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent-green);
  box-shadow: 0 0 6px var(--accent-green);
  animation: pulse 2s ease-in-out infinite;
}

.status-text {
  letter-spacing: 0.05em;
}

.user-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.user-btn:hover {
  background: var(--bg-hover);
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--accent-cyan-dim);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent-cyan);
  font-size: 14px;
}

.user-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.user-arrow {
  font-size: 12px;
  color: var(--text-muted);
}

/* 拖拽调整条 */
.resize-handle {
  width: 4px;
  cursor: col-resize;
  position: relative;
  flex-shrink: 0;
  z-index: 10;
}

.resize-handle:hover,
.resize-handle:active {
  background: var(--accent-cyan-dim);
}

.resize-line {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 2px;
  height: 40px;
  border-radius: 1px;
  background: var(--border-default);
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.resize-handle:hover .resize-line {
  opacity: 1;
  background: var(--accent-cyan);
}

/* 主内容区 */
.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

/* 标签栏 */
.tab-bar {
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-subtle);
  padding: 0 8px;
  flex-shrink: 0;
}

.tab-list {
  display: flex;
  gap: 2px;
  overflow-x: auto;
  scrollbar-width: none;
}

.tab-list::-webkit-scrollbar {
  display: none;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 14px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all var(--transition-fast);
  white-space: nowrap;
  user-select: none;
}

.tab-item:hover {
  color: var(--text-secondary);
  background: var(--bg-tertiary);
}

.tab-item.active {
  color: var(--accent-cyan);
  border-bottom-color: var(--accent-cyan);
}

.tab-icon {
  font-size: 14px;
}

.tab-label {
  letter-spacing: 0.03em;
}

.tab-close {
  font-size: 12px;
  opacity: 0;
  transition: opacity var(--transition-fast);
  margin-left: 2px;
}

.tab-item:hover .tab-close {
  opacity: 0.6;
}

.tab-close:hover {
  opacity: 1;
  color: var(--accent-red);
}

/* 内容区域 */
.content-area {
  flex: 1;
  overflow: auto;
  padding: 20px;
  background: var(--bg-primary);
  position: relative;
}

/* 背景网格 */
.content-area::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(var(--border-subtle) 1px, transparent 1px),
    linear-gradient(90deg, var(--border-subtle) 1px, transparent 1px);
  background-size: 80px 80px;
  opacity: 0.3;
  pointer-events: none;
}

.content-area > * {
  position: relative;
  z-index: 1;
}
</style>
