package service

import (
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// UserContentCleanupPlan 注销账号时的内容处置计划（策略在此，SQL 执行在仓储）。
//
// 口径（2026-09-11 定稿）：
//   - TakeDown 下架：在架/在售/在招的内容置为下架状态（内容保留，只是不再公开）——
//     否则会留下联系不上、永远挂着的僵尸条目；
//   - Wipe 擦除：报名/对接/预约等业务事实行保留，但清空其中的个人联系信息
//     （手机号/邮箱/证件号/执照影像）；
//   - Drop 删除：纯个人信息行整行删除（简历、投递、站内信、文件、收藏、飞手名录档案）；
//   - 不动：工单/合同/资金/保单/审计等履约与合规记录——只让账号身份失效。
func UserContentCleanupPlan() repository.ContentCleanupPlan {
	return repository.ContentCleanupPlan{
		TakeDown: []repository.TakeDownRule{
			{Table: "demands", OwnerColumn: "publisher_id", StatusColumn: "status", OnValue: string(domain.DemandPublished), OffValue: string(domain.DemandCancelled)},
			{Table: "drone_products", OwnerColumn: "seller_id", StatusColumn: "status", OnValue: "listed", OffValue: "removed"},
			{Table: "training_courses", OwnerColumn: "org_id", StatusColumn: "status", OnValue: "published", OffValue: "closed"},
			{Table: "posts", OwnerColumn: "author_id", StatusColumn: "status", OnValue: "published", OffValue: "draft"},
			{Table: "achievements", OwnerColumn: "owner_id", StatusColumn: "status", OnValue: "published", OffValue: "draft"},
		},
		Wipe: []repository.WipeRule{
			// 需求：联系方式（密文手机号）与发布者姓名一并擦掉（内容已下架，不再需要展示名）
			{Table: "demands", OwnerColumn: "publisher_id", Columns: []string{"contact", "publisher_name"}},
			{Table: "demand_intents", OwnerColumn: "intentor_id", Columns: []string{"contact"}},
			// 企业主体保留（工商信息可公开），但法定代表人/联系人/执照影像属于个人信息
			{Table: "enterprises", OwnerColumn: "owner_user_id", Columns: []string{"legal_person", "contact_person", "contact_phone", "email", "license_url"}},
			{Table: "event_registrations", OwnerColumn: "user_id", Columns: []string{"phone"}},
			{Table: "competition_registrations", OwnerColumn: "user_id", Columns: []string{"phone", "id_card"}},
			{Table: "training_enrollments", OwnerColumn: "user_id", Columns: []string{"phone", "email", "id_card"}},
			{Table: "study_tour_enrollments", OwnerColumn: "user_id", Columns: []string{"phone"}},
			{Table: "test_site_bookings", OwnerColumn: "user_id", Columns: []string{"contact_phone", "license_url"}},
			{Table: "industry_resource_bookings", OwnerColumn: "user_id", Columns: []string{"contact_phone"}},
			{Table: "training_courses", OwnerColumn: "org_id", Columns: []string{"phone", "certificate_url"}},
		},
		Drop: []repository.DropRule{
			{Table: "resumes", OwnerColumn: "user_id"},                  // 简历整份都是个人信息
			{Table: "job_applications", OwnerColumn: "applicant_id"},    // 投递记录指向简历
			{Table: "messages", OwnerColumn: "sender_id"},               // 站内信正文可能含联系方式
			{Table: "messages", OwnerColumn: "receiver_id"},
			{Table: "files", OwnerColumn: "owner_id"},                   // 上传文件台账（含私有影像）
			{Table: "uploads", OwnerColumn: "owner_id"},
			{Table: "certified_pilots", OwnerColumn: "user_id"},         // 飞手名录档案（公开名录里的个人资料）
			{Table: "demand_favorites", OwnerColumn: "user_id"},         // 以下为纯偏好数据，无留存价值
			{Table: "product_favorites", OwnerColumn: "user_id"},
			{Table: "listing_favorites", OwnerColumn: "user_id"},
			{Table: "service_listing_favorites", OwnerColumn: "user_id"},
			{Table: "training_course_favorites", OwnerColumn: "user_id"},
			{Table: "achievement_favorites", OwnerColumn: "user_id"},
			{Table: "idempotency_keys", OwnerColumn: "user_id"},
		},
	}
}
