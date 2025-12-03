# 用户注册接口文档

## 接口信息

- **接口路径**: `/api/user/register`
- **请求方法**: `POST`
- **Content-Type**: `application/json`

## 请求参数

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| user_account | string | 是 | 用户账号，至少4个字符 | "testuser" |
| user_password | string | 是 | 用户密码，至少8个字符 | "password123" |
| check_password | string | 是 | 确认密码，需与密码一致 | "password123" |

## 请求示例

```bash
curl -X POST http://localhost:8000/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser",
    "user_password": "password123",
    "check_password": "password123"
  }'
```

## 响应参数

### 成功响应 (200)

```json
{
  "user_id": 1
}
```

| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_id | int64 | 新创建的用户 ID |

### 错误响应

#### 参数错误 (400)

```json
{
  "code": 400,
  "reason": "PARAMS_ERROR",
  "message": "参数为空"
}
```

#### 账号过短 (400)

```json
{
  "code": 400,
  "reason": "ACCOUNT_TOO_SHORT",
  "message": "用户账号过短，至少4个字符"
}
```

#### 密码过短 (400)

```json
{
  "code": 400,
  "reason": "PASSWORD_TOO_SHORT",
  "message": "用户密码过短，至少8个字符"
}
```

#### 密码不一致 (400)

```json
{
  "code": 400,
  "reason": "PASSWORD_NOT_MATCH",
  "message": "两次输入的密码不一致"
}
```

#### 账号重复 (409)

```json
{
  "code": 409,
  "reason": "ACCOUNT_DUPLICATE",
  "message": "账号已存在"
}
```

#### 系统错误 (500)

```json
{
  "code": 500,
  "reason": "SYSTEM_ERROR",
  "message": "注册失败，数据库错误"
}
```

## 业务逻辑

1. **参数校验**
   - 检查账号、密码、确认密码是否为空
   - 检查账号长度 >= 4
   - 检查密码长度 >= 8
   - 检查两次密码是否一致

2. **账号唯一性检查**
   - 查询数据库检查账号是否已存在
   - 如已存在则返回错误

3. **密码加密**
   - 使用 MD5 + 盐值 ("yupi") 加密密码
   - 加密算法: `md5(salt + password)`

4. **创建用户**
   - 默认用户名: "无名"
   - 默认用户角色: "user"
   - 插入数据库并返回用户 ID

## 测试用例

### 正常注册

```bash
curl -X POST http://localhost:8000/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser001",
    "user_password": "password123",
    "check_password": "password123"
  }'
```

预期响应:
```json
{"user_id": 1}
```

### 账号过短

```bash
curl -X POST http://localhost:8000/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "abc",
    "user_password": "password123",
    "check_password": "password123"
  }'
```

预期响应:
```json
{"code": 400, "reason": "ACCOUNT_TOO_SHORT", "message": "用户账号过短，至少4个字符"}
```

### 密码过短

```bash
curl -X POST http://localhost:8000/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser",
    "user_password": "1234567",
    "check_password": "1234567"
  }'
```

预期响应:
```json
{"code": 400, "reason": "PASSWORD_TOO_SHORT", "message": "用户密码过短，至少8个字符"}
```

### 密码不一致

```bash
curl -X POST http://localhost:8000/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser",
    "user_password": "password123",
    "check_password": "password456"
  }'
```

预期响应:
```json
{"code": 400, "reason": "PASSWORD_NOT_MATCH", "message": "两次输入的密码不一致"}
```

## 技术实现

### 项目结构

```
api/user/v1/
├── user.proto              # API 定义
├── error_reason.proto      # 错误码定义
├── user.pb.go             # 生成的消息代码
├── user_http.pb.go        # 生成的 HTTP 路由代码
├── user_grpc.pb.go        # 生成的 gRPC 代码
└── error_reason_errors.go # 错误辅助函数

internal/
├── biz/user.go            # 业务逻辑层
├── data/user.go           # 数据访问层
├── data/user_entity.go    # 数据模型
└── service/user.go        # 服务层
```

### 分层架构

1. **Service 层** (`internal/service/user.go`)
   - 处理 HTTP/gRPC 请求
   - 参数转换和响应封装

2. **Biz 层** (`internal/biz/user.go`)
   - 核心业务逻辑
   - 参数校验
   - 密码加密
   - 调用 Repo 接口

3. **Data 层** (`internal/data/user.go`)
   - 数据库操作
   - GORM ORM
   - 实现 Repo 接口

## 配置要求

确保 `configs/config.yaml` 中配置了正确的数据库连接:

```yaml
data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/your_database?parseTime=True&loc=Local
```

## 运行服务

```bash
# 启动服务
./bin/smart-collab-gallery-server -conf ./configs

# 服务会监听
# HTTP: 0.0.0.0:8000
# gRPC: 0.0.0.0:9000
```
