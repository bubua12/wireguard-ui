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

- 🖥️ **服务器配置管理** - 接口名、地址、端口、DNS、MTU，保存前自动校验；未保存的修改会有提示
- ⏯️ **接口控制** - 在界面里启动 / 停止 / 重启 WireGuard 接口，地址、MTU、NAT 改完后一键重启生效
- 👥 **客户端管理** - 添加、编辑（名称、IP、保活间隔）、删除、启用/禁用，支持搜索、状态筛选、排序和分页
- 🟢 **实时在线状态** - 根据握手时间判断在线，显示最近握手时间和客户端来源地址（默认隐藏）
- 📈 **流量统计** - 累计流量不受接口重启影响；仪表盘有实时速率和 7 / 14 / 30 天流量趋势图，每个客户端可查看按天明细
- 🌐 **自定义 IP** - 手动指定 `/32` 或按网段自动分配，添加时会预览下一个空闲地址
- 📄 **配置下载 / 复制** - 生成客户端 `.conf`，可直接复制到剪贴板
- 📱 **二维码连接** - 手机 App 扫码
- 📥 **配置导入** - 从 `/etc/wireguard/*.conf` 覆盖导入
- ⚡ **热更新** - 接口已启动时增删、改址客户端走 `wg set`，同步配置走 `wg syncconf`，不会断开其他客户端
- 🌍 **分流 / 全局** - 仅内网或 `0.0.0.0/0`
- 🔀 **可选 NAT** - 写入 iptables PostUp/PostDown
- 🌙 **主题切换** - 暗黑 / 亮色，跟随系统，手机和平板上也能正常使用
- 🔐 **JWT 认证** - 密钥可配置，登录失败锁定，修改密码后其他设备自动退出

## 🛠️ 功能展示
客户端管理
<img width="1396" height="914" alt="image" src="https://github.com/user-attachments/assets/e025b8a8-ef44-40f7-9d75-3e2390777860" />



仪表盘
<img width="1430" height="540" alt="image" src="https://github.com/user-attachments/assets/5ea2fc78-73c8-4438-8b34-cf1722d83715" />



设置
<img width="1357" height="711" alt="image" src="https://github.com/user-attachments/assets/1a8ac01d-4c19-477c-b75d-f322175cd7fa" />




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

点击 **保存** 只会写入数据库，接着在页面顶部的接口卡片里让它生效：

| 操作 | 作用 | 是否断线 |
|------|------|:--------:|
| **同步配置** | 写入 `/etc/wireguard/<接口>.conf`，用 `wg syncconf` 应用客户端和监听端口的变化 | 否 |
| **重启接口** | `wg-quick down` 再 `up`，地址、MTU、NAT（PostUp/PostDown）的修改只有它能应用 | 是，几秒 |
| **启动 / 停止接口** | `wg-quick up` / `down` | 是 |

修改了地址、MTU 或 NAT 后，界面会提示「需要重启接口」；重启时还会先把最近一分钟尚未记录的流量落库，避免统计丢失。

> 同步需要本机已安装 WireGuard，并且进程有写 `/etc/wireguard`、操作网卡的权限。如果你正通过这条 VPN 访问本页面，停止或重启接口会让页面短暂失联。

### 3️⃣ 添加客户端

1. 进入 **客户端** 页面
2. 点击 **添加客户端**
3. 输入名称；IP 可留空自动分配（网段内第一个空闲 `/32`，不会占用服务器地址）
4. 手动指定时必须是网段内的 `x.x.x.x/32`

### 4️⃣ 客户端连接

在客户端列表点 **下载配置**，或点二维码图标 **扫码 / 复制配置**。从系统配置导入的客户端没有私钥，无法再生成这两项。

修改客户端的 IP、保活间隔，或服务器的 DNS / 公网地址 / MTU / 全局流量后，客户端需要重新下载配置或重新扫码才会生效。

### 5️⃣ 流量统计说明

- 客户端列表里的 **流量 ↓ / ↑** 站在客户端角度：↓ 是客户端下载（服务器发出），↑ 是客户端上传（服务器收到）
- 后台每分钟采样一次并按天累计，保留最近 90 天；接口重启、停止或禁用客户端前也会补记一次
- 仪表盘的「累计流量」是所有客户端累计值之和，与客户端列表一致（旧版本这里显示的是接口自启动以来的计数）

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
│   ├── stats/        # 后台流量采样
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
| `PUT` | /api/server | 更新服务器配置（只传要改的字段；`dns` 传空串表示不下发，`mtu` 传 0 表示自动），响应里的 `restart_required` 表示需要重启接口 |
| `GET` | /api/status | 接口是否启动、在线数、累计流量、是否需要重启 |
| `GET` | /api/traffic?days=14 | 全部客户端按天合计的流量（缺失的日期补 0） |
| `POST` | /api/interface/up | 启动接口 |
| `POST` | /api/interface/down | 停止接口 |
| `POST` | /api/interface/restart | 重启接口 |

### 客户端接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `GET` | /api/peers | 获取所有客户端（按创建顺序） |
| `GET` | /api/peers/status | 在线状态、累计流量、最近握手和来源地址 |
| `GET` | /api/peers/next-ip | 预览自动分配时的下一个 IP |
| `POST` | /api/peers | 创建客户端 |
| `PUT` | /api/peers/:id | 更新客户端（`name`、`enabled`、`allowed_ips`、`persistent_keepalive`，只传要改的字段） |
| `DELETE` | /api/peers/:id | 删除客户端 |
| `POST` | /api/peers/:id/toggle | 启用/禁用客户端 |
| `GET` | /api/peers/:id/config | 下载客户端配置 |
| `GET` | /api/peers/:id/qrcode | 获取二维码 |
| `GET` | /api/peers/:id/traffic | 单个客户端的按天流量 |

### 其他接口

| 方法 | 路径 | 说明 |
|:----:|------|------|
| `POST` | /api/sync | 同步配置到系统 |
| `POST` | /api/import | 覆盖导入现有配置 |
| `POST` | /api/change-password | 修改密码（返回新 token，其他设备的登录失效） |

## 🔒 安全建议

- 管理面默认只绑回环地址，公网暴露时走反向代理 + HTTPS，并设置 `WG_TRUST_PROXY=true`
- 不要把 SQLite 和 `.jwt_secret` 提交到仓库；数据库权限为 `0600`
- 导入路径被限制在 `WG_CONFIG_DIR` 内
- 接口名必须符合网卡名规则，避免路径穿越；DNS、公网地址会校验格式，不允许换行等字符混进客户端配置
- 响应带有 CSP、`X-Frame-Options` 等安全头；客户端配置和二维码响应设置了 `no-store`
- 修改密码后，旧的登录状态立即失效
- `wg` / `wg-quick` 调用带超时，进程收到 SIGTERM 会先落一次流量采样再退出
- 仓库自带 `.dockerignore`，本地的数据库和 `.jwt_secret` 不会被打进镜像

## 📄 License

本项目采用 [MIT License](LICENSE) 开源协议。
