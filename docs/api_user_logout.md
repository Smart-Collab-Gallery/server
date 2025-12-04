# 用户注销接口文档

## 接口信息

- **接口路径**: `/api/user/logout`
- **请求方法**: `POST`
- **需要认证**: ✅ 是（需要 JWT Token）
- **Content-Type**: `application/json`

## 请求参数

无请求体参数，从 HTTP Header 的 Authorization 中获取 JWT Token。

### HTTP Headers

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| Authorization | string | 是 | JWT Token | "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." |

## 请求示例

```bash
curl -X POST http://localhost:8000/api/user/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

## 响应参数

### 成功响应 (200)

```json
{
  "success": true
}
```

| 参数名 | 类型 | 说明 |
|--------|------|------|
| success | bool | 是否注销成功 |

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

## 业务逻辑

1. **Token 验证**（由 JWT 中间件自动完成）
   - 从 Authorization Header 提取 Token
   - 验证 Token 格式（Bearer 前缀）
   - 验证 Token 签名和有效期
   - 解析用户 ID 并注入到 Context

2. **注销处理**
   - 从 Context 获取当前登录用户 ID
   - 记录注销日志
   - 执行清理工作（可选）
     * 记录注销日志到数据库
     * 清理用户相关缓存
     * 发送注销通知等

3. **返回结果**
   - 返回注销成功标志

## JWT 注销说明

### ⚠️ 重要说明

由于 JWT 是**无状态**的 Token 机制，服务端不存储 Token 信息，因此：

1. **客户端负责删除 Token**
   - 客户端需要在收到注销成功响应后，立即删除本地存储的 Token
   - 删除 localStorage/sessionStorage 中的 Token
   - 清除内存中的 Token 变量

2. **Token 失效时机**
   - 服务端调用注销接口后，Token 本身并不会立即失效
   - Token 会在其过期时间到达后自然失效
   - 如果需要立即失效，需要实现 Token 黑名单机制（可选）

3. **安全建议**
   - 前端注销后立即跳转到登录页
   - 清除所有用户相关的本地数据
   - 如有需要，可以实现 Token 黑名单（Redis）

## Token 黑名单机制（可选扩展）

如果需要服务端主动让 Token 失效，可以实现黑名单：

```
1. 用户注销时，将 Token 加入黑名单（Redis）
2. 设置过期时间为 Token 的剩余有效期
3. JWT 中间件验证时，检查黑名单
4. 如在黑名单中，拒绝请求
```

## 完整使用流程

```bash
# 1. 登录获取 Token
LOGIN_RESPONSE=$(curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_account": "testuser",
    "user_password": "password123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

# 2. 使用 Token 访问需要认证的接口
curl -X GET http://localhost:8000/api/user/get/login \
  -H "Authorization: Bearer $TOKEN"

# 3. 注销
curl -X POST http://localhost:8000/api/user/logout \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'

# 4. 注销后，Token 应该被客户端删除
# 前端代码示例：
# localStorage.removeItem('token');
# sessionStorage.removeItem('token');
# 跳转到登录页
```

## 前端实现示例

### JavaScript/TypeScript

```typescript
// 注销函数
async function logout() {
  try {
    // 调用注销接口
    const response = await fetch('/api/user/logout', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      },
      body: '{}'
    });

    if (response.ok) {
      // 删除本地存储的 Token
      localStorage.removeItem('token');
      sessionStorage.removeItem('token');
      
      // 清除用户信息
      localStorage.removeItem('userInfo');
      
      // 跳转到登录页
      window.location.href = '/login';
    }
  } catch (error) {
    console.error('注销失败:', error);
  }
}
```

### React 示例

```typescript
import { useNavigate } from 'react-router-dom';

function useLogout() {
  const navigate = useNavigate();

  const logout = async () => {
    try {
      const token = localStorage.getItem('token');
      
      await fetch('/api/user/logout', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: '{}'
      });

      // 清除 Token 和用户信息
      localStorage.removeItem('token');
      localStorage.removeItem('userInfo');
      
      // 跳转到登录页
      navigate('/login');
    } catch (error) {
      console.error('注销失败:', error);
    }
  };

  return logout;
}
```

## 测试用例

### 正常注销

```bash
# 先登录
TOKEN=$(curl -X POST http://localhost:8000/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"user_account":"testuser","user_password":"password123"}' \
  | jq -r '.token')

# 注销
curl -X POST http://localhost:8000/api/user/logout \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期响应:
```json
{"success": true}
```

### 未登录注销

```bash
curl -X POST http://localhost:8000/api/user/logout \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期响应:
```json
{"code": 401, "reason": "NOT_LOGIN_ERROR", "message": "未登录"}
```

### Token 无效

```bash
curl -X POST http://localhost:8000/api/user/logout \
  -H "Authorization: Bearer invalid_token" \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期响应:
```json
{"code": 401, "reason": "INVALID_TOKEN", "message": "Token 无效或已过期"}
```

## 注意事项

1. ✅ **客户端必须删除 Token** - 注销成功后立即删除本地存储的 Token
2. ✅ **跳转到登录页** - 删除 Token 后应立即跳转到登录页
3. ✅ **清除用户数据** - 清除所有用户相关的本地缓存数据
4. ⚠️ **Token 仍然有效** - 服务端注销后，Token 在过期前技术上仍然有效
5. 💡 **黑名单机制** - 如需立即使 Token 失效，可实现 Redis 黑名单

## 技术实现

- JWT 中间件自动验证 Token
- 从 Context 获取用户 ID
- 记录注销日志
- 无状态设计，客户端负责删除 Token

## 与 Session 方案的对比

| 特性 | Session 方案 | JWT 方案 (已实现) |
|------|--------------|-------------------|
| 注销方式 | 服务端删除 Session | 客户端删除 Token |
| 服务端存储 | 需要 | 不需要 |
| 立即失效 | ✅ 是 | ❌ 否（除非实现黑名单） |
| 分布式友好 | ❌ 需要共享 Session | ✅ 无状态 |
| 性能 | 需要查询 Session | 无需查询 |
