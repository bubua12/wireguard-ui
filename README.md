<h1 align="center">🛡️ WireGuard UI</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Vue-3.4-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue">
  <img src="https://img.shields.io/badge/Vite-5.0-646CFF?style=flat-square&logo=vite&logoColor=white" alt="Vite">
  <img src="https://img.shields.io/badge/Tailwind_CSS-3.4-38B2AC?style=flat-square&logo=tailwind-css&logoColor=white" alt="Tailwind CSS">
  <img src="https://img.shields.io/badge/SQLite-3-003B57?style=flat-square&logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
</p>

<p align="center">
  一个简洁的 WireGuard VPN 管理界面，用来管理服务器参数、客户端密钥和配置下发。
</p>

---

<img width="2880" height="1478" alt="login" src="https://github.com/user-attachments/assets/b8bb4f66-523c-47be-96b6-31c6a0db1160" />


## ✨ 功能特性

- 🖥️ **服务器配置管理** - 接口名、地址、端口、DNS、MTU
- 👥 **客户端管理** - 添加、编辑、删除、启用/禁用
- 🟢 **实时在线状态** - 根据握手时间判断在线
- 🌐 **自定义 IP** - 手动指定 `/32` 或按网段自动分配
- 📄 **配置下载** - 生成客户端 `.conf`
- 📱 **二维码连接** - 手机 App 扫码
- 📥 **配置导入** - 从 `/etc/wireguard/*.conf` 覆盖导入
- ⚡ **热更新** - 接口已启动时增删客户端走 `wg set`
- 🌍 **分流 / 全局** - 仅内网或 `0.0.0.0/0`
- 🔀 **可选 NAT** - 写入 iptables PostUp/PostDown
- 🌙 **主题切换** - 暗黑 / 亮色
- 🔐 **JWT 认证** - 密钥可配置，登录失败锁定

## 🛠️ 功能展示
客户端管理
<img width="1436" height="695" alt="客户端管理" src="https://github.com/user-attachments/assets/f9aa4669-3419-49aa-8fee-e758fc561e39" />


仪表盘
<img width="1376" height="601" alt="仪表盘" src="https://github.com/user-attachments/assets/2bba0c78-1de9-4a0b-a72e-f2ffcd55178c" />


设置
<img width="1387" height="842" alt="设置" src="https://github.com/user-attachments/assets/c5f577a9-1c85-4222-9fe0-9086ddbf9e44" />



## 🚀 快速开始

### 📋 环境要求

| 依赖 | 版本 |
|------|------|
| Go | 1.24+ |
| Node.js | 18+ |
| npm | 9+ |
| wireguard-tools | 生成密钥、热更新、同步时需要 `wg` / `wg-quick` |

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

### ▶️ 开发模式

**启动后端：**
```bash
cd backend
go run .
# 默认监听 http://127.0.0.1:8081
```

**启动前端：**
```bash
cd frontend
npm run dev
# http://localhost:5173 ，API 代理到 8081
```

生产构建会把前端打进同一个二进制：

```bash
make build
./bin/wireguard-ui
```

### 🔑 初始化管理员

首次可通过网页注册（仅当库里还没有用户）。也可以用命令行创建，并打印一次性随机密码：

```bash
cd backend
go run ./cmd/init
```

```bash
WG_ADMIN_USER=admin WG_ADMIN_PASSWORD='your-strong-password' go run ./cmd/init
```

密码至少 8 位，不能使用 `admin`。

## 📖 使用指南

### 1️⃣ 首次登录

1. 打开 `http://localhost:5173`（开发）或后端监听地址（生产）
2. 用已创建的管理员账户登录
3. 登录后建议立刻修改密码

### 2️⃣ 配置 WireGuard 服务器

进入 **设置** 页面：

| 参数 | 说明 | 示例 |
|------|------|------|
| 显示名称 | 仅用于 UI | My VPN |
| 接口名 | 对应 `wg0` / `/etc/wireguard/wg0.conf` | wg0 |
| 公网地址 | `host:port` | vpn.example.com:51820 |
| 内网地址 | 服务器 CIDR | 10.0.0.1/24 |
| 监听端口 | UDP | 51820 |
| DNS | 写入客户端配置 | 8.8.8.8 |
| MTU | 最大传输单元 | 1420 |
| 全局流量 | 客户端 AllowedIPs 为 `0.0.0.0/0, ::/0` | 默认关闭（仅内网） |
| NAT / 转发 | 写入 iptables PostUp/PostDown | 默认关闭 |

点击 **保存** 写入数据库，再点 **同步到系统** 才会 `wg-quick up` 或 `wg syncconf`。

> 同步需要本机已安装 WireGuard，并且进程有写 `/etc/wireguard`、操作网卡的权限。

### 3️⃣ 添加客户端

1. 进入 **客户端** 页面
2. 点击 **添加客户端**
3. 输入名称；IP 可留空自动分配（网段内第一个空闲 `/32`，不会占用服务器地址）
4. 手动指定时必须是网段内的 `x.x.x.x/32`

### 4️⃣ 客户端连接

**下载配置文件** 或 **扫描二维码**。从系统配置导入的客户端没有私钥，无法再生成这两项。

### 5️⃣ 同步配置到系统

设置页的 **同步到系统**：

- 接口不存在时执行 `wg-quick up <接口名>`
- 接口已存在时执行 `wg syncconf`，尽量不中断现有连接

## ⚙️ 配置说明

通过环境变量配置，可参考仓库根目录 `.env.example`。

| 变量 | 默认 | 说明 |
|------|------|------|
| `WG_LISTEN` | `127.0.0.1:8081` | 监听地址，Docker 示例里改为 `0.0.0.0:8081` |
| `WG_DB_PATH` | `wireguard.db` | SQLite 路径 |
| `WG_JWT_SECRET` | 自动生成到 `.jwt_secret` | JWT 签名密钥，至少 16 位 |
| `WG_JWT_SECRET_FILE` | 数据库同目录 `.jwt_secret` | 密钥文件 |
| `WG_TRUST_PROXY` | `false` | 是否信任反向代理的 `X-Forwarded-For` |
| `WG_CONFIG_DIR` | `/etc/wireguard` | 允许读写的配置目录 |
| `WG_CORS_ORIGINS` | `http://localhost:5173,http://127.0.0.1:5173` | 开发跨域 |
| `WG_DEBUG` | `false` | 打开 gin 调试日志 |

修改端口时，开发模式还要改 `frontend/vite.config.js` 里的代理。

## 🐳 Docker

```bash
docker compose up -d --build
```

默认使用 `network_mode: host` 和 `NET_ADMIN`，以便操作本机 WireGuard 接口。数据在 `./data`，配置目录挂载 `/etc/wireguard`。

## 📁 项目结构

```
wireguard-ui/
├── backend/
│   ├── api/          # HTTP 路由和处理器
│   ├── config/       # 环境变量
│   ├── db/           # SQLite
│   ├── model/
│   ├── wg/           # 密钥、配置生成、wg 命令
│   ├── web/dist/     # 嵌入的前端（make frontend）
│   ├── cmd/init/     # 创建管理员
│   └── main.go
├── frontend/
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## 📡 API 接口

### 认证接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `POST` | /api/login | 用户登录 |
| `POST` | /api/register | 注册（仅首次） |
| `GET` | /api/init | 检查初始化状态 |
| `GET` | /api/health | 存活检查 |

### 服务器接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `GET` | /api/server | 获取服务器配置 |
| `POST` | /api/server | 创建服务器配置 |
| `PUT` | /api/server | 更新服务器配置 |
| `GET` | /api/status | 接口是否启动、在线数、流量 |

### 客户端接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `GET` | /api/peers | 获取所有客户端 |
| `GET` | /api/peers/status | 获取客户端在线状态 |
| `POST` | /api/peers | 创建客户端 |
| `PUT` | /api/peers/:id | 更新客户端信息 |
| `DELETE` | /api/peers/:id | 删除客户端 |
| `POST` | /api/peers/:id/toggle | 启用/禁用客户端 |
| `GET` | /api/peers/:id/config | 下载客户端配置 |
| `GET` | /api/peers/:id/qrcode | 获取二维码 |

### 其他接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `POST` | /api/sync | 同步配置到系统 |
| `POST` | /api/import | 覆盖导入现有配置 |
| `POST` | /api/change-password | 修改密码 |

## 🔒 安全建议

- 管理面默认只绑回环地址，公网暴露时走反向代理 + HTTPS，并设置 `WG_TRUST_PROXY=true`
- 不要把 SQLite 和 `.jwt_secret` 提交到仓库；数据库权限为 `0600`
- 导入路径被限制在 `WG_CONFIG_DIR` 内
- 接口名必须符合网卡名规则，避免路径穿越

## 📄 License

本项目采用 [MIT License](LICENSE) 开源协议。
