package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func setupDemandService(t *testing.T) *service.DemandService {
	t.Helper()
	return service.NewDemandService(memory.NewDemandRepository(nil))
}

func enterpriseActor() domain.Actor {
	return domain.Actor{ID: "ent-001", Role: domain.RoleEnterprise}
}

func individualActor() domain.Actor {
	return domain.Actor{ID: "user-001", Role: domain.RoleIndividual}
}

func platformAdminActor() domain.Actor {
	return domain.Actor{ID: "admin-001", Role: domain.RolePlatformAdmin}
}

// ---- Demand Creation ----

func TestDemandService_Create(t *testing.T) {
	svc := setupDemandService(t)
	d, err := svc.Create(context.Background(), enterpriseActor(), service.CreateDemandInput{
		PublisherName: "测试企业",
		Contact:       "13800001111",
		Title:         "测试需求",
		Description:   "测试描述",
		BizType:       "cable_inspection",
	})
	if err != nil {
		t.Fatalf("failed to create demand: %v", err)
	}
	if d.Status != domain.DemandPending {
		t.Errorf("expected pending status, got %s", d.Status)
	}
	if d.PublisherID != "ent-001" {
		t.Errorf("expected publisher ent-001, got %s", d.PublisherID)
	}
}

func TestDemandService_Create_RequiresTitle(t *testing.T) {
	svc := setupDemandService(t)
	_, err := svc.Create(context.Background(), enterpriseActor(), service.CreateDemandInput{
		Contact: "13800001111",
		Title:   "",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestDemandService_List(t *testing.T) {
	svc := setupDemandService(t)
	// Create and approve a demand so it appears in published list
	d, _ := svc.Create(context.Background(), enterpriseActor(), service.CreateDemandInput{
		PublisherName: "测试", Contact: "13800001111", Title: "已发布需求",
	})
	svc.Submit(context.Background(), enterpriseActor(), d.ID)
	svc.Approve(context.Background(), platformAdminActor(), d.ID)
	// List with no status filter returns published demands
	items, err := svc.List(context.Background(), repository.DemandFilter{})
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected at least one published demand")
	}
}

// ---- Demand Complete ----

func TestDemandService_Complete(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()

	// Publisher completes a published demand → completed
	d, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "完成测试",
	})
	svc.Approve(context.Background(), platformAdminActor(), d.ID)
	done, err := svc.Complete(context.Background(), publisher, d.ID)
	if err != nil {
		t.Fatalf("failed to complete: %v", err)
	}
	if done.Status != domain.DemandCompleted {
		t.Errorf("expected completed, got %s", done.Status)
	}

	// Non-publisher cannot complete
	time.Sleep(time.Microsecond) // ensure unique demand IDs (UnixNano collision)
	d2, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "完成归属测试",
	})
	svc.Approve(context.Background(), platformAdminActor(), d2.ID)
	if _, err := svc.Complete(context.Background(), individualActor(), d2.ID); err == nil {
		t.Fatal("expected error when non-publisher completes")
	}

	// Non-published demand cannot be completed
	time.Sleep(time.Microsecond)
	d3, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "未发布完成测试",
	})
	if _, err := svc.Complete(context.Background(), publisher, d3.ID); err == nil {
		t.Fatal("expected error when completing a pending demand")
	}
}

func TestDemandService_Cancel(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()

	// Pending demand can be cancelled by publisher
	d, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "取消测试",
	})
	cancelled, err := svc.Cancel(context.Background(), publisher, d.ID)
	if err != nil {
		t.Fatalf("failed to cancel pending demand: %v", err)
	}
	if cancelled.Status != domain.DemandCancelled {
		t.Errorf("expected cancelled, got %s", cancelled.Status)
	}

	// Published demand can be cancelled by publisher
	time.Sleep(time.Microsecond) // ensure unique demand IDs (UnixNano collision)
	d2, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "已发布取消测试",
	})
	svc.Approve(context.Background(), platformAdminActor(), d2.ID)
	c2, err := svc.Cancel(context.Background(), publisher, d2.ID)
	if err != nil {
		t.Fatalf("failed to cancel published demand: %v", err)
	}
	if c2.Status != domain.DemandCancelled {
		t.Errorf("expected cancelled, got %s", c2.Status)
	}

	// Non-publisher cannot cancel
	time.Sleep(time.Microsecond)
	d3, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "取消归属测试",
	})
	if _, err := svc.Cancel(context.Background(), individualActor(), d3.ID); err == nil {
		t.Fatal("expected error when non-publisher cancels")
	}

	// Completed demand cannot be cancelled
	time.Sleep(time.Microsecond)
	d4, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "完成不可取消测试",
	})
	svc.Approve(context.Background(), platformAdminActor(), d4.ID)
	svc.Complete(context.Background(), publisher, d4.ID)
	if _, err := svc.Cancel(context.Background(), publisher, d4.ID); err == nil {
		t.Fatal("expected error when cancelling a completed demand")
	}
}

func TestDemandService_Submit_ResubmitRejected(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()
	d, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "重新提交测试",
	})
	// Admin rejects
	svc.Review(context.Background(), platformAdminActor(), d.ID, "reject", "资料不全")
	if rejected, _ := svc.FindByID(context.Background(), d.ID); rejected.Status != domain.DemandRejected {
		t.Fatalf("expected rejected, got %s", rejected.Status)
	}

	// Publisher resubmits a rejected demand → pending
	submitted, err := svc.Submit(context.Background(), publisher, d.ID)
	if err != nil {
		t.Fatalf("failed to resubmit: %v", err)
	}
	if submitted.Status != domain.DemandPending {
		t.Errorf("expected pending after resubmit, got %s", submitted.Status)
	}

	// Non-rejected demand cannot be submitted
	time.Sleep(time.Microsecond) // ensure unique demand IDs (UnixNano collision)
	d2, _ := svc.Create(context.Background(), publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "非法提交测试",
	})
	if _, err := svc.Submit(context.Background(), publisher, d2.ID); err == nil {
		t.Fatal("expected error when submitting a pending demand")
	}
}

// ---- Demand Review & Approval ----

func TestDemandService_Review_Approve(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(context.Background(), enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "待审核需求",
	})
	d, err := svc.Review(context.Background(), platformAdminActor(), d.ID, "approve", "")
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}
	if d.Status != domain.DemandPublished {
		t.Errorf("expected published, got %s", d.Status)
	}
}

func TestDemandService_Review_Reject(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(context.Background(), enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "拒绝测试",
	})
	d, err := svc.Review(context.Background(), platformAdminActor(), d.ID, "reject", "不符合要求")
	if err != nil {
		t.Fatalf("failed to reject: %v", err)
	}
	if d.Status != domain.DemandRejected {
		t.Errorf("expected rejected, got %s", d.Status)
	}
}

func TestDemandService_Review_RequiresReason(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(context.Background(), enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "原因测试",
	})
	_, err := svc.Review(context.Background(), platformAdminActor(), d.ID, "reject", "")
	if err == nil {
		t.Fatal("expected error for reject without reason")
	}
}

func TestDemandService_Review_NonAdmin(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(context.Background(), enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "越权测试",
	})
	_, err := svc.Review(context.Background(), individualActor(), d.ID, "approve", "")
	if err == nil {
		t.Fatal("expected error when non-admin reviews")
	}
}

// ---- Contract Service Tests ----

func setupContractService(t *testing.T) *service.ContractService {
	t.Helper()
	return service.NewContractService(memory.NewContractRepository())
}

func TestContractService_UpdateStatus_ValidTransition(t *testing.T) {
	svc := setupContractService(t)
	c, _ := svc.Create(context.Background(), domain.Actor{ID: "ent-test", Role: domain.RoleEnterprise}, domain.Contract{
		EnterpriseID: "ent-test",
	})
	c, err := svc.UpdateStatus(context.Background(), platformAdminActor(), c.ID, domain.ContractSent)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	if c.Status != domain.ContractSent {
		t.Errorf("expected sent, got %s", c.Status)
	}
}

func TestContractService_UpdateStatus_InvalidTransition(t *testing.T) {
	svc := setupContractService(t)
	c, _ := svc.Create(context.Background(), domain.Actor{ID: "ent-test", Role: domain.RoleEnterprise}, domain.Contract{
		EnterpriseID: "ent-test",
	})
	_, err := svc.UpdateStatus(context.Background(), platformAdminActor(), c.ID, domain.ContractSigned)
	if err == nil {
		t.Fatal("expected error for invalid transition draft -> signed")
	}
}

func TestContractService_UpdateStatus_OwnershipCheck(t *testing.T) {
	svc := setupContractService(t)
	c, _ := svc.Create(context.Background(), domain.Actor{ID: "ent-owner", Role: domain.RoleEnterprise}, domain.Contract{
		EnterpriseID: "ent-owner",
	})
	// Another enterprise tries to void
	otherEnterprise := domain.Actor{ID: "ent-other", Role: domain.RoleEnterprise}
	_, err := svc.UpdateStatus(context.Background(), otherEnterprise, c.ID, domain.ContractVoided)
	if err == nil {
		t.Fatal("expected error when non-owner enterprise voids contract")
	}
}

func TestContractService_Create(t *testing.T) {
	svc := setupContractService(t)
	c, err := svc.Create(context.Background(), domain.Actor{ID: "ent-001", Role: domain.RoleEnterprise}, domain.Contract{
		EnterpriseID: "ent-001",
	})
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}
	if c.Status != domain.ContractDraft {
		t.Errorf("expected draft, got %s", c.Status)
	}
	if c.EnterpriseID != "ent-001" {
		t.Errorf("expected ent-001, got %s", c.EnterpriseID)
	}
}

// ---- Employment Tests ----

func setupEmploymentService(t *testing.T) *service.EmploymentService {
	t.Helper()
	return service.NewEmploymentService(memory.NewEmploymentRepository())
}

func TestEmploymentService_CreateAndList(t *testing.T) {
	svc := setupEmploymentService(t)
	ent := domain.Actor{ID: "ent-emp", Role: domain.RoleEnterprise}
	_, err := svc.Create(context.Background(), ent, domain.EmploymentRequest{
		Position: "测试岗位", Headcount: 3,
	})
	if err != nil {
		t.Fatalf("create employment: %v", err)
	}
	items, total, err := svc.List(context.Background(), ent, 0, 20)
	if err != nil {
		t.Fatalf("list employment: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestEmploymentService_List_NonEnterprise(t *testing.T) {
	svc := setupEmploymentService(t)
	_, _, err := svc.List(context.Background(), individualActor(), 0, 20)
	if err == nil {
		t.Fatal("expected error for individual listing employment")
	}
}

// ---- Contract Full Transition Path ----

func TestContractService_FullLifecycle(t *testing.T) {
	svc := setupContractService(t)
	ent := domain.Actor{ID: "ent-life", Role: domain.RoleEnterprise}
	c, _ := svc.Create(context.Background(), ent, domain.Contract{EnterpriseID: "ent-life"})

	// draft → sent
	c, err := svc.UpdateStatus(context.Background(), ent, c.ID, domain.ContractSent)
	if err != nil {
		t.Fatalf("draft→sent: %v", err)
	}
	// sent → signing
	c, err = svc.UpdateStatus(context.Background(), ent, c.ID, domain.ContractSigning)
	if err != nil {
		t.Fatalf("sent→signing: %v", err)
	}
	// signing → signed
	c, err = svc.UpdateStatus(context.Background(), ent, c.ID, domain.ContractSigned)
	if err != nil {
		t.Fatalf("signing→signed: %v", err)
	}
	// signed → voided
	c, err = svc.UpdateStatus(context.Background(), ent, c.ID, domain.ContractVoided)
	if err != nil {
		t.Fatalf("signed→voided: %v", err)
	}
	if c.Status != domain.ContractVoided {
		t.Errorf("expected voided, got %s", c.Status)
	}
	// voided → anything should fail
	_, err = svc.UpdateStatus(context.Background(), ent, c.ID, domain.ContractSigned)
	if err == nil {
		t.Fatal("expected error transitioning from voided")
	}
}

// ---- Enterprise Tests ----

func setupEnterpriseService(t *testing.T) *service.EnterpriseService {
	t.Helper()
	return service.NewEnterpriseService(memory.NewEnterpriseRepository(nil))
}

func TestEnterpriseService_Pending(t *testing.T) {
	svc := setupEnterpriseService(t)
	items, err := svc.Pending(context.Background(), domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin})
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if items == nil {
		t.Error("expected non-nil result")
	}
}

func TestEnterpriseService_Pending_NonAdmin(t *testing.T) {
	svc := setupEnterpriseService(t)
	_, err := svc.Pending(context.Background(), individualActor())
	if err == nil {
		t.Fatal("expected error for non-admin")
	}
}
