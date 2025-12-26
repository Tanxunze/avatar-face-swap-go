# Avatar Face Swap - Go Backend

[English](#english) | [中文](#中文)

---

## English

A high-performance Go backend for avatar face swap applications with event management, face detection, and multi-authentication support.

### Features

- **Event Management**: Create and manage face swap events with access tokens
- **Face Detection**: Integrated Tencent Cloud face detection API
- **Multi-Auth Support**:
  - Admin password authentication
  - Event-based token access
  - Keycloak SSO integration
- **File Storage**: Automated storage management for event photos and face images
- **QQ Integration**: Fetch QQ avatars and nicknames
- **RESTful API**: Clean API design with JWT-based authentication

### Tech Stack

- Go 1.23.2
- Gin Web Framework
- SQLite Database
- JWT Authentication
- Tencent Cloud Face Detection API
- Keycloak (Optional SSO)

### Project Structure

```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration loading
│   ├── database/        # Database initialization
│   ├── handler/         # HTTP handlers (routes)
│   ├── middleware/      # Authentication middleware
│   ├── model/           # Data models
│   ├── repository/      # Database operations
│   ├── service/         # Business logic
│   └── storage/         # File storage utilities
├── pkg/response/        # Shared response utilities
└── data/                # SQLite DB + file storage
```

### Quick Start

#### Prerequisites

- Go 1.23 or higher
- (Optional) [Air](https://github.com/cosmtrek/air) for hot reload
- (Optional) Docker and Docker Compose for containerized deployment

#### Installation

```bash
# Clone the repository
git clone <repository-url>
cd avatar-face-swap-go

# Copy environment configuration
cp .env.example .env

# Edit .env with your configuration
# Required: PORT, JWT_SECRET, ADMIN_PASSWORD
# Optional: Keycloak credentials, Tencent Cloud credentials
```

#### Development

```bash
# Run with hot reload (recommended)
air

# Or run directly
go run cmd/server/main.go
```

#### Production Build

```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Run
./bin/server
```

#### Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild image
docker-compose up -d --build
```

Or build Docker image manually:

```bash
# Build image
docker build -t avatar-face-swap-go .

# Run container
docker run -d \
  -p 5001:5001 \
  -v $(pwd)/data:/app/data \
  --env-file .env \
  --name avatar-face-swap \
  avatar-face-swap-go
```

### Configuration

Key environment variables:

| Variable | Description | Required | Default |
| -------- | ----------- | -------- | ------- |
| `PORT` | Server port | Yes | `5001` |
| `GIN_MODE` | Gin mode (debug/release) | No | `debug` |
| `DATABASE_URL` | SQLite database file path | No | `./data/app.db` |
| `STORAGE_DIR` | File storage directory | No | `./data/storage` |
| `JWT_SECRET` | JWT signing secret | Yes | - |
| `JWT_EXPIRES_IN` | JWT expiration (seconds) | No | `3600` |
| `ADMIN_PASSWORD` | Admin authentication password | Yes | - |
| `FRONTEND_BASE_URL` | Frontend application URL | No | `http://localhost:5173` |
| `CORS_ALLOWED_ORIGINS` | CORS allowed origins (comma-separated) | No | `http://localhost:5173` |
| `KEYCLOAK_CLIENT_ID` | Keycloak client ID | No | - |
| `KEYCLOAK_CLIENT_SECRET` | Keycloak client secret | No | - |
| `KEYCLOAK_SERVER_URL` | Keycloak OIDC well-known URL | No | - |
| `TENCENTCLOUD_SECRET_ID` | Tencent Cloud API credential | No | - |
| `TENCENTCLOUD_SECRET_KEY` | Tencent Cloud API credential | No | - |

**Security Note**: Generate a strong `JWT_SECRET` using:

```bash
openssl rand -hex 32
```

### API Endpoints

#### Authentication

- `POST /api/verify` - Admin/token login
- `POST /api/verify-token` - Verify JWT token
- `GET /api/login` - Keycloak SSO login
- `GET /api/auth` - Keycloak callback
- `GET /api/logout` - Keycloak logout
- `GET /api/profile` - Get user profile

#### Event Management (Admin only)

- `GET /api/events` - List all events
- `POST /api/events` - Create event
- `PUT /api/events/:id` - Update event
- `DELETE /api/events/:id` - Delete event
- `GET /api/events/:id/token` - Get event access token

#### File Operations

- `POST /api/events/:id/upload-pic` - Upload event photo
- `POST /api/upload/:id/:face` - Upload user avatar
- `GET /api/events/:id/faces` - Get detected faces
- `DELETE /api/events/:id/faces/:filename` - Delete face

---

## 中文

高性能的 Go 语言头像换脸应用后端，支持活动管理、人脸检测和多种认证方式。

### 功能特性

- **活动管理**：创建和管理换脸活动，支持访问令牌
- **人脸检测**：集成腾讯云人脸检测 API
- **多种认证方式**：
  - 管理员密码认证
  - 活动令牌访问
  - Keycloak SSO 单点登录
- **文件存储**：自动化管理活动照片和人脸图像存储
- **QQ 集成**：获取 QQ 头像和昵称
- **RESTful API**：简洁的 API 设计，基于 JWT 的身份认证

### 技术栈

- Go 1.23.2
- Gin Web 框架
- SQLite 数据库
- JWT 身份认证
- 腾讯云人脸检测 API
- Keycloak（可选的 SSO）

### 项目结构

```
.
├── cmd/server/          # 应用程序入口
├── internal/
│   ├── config/          # 配置加载
│   ├── database/        # 数据库初始化
│   ├── handler/         # HTTP 处理器（路由）
│   ├── middleware/      # 认证中间件
│   ├── model/           # 数据模型
│   ├── repository/      # 数据库操作
│   ├── service/         # 业务逻辑
│   └── storage/         # 文件存储工具
├── pkg/response/        # 共享响应工具
└── data/                # SQLite 数据库 + 文件存储
```

### 快速开始

#### 前置要求

- Go 1.23 或更高版本
- （可选）[Air](https://github.com/cosmtrek/air) 用于热重载
- （可选）Docker 和 Docker Compose 用于容器化部署

#### 安装

```bash
# 克隆仓库
git clone <repository-url>
cd avatar-face-swap-go

# 复制环境配置文件
cp .env.example .env

# 编辑 .env 文件配置
# 必需：PORT, JWT_SECRET, ADMIN_PASSWORD
# 可选：Keycloak 凭证、腾讯云凭证
```

#### 开发模式

```bash
# 使用热重载运行（推荐）
air

# 或直接运行
go run cmd/server/main.go
```

#### 生产构建

```bash
# 构建二进制文件
go build -o bin/server cmd/server/main.go

# 运行
./bin/server
```

#### Docker 部署

```bash
# 使用 Docker Compose 构建并运行
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建镜像
docker-compose up -d --build
```

或手动构建 Docker 镜像：

```bash
# 构建镜像
docker build -t avatar-face-swap-go .

# 运行容器
docker run -d \
  -p 5001:5001 \
  -v $(pwd)/data:/app/data \
  --env-file .env \
  --name avatar-face-swap \
  avatar-face-swap-go
```

### 配置说明

主要环境变量：

| 变量名 | 说明 | 是否必需 | 默认值 |
| ------ | ---- | -------- | ------ |
| `PORT` | 服务器端口 | 是 | `5001` |
| `GIN_MODE` | Gin 模式 (debug/release) | 否 | `debug` |
| `DATABASE_URL` | SQLite 数据库文件路径 | 否 | `./data/app.db` |
| `STORAGE_DIR` | 文件存储目录 | 否 | `./data/storage` |
| `JWT_SECRET` | JWT 签名密钥 | 是 | - |
| `JWT_EXPIRES_IN` | JWT 过期时间（秒） | 否 | `3600` |
| `ADMIN_PASSWORD` | 管理员密码 | 是 | - |
| `FRONTEND_BASE_URL` | 前端应用 URL | 否 | `http://localhost:5173` |
| `CORS_ALLOWED_ORIGINS` | CORS 允许的源（逗号分隔） | 否 | `http://localhost:5173` |
| `KEYCLOAK_CLIENT_ID` | Keycloak 客户端 ID | 否 | - |
| `KEYCLOAK_CLIENT_SECRET` | Keycloak 客户端密钥 | 否 | - |
| `KEYCLOAK_SERVER_URL` | Keycloak OIDC 配置地址 | 否 | - |
| `TENCENTCLOUD_SECRET_ID` | 腾讯云 API 凭证 | 否 | - |
| `TENCENTCLOUD_SECRET_KEY` | 腾讯云 API 凭证 | 否 | - |

**安全提示**：使用以下命令生成强密钥：

```bash
openssl rand -hex 32
```

### API 接口

#### 身份认证

- `POST /api/verify` - 管理员/令牌登录
- `POST /api/verify-token` - 验证 JWT 令牌
- `GET /api/login` - Keycloak SSO 登录
- `GET /api/auth` - Keycloak 回调
- `GET /api/logout` - Keycloak 登出
- `GET /api/profile` - 获取用户信息

#### 活动管理（仅管理员）

- `GET /api/events` - 列出所有活动
- `POST /api/events` - 创建活动
- `PUT /api/events/:id` - 更新活动
- `DELETE /api/events/:id` - 删除活动
- `GET /api/events/:id/token` - 获取活动访问令牌

#### 文件操作

- `POST /api/events/:id/upload-pic` - 上传活动照片
- `POST /api/upload/:id/:face` - 上传用户头像
- `GET /api/events/:id/faces` - 获取检测到的人脸
- `DELETE /api/events/:id/faces/:filename` - 删除人脸
