<template>
  <div class="table-data">
    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-select v-model="selectedConn" placeholder="连接" style="width: 160px" size="small">
          <el-option v-for="c in connStore.connections" :key="c.id" :label="c.name" :value="c.id" />
        </el-select>
        <el-select v-model="selectedDb" placeholder="数据库" style="width: 130px" size="small" @change="onDbChange">
          <el-option v-for="db in databases" :key="db" :label="db" :value="db" />
        </el-select>
        <el-select v-model="selectedTable" placeholder="表" style="width: 130px" size="small" @change="loadData">
          <el-option v-for="t in tables" :key="t" :label="t" :value="t" />
        </el-select>
      </div>
      <div class="toolbar-right">
        <el-button size="small" @click="loadData" :loading="loading">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
        <el-button size="small" @click="showAddRow = true" type="primary">
          <el-icon><Plus /></el-icon> 新增
        </el-button>
        <el-button size="small" @click="exportData">
          <el-icon><Download /></el-icon> 导出
        </el-button>
        <span class="total-info" v-if="total > 0">
          共 <strong>{{ total.toLocaleString() }}</strong> 条记录
        </span>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="data-table-container">
      <el-table
        :data="tableData"
        style="width: 100%"
        border
        size="small"
        @selection-change="handleSelectionChange"
        v-loading="loading"
        :empty-text="'暂无数据'"
      >
        <el-table-column type="selection" width="45" align="center" />
        <el-table-column
          v-for="(col, idx) in columns"
          :key="idx"
          :label="col"
          min-width="140"
          show-overflow-tooltip
        >
          <template #header>
            <span class="col-header">{{ col }}</span>
          </template>
          <template #default="{ row, $index }">
            <div v-if="editingRow === $index && editingCol === col" class="cell-editor">
              <el-input
                v-model="editValue"
                size="small"
                @blur="saveCell($index, col)"
                @keyup.enter="saveCell($index, col)"
                autofocus
              />
            </div>
            <div v-else @dblclick="startEdit($index, col, row[idx])" class="cell-display">
              <span :class="{ 'null-value': row[idx] === null }">
                {{ row[idx] === null ? 'NULL' : row[idx] }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="70" align="center" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" size="small" @click="deleteRow(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页 -->
    <div class="pagination-bar">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[20, 50, 100, 200]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadData"
        @current-change="loadData"
        small
      />
    </div>

    <!-- 新增行对话框 -->
    <el-dialog v-model="showAddRow" title="新增记录" width="600px">
      <el-form label-width="100px">
        <el-form-item v-for="col in columns" :key="col" :label="col">
          <el-input v-model="newRow[col]" :placeholder="col" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddRow = false">取消</el-button>
        <el-button type="primary" @click="insertRow">插入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useConnectionStore, api } from '../stores'
import { ElMessage, ElMessageBox } from 'element-plus'

const connStore = useConnectionStore()
const selectedConn = ref('')
const selectedDb = ref('')
const selectedTable = ref('')
const databases = ref([])
const tables = ref([])
const columns = ref([])
const tableData = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)
const loading = ref(false)
const showAddRow = ref(false)
const newRow = reactive({})
const editingRow = ref(-1)
const editingCol = ref('')
const editValue = ref('')

watch(selectedConn, async (val) => {
  if (val) {
    try {
      databases.value = await connStore.fetchDatabases(val)
    } catch (e) {
      ElMessage.error('加载数据库列表失败')
    }
  }
})

async function onDbChange() {
  selectedTable.value = ''
  tables.value = []
  if (selectedConn.value && selectedDb.value) {
    try {
      tables.value = await connStore.fetchTables(selectedConn.value, selectedDb.value)
    } catch (e) {
      ElMessage.error('加载表列表失败')
    }
  }
}

async function loadData() {
  if (!selectedConn.value || !selectedDb.value || !selectedTable.value) return
  loading.value = true
  try {
    const colRes = await api.get(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/columns`)
    columns.value = colRes.data.map(c => c.name)

    const dataRes = await api.get(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/data`, {
      params: { page: currentPage.value, pageSize: pageSize.value }
    })
    tableData.value = dataRes.data.data?.rows || []
    total.value = dataRes.data.total || 0
  } catch (e) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

function startEdit(rowIndex, col, value) {
  editingRow.value = rowIndex
  editingCol.value = col
  editValue.value = value === null ? '' : String(value)
}

async function saveCell(rowIndex, col) {
  const oldValue = tableData.value[rowIndex][columns.value.indexOf(col)]
  const newValue = editValue.value

  if (String(oldValue) !== String(newValue)) {
    try {
      const where = {}
      columns.value.forEach((c, i) => {
        where[c] = tableData.value[rowIndex][i]
      })
      await api.put(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/data`, {
        where,
        set: { [col]: newValue }
      })
      tableData.value[rowIndex][columns.value.indexOf(col)] = newValue
      ElMessage.success('更新成功')
    } catch (e) {
      ElMessage.error('更新失败')
    }
  }
  editingRow.value = -1
  editingCol.value = ''
}

async function insertRow() {
  try {
    const payload = {}
    Object.keys(newRow).forEach(k => {
      if (newRow[k] !== '') payload[k] = newRow[k]
    })
    await api.post(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/data`, payload)
    ElMessage.success('插入成功')
    showAddRow.value = false
    Object.keys(newRow).forEach(k => newRow[k] = '')
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '插入失败')
  }
}

async function deleteRow(row) {
  try {
    await ElMessageBox.confirm('确定要删除这条记录吗？', '确认删除', { type: 'warning' })
    const where = {}
    columns.value.forEach((c, i) => { where[c] = row[i] })
    await api.delete(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/data`, {
      data: where
    })
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

function handleSelectionChange() {}

async function exportData() {
  try {
    const res = await api.post(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/export`, {
      format: 'csv'
    }, { responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedTable.value}.csv`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    ElMessage.error('导出失败')
  }
}

onMounted(async () => {
  // Initialize from store state
  if (connStore.currentConn) {
    selectedConn.value = connStore.currentConn.id
    try {
      databases.value = await connStore.fetchDatabases(selectedConn.value)
    } catch {}
  }
  if (connStore.currentDb) {
    selectedDb.value = connStore.currentDb
    try {
      tables.value = await connStore.fetchTables(selectedConn.value, selectedDb.value)
    } catch {}
  }
  if (connStore.currentTable) {
    selectedTable.value = connStore.currentTable
    await loadData()
  }
})
</script>

<style scoped>
.table-data {
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
  flex-wrap: wrap;
  gap: 10px;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}

.total-info {
  margin-left: 12px;
  font-size: 12px;
  color: var(--text-muted);
}

.total-info strong {
  color: var(--accent-cyan);
  font-family: var(--font-mono);
}

/* 数据表格 */
.data-table-container {
  flex: 1;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.col-header {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.cell-display {
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  min-height: 22px;
}

.cell-display:hover {
  background: var(--accent-cyan-dim);
}

.null-value {
  color: var(--text-muted);
  font-style: italic;
  opacity: 0.5;
}

.cell-editor {
  margin: -4px -8px;
}

/* 分页 */
.pagination-bar {
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  display: flex;
  justify-content: flex-end;
}
</style>
