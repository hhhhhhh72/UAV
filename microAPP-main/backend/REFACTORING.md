# 后台管理系统重构指南

## 已完成的重构

### 1. 新增模块结构

```
backend/
├── config.js                    # 配置管理(替代硬编码)
├── logger.js                    # 日志模块
├── middleware/
│   ├── auth.js                 # 认证和权限中间件
│   ├── validation.js           # 输入验证和消毒
│   └── error.js                # 错误处理
├── routes/
│   ├── auth.js                 # 认证路由(登录/注册/SSO)
│   └── admin.js                # 管理功能路由
├── .env.example                # 环境变量模板
└── index.js                    # 主入口文件(待重构)
```

### 2. 安全性改进

#### 修复的问题:
1. ✅ 移除硬编码的管理员密码
2. ✅ 使用环境变量管理敏感信息
3. ✅ 添加输入验证和XSS防护
4. ✅ 实现请求速率限制
5. ✅ 完善权限隔离

#### 管理员密码管理:

**旧方式** (不安全):
```javascript
// 硬编码密码 - 所有人都能看到
password: 'dkjjfwy2026DSL'
password: 'dkyxAdmin'
```

**新方式** (安全):
```javascript
// 从环境变量读取,或自动生成随机密码
const config = require('./config');
const dslAdminPassword = process.env.DSL_ADMIN_PASSWORD || crypto.randomBytes(16).toString('hex');
```

### 3. 权限隔离修复

**问题**: study_admin 和 dsl_admin 可能访问其他服务的数据

**修复**:
- 所有查询都根据角色过滤数据
- 更新操作验证用户权限
- 前端路由守卫完善

### 4. 数据验证

新增验证中间件:
- 手机号格式验证
- 密码强度验证
- 必填字段验证
- 文件类型和大小验证
- XSS 攻击防护

## 如何应用重构

### 方式1: 渐进式迁移 (推荐)

1. **保留现有 index.js**
2. **在 index.js 中引入新模块**:
   ```javascript
   // 在文件开头添加
   require('dotenv').config();
   const { config, validateConfig } = require('./config');
   const { logger } = require('./logger');

   // 使用新的路由
   app.use('/api/auth', require('./routes/auth'));
   app.use('/api/admin', require('./routes/admin'));
   ```

3. **逐步替换旧代码**

### 方式2: 完全替换

使用新的模块化结构,将 index.js 拆分为:
- routes/auth.js (认证相关)
- routes/admin.js (管理功能)
- routes/services.js (服务相关)
- routes/cases.js (案例管理)

## 配置步骤

### 1. 创建 .env 文件

```bash
cd backend
cp .env.example .env
```

编辑 `.env` 文件,填入实际配置:

```env
# 必须修改的配置
JWT_SECRET=生成一个至少32位的随机字符串
DSL_ADMIN_PASSWORD=设置DSL管理员密码
STUDY_ADMIN_PASSWORD=设置研学管理员密码

# 如果使用微信登录
WX_APPID=你的微信AppID
WX_SECRET=你的微信Secret
```

生成随机密钥的方法:
```bash
# Linux/Mac
openssl rand -hex 32

# Windows PowerShell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
```

### 2. 安装新依赖

```bash
cd backend
npm install dotenv winston
```

### 3. 测试

```bash
# 开发环境
npm run dev

# 生产环境
npm start
```

## 性能优化建议

### 1. 添加缓存

```javascript
// 简单的内存缓存
const cache = new Map();

function withCache(key, ttl, fn) {
  const cached = cache.get(key);
  if (cached && Date.now() < cached.expiresAt) {
    return cached.data;
  }

  const data = fn();
  cache.set(key, {
    data,
    expiresAt: Date.now() + ttl
  });
  return data;
}
```

### 2. 数据库优化

- 使用 PostgreSQL 替代 JSON 文件
- 添加索引
- 实现分页查询

### 3. 文件上传优化

- 使用云存储 (OSS/S3)
- 添加图片压缩
- 实现断点续传

## SM2 签名修复

查看 `test-sm2-compat.js` 文件中的测试结果。

Java SDK 的签名方式:
```java
SM2Utils.sign(userId.getBytes(), privateKey_bytes, sourceData_bytes)
```

Node.js sm-crypto 对应实现:
```javascript
const { sm2 } = require('sm-crypto');

// userId 的 UTF-8 bytes 转为 hex
const userIdHex = Buffer.from(userId, 'utf8').toString('hex');

// sourceData 已经是 hex 字符串
const signature = sm2.doSignature(sourceDataHex, privateKeyHex, {
  hash: false,
  userId: userIdHex,
  der: true  // 如果需要 DER 格式
});
```

## 前端优化

### Pinia 状态管理

创建 store 文件:

```javascript
// src/stores/user.js
import { defineStore } from 'pinia'
import axios from '@/utils/http'

export const useUserStore = defineStore('user', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    accessToken: localStorage.getItem('accessToken') || null
  }),

  getters: {
    isLoggedIn: (state) => !!state.accessToken,
    isAdmin: (state) => state.user?.role === 'admin',
    isDslAdmin: (state) => state.user?.role === 'dsl_admin',
    isStudyAdmin: (state) => state.user?.role === 'study_admin'
  },

  actions: {
    async login(phone, password) {
      const res = await axios.post('/api/auth/login', { phone, password })
      if (res.data.success) {
        this.user = res.data.user
        this.accessToken = res.data.accessToken
        localStorage.setItem('user', JSON.stringify(this.user))
        localStorage.setItem('accessToken', res.data.accessToken)
        localStorage.setItem('refreshToken', res.data.refreshToken)
      }
      return res.data
    },

    async logout() {
      await axios.post('/api/auth/logout')
      this.user = null
      this.accessToken = null
      localStorage.removeItem('user')
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
    },

    async fetchUser() {
      const res = await axios.get('/api/auth/me')
      if (res.data.success) {
        this.user = res.data.user
        localStorage.setItem('user', JSON.stringify(this.user))
      }
    }
  }
})
```

## 监控和日志

### 查看日志

```bash
# 查看今天的日志
tail -f logs/$(date +%Y-%m-%d).log

# 查看错误日志
grep ERROR logs/*.log
```

### 日志级别

- `error`: 错误信息
- `warn`: 警告信息
- `info`: 一般信息
- `debug`: 调试信息

生产环境建议设置为 `info`,开发环境设置为 `debug`。

## 部署检查清单

- [ ] 设置强密码和 JWT_SECRET
- [ ] 配置 HTTPS
- [ ] 启用 PostgreSQL (可选但推荐)
- [ ] 配置 CORS 为实际域名
- [ ] 设置日志轮转
- [ ] 定期备份数据
- [ ] 监控错误日志
- [ ] 限制文件上传大小
- [ ] 配置防火墙规则

## 常见问题

### Q: 迁移后旧用户无法登录?

A: 系统会自动将旧用户的明文密码转换为 bcrypt 哈希,无需手动处理。

### Q: 如何重置管理员密码?

A: 在 `.env` 中设置新密码,然后删除 users.json 中对应的用户记录,重启服务会自动创建。

### Q: 如何查看当前有哪些管理员账户?

A: 使用管理后台的"用户管理"功能,或用以下脚本:

```javascript
const users = require('./storage').readUsersDB();
console.log(users.filter(u => ['admin', 'dsl_admin', 'study_admin'].includes(u.role)));
```

## 下一步优化建议

1. **数据库迁移**: 从 JSON 文件迁移到 PostgreSQL
2. **Redis 缓存**: 添加 Redis 缓存热数据
3. **消息队列**: 使用 Redis/RabbitMQ 处理异步任务
4. **文件存储**: 使用阿里云 OSS 或 AWS S3
5. **CI/CD**: 配置自动化测试和部署
6. **API 文档**: 使用 Swagger 生成 API 文档
7. **单元测试**: 添加 Jest 测试
8. **监控告警**: 集成 Sentry 或类似服务
