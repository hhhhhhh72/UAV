---
name: architecture-conventions
description: 分层架构约束、中间件链、命名约定、错误处理规范
metadata:
  type: reference
---

# 架构与编码规范

## 严格分层（铁律）

```
Handler     → 只做：解析请求、调 Service、调 respond/fail
              禁做：SQL 查询、业务规则判断、直接读写库

Service     → 只做：角色校验、归属校验、状态机迁移、调 Repository 接口
              禁做：http.Request、http.ResponseWriter、JSON 编解码、环境变量

Repository  → 只做：SQL 执行、数据映射、返回 Domain 对象
              禁做：业务规则、权限判断、HTTP 相关
```

违反案例：
- Handler 里写 `db.Query(...)` — 越界
- Service 里写 `json.NewEncoder(w).Encode(...)` — 越界
- Repository 里写 `if a.Role != "admin"` — 越界

## 中间件链

请求 → 幂等去重(24h) → 限流(100/s, Token Bucket) → 链路追踪(X-Request-ID) → Panic恢复 → CORS(白名单) → Token鉴权(Bearer HMAC-SHA256, 15min过期) → 业务处理

## 命名约定

| 层级 | 约定 | 示例 |
|------|------|------|
| 路由 | `/api/v1/资源复数` | `/api/v1/demands` |
| Handler | `func (s *Server) verbNoun(...)` | `createDemand` |
| Service | `type NounService struct` | `DemandService` |
| Repository接口 | `type NounRepository interface` | `DemandRepository` |
| PG实现 | `type nounRepo struct` (小写) | `demandRepo` |
| 迁移文件 | `00000X_description.up.sql` | `000001_init.up.sql` |

## 响应格式

```go
// 成功
respond(w, r, http.StatusOK, data)          // → {"data": ..., "request_id": "..."}
paginatedRespond(w, r, items, totalCount)   // → {"data": [...], "page": 1, "page_size": 20, "total": N, ...}
// 失败
fail(w, r, statusCode, err)                 // → {"error": {"code": "...", "message": "..."}, ...}
```

错误码: 400 VALIDATION_ERROR / 401 UNAUTHENTICATED / 403 FORBIDDEN / 404 NOT_FOUND / 409 CONFLICT / 422 STATE_INVALID / 429 RATE_LIMITED / 500 INTERNAL

## 认证

- Token: HMAC-SHA256 签名，格式 `base64(payload).base64(sig)`，payload 含 `{sub, role, exp}`
- TokenManager 在 internal/httpapi/auth.go
- 公开路径白名单在 isPublicPath() 函数中（16个GET端点）
- authenticatedActor(r) 从 context 获取当前用户
- 管理员: platform_admin > association_admin > enterprise > individual

## 启动与配置

- 入口: cmd/api/main.go → config.Load() → 选择PG/内存存储 → NewServer(所有Service) → ListenAndServe
- 环境变量: AUTH_SECRET(必填32字节+), DATABASE_URL(可选，不设则内存), WECHAT_APPID/APPSECRET, ADMIN_DEV_MODE, ENCRYPTION_KEY, CORS_ORIGINS, LOG_DIR/LEVEL, HTTP_ADDR
- start.bat: Windows开发启动脚本(编译+运行)
- stop.bat: kill drone-api.exe
