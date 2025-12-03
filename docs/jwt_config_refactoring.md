# JWT 配置重构说明

## 修改内容

将 JWT 密钥和过期时间从硬编码改为从配置文件读取，提高了系统的安全性和可维护性。

## 变更文件

### 1. 配置文件定义 (`internal/conf/conf.proto`)

新增 `Auth` 配置消息：

```protobuf
message Auth {
  string jwt_secret = 1;                        // JWT 密钥
  google.protobuf.Duration jwt_expire = 2;      // JWT 过期时间
}
```

在 `Bootstrap` 中添加：

```protobuf
message Bootstrap {
  Server server = 1;
  Data data = 2;
  Auth auth = 3;  // 新增
}
```

### 2. 配置文件 (`configs/config.yaml`)

新增 auth 配置段：

```yaml
auth:
  jwt_secret: "smart-collab-gallery-secret-key-change-in-production"
  jwt_expire: 168h  # 7天 = 7 * 24h = 168h
```

### 3. JWT 工具类重构 (`internal/pkg/jwt.go`)

**修改前**：
```go
const (
    JWTSecret = "your-secret-key-change-in-production"
    TokenExpireDuration = time.Hour * 24 * 7
)

func GenerateToken(userID int64, userAccount string, userRole string) (string, error) {
    // 使用全局常量
}
```

**修改后**：
```go
type JWTManager struct {
    secret string
    expire time.Duration
}

func NewJWTManager(secret string, expire time.Duration) *JWTManager {
    return &JWTManager{
        secret: secret,
        expire: expire,
    }
}

func (m *JWTManager) GenerateToken(userID int64, userAccount string, userRole string) (string, error) {
    // 使用实例的配置
}
```

### 4. Service 层更新 (`internal/service/service.go`)

添加 JWT 管理器的 Provider：

```go
func NewJWTManager(c *conf.Auth) *pkg.JWTManager {
    return pkg.NewJWTManager(c.JwtSecret, c.JwtExpire.AsDuration())
}
```

### 5. UserService 更新 (`internal/service/user.go`)

注入 JWTManager：

```go
type UserService struct {
    v1.UnimplementedUserServer
    
    uc         *biz.UserUsecase
    jwtManager *pkg.JWTManager  // 新增
    log        *log.Helper
}

func NewUserService(uc *biz.UserUsecase, jwtManager *pkg.JWTManager, logger log.Logger) *UserService {
    return &UserService{
        uc:         uc,
        jwtManager: jwtManager,  // 注入
        log:        log.NewHelper(logger),
    }
}
```

使用 JWTManager：

```go
// 修改前
token, err := pkg.GenerateToken(user.ID, user.UserAccount, user.UserRole)

// 修改后
token, err := s.jwtManager.GenerateToken(user.ID, user.UserAccount, user.UserRole)
```

### 6. Wire 依赖注入更新

**wire.go**:
```go
func wireApp(*conf.Server, *conf.Data, *conf.Auth, log.Logger) (*kratos.App, func(), error)
```

**main.go**:
```go
app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, logger)
```

## 优势

### ✅ 安全性提升
- JWT 密钥不再硬编码在代码中
- 生产环境可以使用不同的密钥
- 密钥可以通过环境变量或加密配置管理

### ✅ 可维护性提升
- 配置与代码分离
- 不同环境可以使用不同的配置
- 修改配置无需重新编译代码

### ✅ 灵活性提升
- Token 过期时间可配置
- 开发环境可以使用较长的过期时间
- 生产环境可以使用较短的过期时间

## 配置建议

### 开发环境

```yaml
auth:
  jwt_secret: "dev-secret-key"
  jwt_expire: 720h  # 30天，方便开发
```

### 生产环境

```yaml
auth:
  jwt_secret: "${JWT_SECRET}"  # 从环境变量读取
  jwt_expire: 168h  # 7天
```

### 安全建议

1. **密钥强度**：使用至少 32 字符的随机字符串
2. **密钥轮换**：定期更换 JWT 密钥
3. **环境变量**：生产环境通过环境变量注入密钥
4. **密钥管理**：使用密钥管理服务（如 AWS KMS, HashiCorp Vault）

## 生成强密钥

```bash
# 使用 openssl 生成随机密钥
openssl rand -base64 32

# 或使用 Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

## 测试验证

```bash
# 构建项目
go build -o ./bin/ ./...

# 运行服务（确保 config.yaml 中配置了 auth 段）
./bin/smart-collab-gallery-server -conf ./configs

# 测试登录接口
curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"user_account":"testuser","user_password":"password123"}'
```

## 迁移检查清单

- [x] 更新 conf.proto 添加 Auth 配置
- [x] 生成配置代码 `make config`
- [x] 更新 config.yaml 添加 auth 配置
- [x] 重构 JWT 工具类为可配置的 Manager
- [x] 创建 JWTManager Provider
- [x] 更新 UserService 注入 JWTManager
- [x] 更新 Wire 依赖注入配置
- [x] 重新生成 Wire 代码
- [x] 构建验证
- [x] 测试验证

## 总结

通过这次重构，我们将 JWT 配置从硬编码迁移到了配置文件，使系统更加安全、灵活和易于维护。这是一个标准的最佳实践，符合 12-Factor App 的配置管理原则。
