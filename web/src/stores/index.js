import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')

  const isLoggedIn = computed(() => !!token.value)

  async function login(user, pass) {
    const res = await axios.post('/api/auth/login', {
      username: user,
      password: pass
    })
    token.value = res.data.token
    username.value = res.data.username
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('username', res.data.username)
    return res.data
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('username')
  }

  return { token, username, isLoggedIn, login, logout }
})

export const useConnectionStore = defineStore('connections', () => {
  const connections = ref([])
  const currentConn = ref(null)
  const currentDb = ref('')
  const currentTable = ref('')
  const databases = ref([])
  const tables = ref([])

  async function fetchConnections() {
    const res = await api.get('/api/connections')
    connections.value = res.data
  }

  async function addConnection(conn) {
    const res = await api.post('/api/connections', conn)
    connections.value.push(res.data)
    return res.data
  }

  async function updateConnection(conn) {
    const res = await api.put(`/api/connections/${conn.id}`, conn)
    const idx = connections.value.findIndex(c => c.id === conn.id)
    if (idx >= 0) connections.value[idx] = res.data
    return res.data
  }

  async function deleteConnection(id) {
    await api.delete(`/api/connections/${id}`)
    connections.value = connections.value.filter(c => c.id !== id)
    if (currentConn.value?.id === id) {
      currentConn.value = null
      databases.value = []
      tables.value = []
    }
  }

  async function testConnection(id) {
    const res = await api.post(`/api/connections/${id}/test`)
    return res.data
  }

  async function fetchDatabases(connId) {
    const res = await api.get(`/api/conn/${connId}/databases`)
    databases.value = res.data
    return res.data
  }

  async function fetchTables(connId, dbName) {
    const res = await api.get(`/api/conn/${connId}/databases/${dbName}/tables`)
    tables.value = res.data
    return res.data
  }

  function setCurrentConn(conn) {
    currentConn.value = conn
    currentDb.value = ''
    currentTable.value = ''
    databases.value = []
    tables.value = []
  }

  function setCurrentDb(db) {
    currentDb.value = db
    currentTable.value = ''
    tables.value = []
  }

  function setCurrentTable(table) {
    currentTable.value = table
  }

  return {
    connections, currentConn, currentDb, currentTable,
    databases, tables,
    fetchConnections, addConnection, updateConnection, deleteConnection,
    testConnection, fetchDatabases, fetchTables,
    setCurrentConn, setCurrentDb, setCurrentTable
  }
})

// axios instance with auth
const api = axios.create()

api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export { api }
