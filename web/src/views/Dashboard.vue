<template>
  <div class="dashboard">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-select v-model="selectedConn" placeholder="连接" style="width: 160px" size="small">
        <el-option v-for="c in connStore.connections" :key="c.id" :label="c.name" :value="c.id" />
      </el-select>
      <el-select v-model="selectedDb" placeholder="数据库" style="width: 130px" size="small" @change="loadTables">
        <el-option v-for="db in databases" :key="db" :label="db" :value="db" />
      </el-select>
      <el-select v-model="selectedTable" placeholder="表" style="width: 130px" size="small" @change="loadStats">
        <el-option v-for="t in tables" :key="t" :label="t" :value="t" />
      </el-select>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid" v-if="stats.totalRows > 0">
      <div class="stat-card">
        <div class="stat-icon">
          <el-icon><Grid /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalRows.toLocaleString() }}</div>
          <div class="stat-label">总行数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon purple">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.columnCount }}</div>
          <div class="stat-label">字段数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon amber">
          <el-icon><Coin /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.tableSize }}</div>
          <div class="stat-label">表大小</div>
        </div>
      </div>
    </div>

    <!-- 图表区 -->
    <div class="chart-section" v-if="chartData">
      <div class="section-card">
        <div class="section-header">
          <span class="section-title">数据分布</span>
          <el-radio-group v-model="chartType" size="small">
            <el-radio-button label="bar">柱状图</el-radio-button>
            <el-radio-button label="line">折线图</el-radio-button>
            <el-radio-button label="pie">饼图</el-radio-button>
          </el-radio-group>
        </div>
        <div class="chart-container">
          <div ref="chartRef" class="chart"></div>
        </div>
      </div>
    </div>

    <!-- 自定义分析 -->
    <div class="analysis-section">
      <div class="section-card">
        <div class="section-header">
          <span class="section-title">自定义分析</span>
        </div>
        <div class="section-body">
          <div class="query-input">
            <el-input
              v-model="analysisQuery"
              type="textarea"
              :rows="3"
              placeholder="输入聚合查询语句，例如：SELECT status, COUNT(*) as count FROM orders GROUP BY status"
              class="sql-textarea"
            />
          </div>
          <el-button type="primary" style="margin-top: 12px" @click="runAnalysis" :loading="analyzing">
            <el-icon><CaretRight /></el-icon> 执行分析
          </el-button>
        </div>
      </div>
    </div>

    <!-- 分析结果 -->
    <div v-if="analysisResult" class="result-section">
      <div class="section-card">
        <div class="section-header">
          <span class="section-title">分析结果</span>
        </div>
        <div class="section-body">
          <el-table :data="analysisResult.rows" style="width: 100%" border size="small">
            <el-table-column
              v-for="(col, idx) in analysisResult.columns"
              :key="idx"
              :prop="String(idx)"
              :label="col"
              min-width="120"
            />
          </el-table>
          <div ref="analysisChartRef" class="chart" style="margin-top: 20px; height: 350px;"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { useConnectionStore, api } from '../stores'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'

const connStore = useConnectionStore()
const selectedConn = ref('')
const selectedDb = ref('')
const selectedTable = ref('')
const databases = ref([])
const tables = ref([])
const stats = ref({ totalRows: 0, columnCount: 0, tableSize: 'N/A' })
const chartType = ref('bar')
const chartData = ref(null)
const chartRef = ref(null)
const analysisQuery = ref('')
const analyzing = ref(false)
const analysisResult = ref(null)
const analysisChartRef = ref(null)

let chart = null
let analysisChart = null

watch(selectedConn, async (val) => {
  if (val) {
    const res = await api.get(`/api/conn/${val}/databases`)
    databases.value = res.data
  }
})

watch(chartType, () => {
  if (chartData.value) renderChart()
})

async function loadTables() {
  if (!selectedConn.value || !selectedDb.value) return
  const res = await api.get(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables`)
  tables.value = res.data
}

async function loadStats() {
  if (!selectedConn.value || !selectedDb.value || !selectedTable.value) return
  try {
    const countRes = await api.post(`/api/conn/${selectedConn.value}/query`, {
      sql: `SELECT COUNT(*) as total FROM \`${selectedTable.value}\``,
      database: selectedDb.value
    })
    stats.value.totalRows = countRes.data.rows?.[0]?.[0] || 0

    const colRes = await api.get(`/api/conn/${selectedConn.value}/databases/${selectedDb.value}/tables/${selectedTable.value}/columns`)
    stats.value.columnCount = colRes.data?.length || 0

    await generateChartData()
  } catch (e) {
    console.error('Failed to load stats', e)
  }
}

async function generateChartData() {
  try {
    const res = await api.post(`/api/conn/${selectedConn.value}/query`, {
      sql: `SELECT * FROM \`${selectedTable.value}\` LIMIT 100`,
      database: selectedDb.value
    })

    if (res.data.columns && res.data.rows) {
      chartData.value = { columns: res.data.columns, rows: res.data.rows }
      await nextTick()
      renderChart()
    }
  } catch (e) {
    console.error('Failed to generate chart data')
  }
}

function renderChart() {
  if (!chartRef.value || !chartData.value) return

  if (chart) chart.dispose()
  chart = echarts.init(chartRef.value)

  const { columns, rows } = chartData.value
  const option = {
    backgroundColor: 'transparent',
    textStyle: { fontFamily: "'IBM Plex Sans', sans-serif" },
    tooltip: { trigger: 'axis' },
    grid: { top: 30, right: 20, bottom: 30, left: 50 },
    xAxis: { type: 'category', data: [], axisLine: { lineStyle: { color: '#2d2d3f' } }, axisLabel: { color: '#64748b' } },
    yAxis: { type: 'value', axisLine: { lineStyle: { color: '#2d2d3f' } }, axisLabel: { color: '#64748b' }, splitLine: { lineStyle: { color: '#1a1a2e' } } },
    series: []
  }

  const numericCols = []
  columns.forEach((col, idx) => {
    if (rows.length > 0 && typeof rows[0][idx] === 'number') {
      numericCols.push(idx)
    }
  })

  if (chartType.value === 'pie') {
    const nameIdx = 0
    const valueIdx = numericCols[0] || 1
    option.series = [{
      type: 'pie',
      data: rows.slice(0, 20).map(row => ({
        name: String(row[nameIdx] || 'N/A'),
        value: row[valueIdx] || 0
      })),
      itemStyle: { borderRadius: 4, borderColor: '#0f0f17', borderWidth: 2 }
    }]
    delete option.xAxis
    delete option.yAxis
  } else {
    option.xAxis.data = rows.slice(0, 50).map((row, i) => row[0] || `Row ${i + 1}`)
    option.series = numericCols.slice(0, 5).map(idx => ({
      name: columns[idx],
      type: chartType.value,
      data: rows.slice(0, 50).map(row => row[idx] || 0),
      smooth: true,
      areaStyle: chartType.value === 'line' ? { opacity: 0.1 } : undefined
    }))
    option.legend = { textStyle: { color: '#94a3b8' } }
  }

  chart.setOption(option)
}

async function runAnalysis() {
  if (!selectedConn.value || !analysisQuery.value.trim()) {
    ElMessage.warning('请先选择连接并输入查询')
    return
  }

  analyzing.value = true
  try {
    const res = await api.post(`/api/conn/${selectedConn.value}/query`, {
      sql: analysisQuery.value.trim(),
      database: selectedDb.value
    })

    if (res.data.error) {
      ElMessage.error(res.data.error)
      return
    }

    analysisResult.value = res.data
    await nextTick()
    renderAnalysisChart()
  } catch (e) {
    ElMessage.error('查询失败')
  } finally {
    analyzing.value = false
  }
}

function renderAnalysisChart() {
  if (!analysisChartRef.value || !analysisResult.value) return

  if (analysisChart) analysisChart.dispose()
  analysisChart = echarts.init(analysisChartRef.value)

  const { columns, rows } = analysisResult.value
  const numericCols = []
  columns.forEach((col, idx) => {
    if (rows.length > 0 && typeof rows[0][idx] === 'number') {
      numericCols.push(idx)
    }
  })

  if (numericCols.length === 0) return

  const option = {
    backgroundColor: 'transparent',
    textStyle: { fontFamily: "'IBM Plex Sans', sans-serif" },
    tooltip: { trigger: 'axis' },
    grid: { top: 30, right: 20, bottom: 30, left: 50 },
    xAxis: {
      type: 'category',
      data: rows.map(row => String(row[0] || '')),
      axisLine: { lineStyle: { color: '#2d2d3f' } },
      axisLabel: { color: '#64748b' }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#2d2d3f' } },
      axisLabel: { color: '#64748b' },
      splitLine: { lineStyle: { color: '#1a1a2e' } }
    },
    series: numericCols.map(idx => ({
      name: columns[idx],
      type: 'bar',
      data: rows.map(row => row[idx] || 0),
      barWidth: '60%'
    })),
    legend: { textStyle: { color: '#94a3b8' } }
  }

  analysisChart.setOption(option)
}

onMounted(async () => {
  if (connStore.currentConn) selectedConn.value = connStore.currentConn.id
  if (connStore.currentDb) {
    selectedDb.value = connStore.currentDb
    await loadTables()
  }
  if (connStore.currentTable) {
    selectedTable.value = connStore.currentTable
    await loadStats()
  }
})
</script>

<style scoped>
.dashboard {
  height: 100%;
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
  margin-bottom: 16px;
}

/* 统计网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.stat-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all var(--transition-fast);
}

.stat-card:hover {
  border-color: var(--border-hover);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: var(--accent-cyan-dim);
  color: var(--accent-cyan);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.stat-icon.purple {
  background: rgba(168, 85, 247, 0.15);
  color: var(--accent-purple);
}

.stat-icon.amber {
  background: var(--accent-amber-dim);
  color: var(--accent-amber);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
}

/* 区块卡片 */
.section-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
  margin-bottom: 16px;
}

.section-header {
  padding: 14px 20px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.section-body {
  padding: 20px;
}

.chart-container {
  padding: 16px;
}

.chart {
  width: 100%;
  height: 350px;
}

/* SQL 文本框 */
.sql-textarea :deep(.el-textarea__inner) {
  font-family: var(--font-mono) !important;
  font-size: 13px !important;
}
</style>
