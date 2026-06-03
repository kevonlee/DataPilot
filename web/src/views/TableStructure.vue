<template>
  <div class="table-structure">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-select v-model="selectedConn" placeholder="连接" style="width: 160px" size="small">
        <el-option v-for="c in connStore.connections" :key="c.id" :label="c.name" :value="c.id" />
      </el-select>
      <el-select v-model="selectedDb" placeholder="数据库" style="width: 130px" size="small" @change="onDbChange">
        <el-option v-for="db in databases" :key="db" :label="db" :value="db" />
      </el-select>
      <el-select v-model="selectedTable" placeholder="表" style="width: 130px" size="small" @change="loadStructure">
        <el-option v-for="t in tables" :key="t" :label="t" :value="t" />
      </el-select>
    </div>

    <!-- 内容区 -->
    <div class="structure-content">
      <el-tabs v-model="activeTab" class="custom-tabs">
        <!-- 字段列表 -->
        <el-tab-pane label="字段" name="columns">
          <div class="tab-content">
            <el-table :data="columns" style="width: 100%" border size="small">
              <el-table-column type="index" width="50" align="center" label="#" />
              <el-table-column prop="name" label="字段名" min-width="150">
                <template #default="{ row }">
                  <span class="field-name">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="type" label="类型" min-width="120">
                <template #default="{ row }">
                  <span class="field-type">{{ row.type }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="nullable" label="允许空" width="80" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.nullable === 'YES' ? 'success' : 'danger'" size="small" effect="dark">
                    {{ row.nullable === 'YES' ? '是' : '否' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="key" label="键" width="80" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.key === 'PRI'" type="warning" size="small" effect="dark">PK</el-tag>
                  <el-tag v-else-if="row.key === 'UNI'" type="info" size="small" effect="dark">UNI</el-tag>
                  <el-tag v-else-if="row.key === 'MUL'" type="info" size="small" effect="dark">MUL</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="defaultValue" label="默认值" min-width="120">
                <template #default="{ row }">
                  <span class="default-value">{{ row.defaultValue ?? 'NULL' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="extra" label="额外" min-width="120" />
              <el-table-column prop="comment" label="注释" min-width="150" />
            </el-table>
          </div>
        </el-tab-pane>

        <!-- 索引 -->
        <el-tab-pane label="索引" name="indexes">
          <div class="tab-content">
            <el-table :data="indexes" style="width: 100%" border size="small">
              <el-table-column prop="name" label="索引名" min-width="150">
                <template #default="{ row }">
                  <span class="index-name">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column label="字段" min-width="200">
                <template #default="{ row }">
                  <el-tag v-for="col in row.columns" :key="col" size="small" style="margin: 2px;" effect="plain">
                    {{ col }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="unique" label="唯一" width="80" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.unique ? 'success' : 'info'" size="small" effect="dark">
                    {{ row.unique ? '是' : '否' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="primary" label="主键" width="80" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.primary" type="warning" size="small" effect="dark">是</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- DDL -->
        <el-tab-pane label="DDL" name="ddl">
          <div class="tab-content">
            <div class="ddl-container">
              <div class="ddl-header">
                <span class="ddl-title">CREATE TABLE 语句</span>
                <el-button size="small" @click="copyDDL">
                  <el-icon><CopyDocument /></el-icon> 复制
                </el-button>
              </div>
              <pre class="ddl-code"><code>{{ ddl || '-- 请先选择表' }}</code></pre>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useConnectionStore, api } from '../stores'
import { ElMessage } from 'element-plus'

const connStore = useConnectionStore()
const selectedConn = ref('')
const selectedDb = ref('')
const selectedTable = ref('')
const databases = ref([])
const tables = ref([])
const columns = ref([])
const indexes = ref([])
const ddl = ref('')
const activeTab = ref('columns')

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
  columns.value = []
  indexes.value = []
  ddl.value = ''
  if (selectedConn.value && selectedDb.value) {
    try {
      tables.value = await connStore.fetchTables(selectedConn.value, selectedDb.value)
    } catch (e) {
      ElMessage.error('加载表列表失败')
    }
  }
}

async function loadStructure() {
  if (!selectedConn.value || !selectedDb.value || !selectedTable.value) return
  try {
    const colRes = await api.get(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/columns`)
    columns.value = colRes.data

    const idxRes = await api.get(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/indexes`)
    indexes.value = idxRes.data || []

    const ddlRes = await api.get(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/ddl`)
    ddl.value = ddlRes.data?.ddl || ''
  } catch (e) {
    ElMessage.error('加载结构失败')
  }
}

function copyDDL() {
  navigator.clipboard?.writeText(ddl.value)
  ElMessage.success('DDL 已复制到剪贴板')
}

onMounted(async () => {
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
    await loadStructure()
  }
})
</script>

<style scoped>
.table-structure {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

/* 内容区 */
.structure-content {
  flex: 1;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.custom-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.custom-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 16px;
  background: var(--bg-tertiary);
}

.custom-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
}

.custom-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.tab-content {
  height: 100%;
  overflow: auto;
  padding: 16px;
}

/* 字段样式 */
.field-name {
  font-family: var(--font-mono);
  font-weight: 500;
  color: var(--accent-cyan);
}

.field-type {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--accent-amber);
}

.default-value {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-muted);
}

.index-name {
  font-family: var(--font-mono);
  font-weight: 500;
}

/* DDL 容器 */
.ddl-container {
  background: var(--bg-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.ddl-header {
  padding: 10px 16px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.ddl-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.ddl-code {
  padding: 20px;
  margin: 0;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-primary);
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
