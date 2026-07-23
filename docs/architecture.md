# 架构设计

## 技术栈

| 端 | 技术 |
|----|------|
| 微信小程序 | 原生 + Vant Weapp 1.11，27 页面 + 7 分包 |
| Web 管理后台 | 嵌入式 SPA（Go embed），ECharts 5 图表 |
| 后端 API | Go 1.22+，标准库 net/http，120 条端点 |
| 数据库 | PostgreSQL 15+（生产）/ 内存存储（开发），30 张表 |
| 基础设施 | config + logger + cache + middleware + crypto |

## 分层架构

```
cmd/api          启动入口，组装依赖
internal/httpapi 路由、中间件、请求响应处理（仅做：解析请求→调Service→respond/fail）
internal/service 业务规则、权限校验、状态流转（禁做：HTTP 相关、JSON 编解码）
internal/repository 数据持久化（PostgreSQL/内存双模式，禁做：业务规则、权限判断）
internal/domain  业务实体与常量定义
```

## 中间件链

```
请求 → 幂等去重(24h) → 限流(100/s) → 链路追踪 → Panic恢复 → CORS → Token鉴权 → 业务处理
```

## 权限矩阵

| 角色 | 权限范围 |
|------|------|
| `platform_admin` | 全部管理权限（需求/用工/合同审核、用户管理、数据导出） |
| `association_admin` | 企业入驻审核、会员管理、行业资讯、内容审核 |
| `enterprise` | 发布/承接需求、招聘、用工申请、合同创建 |
| `individual` | 接单、求职、二手交易、社区发布 |

## 公开端点（无需认证）

16 个 GET 端点免登录访问：`/home`、`/search`、`/demands`、`/posts`、`/comments`、`/jobs`、`/listings`、`/training-courses`、`/instructors`、`/certified-pilots`、`/products`、`/articles`、`/venues`、`/reviews`、`/contract-templates`、`/image`。

## 安全能力

| 能力 | 实现 |
|------|------|
| 身份认证 | HMAC-SHA256 Bearer Token，15min 过期，支持 Refresh |
| 角色控制 | 4 级 RBAC，服务端强制校验 |
| 数据隔离 | 企业仅查看自身合同和用工数据 |
| 操作审计 | 所有写操作自动记录 actor/action/resource/result |
| 接口防刷 | Token bucket 100/s，超限 429 |
| 幂等保护 | Idempotency-Key，24h 内去重 |
| 数据加密 | AES-256-GCM 加密手机号、身份证 |
| 输出脱敏 | 公开接口屏蔽联系方式、精确坐标 |
| 管理后台保护 | `ADMIN_DEV_MODE` 环境变量，生产 403 |

## 生产部署

```bash
# 需要的环境变量
AUTH_SECRET=至少32字节随机密钥
DATABASE_URL=postgres://user:pass@host:5432/drone_platform
WECHAT_APPID=小程序AppID
WECHAT_APPSECRET=小程序AppSecret
CORS_ORIGINS=https://your-domain.com
LOG_DIR=./logs
LOG_LEVEL=info
# ADMIN_DEV_MODE 不设置（生产禁用管理后台）
```

## 配置模块

`internal/config/config.go` 提供集中配置管理：`Load()` 从环境变量读取，`Validate()` 启动时检查必填项和安全性，`Print()` 脱敏打印配置摘要。
