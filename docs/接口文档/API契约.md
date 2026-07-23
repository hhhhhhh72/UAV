# API 契约

## 规范

- 基础路径: `/api/v1/`
- 认证: `Authorization: Bearer <token>`
- 分页: `?page=1&page_size=20` (最大100)

## 响应格式

```json
// 成功
{"data": {...}, "request_id": "req_xxx"}

// 分页
{"data": [...], "page": 1, "page_size": 20, "total": 100, "request_id": "req_xxx"}

// 失败
{"error": {"code": "FORBIDDEN", "message": "..."}, "request_id": "req_xxx"}
```

## 端点清单 (212条)

| 模块 | 端点 | 说明 |
|------|------|------|
| **认证** | `POST /auth/wechat/login` `POST /auth/refresh` `POST /auth/logout` `GET /me` | 微信登录/令牌刷新/登出 |
| **企业** | `GET/POST/PATCH /enterprises` `POST /enterprises/{id}/submit` `POST /admin/enterprises/{id}/review` `POST /admin/enterprises/batch-review` | 入驻/审核/批量 |
| **需求** | `GET/POST /demands` `POST /demands/{id}/applications` `POST /demands/{id}/applications/{aid}/select` `POST /demands/{id}/complete` `POST /demands/{id}/dispute` | 发布/竞标/选中/完成/争议 |
| **招聘** | `GET/POST /jobs` `GET/POST /resumes` `POST /applications` `PATCH /applications/{id}/status` | 职位/简历/投递 |
| **社区** | `GET/POST /posts` `POST /posts/{id}/comments` `POST /reports` | 帖子/评论/举报 |
| **二手** | `GET/POST /listings` `POST /listings/{id}/favorites` | 商品/收藏 |
| **用工** | `GET/POST /labour-orders` `POST /labour-orders/{id}/quote` | 订单/报价 |
| **合同** | `GET/POST /contracts` `POST /contracts/{id}/void` `POST /webhooks/signing` | 创建/作废/签章回调 |
| **托管** | `POST /escrow/deposit` `POST /escrow/freeze` `POST /escrow/release` `POST /escrow/refund` `GET /escrow/balance` | 充值/冻结/释放/退款 |
| **培训** | `GET/POST /certificates` `GET/POST /training-courses` `GET/POST /instructors` `GET/POST /certified-pilots` `POST /training-courses/{id}/pay-and-enroll` | 证书/课程/教员/飞手 |
| **交易** | `GET/POST /products` `GET/POST /repairs` | 商品/维修 |
| **保险** | `GET/POST /policies` `GET/POST /inspections` | 保单/年审 |
| **金融** | `GET/POST /loans` | 贷款 |
| **资讯** | `GET/POST /articles` `POST /articles/{id}/publish` | 发布/列表 |
| **评价** | `GET/POST /reviews` `POST /admin/reviews/{id}/approve` | 提交/审核 |
| **场地** | `GET/POST /venues` `POST /venues/{id}/book` | 场地/预约 |
| **消息** | `GET /messages` `POST /messages/{id}/read` `GET /messages/unread-count` | 列表/已读 |
| **管理** | `GET /admin/dashboard` `GET /admin/export/demands` `GET /admin/export/enterprises` `POST /admin/demands/batch-approve` | 看板/导出/批量 |
| **专家** | `GET /experts` `POST/PUT/DELETE /admin/experts` | 智库 |
| **案例** | `GET /cases` `POST/PUT/DELETE /admin/cases` | 案例库 |
| **合规** | `GET /compliance-docs` `GET /compliance-standards` `POST /admin/compliance-*` | 法规/标准 |
| **报告** | `GET /industry-reports` `POST/DELETE /admin/industry-reports` | 行业报告 |
| **品牌** | `GET /portfolios` `GET /portfolios/mine` `POST/PUT /portfolios` | 会员品牌 |
| **成果** | `GET /achievements` `POST/PUT/DELETE /achievements` | 科研成果 |
| **难题** | `GET /rd-challenges` `POST/PUT /rd-challenges` | 研发难题 |
| **攻关** | `GET /research-projects` `POST/PUT /admin/research-projects` | 联合攻关 |
| **申报** | `GET /project-applications/mine` `GET/POST /admin/project-applications` `POST /admin/project-applications/{id}/review` | 项目申报 |
| **赛事** | `GET /competitions` `POST /admin/competitions` `POST /competitions/{id}/register` | 赛事/报名 |
| **活动** | `GET /events` `POST /admin/events` `POST /events/{id}/register` | 活动/报名 |
| **资源** | `GET /industry-resources` `POST/PUT /admin/industry-resources` | 产业资源 |
| **应急** | `GET /emergency-resources` `GET /emergency-dispatches` `POST /admin/emergency-*` `GET /emergency-depts` `GET /emergency-drills` `GET /rescue-cases` | 应急协同 |
| **展会** | `GET /exhibitions` `POST /admin/exhibitions` `POST /exhibitions/{id}/booths` | 展会/展位 |
| **测试** | `GET /test-sites` `POST /admin/test-sites` `POST /test-sites/{id}/book` | 测试预约 |
| **转化** | `GET /transformations` `POST /admin/transformations` `POST /transformations/{id}/advance` | 成果转化 |
| **院校** | `GET /colleges` `POST /admin/colleges` | 院校展示 |
| **校企** | `GET /cooperation-programs` `POST /cooperation-programs` | 校企共建 |
| **院校** | `GET /resource-pools` `POST /admin/resource-pools` | 资源池 |
| **协会** | `GET /association-members` `GET /association-members/me` `POST /admin/association-members` | 协会权限 |
| **匹配** | `GET /recommendations` `GET /match` | 智能匹配 |
