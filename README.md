<h1 align="center">🛡️ WireGuard UI</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3.4-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue">
  <img src="https://img.shields.io/badge/Vite-5.0-646CFF?style=flat-square&logo=vite&logoColor=white" alt="Vite">
  <img src="https://img.shields.io/badge/Tailwind_CSS-3.4-38B2AC?style=flat-square&logo=tailwind-css&logoColor=white" alt="Tailwind CSS">
  <img src="https://img.shields.io/badge/SQLite-3-003B57?style=flat-square&logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
</p>

<p align="center">
  一个简洁美观的 WireGuard VPN 管理界面，提供 Web UI 来管理 WireGuard 服务器和客户端配置。
</p>

---

<img width="2880" height="1478" alt="login" src="https://github.com/user-attachments/assets/b8bb4f66-523c-47be-96b6-31c6a0db1160" />


## ✨ 功能特性

- 🖥️ **服务器配置管理** - 轻松配置 WireGuard 服务器参数
- 👥 **客户端管理** - 添加、删除、启用/禁用客户端
- 📄 **配置下载** - 一键生成客户端配置文件
- 📱 **二维码连接** - 扫码即可连接 VPN
- 🌙 **主题切换** - 支持暗黑/亮色主题
- 🔐 **安全认证** - JWT Token 登录认证

## 🛠️ 功能展示
客户端管理
<img width="2880" height="1478" alt="image" src="https://github.com/user-attachments/assets/7bb7e4bc-387d-445b-bce6-671c4bf16bd0" />


仪表盘
<img width="2880" height="1478" alt="dashboard" src="https://github.com/user-attachments/assets/24da812c-ddaf-4ee5-a5dc-853f65709d59" />



## 🚀 快速开始

### 📋 环境要求

| 依赖 | 版本 |
|------|------|
| Go | 1.21+ |
| Node.js | 18+ |
| npm | 9+ |

### 📦 安装依赖

**后端：**
```bash
cd backend
go mod tidy
```

**前端：**
```bash
cd frontend
npm install
```

### ▶️ 运行项目

**启动后端：**
```bash
cd backend
go run main.go
# 服务运行在 http://localhost:8081
```

**启动前端：**
```bash
cd frontend
npm run dev
# 服务运行在 http://localhost:5173
```

### 🔑 初始化管理员

首次使用需要创建管理员账户：

```bash
cd backend
go run cmd/init/main.go
```

> 默认账户：**admin** / **admin**

## 📖 使用指南

### 1️⃣ 首次登录

1. 打开浏览器访问 `http://localhost:5173`
2. 使用管理员账户登录（默认：admin / admin）
3. 建议登录后修改默认密码

### 2️⃣ 配置 WireGuard 服务器

进入 **Settings** 页面，配置服务器参数：

| 参数 | 说明 | 示例 |
|------|------|------|
| Name | 服务器名称 | My VPN Server |
| Endpoint | 公网地址:端口 | vpn.example.com:51820 |
| Address | 服务器内网地址 | 10.0.0.1/24 |
| Listen Port | 监听端口 | 51820 |
| DNS | 客户端 DNS | 8.8.8.8 |
| MTU | 最大传输单元 | 1420 |

点击 **Save** 保存配置。

### 3️⃣ 添加客户端

1. 进入 **Peers** 页面
2. 点击 **Add Peer** 按钮
3. 输入客户端名称（如：iPhone、MacBook）
4. 系统自动生成密钥对和分配 IP

### 4️⃣ 客户端连接

**方式一：下载配置文件**
- 点击客户端的 **Download** 按钮
- 将 `.conf` 文件导入 WireGuard 客户端

**方式二：扫描二维码**
- 点击客户端的 **QR** 按钮
- 使用 WireGuard 手机 App 扫码连接

### 5️⃣ 客户端管理

| 操作 | 说明 |
|------|------|
| Enable/Disable | 启用或禁用客户端 |
| Download | 下载客户端配置文件 |
| QR | 显示连接二维码 |
| Delete | 删除客户端 |

### 6️⃣ 同步配置到系统

在 **Settings** 页面点击 **Sync to System** 按钮，将配置同步到 WireGuard 服务。

> ⚠️ **注意**：同步功能需要服务器已安装 WireGuard，并具有相应权限。

## 📁 项目结构

```
wireguard-ui/
├── 📂 backend/
│   ├── 📂 api/          # API 路由和处理器
│   ├── 📂 db/           # 数据库操作
│   ├── 📂 model/        # 数据模型
│   ├── 📂 wg/           # WireGuard 工具函数
│   ├── 📂 cmd/init/     # 初始化脚本
│   └── 📄 main.go
├── 📂 frontend/
│   ├── 📂 src/
│   │   ├── 📂 views/    # 页面组件
│   │   ├── 📄 App.vue
│   │   ├── 📄 main.js
│   │   └── 📄 router.js
│   └── 📄 vite.config.js
└── 📄 README.md
```

## 📡 API 接口

### 认证接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `POST` | /api/login | 用户登录 |
| `POST` | /api/register | 注册（仅首次） |
| `GET` | /api/init | 检查初始化状态 |

### 服务器接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `GET` | /api/server | 获取服务器配置 |
| `POST` | /api/server | 创建服务器配置 |
| `PUT` | /api/server | 更新服务器配置 |

### 客户端接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `GET` | /api/peers | 获取所有客户端 |
| `POST` | /api/peers | 创建客户端 |
| `DELETE` | /api/peers/:id | 删除客户端 |
| `GET` | /api/peers/:id/config | 下载客户端配置 |
| `GET` | /api/peers/:id/qrcode | 获取二维码 |

## ⚙️ 配置说明

### 后端端口

修改 `backend/main.go`：

```go
r.Run(":8081")  // 修改为你需要的端口
```

### 前端代理

修改 `frontend/vite.config.js`：

```js
proxy: {
  '/api': 'http://localhost:8081'  // 对应后端端口
}
```

## 📄 License

本项目采用 [MIT License](LICENSE) 开源协议。

---

<p align="center">
  Made with ❤️ by WireGuard UI Team
</p>
