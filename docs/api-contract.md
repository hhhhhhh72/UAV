# API 契约草案（v1）

本文件定义 P0/P1 的实现边界。正式实现后应由 OpenAPI 3.1 文档生成 SDK 和接口测试；本文件不替代 OpenAPI。

## 通用约定

- 基础路径：`/api/v1`；仅 `GET /healthz` 不需要认证。
- 认证：`Authorization: Bearer <access_token>`。access token 不携带可被客户端修改的业务权限；服务端仍须查询资源归属。
- 内容类型：写请求使用 `application/json`。文件先申请私有上传凭证，再写入文件 ID。
- 幂等：创建订单、提交审核、发起签约等不可重复副作用的请求必须带 `Idempotency-Key`（8–128 个可打印字符）。同一用户相同键在 24 小时内返回首次结果。
- 时间：请求/响应均使用 RFC 3339 UTC；金额使用最小货币单位整数，例如 `amount_fen`。
- 分页：列表使用 `cursor`、`page_size`（默认 20，最大 100）和稳定排序；响应返回 `items`、`next_cursor`。

## 响应与错误

成功：

```json
{"data": {}, "request_id": "req_..."}
```

失败：

```json
{"error":{"code":"FORBIDDEN","message":"没有该资源的访问权限"},"request_id":"req_..."}
```

| HTTP | 错误码 | 适用场景 |
|---|---|---|
| 400 | `VALIDATION_ERROR` | 字段格式、枚举、时间范围错误 |
| 401 | `UNAUTHENTICATED` | 缺失、失效或撤销令牌 |
| 403 | `FORBIDDEN` | 角色或资源归属不满足 |
| 404 | `NOT_FOUND` | 资源不存在，或为避免枚举而隐藏资源 |
| 409 | `CONFLICT` | 状态迁移冲突、重复提交、乐观锁冲突 |
| 422 | `STATE_INVALID` | 业务状态不允许当前操作 |
| 429 | `RATE_LIMITED` | 限流触发 |
| 500 | `INTERNAL` | 未预期错误；不得返回内部细节 |

## 认证与个人资料

| 方法 | 路径 | 权限 | 说明 | 状态 |
|---|---|---|---|---|
| POST | `/auth/wechat/login` | 匿名 | 微信 code → openid → 创建/登录 → Token | ✅ |
| POST | `/auth/refresh` | refresh token | 轮换 refresh token | ✅ |
| POST | `/auth/logout` | 已登录 | 撤销 refresh token | ✅ |
| GET | `/me` | 已登录 | 当前用户信息汇总 | ✅ |
| PATCH | `/me` | 已登录 | 修改昵称等公开资料 | ✅ |

## 公开端点（无需认证）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 健康检查（server + storage 探针） |
| GET | `/admin` | 管理后台 SPA（`ADMIN_DEV_MODE=true` 时可用） |
| GET | `/api/v1/home` | 首页数据（城市+Banner+快捷入口+热门需求） |
| GET | `/api/v1/search` | 全局搜索 |
| GET | `/api/v1/image` | 图片缩放/格式转换 + 磁盘缓存 |

## 企业与协会

| 方法 | 路径 | 权限 | 状态 |
|---|---|---|---|
| GET | `/enterprises` | 已登录 | ✅ |
| POST | `/enterprises` | 已登录 | ✅ |
| PATCH | `/enterprises/{id}` | 企业所有者 | ✅ |
| POST | `/enterprises/{id}/submit` | 企业所有者 | ✅ |
| POST | `/enterprises/{id}/documents` | 企业所有者 | ✅ |
| GET | `/admin/enterprises` | 协会/平台管理员 | ✅ |
| GET | `/admin/enterprises/pending` | 协会/平台管理员 | ✅ |
| POST | `/admin/enterprises/{id}/review` | 协会管理员 | ✅ |
| POST | `/admin/members/import` | 协会管理员 | ✅ |

## 需求大厅

| 方法 | 路径 | 权限 | 状态 |
|---|---|---|---|
| GET | `/demands` | 公开 | ✅ |
| POST | `/demands` | 企业/个人 | ✅ |
| PATCH | `/demands/{id}` | 发布者 | ✅ |
| POST | `/demands/{id}/submit` | 发布者 | ✅ |
| POST | `/admin/demands/{id}/review` | 协会管理员 | ✅ |
| POST | `/admin/demands/{id}/approve` | 协会管理员 | ✅ |
| POST | `/demands/{id}/applications` | 企业/个人（竞标） | ✅ |
| POST | `/demands/{id}/applications/{appId}/select` | 发布者（选标） | ✅ 含归属校验+原子 CAS |
| POST | `/demands/{id}/complete` | 参与方（双确认） | ✅ |
| POST | `/demands/{id}/dispute` | 参与方（争议） | ✅ |

> **2026-07-21 更新**: 竞标报价(Bid) 已持久化到 `demand_bids` 表，SelectBid 使用 CompareAndSetStatus 原子操作防并发双接受。

## 招聘求职 / 培训认证 / 交易 / 保险 / 金融

| 资源 | 端点 | 状态 |
|---|---|---|
| 职位 | `POST/GET /jobs`, `POST /jobs/{id}/publish`, `POST /jobs/{id}/close` | ✅ |
| 简历 | `POST/GET/PATCH /resumes` | ✅ |
| 投递 | `POST /applications`, `PATCH /applications/{id}/status`, `GET /applications` | ✅ |
| 课程 | `POST/GET /training-courses`, `POST /training-courses/{id}/enroll`, `GET {id}/enrollments` | ✅ |
| 证书 | `POST /certificates`, `GET /certificates/mine`, `GET /certificates/expiring` | ✅ |
| 教练/飞手 | `POST/GET /instructors`, `POST/GET /certified-pilots` | ✅ |
| 商品 | `POST/GET /products` | ✅ |
| 维修 | `POST /repairs`, `GET /repairs/mine` | ✅ |
| 订单 | `POST /trade-orders`, `GET /trade-orders/mine`, `PATCH {id}/status` | ✅ |
| 保单 | `POST /policies`, `GET /policies/mine` | ✅ |
| 年审 | `POST /inspections`, `GET /inspections/mine`, `GET /inspections/expiring` | ✅ |
| 贷款 | `POST /loans`, `GET /loans/mine` | ✅ |

## 内容 / 交易 / 资金

| 资源 | 端点 | 状态 |
|---|---|---|
| 社区 | `POST/GET /posts`, `POST {id}/publish/remove/comments`, `GET /comments`, `POST /reports` | ✅ |
| 二手 | `POST/GET /listings`, `POST {id}/close`, `POST {id}/favorites` | ✅ |
| 用工 | `POST/GET /labour-orders`, `POST {id}/quote`, `GET /quotes`, `POST /assignments` | ✅ |
| 合同 | `GET /contract-templates`, `GET/POST /contracts`, `POST {id}/void`, `POST /webhooks/signing` | ✅ Webhook 签章回调更新合同状态+状态机校验 |
| 托管 | `POST /escrow/deposit/freeze/release/refund`, `GET /escrow/balance/transactions` | ✅ |
| 场地 | `POST/GET /venues`, `POST {id}/book` | ✅ |
| 评价 | `POST/GET /reviews`, `GET /admin/reviews`, `POST /admin/reviews/{id}/approve/reject` | ✅ |
| 资讯 | `POST/GET /articles`, `POST {id}/publish` | ✅ |
| 消息 | `GET /messages`, `POST {id}/read`, `GET /unread-count` | ✅ |
| 文件 | `POST /files/upload` | ✅ |
| 管理 | `GET /admin/dashboard`, `GET /admin/export/demands`, `GET /admin/export/enterprises`, `POST /admin/demands/batch-approve` | ✅ |

| 匹配 | `GET /recommendations` (智能推荐), `GET /match` (搜索+评分) | ✅ |
| 专家 | `GET /experts`, `POST/GET/PUT/DELETE /admin/experts` | ✅ |
| 案例 | `GET /cases`, `POST/PUT/DELETE /admin/cases` | ✅ |
| 合规 | `GET /compliance-docs`, `GET /compliance-standards`, `POST /admin/compliance-*` | ✅ |
| 报告 | `GET /industry-reports`, `POST/DELETE /admin/industry-reports` | ✅ |
| 品牌 | `GET /portfolios`, `GET /portfolios/mine`, `POST /portfolios`, `PUT /portfolios/{id}` | ✅ |
| 成果 | `GET /achievements`, `POST/PUT/DELETE /achievements` | ✅ |
| 难题 | `GET /rd-challenges`, `POST/PUT /rd-challenges` | ✅ |
| 攻关 | `GET /research-projects`, `POST/PUT /admin/research-projects` | ✅ |
| 申报 | `GET /project-applications/mine`, `GET/POST /admin/project-applications`, `POST {id}/review` | ✅ |
| 赛事 | `GET /competitions`, `POST /admin/competitions`, `GET {id}/registrations`, `POST {id}/register` | ✅ |
| 活动 | `GET /events`, `POST /admin/events`, `GET {id}/registrations`, `POST {id}/register` | ✅ |
| 资源 | `GET /industry-resources`, `POST/PUT /admin/industry-resources` | ✅ |
| 应急 | `GET /emergency-resources`, `GET /emergency-dispatches`, `POST /admin/emergency-*` | ✅ |

> **总计：172 条 API 端点，全部已实现。**
