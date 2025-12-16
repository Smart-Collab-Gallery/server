# Smart Collab Gallery Server

智能协同云图库后端服务 - 基于 Kratos 微服务框架

## 🚀 快速开始

### 环境要求

- Go 1.23.12+
- MySQL 5.7+
- Redis (可选)
- Consul (可选，用于配置中心)

### 安装依赖

```bash
# 安装 Kratos CLI
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest

# 安装开发工具
make init
```

### 构建运行

```bash
# 方式 1：直接构建运行（本地开发）
make build
./bin/smart-collab-gallery-server -conf ./configs

# 方式 2：使用环境变量（推荐）
export APP_NAME=smart-collab-gallery-server
export APP_VERSION=v1.0.0
export CONSUL_ADDRESS=127.0.0.1:8500
./bin/smart-collab-gallery-server -conf ./configs

# 方式 3：使用启动脚本（自动加载 .env 文件）
cp .env.example .env
# 编辑 .env 文件
./scripts/start.sh
```

服务端口：
- HTTP: `http://localhost:8000`
- gRPC: `localhost:9000`

## 📚 功能特性

### ✅ 已实现功能

- **用户认证**
  - 用户注册（MD5+盐加密）
  - 用户登录（JWT Token）
  - 获取当前登录用户
  - 用户注销

- **权限控制**
  - 基于角色的访问控制（RBAC）
  - 支持普通用户（user）和管理员（admin）角色
  - 中间件级别权限校验（类似 Java Spring `@AuthCheck`）
  - 灵活的权限匹配策略

- **配置管理**
  - 支持本地 YAML 配置
  - 集成 Consul 配置中心
  - JWT 配置化管理

### 🔧 技术栈

- **框架**: Kratos v2 (Go 微服务框架)
- **数据库**: MySQL + GORM v1.25.12
- **认证**: JWT (golang-jwt/jwt/v5)
- **授权**: 基于中间件的角色权限控制
- **配置中心**: Consul
- **依赖注入**: Google Wire
- **架构**: Clean Architecture (Service → Biz → Data)

## 📖 API 文档

### 用户接口

#### 1. 用户注册
```http
POST /api/v1/user/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

#### 2. 用户登录
```http
POST /api/v1/user/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}

Response:
{
  "token": "eyJhbGc..."
}
```

#### 3. 获取当前用户
```http
GET /api/v1/user/current
Authorization: Bearer <token>

Response:
{
  "id": "1",
  "username": "testuser",
  ...
}
```

#### 4. 用户注销
```http
POST /api/v1/user/logout
Authorization: Bearer <token>
```

## 🔧 配置说明

### 环境变量配置（推荐）

支持通过环境变量进行配置，适用于容器化部署：

| 环境变量 | 说明 | 示例 |
|---------|------|------|
| `APP_NAME` | 应用名称（Consul Key） | `smart-collab-gallery-server` |
| `APP_VERSION` | 应用版本 | `v1.0.0` |
| `CONSUL_ADDRESS` | Consul 地址 | `127.0.0.1:8500` |
| `CONSUL_TOKEN` | Consul 令牌（可选） | `your-token` |

详细说明：[环境变量配置文档](docs/environment-variables.md)

### 基础配置 (configs/config.yaml)

**⚠️ 重要：时间配置必须使用秒数格式（如 `604800s`），不能使用 Go 格式（如 `168h`）**

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s

data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local

auth:
  jwt_secret: "your-secret-key"
  jwt_expire: 604800s  # 7天 = 604800秒

consul:
  enabled: false              # 是否启用 Consul
  address: "127.0.0.1:8500"   # Consul 地址
```

**时间换算表**：
- 1小时 = `3600s`
- 1天 = `86400s`
- 7天 = `604800s`
- 30天 = `2592000s`

详细说明：[配置文件格式文档](docs/config-format.md)

### Consul 配置中心

服务支持从 Consul 动态加载配置：

1. **构建时注入服务名称**：
   ```bash
   make build  # 注入 Name=smart-collab-gallery-server
   ```

2. **Consul 中存储配置**：
   ```bash
   # Key 必须与服务名称一致
   consul kv put smart-collab-gallery-server @configs/config.yaml
   ```

3. **启用 Consul**：
   ```yaml
   consul:
     enabled: true
     address: "127.0.0.1:8500"
   ```

详细说明请参考：
- [Consul 配置详细文档](docs/consul-config.md)
- [Consul 快速开始](docs/consul-quickstart.md)
- [Consul 流程图](docs/consul-flow.md)

### 权限控制

系统实现了基于中间件的角色权限控制，类似于 Java Spring 的 `@AuthCheck` 注解功能：

**角色定义**：
- `user` - 普通用户
- `admin` - 管理员

**使用示例**：

```go
// 在 HTTP Server 中配置管理员权限
selector.Server(
    middleware.RequireAdmin(),
).Match(NewAdminOnlyMatcher()).Build()

// 在业务层手动校验
userRole := middleware.GetUserRoleFromContext(ctx)
if middleware.UserRole(userRole) != middleware.RoleAdmin {
    return ErrorNoAuth
}
```

**详细文档**：
- [权限校验详细文档](docs/role-authorization.md) - 完整指南
- [权限校验快速参考](docs/role-authorization-quickstart.md) - 快速上手

## 🛠️ 开发指南

### 生成代码

```bash
# 生成 API 代码（proto → Go）
make api

# 生成配置代码
make config

# 生成依赖注入代码（Wire）
make generate

# 一次性生成所有代码
make all
make all
```
## Automated Initialization (wire)
```
# install wire
go get github.com/google/wire/cmd/wire

# generate wire
cd cmd/server
wire
```

## Docker
```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

