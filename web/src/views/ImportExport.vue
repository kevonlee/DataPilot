<template>
  <div class="import-export">
    <el-tabs v-model="activeTab" class="main-tabs">
      <!-- 导出 -->
      <el-tab-pane label="导出数据" name="export">
        <div class="tab-content">
          <div class="section-card">
            <div class="section-header">
              <h3 class="section-title">导出配置</h3>
            </div>
            <div class="section-body">
              <el-form label-width="80px">
                <el-form-item label="连接">
                  <el-select v-model="exportForm.connId" placeholder="选择连接" style="width: 100%">
                    <el-option v-for="c in connStore.connections" :key="c.id" :label="c.name" :value="c.id" />
                  </el-select>
                </el-form-item>
                <el-form-item label="数据库">
                  <el-select v-model="exportForm.database" placeholder="选择数据库" style="width: 100%" @change="loadExportTables">
                    <el-option v-for="db in exportDatabases" :key="db" :label="db" :value="db" />
                  </el-select>
                </el-form-item>
                <el-form-item label="表">
                  <el-select v-model="exportForm.table" placeholder="选择表" style="width: 100%">
                    <el-option v-for="t in exportTables" :key="t" :label="t" :value="t" />
                  </el-select>
                </el-form-item>
                <el-form-item label="格式">
                  <el-radio-group v-model="exportForm.format">
                    <el-radio label="csv">CSV</el-radio>
                    <el-radio label="json">JSON</el-radio>
                    <el-radio label="sql">SQL INSERT</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item label="自定义 SQL">
                  <el-input
                    v-model="exportForm.sql"
                    type="textarea"
                    :rows="3"
                    placeholder="可选：输入自定义查询语句，留空则导出全表数据"
                  />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="doExport" :loading="exporting">
                    <el-icon><Download /></el-icon> 开始导出
                  </el-button>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- 导入 -->
      <el-tab-pane label="导入数据" name="import">
        <div class="tab-content">
          <div class="section-card">
            <div class="section-header">
              <h3 class="section-title">导入配置</h3>
            </div>
            <div class="section-body">
              <el-form label-width="80px">
                <el-form-item label="连接">
                  <el-select v-model="importForm.connId" placeholder="选择连接" style="width: 100%">
                    <el-option v-for="c in connStore.connections" :key="c.id" :label="c.name" :value="c.id" />
                  </el-select>
                </el-form-item>
                <el-form-item label="数据库">
                  <el-select v-model="importForm.database" placeholder="选择数据库" style="width: 100%">
                    <el-option v-for="db in importDatabases" :key="db" :label="db" :value="db" />
                  </el-select>
                </el-form-item>
                <el-form-item label="类型">
                  <el-radio-group v-model="importForm.type">
                    <el-radio label="sql">SQL 文件</el-radio>
                    <el-radio label="csv">CSV 文件</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item label="文件">
                  <el-upload
                    :auto-upload="false"
                    :limit="1"
                    :on-change="handleFileChange"
                    accept=".sql,.csv,.txt"
                    class="upload-area"
                  >
                    <div class="upload-trigger">
                      <el-icon class="upload-icon"><Upload /></el-icon>
                      <div class="upload-text">点击选择文件</div>
                      <div class="upload-hint">支持 .sql, .csv, .txt 格式</div>
                    </div>
                  </el-upload>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="doImport" :loading="importing">
                    <el-icon><Upload /></el-icon> 开始导入
                  </el-button>
                </el-form-item>
              </el-form>
            </div>
          </div>

          <div class="section-card" style="margin-top: 16px;">
            <div class="section-header">
              <h3 class="section-title">执行 SQL 脚本</h3>
            </div>
            <div class="section-body">
              <el-form label-width="80px">
                <el-form-item label="SQL 脚本">
                  <el-input
                    v-model="importForm.sql"
                    type="textarea"
                    :rows="6"
                    placeholder="粘贴 SQL 脚本..."
                    class="sql-textarea"
                  />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="executeScript" :loading="executing">
                    <el-icon><CaretRight /></el-icon> 执行脚本
                  </el-button>
                </el-form-item>
              </el-form>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useConnectionStore, api } from '../stores'
import { ElMessage } from 'element-plus'

const connStore = useConnectionStore()
const activeTab = ref('export')
const exporting = ref(false)
const importing = ref(false)
const executing = ref(false)

const exportDatabases = ref([])
const exportTables = ref([])
const importDatabases = ref([])

const exportForm = reactive({
  connId: '',
  database: '',
  table: '',
  format: 'csv',
  sql: ''
})

const importForm = reactive({
  connId: '',
  database: '',
  type: 'sql',
  sql: '',
  file: null
})

watch(() => exportForm.connId, async (val) => {
  if (val) {
    const res = await api.get(`/api/conn/${val}/databases`)
    exportDatabases.value = res.data
  }
})

watch(() => importForm.connId, async (val) => {
  if (val) {
    const res = await api.get(`/api/conn/${val}/databases`)
    importDatabases.value = res.data
  }
})

async function loadExportTables() {
  if (!exportForm.connId || !exportForm.database) return
  const res = await api.get(`/api/conn/${exportForm.connId}/databases/${exportForm.database}/tables`)
  exportTables.value = res.data
}

async function doExport() {
  if (!exportForm.connId) {
    ElMessage.warning('请先选择连接')
    return
  }
  exporting.value = true
  try {
    let sql = exportForm.sql
    if (!sql && exportForm.table) {
      sql = `SELECT * FROM \`${exportForm.table}\``
    }
    if (!sql) {
      ElMessage.warning('请选择表或输入查询')
      return
    }

    const queryRes = await api.post(`/api/conn/${exportForm.connId}/query`, {
      sql,
      database: exportForm.database
    })

    const result = queryRes.data
    if (result.error) {
      ElMessage.error(result.error)
      return
    }

    let content, filename, mimeType
    switch (exportForm.format) {
      case 'csv':
        content = convertToCSV(result)
        filename = `${exportForm.table || 'query'}.csv`
        mimeType = 'text/csv'
        break
      case 'json':
        content = convertToJSON(result)
        filename = `${exportForm.table || 'query'}.json`
        mimeType = 'application/json'
        break
      case 'sql':
        content = convertToSQL(result, exportForm.table)
        filename = `${exportForm.table || 'query'}.sql`
        mimeType = 'text/plain'
        break
    }

    downloadFile(content, filename, mimeType)
    ElMessage.success('导出完成')
  } catch (e) {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

function convertToCSV(result) {
  if (!result.columns || !result.rows) return ''
  let csv = result.columns.join(',') + '\n'
  result.rows.forEach(row => {
    csv += row.map(v => v === null ? '' : `"${String(v).replace(/"/g, '""')}"`).join(',') + '\n'
  })
  return csv
}

function convertToJSON(result) {
  if (!result.columns || !result.rows) return '[]'
  const data = result.rows.map(row => {
    const obj = {}
    result.columns.forEach((col, i) => { obj[col] = row[i] })
    return obj
  })
  return JSON.stringify(data, null, 2)
}

function convertToSQL(result, tableName) {
  if (!result.columns || !result.rows) return ''
  let sql = ''
  result.rows.forEach(row => {
    const cols = result.columns.map(c => `\`${c}\``).join(', ')
    const vals = row.map(v => v === null ? 'NULL' : typeof v === 'string' ? `'${v.replace(/'/g, "\\'")}'` : v).join(', ')
    sql += `INSERT INTO \`${tableName || 'table'}\` (${cols}) VALUES (${vals});\n`
  })
  return sql
}

function downloadFile(content, filename, mimeType) {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function handleFileChange(file) {
  importForm.file = file.raw
}

async function doImport() {
  if (!importForm.connId || !importForm.database) {
    ElMessage.warning('请先选择连接和数据库')
    return
  }
  if (!importForm.file) {
    ElMessage.warning('请先选择文件')
    return
  }

  importing.value = true
  try {
    const content = await readFile(importForm.file)
    if (importForm.type === 'sql') {
      const statements = content.split(';').filter(s => s.trim())
      let success = 0, failed = 0
      for (const sql of statements) {
        if (!sql.trim()) continue
        try {
          await api.post(`/api/conn/${importForm.connId}/query`, {
            sql: sql.trim(),
            database: importForm.database
          })
          success++
        } catch { failed++ }
      }
      ElMessage.success(`导入完成：成功 ${success} 条，失败 ${failed} 条`)
    } else {
      ElMessage.warning('CSV 导入请使用 SQL 脚本方式')
    }
  } catch (e) {
    ElMessage.error('导入失败')
  } finally {
    importing.value = false
  }
}

async function executeScript() {
  if (!importForm.connId || !importForm.database) {
    ElMessage.warning('请先选择连接和数据库')
    return
  }
  if (!importForm.sql?.trim()) {
    ElMessage.warning('请输入 SQL 脚本')
    return
  }

  executing.value = true
  try {
    const statements = importForm.sql.split(';').filter(s => s.trim())
    let success = 0, failed = 0
    for (const sql of statements) {
      if (!sql.trim()) continue
      try {
        await api.post(`/api/conn/${importForm.connId}/query`, {
          sql: sql.trim(),
          database: importForm.database
        })
        success++
      } catch { failed++ }
    }
    ElMessage.success(`执行完成：成功 ${success} 条，失败 ${failed} 条`)
  } catch (e) {
    ElMessage.error('执行失败')
  } finally {
    executing.value = false
  }
}

function readFile(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => resolve(e.target.result)
    reader.onerror = reject
    reader.readAsText(file)
  })
}

onMounted(() => {
  if (connStore.currentConn) {
    exportForm.connId = connStore.currentConn.id
    importForm.connId = connStore.currentConn.id
  }
  if (connStore.currentDb) {
    exportForm.database = connStore.currentDb
    importForm.database = connStore.currentDb
  }
  if (connStore.currentTable) {
    exportForm.table = connStore.currentTable
  }
})
</script>

<style scoped>
.import-export {
  height: 100%;
}

.main-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.main-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.main-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
}

.main-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.tab-content {
  height: 100%;
  overflow: auto;
  padding: 16px;
}

/* 区块卡片 */
.section-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.section-header {
  padding: 14px 20px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-subtle);
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.section-body {
  padding: 20px;
}

/* 上传区域 */
.upload-area {
  width: 100%;
}

.upload-trigger {
  padding: 24px;
  border: 2px dashed var(--border-default);
  border-radius: var(--radius-md);
  text-align: center;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.upload-trigger:hover {
  border-color: var(--accent-cyan);
  background: var(--accent-cyan-dim);
}

.upload-icon {
  font-size: 32px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.upload-text {
  font-size: 14px;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.upload-hint {
  font-size: 12px;
  color: var(--text-muted);
}

/* SQL 文本框 */
.sql-textarea :deep(.el-textarea__inner) {
  font-family: var(--font-mono) !important;
  font-size: 13px !important;
}
</style>
