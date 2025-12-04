# 获取当前登录用户接口文档

## 接口信息

- **接口路径**: `/api/user/get/login`
- **请求方法**: `GET`
- **需要认证**: ✅ 是（需要在 Header 中携带 JWT Token）

## 请求头

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| Authorization | string | 是 | JWT Token，格式为 "Bearer {token}" | "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." |

## 请求参数

无需请求参数，用户信息从 JWT Token 中解析获取。

## 请求示例

```bash
curl -X GET http://localhost:8000/api/user/get/login \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 响应参数

### 成功响应 (200)

```json
{
  "user": {
    "id": 1,
    "user_account": "testuser",
    "user_name": "无名",
    "user_avatar": "",
    "user_profile": "",
    "user_role": "user",
    "vip_number": 0,
    "vip_expire_time": "",
    "create_time": "2025-12-03T14:00:00Z",
    "update_time": "2025-12-03T14:00:00Z"
  }
}
```

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user | LoginUserVO | 用户信息（已脱敏） |

### LoginUserVO 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int64 | 用户 ID |
| user_account | string | 用户账号 |
| user_name | string | 用户昵称 |
| user_avatar | string | 用户头像 URL |
| user_profile | string | 用户简介 |
| user_role | string | 用户角色（user/admin） |
| vip_number | int64 | 会员编号 |
| vip_expire_time | string | 会员过期时间（RFC3339 格式） |
| create_time | string | 创建时间（RFC3339 格式） |
| update_time | string | 更新时间（RFC3339 格式） |

**注意**: 响应中不包含敏感信息（如密码）。

### 错误响应

#### 未登录 (401)

```json
{
  "code": 401,
  "reason": "NOT_LOGIN_ERROR",
  "message": "未登录"
}
```

#### Token 无效 (401)

```json
{
  "code": 401,
  "reason": "INVALID_TOKEN",
  "message": "Token 无效或已过期"
}
```

#### 用户不存在 (404)

```json
{
  "code": 404,
  "reason": "USER_NOT_FOUND",
  "message": "用户不存在"
}
```

## 业务逻辑

1. **Token 验证**
   - 从 HTTP Header 中提取 Authorization 字段
   - 验证 Token 格式（Bearer 前缀）
   - 解析 JWT Token
   - 验证 Token 签名和有效期

2. **用户信息提取**
   - 从 Token Claims 中获取 user_id
   - 从上下文中读取用户 ID

3. **数据库查询**
   - 根据 user_id 从数据库查询用户完整信息
   - 确保用户未被删除（isDelete = 0）

4. **数据脱敏**
   - 不返回密码字段
   - 返回 LoginUserVO 视图对象

## 认证流程

```
客户端请求（带 Token）
    ↓
JWT Auth Middleware
    ├── 提取 Token
    ├── 验证 Token
    └── 解析用户信息到 Context
    ↓
Service Layer
    ├── 从 Context 获取 user_id
    └── 调用 Biz 层
    ↓
Biz Layer
    └── 调用 Data 层查询用户
    ↓
Data Layer
    └── GORM 查询数据库
    ↓
返回脱敏的用户信息
```

## 测试用例

### 正常获取用户信息

```bash
# 1. 先登录获取 Token
TOKEN=$(curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"user_account":"testuser","user_password":"password123"}' \
  | jq -r '.token')

# 2. 使用 Token 获取用户信息
curl -X GET http://localhost:8000/api/user/get/login \
  -H "Authorization: Bearer $TOKEN"
```

### 未携带 Token

```bash
curl -X GET http://localhost:8000/api/user/get/login
```

预期响应:
```json
{"code": 401, "reason": "NOT_LOGIN_ERROR", "message": "未登录"}
```

### Token 格式错误

```bash
curl -X GET http://localhost:8000/api/user/get/login \
  -H "Authorization: InvalidToken"
```

预期响应:
```json
{"code": 401, "reason": "INVALID_TOKEN", "message": "Token 格式错误"}
```

### Token 已过期

```bash
curl -X GET http://localhost:8000/api/user/get/login \
  -H "Authorization: Bearer expired_token"
```

预期响应:
```json
{"code": 401, "reason": "INVALID_TOKEN", "message": "Token 无效或已过期"}
```

## 技术实现

### 1. JWT 中间件

`internal/middleware/auth.go`:
- 实现 JWT 认证中间件
- 从 Header 提取和验证 Token
- 将用户信息注入到 Context

### 2. 白名单机制

不需要认证的接口:
- `/api/user/register` - 用户注册
- `/api/user/login` - 用户登录
- `/api/helloworld/greeter/SayHello` - Hello World

需要认证的接口:
- `/api/user/get/login` - 获取当前登录用户
- 其他所有接口（默认）

### 3. 选择器模式

使用 Kratos 的 `selector.Server()` 实现选择性应用中间件:

```go
selector.Server(
    middleware.JWTAuth(jwtManager),
).Match(NewWhiteListMatcher()).Build()
```

## 安全考虑

1. ✅ **Token 验证**: 每次请求都验证 Token 有效性
2. ✅ **数据脱敏**: 不返回密码等敏感信息
3. ✅ **数据库查询**: 实时从数据库获取最新用户信息
4. ✅ **软删除检查**: 确保用户未被删除
5. ✅ **Token 过期**: Token 有效期为 7 天（可配置）

## 性能优化建议

### 方案 1: 直接返回 Token 中的信息（性能优先）

```go
// 不查询数据库，直接从 Token 构建用户信息
func (s *UserService) GetLoginUser(ctx context.Context, req *v1.GetLoginUserRequest) (*v1.GetLoginUserReply, error) {
    userID := middleware.GetUserIDFromContext(ctx)
    userAccount := middleware.GetUserAccountFromContext(ctx)
    userRole := middleware.GetUserRoleFromContext(ctx)
    
    // 直接构建响应，不查数据库
    return &v1.GetLoginUserReply{
        User: &v1.LoginUserVO{
            Id: userID,
            UserAccount: userAccount,
            UserRole: userRole,
            // 其他字段可能不完整
        },
    }, nil
}
```

### 方案 2: Redis 缓存（平衡性能和数据一致性）

```go
// 1. 先从 Redis 获取
// 2. 如果 Redis 没有，从数据库查询
// 3. 将结果缓存到 Redis（TTL 5分钟）
```

### 方案 3: 数据库查询（数据一致性优先） ✅ 当前实现

每次都从数据库查询，确保数据是最新的。

## 使用场景

- 前端页面加载时获取当前用户信息
- 验证用户登录状态
- 获取用户权限信息
- 个人中心页面展示

## 接口对比

| 对比项 | Java Session 方案 | Go JWT 方案（已实现） |
|--------|------------------|----------------------|
| 用户状态存储 | 服务端 Session | JWT Token |
| 请求携带 | Cookie | Authorization Header |
| 扩展性 | 单机或需要 Session 共享 | 无状态，天然支持分布式 |
| 获取方式 | 从 Session 获取 | 解析 Token + 查数据库 |
| 数据实时性 | 高（直接查数据库） | 高（每次查数据库） |

## 完整测试脚本

```bash
#!/bin/bash

# 测试获取登录用户接口

BASE_URL="http://localhost:8000"

echo "===== 1. 用户登录 ====="
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/login" \
  -H "Content-Type: application/json" \
  -d '{"user_account":"testuser","user_password":"password123"}')

echo "$LOGIN_RESPONSE" | jq '.'

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "登录失败，请先注册用户"
    exit 1
fi

echo ""
echo "===== 2. 获取当前登录用户 ====="
curl -s -X GET "$BASE_URL/api/user/get/login" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo ""
echo "===== 3. 测试未携带 Token ====="
curl -s -X GET "$BASE_URL/api/user/get/login" | jq '.'

echo ""
echo "===== 4. 测试错误的 Token ====="
curl -s -X GET "$BASE_URL/api/user/get/login" \
  -H "Authorization: Bearer invalid_token" | jq '.'
```

保存为 `test_get_login_user.sh`，运行：

```bash
chmod +x test_get_login_user.sh
./test_get_login_user.sh
```
