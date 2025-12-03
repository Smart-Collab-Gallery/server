# 用户登录接口文档

## 接口信息

- **接口路径**: `/api/user/login`
- **请求方法**: `POST`
- **Content-Type**: `application/json`
- **认证方式**: JWT Token

## 请求参数

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| user_account | string | 是 | 用户账号，至少4个字符 | "testuser" |
| user_password | string | 是 | 用户密码，至少8个字符 | "password123" |

## 请求示例

```bash
curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser",
    "user_password": "password123"
  }'
```

## 响应参数

### 成功响应 (200)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "user_account": "testuser",
    "user_name": "无名",
    "user_avatar": "",
    "user_profile": "",
    "user_role": "user",
    "vip_number": 0,
    "vip_expire_time": "",
    "create_time": "2025-12-03T12:00:00Z"
  }
}
```

| 参数名 | 类型 | 说明 |
|--------|------|------|
| token | string | JWT 认证令牌，有效期 7 天 |
| user | object | 用户信息对象 |
| user.id | int64 | 用户 ID |
| user.user_account | string | 用户账号 |
| user.user_name | string | 用户昵称 |
| user.user_avatar | string | 用户头像 URL |
| user.user_profile | string | 用户简介 |
| user.user_role | string | 用户角色 (user/admin) |
| user.vip_number | int64 | 会员编号 |
| user.vip_expire_time | string | 会员过期时间 (RFC3339 格式) |
| user.create_time | string | 创建时间 (RFC3339 格式) |

### 错误响应

#### 参数错误 (400)

```json
{
  "code": 400,
  "reason": "PARAMS_ERROR",
  "message": "参数为空"
}
```

#### 账号错误 (400)

```json
{
  "code": 400,
  "reason": "ACCOUNT_ERROR",
  "message": "账号错误"
}
```

#### 密码错误 (400)

```json
{
  "code": 400,
  "reason": "PASSWORD_ERROR",
  "message": "密码错误"
}
```

#### 用户不存在或密码错误 (401)

```json
{
  "code": 401,
  "reason": "USER_NOT_EXIST_OR_PASSWORD_ERROR",
  "message": "用户不存在或密码错误"
}
```

## 业务逻辑

1. **参数校验**
   - 检查账号、密码是否为空
   - 检查账号长度 >= 4
   - 检查密码长度 >= 8

2. **密码加密**
   - 使用 MD5 + 盐值 ("yupi") 加密用户输入的密码
   - 加密算法: `md5(salt + password)`

3. **查询用户**
   - 根据账号和加密后的密码查询数据库
   - 查询用户是否存在且未被删除

4. **生成 JWT Token**
   - 生成包含用户信息的 JWT Token
   - Token 有效期: 7 天
   - Token 包含: user_id, user_account, user_role

5. **返回登录信息**
   - 返回 Token 供后续请求认证使用
   - 返回用户基本信息（不包含密码等敏感信息）

## JWT Token 说明

### Token 结构

```
Header.Payload.Signature
```

### Payload 包含的信息

```json
{
  "user_id": 1,
  "user_account": "testuser",
  "user_role": "user",
  "exp": 1733234567,  // 过期时间
  "iat": 1732629767,  // 签发时间
  "nbf": 1732629767,  // 生效时间
  "iss": "smart-collab-gallery"  // 签发者
}
```

### Token 使用方式

在后续需要认证的请求中，在 Header 中携带 Token：

```bash
curl -X GET http://localhost:8000/api/user/info \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 测试用例

### 正常登录

```bash
curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser",
    "user_password": "password123"
  }'
```

预期响应: 包含 token 和用户信息的 JSON 对象

### 参数为空

```bash
curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "",
    "user_password": ""
  }'
```

预期响应:
```json
{"code": 400, "reason": "PARAMS_ERROR", "message": "参数为空"}
```

### 账号过短

```bash
curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "abc",
    "user_password": "password123"
  }'
```

预期响应:
```json
{"code": 400, "reason": "ACCOUNT_ERROR", "message": "账号错误"}
```

### 密码过短

```bash
curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser",
    "user_password": "1234567"
  }'
```

预期响应:
```json
{"code": 400, "reason": "PASSWORD_ERROR", "message": "密码错误"}
```

### 账号或密码错误

```bash
curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "wronguser",
    "user_password": "wrongpassword"
  }'
```

预期响应:
```json
{"code": 401, "reason": "USER_NOT_EXIST_OR_PASSWORD_ERROR", "message": "用户不存在或密码错误"}
```

## 安全注意事项

1. **密码加密**: 用户密码在数据库中存储的是加密后的值，登录时需对输入的密码进行相同算法加密后比对
2. **Token 安全**: 
   - Token 包含敏感信息，前端应安全存储（如 httpOnly cookie 或加密的 localStorage）
   - Token 有效期为 7 天，过期后需重新登录
3. **防暴力破解**: 生产环境建议添加登录频率限制
4. **HTTPS**: 生产环境必须使用 HTTPS 传输
5. **错误信息**: 账号不存在和密码错误返回相同错误信息，防止账号枚举

## 技术实现

### 项目结构

```
api/user/v1/
├── user.proto              # 添加 Login 接口定义
└── error_reason.proto      # 添加登录相关错误码

internal/
├── biz/user.go            # 添加 Login() 方法
├── data/user.go           # 添加 GetUserByAccountAndPassword()
├── service/user.go        # 添加 Login() 服务方法
└── pkg/jwt.go             # JWT 工具类 (新增)
```

### 登录流程

```
客户端请求 (账号+密码)
    ↓
HTTP/gRPC Server
    ↓
Service Layer
    ├── 调用 Biz 层登录
    ├── 生成 JWT Token
    └── 构建返回数据
    ↓
Biz Layer
    ├── 参数校验
    ├── 密码加密
    └── 调用 Repo 查询用户
    ↓
Data Layer
    └── 根据账号和加密密码查询数据库
    ↓
返回 Token + 用户信息
```

## 配置要求

无需额外配置，使用现有数据库配置即可。

**注意**: JWT 密钥在 `internal/pkg/jwt.go` 中定义，生产环境请修改为安全的密钥。

## 与注册接口的配合使用

1. 用户先调用注册接口创建账号
2. 注册成功后调用登录接口获取 Token
3. 使用 Token 访问需要认证的接口
