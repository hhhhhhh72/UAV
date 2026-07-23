---
name: data-models
description: 核心业务实体、状态枚举、状态机流转规则
metadata:
  type: reference
---

# 数据模型与状态机

所有实体在 `internal/domain/models.go` (612行)，使用 UUID/ULID 主键、created_at/updated_at/version(乐观锁)、deleted_at 软删除。金额用 BIGINT 分(`amount_fen`)。

## 核心实体

- **User**: id, wechat_openid, phone_ciphertext, status, Role
- **Actor**: {ID, Role} — 认证后注入 context
- **Enterprise**: owner_user_id, name, license_url, account_name, status(5态)
- **Demand**: publisher_id, title, biz_type(6类), district, budget_fen, status(6态)
- **DemandBid**: demand_id, bidder_id, amount_fen, proposal, status
- **Job/Resume/JobApplication**: 招聘三件套，AppStatus 6态
- **Post/Comment/Report**: 社区三件套，先审后发
- **Listing**: 二手商品，listed/sold/removed
- **LabourOrder/Quote/Assignment**: 用工派遣
- **Contract**: 合同+签章回调，6态
- **TrainingCourse/Certificate/Instructor/CertifiedPilot/Enrollment**: 培训认证
- **DroneProduct/RepairOrder/TradeOrder**: 无人机交易
- **InsurancePolicy/AnnualInspection**: 保险年审
- **LoanApplication**: 金融贷款
- **EscrowAccount/EscrowTransaction**: 资金托管
- **Article/Review/Venue/VenueBooking**: 资讯/评价/场地
- **Banner/HomeQuickEntry**: 首页配置
- **FileRecord**: 文件上传记录(含SHA256去重)
- **Message**: 站内信

## 关键状态机

| 领域 | 状态 | 合法迁移 |
|------|------|------|
| 企业入驻 | draft→submitted→approved/rejected/supplement_required | supplement→submitted |
| 需求 | pending→published→matched→completed; 可cancelled/rejected | 双确认完成 |
| 职位 | draft→published→closed→archived | closed→published |
| 投递 | submitted→viewed→interviewing→offered→rejected/withdrawn | |
| 用工 | draft→submitted→quoted→confirmed→fulfilled→settled; 可cancelled | |
| 合同 | draft→sent→signing→signed→voided/expired | |

## 角色 (4级RBAC)

platform_admin > association_admin > enterprise > individual
权限矩阵见 docs/architecture.md
