package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 课程剩余名额（remain）是**派生值**：max(0, 总名额 - 已报)，不是表单字段。
//
// 回归背景（生产实测）：建课时 remain 直接采信表单里手填的值，与 max_students 无约束，
// 于是出现 max_students=11 / remain=10 的课程。而仓储的 BumpEnrolled 只在**有人报名时**
// 才按公式纠正——这门课从没人报名，那个错值就永远挂着：课程列表据此显示「仅剩 10 个」，
// 而详情页按容量算出「已报 0 / 11」，同一门课两个答案。
func TestCourseRemainIsDerived(t *testing.T) {
	ctx := context.Background()
	courseRepo := memory.NewCourseRepository()
	svc := service.NewTrainingService(
		memory.NewCertificateRepository(), courseRepo,
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil),
	)
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}

	// ① 建课：表单里塞一个与 max_students 不符的 remain，必须被公式覆盖
	c, err := svc.CreateCourse(ctx, admin, domain.TrainingCourse{
		Title: "名额一致性", MaxStudents: 11, EnrolledCount: 0, Remain: 10,
	})
	if err != nil {
		t.Fatalf("CreateCourse: %v", err)
	}
	if c.Remain != 11 {
		t.Fatalf("建课后 Remain=%d，want 11（= max(0, 11-0)）；不得采信表单传入的 10", c.Remain)
	}

	// ② 改课：总名额改了，剩余名额必须跟着变
	c.MaxStudents = 20
	c.Remain = 999 // 表单里乱填一个
	upd, err := svc.UpdateCourse(ctx, c)
	if err != nil {
		t.Fatalf("UpdateCourse: %v", err)
	}
	if upd.Remain != 20 {
		t.Fatalf("改课后 Remain=%d，want 20（= max(0, 20-0)）；不得采信表单传入的 999", upd.Remain)
	}

	// ③ 改课不得清零已报数：管理端表单不带 enrolled_count，
	//    采信它会把已报数打回 0（内存实现是整条替换，尤其明显），remain 也跟着失真。
	if err := courseRepo.BumpEnrolled(ctx, c.ID, 5); err != nil {
		t.Fatalf("BumpEnrolled: %v", err)
	}
	c.MaxStudents = 20
	c.EnrolledCount = 0 // 模拟管理端表单不带该字段
	c.Remain = 0
	upd2, err := svc.UpdateCourse(ctx, c)
	if err != nil {
		t.Fatalf("UpdateCourse(2): %v", err)
	}
	if upd2.EnrolledCount != 5 {
		t.Fatalf("改课把已报数改成了 %d，want 5（必须沿用旧值）", upd2.EnrolledCount)
	}
	if upd2.Remain != 15 {
		t.Fatalf("改课后 Remain=%d，want 15（= 20 - 5）", upd2.Remain)
	}

	// ④ 不限额（max_students=0）：剩余名额无意义，恒为 0，不得采信表单
	c3, err := svc.CreateCourse(ctx, admin, domain.TrainingCourse{
		Title: "不限额课程", MaxStudents: 0, Remain: 99,
	})
	if err != nil {
		t.Fatalf("CreateCourse(不限额): %v", err)
	}
	if c3.Remain != 0 {
		t.Fatalf("不限额课程 Remain=%d，want 0", c3.Remain)
	}

	// ⑤ 报名把名额吃满后，剩余不得为负
	full, err := svc.CreateCourse(ctx, admin, domain.TrainingCourse{Title: "两人课", MaxStudents: 2})
	if err != nil {
		t.Fatalf("CreateCourse(两人课): %v", err)
	}
	if err := courseRepo.BumpEnrolled(ctx, full.ID, 5); err != nil { // 超卖由别处拦，这里只验不为负
		t.Fatalf("BumpEnrolled(超量): %v", err)
	}
	got, err := courseRepo.FindByID(ctx, full.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Remain != 0 {
		t.Fatalf("超卖后 Remain=%d，want 0（不得为负）", got.Remain)
	}
}
