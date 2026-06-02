# DataPilot

<div align="center">

**数据库管理工具**

一个基于 Go 和 Vue 3 构建的 Web 端数据库管理工具，类似 Navicat，开箱即用。

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D?style=flat-square&logo=vue.js&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)

</div>

---

## 功能特性

### 核心功能

| 功能 | 说明 |
|------|------|
| 连接管理 | 保存和管理多个数据库连接配置 |
| 数据库浏览器 | 树形结构浏览数据库、表、字段 |
| SQL 编辑器 | Monaco Editor，语法高亮，自动补全 |
| 表数据管理 | 查看、编辑、新增、删除表数据 |
| 表结构查看 | 查看字段、索引、DDL 语句 |
| 数据导入导出 | 支持 CSV/JSON/SQL 格式导入导出 |
| 数据可视化 | ECharts 图表展示查询结果 |
| 用户认证 | JWT Token 安全认证 |

### 支持的数据库

| 数据库 | 驱动 | 默认端口 |
|--------|------|----------|
| MySQL / MariaDB | go-sql-driver/mysql | 3306 |
| PostgreSQL | lib/pq | 5432 |
| SQLite | mattn/go-sqlite3 | - |
| SQL Server | denisenkom/go-mssqldb | 1433 |
| Oracle | sijms/go-ora | 1521 |

## 快速开始

### 环境要求

- **Go 1.21+** - [下载](https://go.dev/dl/)
- **Node.js 18+** - [下载](https://nodejs.org/)

### 构建与运行

#### Windows

```batch
build.bat
dbmanager.exe
```

#### Linux / macOS

```bash
chmod +x build.sh
./build.sh
./dbmanager
```

#### 手动构建

```bash
# 1. 构建前端
cd web
npm install
npm run build
cd ..

# 2. 构建后端
go mod tidy
go build -o dbmanager .

# 3. 运行
./dbmanager
```

### 访问

打开浏览器访问：http://localhost:9090

**默认登录账号：**
- 用户名：`admin`
- 密码：`admin`

## 使用指南

### 创建数据库连接

1. 点击侧边栏「新建连接」按钮
2. 选择数据库类型（MySQL、PostgreSQL 等）
3. 填写连接信息（主机、端口、用户名、密码）
4. 点击「测试连接」验证配置
5. 点击「保存」

### 执行 SQL 查询

1. 在侧边栏选择连接和数据库
2. 双击数据库打开 SQL 编辑器
3. 输入 SQL 语句
4. 按 `Ctrl + Enter` 或点击「执行」按钮
5. 在下方查看查询结果

### 管理表数据

1. 在侧边栏双击表名打开数据视图
2. 双击单元格可编辑数据
3. 使用工具栏按钮新增或删除行
4. 支持分页浏览和数据导出

### 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Ctrl + Enter` | 执行 SQL 查询 |
| `Ctrl + Q` | 新建查询 |
| `Ctrl + C` | 复制表名/字段名 |
| `Enter` | 查看表数据 |

## 配置说明

配置文件位于 `data/config.json`，可设置：

```json
{
  "port": 9090,
  "jwtSecret": "your-secret-key",
  "user": {
    "username": "admin",
    "password": "admin"
  }
}
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DB_MANAGER_PORT` | 服务端口 | 9090 |
| `DB_MANAGER_DATA` | 数据目录 | data |

## 项目结构

```
DataPilot/
├── main.go                          # 程序入口
├── go.mod                           # Go 模块配置
├── build.bat                        # Windows 构建脚本
├── build.sh                         # Linux/Mac 构建脚本
├── internal/
│   ├── config/
│   │   └── config.go                # 配置管理
│   ├── handler/
│   │   ├── router.go                # 路由注册
│   │   ├── auth.go                  # 认证接口
│   │   ├── connection.go            # 连接管理接口
│   │   └── database.go              # 数据库操作接口
│   ├── middleware/
│   │   ├── auth.go                  # JWT 认证中间件
│   │   └── cors.go                  # 跨域中间件
│   ├── model/
│   │   ├── connection.go            # 连接模型
│   │   └── user.go                  # 用户模型
│   └── service/
│       ├── dbmanager.go             # 数据库连接池管理
│       ├── executor.go              # SQL 执行引擎
│       └── exporter.go              # 导出服务
├── web/                             # Vue 3 前端
│   ├── src/
│   │   ├── views/                   # 页面组件
│   │   ├── stores/                  # 状态管理
│   │   ├── router/                  # 路由配置
│   │   └── styles/                  # 全局样式
│   ├── package.json
│   └── vite.config.js
└── data/                            # 本地数据存储
```

## 技术栈

### 后端

- **Go** - 主语言
- **net/http** - HTTP 服务
- **database/sql** - 数据库操作
- **JWT** - 身份认证
- **embed** - 静态文件嵌入

### 前端

- **Vue 3** - UI 框架
- **Vite** - 构建工具
- **Element Plus** - UI 组件库
- **Monaco Editor** - SQL 编辑器
- **ECharts** - 数据可视化
- **Pinia** - 状态管理
- **Vue Router** - 路由管理

## 开发说明

### 开发模式

```bash
# 启动后端
go run main.go

# 另开终端，启动前端开发服务器
cd web
npm run dev
```

前端开发服务器会自动代理 `/api` 请求到后端 `http://localhost:9090`。

### 构建发布

```bash
# 一键构建（前端 + 后端）
build.bat    # Windows
./build.sh   # Linux/Mac
```

构建产物为单个可执行文件 `dbmanager`（或 `dbmanager.exe`），无需额外依赖，直接运行即可。

## 常见问题

### 端口被占用

设置环境变量更改端口：

```bash
# Windows
set DB_MANAGER_PORT=8080

# Linux/Mac
export DB_MANAGER_PORT=8080
```

### 忘记密码

删除 `data/config.json` 文件，重启服务后会重置为默认账号 `admin/admin`。

### 数据库连接失败

1. 检查数据库服务是否启动
2. 检查主机、端口、用户名、密码是否正确
3. 检查防火墙设置
4. 检查数据库用户是否有远程访问权限

## 许可证

MIT License

## 致谢

感谢以下开源项目：

- [Go](https://go.dev/)
- [Vue.js](https://vuejs.org/)
- [Element Plus](https://element-plus.org/)
- [Monaco Editor](https://microsoft.github.io/monaco-editor/)
- [ECharts](https://echarts.apache.org/)
