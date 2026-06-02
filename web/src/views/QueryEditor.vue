<template>
  <div class="query-editor">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-select v-model="selectedConn" placeholder="选择连接" style="width: 180px" size="small">
          <el-option
            v-for="c in connStore.connections"
            :key="c.id"
            :label="c.name"
            :value="c.id"
          />
        </el-select>
        <el-select v-model="selectedDb" placeholder="选择数据库" style="width: 150px" size="small" @change="onDbChange">
          <el-option v-for="db in connStore.databases" :key="db" :label="db" :value="db" />
        </el-select>
      </div>
      <div class="toolbar-right">
        <el-button type="primary" size="small" @click="executeQuery" :loading="executing">
          <el-icon><CaretRight /></el-icon> 执行
        </el-button>
        <el-button size="small" @click="formatSQL">
          <el-icon><Document /></el-icon> 格式化
        </el-button>
        <el-button size="small" @click="clearEditor">
          <el-icon><Delete /></el-icon> 清空
        </el-button>
        <el-button size="small" @click="exportResult" :disabled="!result">
          <el-icon><Download /></el-icon> 导出
        </el-button>
      </div>
    </div>

    <!-- 编辑器区域 -->
    <div class="editor-container">
      <!-- SQL 编辑器 -->
      <div class="editor-panel">
        <div class="panel-header">
          <span class="panel-title">SQL 编辑器</span>
          <span class="panel-hint">Ctrl+Enter 执行</span>
        </div>
        <div ref="editorRef" class="monaco-editor"></div>
      </div>

      <!-- 调整大小 -->
      <div class="resize-bar" @mousedown="startResize">
        <div class="resize-dots">
          <span></span><span></span><span></span>
        </div>
      </div>

      <!-- 结果面板 -->
      <div class="result-panel" :style="{ height: resultHeight + 'px' }">
        <div class="panel-header">
          <span class="panel-title">执行结果</span>
          <span v-if="result" class="result-stats">
            <template v-if="result.isSelect">
              <span class="stat-item">
                <el-icon><Grid /></el-icon>
                {{ result.rows?.length || 0 }} 行
              </span>
            </template>
            <template v-else>
              <span class="stat-item success">
                <el-icon><CircleCheck /></el-icon>
                {{ result.rowsAffected }} 行受影响
              </span>
            </template>
            <span class="stat-item">
              <el-icon><Timer /></el-icon>
              {{ result.duration }}ms
            </span>
          </span>
        </div>

        <div class="result-content">
          <!-- 错误信息 -->
          <div v-if="result?.error" class="error-display">
            <div class="error-icon">
              <el-icon><WarningFilled /></el-icon>
            </div>
            <div class="error-message">{{ result.error }}</div>
          </div>

          <!-- 查询结果表格 -->
          <div v-else-if="result?.isSelect && result.rows" class="result-table-wrapper">
            <el-table :data="result.rows" style="width: 100%" max-height="100%" border size="small">
              <el-table-column
                v-for="(col, idx) in result.columns"
                :key="idx"
                :prop="String(idx)"
                :label="col"
                min-width="120"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <span :class="{ 'null-cell': row[idx] === null }">
                    {{ row[idx] === null ? 'NULL' : row[idx] }}
                  </span>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <!-- 执行成功 -->
          <div v-else-if="result && !result.isSelect" class="success-display">
            <div class="success-icon">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="success-text">查询执行成功</div>
          </div>

          <!-- 空状态 -->
          <div v-else class="empty-display">
            <div class="empty-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="empty-text">执行查询查看结果</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { useConnectionStore, api } from '../stores'
import { ElMessage } from 'element-plus'
import * as monaco from 'monaco-editor'

const connStore = useConnectionStore()
const editorRef = ref(null)
const selectedConn = ref('')
const selectedDb = ref('')
const executing = ref(false)
const result = ref(null)
const resultHeight = ref(280)

let editor = null

onMounted(async () => {
  await nextTick()
  initEditor()
  if (connStore.currentConn) {
    selectedConn.value = connStore.currentConn.id
    loadDatabases()
  }
})

function initEditor() {
  if (!editorRef.value) return

  editor = monaco.editor.create(editorRef.value, {
    value: 'SELECT * FROM ',
    language: 'sql',
    theme: 'vs-dark',
    minimap: { enabled: false },
    fontSize: 13,
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    lineNumbers: 'on',
    scrollBeyondLastLine: false,
    automaticLayout: true,
    tabSize: 2,
    wordWrap: 'on',
    padding: { top: 12, bottom: 12 },
    renderLineHighlight: 'line',
    lineDecorationsWidth: 0,
    lineNumbersMinChars: 3,
    cursorBlinking: 'smooth',
    cursorSmoothCaretAnimation: 'on',
    smoothScrolling: true,
    contextmenu: true,
    bracketPairColorization: { enabled: true }
  })

  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
    executeQuery()
  })
}

watch(selectedConn, async (val) => {
  if (val) {
    await loadDatabases()
  }
})

async function loadDatabases() {
  if (!selectedConn.value) return
  try {
    await connStore.fetchDatabases(selectedConn.value)
    if (connStore.databases.length > 0 && !selectedDb.value) {
      selectedDb.value = connStore.databases[0]
    }
  } catch (e) {
    ElMessage.error('加载数据库列表失败')
  }
}

function onDbChange() {
  connStore.setCurrentDb(selectedDb.value)
}

async function executeQuery() {
  if (!selectedConn.value) {
    ElMessage.warning('请先选择连接')
    return
  }

  const sql = editor?.getSelection()?.isEmpty()
    ? editor.getValue()
    : editor.getModel().getValueInRange(editor.getSelection())

  if (!sql?.trim()) {
    ElMessage.warning('请输入查询语句')
    return
  }

  executing.value = true
  try {
    const res = await api.post(`/api/conn/${selectedConn.value}/query`, {
      sql: sql.trim(),
      database: selectedDb.value
    })
    result.value = res.data
    if (res.data.error) {
      ElMessage.error('查询执行失败')
    }
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '查询失败')
  } finally {
    executing.value = false
  }
}

function formatSQL() {
  if (!editor) return
  let sql = editor.getValue()
  sql = sql.replace(/\s+/g, ' ').trim()
  const keywords = [
    'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'ORDER BY', 'GROUP BY',
    'HAVING', 'LIMIT', 'OFFSET', 'JOIN', 'LEFT JOIN', 'RIGHT JOIN',
    'INNER JOIN', 'ON', 'INSERT INTO', 'VALUES', 'UPDATE', 'SET',
    'DELETE FROM', 'CREATE TABLE', 'ALTER TABLE', 'DROP TABLE'
  ]
  keywords.forEach(kw => {
    const regex = new RegExp(`\\b${kw}\\b`, 'gi')
    sql = sql.replace(regex, '\n' + kw)
  })
  editor.setValue(sql.trim())
}

function clearEditor() {
  editor?.setValue('')
  result.value = null
}

function exportResult() {
  if (!result.value?.isSelect || !result.value?.rows) return

  const data = result.value
  let csv = data.columns.join(',') + '\n'
  data.rows.forEach(row => {
    csv += row.map(v => v === null ? '' : `"${String(v).replace(/"/g, '""')}"`).join(',') + '\n'
  })

  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'query_result.csv'
  a.click()
  URL.revokeObjectURL(url)
}

let resizing = false
function startResize(e) {
  resizing = true
  const startY = e.clientY
  const startHeight = resultHeight.value
  const onMove = (e) => {
    if (!resizing) return
    resultHeight.value = Math.max(100, Math.min(600, startHeight - (e.clientY - startY)))
  }
  const onUp = () => {
    resizing = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}
</script>

<style scoped>
.query-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* 编辑器容器 */
.editor-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

/* 面板头部 */
.panel-header {
  padding: 10px 16px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.panel-hint {
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

/* 编辑器面板 */
.editor-panel {
  flex: 1;
  min-height: 200px;
  display: flex;
  flex-direction: column;
}

.monaco-editor {
  flex: 1;
  min-height: 0;
}

/* 调整大小 */
.resize-bar {
  height: 8px;
  background: var(--bg-tertiary);
  cursor: row-resize;
  display: flex;
  align-items: center;
  justify-content: center;
  border-top: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
  transition: background var(--transition-fast);
}

.resize-bar:hover {
  background: var(--accent-cyan-dim);
}

.resize-dots {
  display: flex;
  gap: 3px;
}

.resize-dots span {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: var(--text-muted);
}

/* 结果面板 */
.result-panel {
  display: flex;
  flex-direction: column;
  min-height: 100px;
}

.result-content {
  flex: 1;
  overflow: auto;
}

.result-stats {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--text-muted);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.stat-item.success {
  color: var(--accent-green);
}

/* 错误显示 */
.error-display {
  padding: 20px;
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.error-icon {
  color: var(--accent-red);
  font-size: 18px;
  flex-shrink: 0;
  margin-top: 2px;
}

.error-message {
  color: var(--accent-red);
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 成功显示 */
.success-display {
  padding: 30px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.success-icon {
  color: var(--accent-green);
  font-size: 32px;
}

.success-text {
  color: var(--text-secondary);
  font-size: 14px;
}

/* 空状态 */
.empty-display {
  padding: 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 32px;
  opacity: 0.3;
}

.empty-text {
  font-size: 13px;
}

/* 结果表格 */
.result-table-wrapper {
  height: 100%;
}

/* NULL 单元格 */
.null-cell {
  color: var(--text-muted);
  font-style: italic;
  opacity: 0.6;
}
</style>
