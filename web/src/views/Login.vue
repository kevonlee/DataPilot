<template>
  <div class="login-page">
    <!-- 背景网格 -->
    <div class="grid-bg"></div>

    <!-- 装饰线条 -->
    <div class="deco-line deco-line-1"></div>
    <div class="deco-line deco-line-2"></div>

    <!-- 登录卡片 -->
    <div class="login-container">
      <div class="login-card">
        <!-- 顶部装饰条 -->
        <div class="card-accent"></div>

        <!-- Logo 区域 -->
        <div class="logo-section">
          <div class="logo-icon">
            <svg viewBox="0 0 40 40" fill="none">
              <rect x="4" y="8" width="32" height="24" rx="3" stroke="currentColor" stroke-width="2"/>
              <ellipse cx="20" cy="16" rx="10" ry="4" stroke="currentColor" stroke-width="1.5"/>
              <path d="M10 16v8c0 2.2 4.5 4 10 4s10-1.8 10-4v-8" stroke="currentColor" stroke-width="1.5"/>
              <path d="M10 20c0 2.2 4.5 4 10 4s10-1.8 10-4" stroke="currentColor" stroke-width="1.5" opacity="0.5"/>
            </svg>
          </div>
          <h1 class="logo-title">DataPilot</h1>
          <p class="logo-subtitle">数据库管理工具</p>
        </div>

        <!-- 登录表单 -->
        <el-form :model="form" @submit.prevent="handleLogin" class="login-form">
          <div class="input-group">
            <label class="input-label">用户名</label>
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              prefix-icon="User"
              size="large"
            />
          </div>

          <div class="input-group">
            <label class="input-label">密码</label>
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="Lock"
              size="large"
              show-password
              @keyup.enter="handleLogin"
            />
          </div>

          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleLogin"
            class="login-btn"
          >
            <span v-if="!loading">登 录</span>
            <span v-else>验证中...</span>
          </el-button>
        </el-form>

        <!-- 底部信息 -->
        <div class="login-footer">
          <div class="hint-line">
            <span class="hint-dot"></span>
            <span>默认账号: admin / admin</span>
          </div>
        </div>
      </div>

      <!-- 版本信息 -->
      <div class="version-info">
        <span class="version-tag">v1.0.0</span>
        <span class="version-sep">·</span>
        <span>Database Manager</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

async function handleLogin() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: var(--bg-primary);
}

/* 网格背景 */
.grid-bg {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 212, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
}

/* 装饰线条 */
.deco-line {
  position: absolute;
  background: linear-gradient(90deg, transparent, var(--accent-cyan), transparent);
  height: 1px;
  opacity: 0.3;
}

.deco-line-1 {
  top: 30%;
  left: 0;
  right: 0;
  animation: scanLine 8s linear infinite;
}

.deco-line-2 {
  bottom: 25%;
  left: 0;
  right: 0;
  animation: scanLine 12s linear infinite reverse;
}

@keyframes scanLine {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* 登录容器 */
.login-container {
  position: relative;
  z-index: 10;
  display: flex;
  flex-direction: column;
  align-items: center;
  animation: fadeIn 0.6s ease-out;
}

/* 登录卡片 */
.login-card {
  width: 400px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-xl);
  padding: 0 36px 36px;
  position: relative;
  overflow: hidden;
  box-shadow: var(--shadow-lg), 0 0 60px rgba(0, 212, 255, 0.05);
}

.card-accent {
  height: 3px;
  background: linear-gradient(90deg, var(--accent-cyan), var(--accent-purple), var(--accent-cyan));
  background-size: 200% 100%;
  animation: shimmer 3s linear infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Logo */
.logo-section {
  text-align: center;
  padding: 32px 0 28px;
}

.logo-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 16px;
  color: var(--accent-cyan);
  filter: drop-shadow(0 0 10px var(--accent-cyan-glow));
}

.logo-title {
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: 0.15em;
  margin-bottom: 6px;
}

.logo-subtitle {
  font-size: 13px;
  color: var(--text-muted);
  letter-spacing: 0.2em;
  text-transform: uppercase;
}

/* 表单 */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.login-btn {
  width: 100%;
  height: 48px;
  margin-top: 8px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.2em;
  border-radius: var(--radius-md) !important;
  position: relative;
  overflow: hidden;
}

.login-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
  transform: translateX(-100%);
  transition: transform 0.6s;
}

.login-btn:hover::before {
  transform: translateX(100%);
}

/* 底部 */
.login-footer {
  margin-top: 28px;
  padding-top: 20px;
  border-top: 1px solid var(--border-subtle);
}

.hint-line {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.hint-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent-green);
  box-shadow: 0 0 8px var(--accent-green);
}

/* 版本信息 */
.version-info {
  margin-top: 24px;
  font-size: 11px;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 8px;
}

.version-tag {
  font-family: var(--font-mono);
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: 10px;
}

.version-sep {
  opacity: 0.3;
}
</style>
