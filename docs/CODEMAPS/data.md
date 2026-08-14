<!-- Generated: 2026-08-19 | Files scanned: 126 SQL | Token estimate: ~700 -->

# 数据架构地图（PostgreSQL 16 / 内存存储）

## 85 张表（63 组迁移，126 个 SQL 文件；demand_bids 历史重复建表 2 次）

| 域 | 表 |
|----|----|
| 认证/用户 | users, user_roles(⚠从未写入,历史遗留), roles, refresh_tokens, idempotency_keys |
| 企业 | enterprises, enterprise_documents, review_records |
| 供需 | demands, demand_bids(⚠两次建表:000002有外键/000012无), demand_intents |
| 招聘 | jobs, resumes, job_applications |
| 社区/二手/用工 | posts, comments, reports / listings, listing_favorites / labour_orders, labour_quotes, assignments, employment_requests |
| 合同 | contracts, contract_templates, contract_events |
| 培训 | certificates, training_courses, instructors, certified_pilots, training_enrollments |
| 交易 | drone_products, repair_orders, trade_orders |
| 保险/金融/托管 | insurance_policies, annual_inspections / loan_applications / escrow_accounts, escrow_transactions |
| 内容 | articles, banners, home_quick_entries, messages |
| 评价/场地 | reviews, venues, venue_bookings |
| 会员/智库 | experts, case_entries, industry_resources, member_portfolios |
| 创新 | achievements, rd_challenges, research_projects, project_applications |
| 运营 | competitions, competition_registrations, association_events, event_registrations, industry_reports |
| 应急 | emergency_resources, emergency_dispatches, emergency_depts, emergency_drills, rescue_cases |
| 资源池/展会/测试/转化 | resource_pools, resource_pool_members / exhibitions, exhibition_booths / test_sites, test_site_bookings / transformations |
| 院校/校企/店铺/协会 | colleges, study_tours / cooperation_programs / shops / association_members |
| 基建 | audit_logs, outbox_events, files |

## 迁移历史（63 组）

```
000001-000005  基础+业务模块+审核+社区+审计（users/enterprises/demands/contracts...）
000006-000012  首页配置/消息/报名订单/托管/资讯/评价场地/竞标重建
000013-000017  user_roles 列 + 17 新业务表 + 转化 + batch 13 表 + shops
000018-000032  补列（双确认/密码bcrypt/评价status/头像/资源可见级别/院校coop_type/
               企业档案7字段/简历7字段/飞手bio/报名12字段/线下成交金额...）
000033         意向登记 demand_intents
000034-000037  seed/级联（院校/应急初始数据、需求删除级联、研学数据）
000038-000043  转化阶段/应急状态归一/服务申报/展会封面/用户昵称/小程序页面字段对齐
000044-000047  小程序字段二轮/工单 work_orders/服务能力 service_listings
000048-000052  赛事种子/院校种子二轮/服务展示种子/服务图集/企业审核意见
000053-000057  废弃 shops/企业PRD字段/研学扩展/飞手头像地区/交易售后
000058-000063  用户档案/标准文档分类/资源预约/赛事报名扩展/合同模板种子/应急资源种子
```

## 关键设计

- **加密字段**：Contact/IDCard/LicenseURL/AccountName/PhoneCipher — memory 锁内加解密，PG SQL 前后加解密（AES-256-GCM）
- **分页规范**：page/page_size（max 100），PG 用 COUNT(*)+LIMIT，memory 全量过滤切片
- **迁移执行**：启动时 RunMigrationsFromDir 按文件名排序跑 .up.sql（幂等 IF NOT EXISTS；pg_advisory_lock 防多实例竞争 + schema_migrations 版本表）
- **级联**：需求删除级联（000036）、院校/应急 seed 数据（000034/000037）

## 迁移文件命名

`migrations/0000XX_description.up.sql` + 同名 `.down.sql`（成对）
