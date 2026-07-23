# 数据模型与存储设计

推荐 PostgreSQL 15+。所有业务表使用 UUID/ULID 主键、`created_at`、`updated_at`、`version`（乐观锁）；删除优先采用 `deleted_at` 软删除。金额使用 `BIGINT` 分，状态使用受约束的文本枚举。

## 身份与权限

| 表 | 关键字段 | 约束/索引 |
|---|---|---|
| `users` | id、wechat_openid、phone_ciphertext、status | `wechat_openid` 唯一；手机号密文不做普通索引 |
| `roles` | id、code、name | `code` 唯一 |
| `user_roles` | user_id、role_id、scope_type、scope_id | 唯一组合；支持协会/企业作用域 |
| `refresh_tokens` | id、user_id、token_hash、expires_at、revoked_at | `token_hash` 唯一；按 user_id、expires_at 索引 |
| `audit_logs` | actor_id、action、resource_type、resource_id、result、request_id、metadata | 按 resource 与时间、actor 与时间索引；metadata 禁止存敏感原文 |

## 企业与需求

| 表 | 关键字段 | 说明 |
|---|---|---|
| `enterprises` | owner_user_id、name、credit_code_ciphertext、status、city_code | 信用代码展示时掩码；owner 索引、状态+城市索引 |
| `enterprise_documents` | enterprise_id、file_id、document_type、review_status | 私有文件；审核结果与原因独立记录 |
| `review_records` | resource_type、resource_id、action、reason、reviewer_id | 企业、需求、内容复用的不可变审核历史 |
| `demands` | publisher_id、enterprise_id、category、city_code、district、budget_fen、status、accepted_application_id | `status, city_code, created_at desc` 复合索引 |
| `demand_bids` | demand_id、bidder_id、bidder_name、amount_fen、proposal、status | `(demand_id)` 索引，`(bidder_id)` 索引；竞标报价持久化 |
| `demand_applications` | demand_id、applicant_id、proposal、status | `(demand_id, applicant_id)` 唯一，防止重复承接 |

## 招聘、内容、订单与合同

| 表 | 关键字段 | 说明 |
|---|---|---|
| `jobs` / `resumes` / `job_applications` | 所属方、可见范围、状态、内容版本 | 简历正文和附件均受访问控制 |
| `posts` / `comments` / `reports` | 作者、内容、审核状态、父资源 | 评论不可越过帖子审核状态展示 |
| `listings` / `listing_favorites` | 作者、分类、状态、位置概略、价格 | 收藏唯一索引 `(listing_id, user_id)` |
| `labour_orders` / `quotes` / `assignments` | 需求方、供给方、人数、日期、金额、状态 | 日期范围校验；订单参与者索引 |
| `contract_templates` / `contracts` / `contract_events` | 模板版本、参与者、签章单号、状态、事件 ID | 外部事件 ID 唯一，回调可幂等 |

## 扩展业务表（Phase 3+）

| 表 | 关键字段 | 说明 |
|---|---|---|
| `training_courses` | org_id、title、cert_type、start/end_date、max_students、price_fen | 培训课程 |
| `enrollments` | course_id、user_id、status | 课程报名 |
| `certificates` | user_id、cert_type、cert_number、level、issue/expire_date、status | 证书管理（CAAC/UTC/人社） |
| `instructors` | user_id、name、cert_types、bio、org_id、status | 教练认证 |
| `certified_pilots` | user_id、real_name、cert_ids、flight_hours、rating | 飞手认证 |
| `drone_products` | seller_id、prod_type、title、brand、model、condition、price_fen | 无人机商品 |
| `repair_orders` | customer_id、product_desc、fault_desc、quote_fen、status | 维修服务 |
| `trade_orders` | product_id、buyer_id、seller_id、amount_fen、status | 交易订单 |
| `insurance_policies` | user_id、drone_model、drone_sn、policy_type、premium_fen、coverage_fen | 保险保单 |
| `annual_inspections` | user_id、drone_model、drone_sn、inspect/expire_date、result | 年审记录 |
| `loan_applications` | user_id、amount_fen、term_months、purpose、status | 贷款申请 |
| `escrow_accounts` | user_id、balance_fen、frozen_fen | 资金托管账户 |
| `escrow_transactions` | from_user、to_user、amount_fen、tx_type、reference | 资金交易记录 |
| `articles` | title、content、summary、category、source、is_pinned、status | 行业资讯 |
| `reviews` | reviewer_id、target_type、target_id、rating、content、status | 信用评价（先审后发） |
| `venues` | owner_id、name、venue_type、location、price_fen | 场地管理 |
| `venue_bookings` | venue_id、user_id、start_time、end_time、status | 场地预约 |
| `banners` | image_url、link_url、sort_order | 首页轮播 |
| `home_quick_entries` | key、name、icon_url、link_url、sort_order | 首页快捷入口 |

## 文件、事件与幂等

| 表 | 关键字段 | 说明 |
|---|---|---|
| `files` | storage_key、sha256、content_type、size_bytes、visibility、owner_id | 对象存储 key 不直接作为公开 URL；按 sha256 可做去重 |
| `idempotency_keys` | user_id、key、request_hash、response_status、response_body、expires_at | `(user_id, key)` 唯一；请求体不一致返回 409 |
| `outbox_events` | aggregate_type、aggregate_id、event_type、payload、published_at | 与业务事务同写，异步投递通知/搜索索引 |

## 数据保留与访问

- 日志默认保留 180 天，审计日志保留期限按合规要求另行配置，且需防篡改存储。
- 用户注销后先冻结账号，异步删除或匿名化可删除信息；法定保留的合同、结算、审计数据应隔离保存。
- 生产环境密钥由密钥管理服务托管；字段加密须使用信封加密并支持轮换。
- 备份必须加密，定期进行恢复演练；测试环境不得使用生产数据或未脱敏副本。

## 新增业务模块表（第3步开发）

| 表名 | 说明 | 关键字段 |
|------|------|------|
| `experts` | 专家智库 | id, name, title, org, field, tags(JSONB), bio |
| `case_entries` | 企业案例库 | id, title, category, images(JSONB), client_name, result |
| `compliance_docs` | 合规知识库 | id, title, category, content, tags(JSONB) |
| `standard_docs` | 团体标准库 | id, title, std_number, version, issue_date, file_url |
| `project_applications` | 项目申报 | id, applicant_id, project_name, budget_fen, attachments(JSONB), review_note |
| `achievements` | 科技成果库 | id, owner_id, title, achieve_type, field, stage, contact_info |
| `rd_challenges` | 研发难题广场 | id, poster_id, title, field, budget_fen, deadline |
| `research_projects` | 课题联合攻关 | id, title, field, lead_org, members(JSONB), milestones |
| `competitions` | 赛事管理 | id, title, category, location, start_date, max_teams, sponsor |
| `competition_registrations` | 赛事报名 | id, competition_id, user_id, team_name, member_count |
| `association_events` | 协会活动 | id, title, event_type, location, start_time, max_attendees, cover_url |
| `event_registrations` | 活动报名 | id, event_id, user_id, name, phone |
| `member_portfolios` | 会员品牌展示 | id, enterprise_id, name, logo_url, products(JSONB), honors(JSONB) |
| `industry_reports` | 行业报告 | id, title, period, category, summary, file_url |
| `industry_resources` | 产业资源台账 | id, owner_id, name, res_type, model, location, price_fen |
| `emergency_resources` | 应急资源 | id, owner_id, name, res_type, specs, quantity, contact_info |
| `emergency_dispatches` | 应急调度记录 | id, resource_id, event_desc, location, commander, result |
