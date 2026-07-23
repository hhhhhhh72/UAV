package service_test

import (
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func setupDemandService(t *testing.T) *service.DemandService {
	t.Helper()
	return service.NewDemandService(memory.NewDemandRepository(nil), memory.NewBidRepository())
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
	d, err := svc.Create(enterpriseActor(), service.CreateDemandInput{
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
	_, err := svc.Create(enterpriseActor(), service.CreateDemandInput{
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
	d, _ := svc.Create(enterpriseActor(), service.CreateDemandInput{
		PublisherName: "测试", Contact: "13800001111", Title: "已发布需求",
	})
	svc.Submit(enterpriseActor(), d.ID)
	svc.Approve(platformAdminActor(), d.ID)
	// List with no status filter returns published demands
	items, err := svc.List(repository.DemandFilter{})
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected at least one published demand")
	}
}

// ---- Bid Tests ----

func TestDemandService_CreateBid(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "需求标题",
	})
	// Submit and approve to make it published
	svc.Submit(enterpriseActor(), d.ID)
	svc.Approve(platformAdminActor(), d.ID)

	bid, err := svc.CreateBid(individualActor(), d.ID, 50000, "我可以做")
	if err != nil {
		t.Fatalf("failed to create bid: %v", err)
	}
	if bid.Status != "pending" {
		t.Errorf("expected pending bid, got %s", bid.Status)
	}
	if bid.AmountFen != 50000 {
		t.Errorf("expected 50000 fen, got %d", bid.AmountFen)
	}
}

func TestDemandService_CreateBid_CannotBidOnOwnDemand(t *testing.T) {
	svc := setupDemandService(t)
	ent := enterpriseActor()
	d, _ := svc.Create(ent, service.CreateDemandInput{
		PublisherName: "自己", Contact: "13800001111", Title: "自己的需求",
	})
	svc.Submit(ent, d.ID)
	svc.Approve(platformAdminActor(), d.ID)

	_, err := svc.CreateBid(ent, d.ID, 50000, "自己投自己")
	if err == nil {
		t.Fatal("expected error when bidding on own demand")
	}
}

func TestDemandService_SelectBid(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "选标测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(platformAdminActor(), d.ID)

	bid, _ := svc.CreateBid(individualActor(), d.ID, 30000, "报价300")

	matched, err := svc.SelectBid(publisher, d.ID, bid.ID)
	if err != nil {
		t.Fatalf("failed to select bid: %v", err)
	}
	if matched.Status != domain.DemandMatched {
		t.Errorf("expected matched status, got %s", matched.Status)
	}
}

func TestDemandService_SelectBid_WrongOwner(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "归属测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(platformAdminActor(), d.ID)

	bid, _ := svc.CreateBid(individualActor(), d.ID, 30000, "报价")

	otherEnterprise := domain.Actor{ID: "ent-other", Role: domain.RoleEnterprise}
	_, err := svc.SelectBid(otherEnterprise, d.ID, bid.ID)
	if err == nil {
		t.Fatal("expected error when non-owner selects bid")
	}
}

func TestDemandService_SelectBid_BidMustBelongToDemand(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()

	// Create two demands (sleep to ensure unique IDs — UnixNano collision in tests)
	d1, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "需求1",
	})
	time.Sleep(time.Microsecond)
	d2, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "需求2",
	})
	svc.Submit(publisher, d1.ID)
	svc.Submit(publisher, d2.ID)
	svc.Approve(platformAdminActor(), d1.ID)
	svc.Approve(platformAdminActor(), d2.ID)

	bid, _ := svc.CreateBid(individualActor(), d1.ID, 30000, "报价给d1")

	// Try to select bid for wrong demand
	_, err := svc.SelectBid(publisher, d2.ID, bid.ID)
	if err == nil {
		t.Fatal("expected error when selecting bid for wrong demand")
	}
}

func TestDemandService_SelectBid_AlreadyMatched(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "匹配测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(platformAdminActor(), d.ID)

	bid1, _ := svc.CreateBid(individualActor(), d.ID, 20000, "第一个报价")
	svc.SelectBid(publisher, d.ID, bid1.ID)

	// Second bid should fail because demand is already matched (add sleep for unique bid ID)
	time.Sleep(time.Microsecond)
	bid2, _ := svc.CreateBid(domain.Actor{ID: "user-002", Role: domain.RoleIndividual}, d.ID, 30000, "第二个报价")
	_, err := svc.SelectBid(publisher, d.ID, bid2.ID)
	if err == nil {
		t.Fatal("expected error when selecting bid for already-matched demand")
	}
}

// ---- Demand Review & Approval ----

func TestDemandService_Review_Approve(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "待审核需求",
	})
	d, err := svc.Review(platformAdminActor(), d.ID, "approve", "")
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}
	if d.Status != domain.DemandPublished {
		t.Errorf("expected published, got %s", d.Status)
	}
}

func TestDemandService_Review_Reject(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "拒绝测试",
	})
	d, err := svc.Review(platformAdminActor(), d.ID, "reject", "不符合要求")
	if err != nil {
		t.Fatalf("failed to reject: %v", err)
	}
	if d.Status != domain.DemandRejected {
		t.Errorf("expected rejected, got %s", d.Status)
	}
}

func TestDemandService_Review_RequiresReason(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "原因测试",
	})
	_, err := svc.Review(platformAdminActor(), d.ID, "reject", "")
	if err == nil {
		t.Fatal("expected error for reject without reason")
	}
}

func TestDemandService_Review_NonAdmin(t *testing.T) {
	svc := setupDemandService(t)
	d, _ := svc.Create(enterpriseActor(), service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "越权测试",
	})
	_, err := svc.Review(individualActor(), d.ID, "approve", "")
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
	c, _ := svc.Create(domain.Actor{ID: "ent-test", Role: domain.RoleEnterprise}, domain.Contract{
		EnterpriseID: "ent-test",
	})
	c, err := svc.UpdateStatus(platformAdminActor(), c.ID, domain.ContractSent)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	if c.Status != domain.ContractSent {
		t.Errorf("expected sent, got %s", c.Status)
	}
}

func TestContractService_UpdateStatus_InvalidTransition(t *testing.T) {
	svc := setupContractService(t)
	c, _ := svc.Create(domain.Actor{ID: "ent-test", Role: domain.RoleEnterprise}, domain.Contract{
		EnterpriseID: "ent-test",
	})
	_, err := svc.UpdateStatus(platformAdminActor(), c.ID, domain.ContractSigned)
	if err == nil {
		t.Fatal("expected error for invalid transition draft -> signed")
	}
}

func TestContractService_UpdateStatus_OwnershipCheck(t *testing.T) {
	svc := setupContractService(t)
	c, _ := svc.Create(domain.Actor{ID: "ent-owner", Role: domain.RoleEnterprise}, domain.Contract{
		EnterpriseID: "ent-owner",
	})
	// Another enterprise tries to void
	otherEnterprise := domain.Actor{ID: "ent-other", Role: domain.RoleEnterprise}
	_, err := svc.UpdateStatus(otherEnterprise, c.ID, domain.ContractVoided)
	if err == nil {
		t.Fatal("expected error when non-owner enterprise voids contract")
	}
}

func TestContractService_Create(t *testing.T) {
	svc := setupContractService(t)
	c, err := svc.Create(domain.Actor{ID: "ent-001", Role: domain.RoleEnterprise}, domain.Contract{
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

// ---- Confirm Complete ----

func TestDemandService_ConfirmComplete_DualConfirm(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "双确认测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(platformAdminActor(), d.ID)

	bidder := individualActor()
	bid, _ := svc.CreateBid(bidder, d.ID, 10000, "报价")
	svc.SelectBid(publisher, d.ID, bid.ID)

	// First confirm (publisher)
	d, completed, err := svc.ConfirmComplete(publisher, d.ID)
	if err != nil {
		t.Fatalf("first confirm failed: %v", err)
	}
	if completed {
		t.Error("should not be completed after first confirm")
	}

	// Second confirm (bidder)
	d, completed, err = svc.ConfirmComplete(bidder, d.ID)
	if err != nil {
		t.Fatalf("second confirm failed: %v", err)
	}
	if !completed {
		t.Error("should be completed after second confirm")
	}
	if d.Status != domain.DemandCompleted {
		t.Errorf("expected completed, got %s", d.Status)
	}
}

// ---- Dispute ----

func TestDemandService_Dispute(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "争议测试",
	})
	// Submit and approve so demand is published (required for dispute).
	svc.Submit(publisher, d.ID)
	svc.Approve(platformAdminActor(), d.ID)
	_, err := svc.Dispute(publisher, d.ID, "不满意")
	if err != nil {
		t.Fatalf("failed to dispute: %v", err)
	}
}

func TestDemandService_Dispute_WrongOwner(t *testing.T) {
	svc := setupDemandService(t)
	publisher := enterpriseActor()
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布者", Contact: "13800001111", Title: "争议归属测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(platformAdminActor(), d.ID)
	_, err := svc.Dispute(individualActor(), d.ID, "不满意")
	if err == nil {
		t.Fatal("expected error when non-owner disputes")
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
	_, err := svc.Create(ent, domain.EmploymentRequest{
		Position: "测试岗位", Headcount: 3,
	})
	if err != nil {
		t.Fatalf("create employment: %v", err)
	}
	items, total, err := svc.List(ent, 0, 20)
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
	_, _, err := svc.List(individualActor(), 0, 20)
	if err == nil {
		t.Fatal("expected error for individual listing employment")
	}
}

// ---- Contract Full Transition Path ----

func TestContractService_FullLifecycle(t *testing.T) {
	svc := setupContractService(t)
	ent := domain.Actor{ID: "ent-life", Role: domain.RoleEnterprise}
	c, _ := svc.Create(ent, domain.Contract{EnterpriseID: "ent-life"})

	// draft → sent
	c, err := svc.UpdateStatus(ent, c.ID, domain.ContractSent)
	if err != nil {
		t.Fatalf("draft→sent: %v", err)
	}
	// sent → signing
	c, err = svc.UpdateStatus(ent, c.ID, domain.ContractSigning)
	if err != nil {
		t.Fatalf("sent→signing: %v", err)
	}
	// signing → signed
	c, err = svc.UpdateStatus(ent, c.ID, domain.ContractSigned)
	if err != nil {
		t.Fatalf("signing→signed: %v", err)
	}
	// signed → voided
	c, err = svc.UpdateStatus(ent, c.ID, domain.ContractVoided)
	if err != nil {
		t.Fatalf("signed→voided: %v", err)
	}
	if c.Status != domain.ContractVoided {
		t.Errorf("expected voided, got %s", c.Status)
	}
	// voided → anything should fail
	_, err = svc.UpdateStatus(ent, c.ID, domain.ContractSigned)
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
	items, err := svc.Pending(domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin})
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if items == nil {
		t.Error("expected non-nil result")
	}
}

func TestEnterpriseService_Pending_NonAdmin(t *testing.T) {
	svc := setupEnterpriseService(t)
	_, err := svc.Pending(individualActor())
	if err == nil {
		t.Fatal("expected error for non-admin")
	}
}
