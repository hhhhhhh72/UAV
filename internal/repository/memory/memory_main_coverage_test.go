package memory_test

import (
	"context"
	"errors"
	"testing"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
)

// ---- 断言 helpers（方法名 + got/want）----

func mmErr(t *testing.T, method string, err error, wantErr bool) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatalf("%s: got nil error, want error", method)
		}
		return
	}
	if err != nil {
		t.Fatalf("%s: got error %v, want nil", method, err)
	}
}

func mmStr(t *testing.T, method, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got=%q want=%q", method, got, want)
	}
}

func mmInt(t *testing.T, method string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got=%d want=%d", method, got, want)
	}
}

func mmInt64(t *testing.T, method string, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got=%d want=%d", method, got, want)
	}
}

// mmCipher 构造 32 字节 base64 密钥的 AES-256-GCM Cipher（与 comp_encryption_internal_test.go 相同密钥）。
const mmCipherKey = "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI="

func mmCipher(t *testing.T) *crypto.Cipher {
	t.Helper()
	c, err := crypto.NewCipher(mmCipherKey)
	if err != nil {
		t.Fatalf("crypto.NewCipher: %v", err)
	}
	return c
}

// ======================= Demand =======================

func TestDemandRepoListAllCoverage(t *testing.T) {
	r := memory.NewDemandRepository(nil) // seed: demand-001 pending/南岸区/cable_inspection/publisher=enterprise-001
	mustCreate := func(id, publisher, district string, bt domain.BizType, st domain.DemandStatus) {
		_, err := r.Create(context.Background(), domain.Demand{ID: id, PublisherID: publisher, District: district, BizType: bt, Status: st})
		mmErr(t, "demand.Create", err, false)
	}
	mustCreate("d-pub-1", "enterprise-001", "渝北区", domain.BizPlantTransport, domain.DemandPublished)
	mustCreate("d-pub-2", "enterprise-002", "南岸区", domain.BizCableInspection, domain.DemandPublished)
	mustCreate("d-pub-3", "enterprise-003", "南岸区", domain.BizSprayPesticide, domain.DemandPublished)
	mustCreate("d-cancel", "enterprise-004", "南岸区", domain.BizOther, domain.DemandCancelled)

	all, err := r.ListAll(context.Background(), repository.DemandFilter{})
	mmErr(t, "demand.ListAll(empty)", err, false)
	mmInt(t, "demand.ListAll(empty).len", len(all), 5)

	pub, err := r.ListAll(context.Background(), repository.DemandFilter{Status: "published"})
	mmErr(t, "demand.ListAll(status)", err, false)
	mmInt(t, "demand.ListAll(status=published).len", len(pub), 3)

	allStatus, err := r.ListAll(context.Background(), repository.DemandFilter{Status: "all"})
	mmErr(t, "demand.ListAll(status=all)", err, false)
	mmInt(t, "demand.ListAll(status=all).len", len(allStatus), 5)

	dist, err := r.ListAll(context.Background(), repository.DemandFilter{District: "南岸区"})
	mmErr(t, "demand.ListAll(district)", err, false)
	mmInt(t, "demand.ListAll(district).len", len(dist), 4)

	biz, err := r.ListAll(context.Background(), repository.DemandFilter{BizType: "cable_inspection"})
	mmErr(t, "demand.ListAll(biz_type)", err, false)
	mmInt(t, "demand.ListAll(biz_type).len", len(biz), 2)

	comb, err := r.ListAll(context.Background(), repository.DemandFilter{Status: "published", District: "南岸区", BizType: "cable_inspection"})
	mmErr(t, "demand.ListAll(combined)", err, false)
	mmInt(t, "demand.ListAll(combined).len", len(comb), 1)
	mmStr(t, "demand.ListAll(combined)[0].ID", comb[0].ID, "d-pub-2")
}

func TestDemandRepoListByPublisherCoverage(t *testing.T) {
	r := memory.NewDemandRepository(nil) // seed publisher=enterprise-001
	_, err := r.Create(context.Background(), domain.Demand{ID: "d-pub-9", PublisherID: "enterprise-001", Status: domain.DemandPublished})
	mmErr(t, "demand.Create(ent-001)", err, false)
	_, err = r.Create(context.Background(), domain.Demand{ID: "d-pub-10", PublisherID: "enterprise-002", Status: domain.DemandPublished})
	mmErr(t, "demand.Create(ent-002)", err, false)

	got, err := r.ListByPublisher(context.Background(), "enterprise-001")
	mmErr(t, "demand.ListByPublisher", err, false)
	mmInt(t, "demand.ListByPublisher.len", len(got), 2)

	none, err := r.ListByPublisher(context.Background(), "nobody")
	mmErr(t, "demand.ListByPublisher(empty)", err, false)
	mmInt(t, "demand.ListByPublisher(empty).len", len(none), 0)
}

func TestDemandRepoDeleteCoverage(t *testing.T) {
	r := memory.NewDemandRepository(nil)
	err := r.Delete(context.Background(), "demand-001") // hit seed
	mmErr(t, "demand.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "demand-001")
	mmErr(t, "demand.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "demand.Delete(miss)", err, true)
}

func TestDemandRepoCipherCoverage(t *testing.T) {
	r := memory.NewDemandRepository(mmCipher(t))
	// seed demand-001 的 Contact 已加密，FindByID 应解密回明文
	d, err := r.FindByID(context.Background(), "demand-001")
	mmErr(t, "demand.FindByID(seed)", err, false)
	mmStr(t, "demand.FindByID(seed).Contact", d.Contact, "138****8888")

	// Create 加密 Contact；返回值为密文，FindByID 解密回明文
	created, err := r.Create(context.Background(), domain.Demand{ID: "d-ciph", Contact: "13900000000", PublisherID: "u-ciph", Status: domain.DemandPublished})
	mmErr(t, "demand.Create(cipher)", err, false)
	if created.Contact == "13900000000" {
		t.Fatalf("demand.Create(cipher): got plaintext Contact %q, want ciphertext", created.Contact)
	}
	f, err := r.FindByID(context.Background(), "d-ciph")
	mmErr(t, "demand.FindByID(cipher)", err, false)
	mmStr(t, "demand.FindByID(cipher).Contact", f.Contact, "13900000000")
}

// ======================= Enterprise =======================

func TestEnterpriseRepoMainCoverage(t *testing.T) {
	r := memory.NewEnterpriseRepository(nil)
	_, _ = r.Create(context.Background(), domain.Enterprise{ID: "e-1", OwnerUserID: "u-1", Name: "一", Status: domain.EnterpriseSubmitted})
	_, _ = r.Create(context.Background(), domain.Enterprise{ID: "e-2", OwnerUserID: "u-2", Name: "二", Status: domain.EnterpriseApproved})
	_, _ = r.Create(context.Background(), domain.Enterprise{ID: "e-3", OwnerUserID: "u-3", Name: "三", Status: domain.EnterpriseSubmitted})

	// ListByStatus 分页：status 过滤 / 空 status 全量 / offset 越界 / limit 截断
	list, total, err := r.ListByStatus(context.Background(), "submitted", 0, 10)
	mmErr(t, "enterprise.ListByStatus(submitted)", err, false)
	mmInt(t, "enterprise.ListByStatus(submitted).total", total, 2)
	mmInt(t, "enterprise.ListByStatus(submitted).len", len(list), 2)

	all, total2, err := r.ListByStatus(context.Background(), "", 0, 10)
	mmErr(t, "enterprise.ListByStatus(all)", err, false)
	mmInt(t, "enterprise.ListByStatus(all).total", total2, 3)
	mmInt(t, "enterprise.ListByStatus(all).len", len(all), 3)

	none, total3, err := r.ListByStatus(context.Background(), "submitted", 10, 10)
	mmErr(t, "enterprise.ListByStatus(offset overflow)", err, false)
	mmInt(t, "enterprise.ListByStatus(offset overflow).total", total3, 2)
	mmInt(t, "enterprise.ListByStatus(offset overflow).len", len(none), 0)

	page, total4, err := r.ListByStatus(context.Background(), "submitted", 0, 1)
	mmErr(t, "enterprise.ListByStatus(limit truncate)", err, false)
	mmInt(t, "enterprise.ListByStatus(limit truncate).total", total4, 2)
	mmInt(t, "enterprise.ListByStatus(limit truncate).len", len(page), 1)

	// Delete（实现恒返回 nil，未命中不报错）
	err = r.Delete(context.Background(), "e-1")
	mmErr(t, "enterprise.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "e-1")
	mmErr(t, "enterprise.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "enterprise.Delete(miss)", err, false)

	// AddDocument + ListDocuments
	doc, err := r.AddDocument(context.Background(), domain.EnterpriseDocument{ID: "doc-1", EnterpriseID: "e-2", DocumentType: "license"})
	mmErr(t, "enterprise.AddDocument", err, false)
	mmStr(t, "enterprise.AddDocument.ID", doc.ID, "doc-1")
	docs, err := r.ListDocuments(context.Background(), "e-2")
	mmErr(t, "enterprise.ListDocuments", err, false)
	mmInt(t, "enterprise.ListDocuments.len", len(docs), 1)
	docsNone, err := r.ListDocuments(context.Background(), "nobody")
	mmErr(t, "enterprise.ListDocuments(empty)", err, false)
	mmInt(t, "enterprise.ListDocuments(empty).len", len(docsNone), 0)
}

func TestEnterpriseRepoCipherCoverage(t *testing.T) {
	r := memory.NewEnterpriseRepository(mmCipher(t))
	e, err := r.Create(context.Background(), domain.Enterprise{ID: "ent-ciph", OwnerUserID: "u-1", Name: "加密企业", LicenseURL: "http://lic", AccountName: "账户名", Status: domain.EnterpriseApproved})
	mmErr(t, "enterprise.Create(cipher)", err, false)
	if e.LicenseURL == "http://lic" || e.AccountName == "账户名" {
		t.Fatalf("enterprise.Create(cipher): sensitive fields stored in plaintext: %+v", e)
	}
	f, err := r.FindByID(context.Background(), "ent-ciph")
	mmErr(t, "enterprise.FindByID(cipher)", err, false)
	mmStr(t, "enterprise.FindByID.LicenseURL", f.LicenseURL, "http://lic")
	mmStr(t, "enterprise.FindByID.AccountName", f.AccountName, "账户名")

	// Search 含 cipher 分支：解密后返回
	results, err := r.Search(context.Background(), "加密")
	mmErr(t, "enterprise.Search(cipher)", err, false)
	mmInt(t, "enterprise.Search.len", len(results), 1)
	mmStr(t, "enterprise.Search[0].LicenseURL", results[0].LicenseURL, "http://lic")
}

// ======================= Job =======================

func TestJobRepoMainCoverage(t *testing.T) {
	r := memory.NewJobRepository()
	_, _ = r.Create(context.Background(), domain.Job{ID: "j-1", EnterpriseID: "ent-1", Title: "A", Status: domain.JobPublished})
	_, _ = r.Create(context.Background(), domain.Job{ID: "j-2", EnterpriseID: "ent-2", Title: "B", Status: domain.JobDraft})

	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "job.ListAll", err, false)
	mmInt(t, "job.ListAll.total", total, 2)
	mmInt(t, "job.ListAll.len", len(all), 2)

	none, total2, err := r.ListAll(context.Background(), 10, 10)
	mmErr(t, "job.ListAll(offset overflow)", err, false)
	mmInt(t, "job.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "job.ListAll(offset overflow).len", len(none), 0)

	page, total3, err := r.ListAll(context.Background(), 0, 1)
	mmErr(t, "job.ListAll(limit truncate)", err, false)
	mmInt(t, "job.ListAll(limit truncate).total", total3, 2)
	mmInt(t, "job.ListAll(limit truncate).len", len(page), 1)

	byEnt, err := r.ListByEnterprise(context.Background(), "ent-1")
	mmErr(t, "job.ListByEnterprise", err, false)
	mmInt(t, "job.ListByEnterprise.len", len(byEnt), 1)
	byEntNone, err := r.ListByEnterprise(context.Background(), "nobody")
	mmErr(t, "job.ListByEnterprise(empty)", err, false)
	mmInt(t, "job.ListByEnterprise(empty).len", len(byEntNone), 0)

	err = r.Delete(context.Background(), "j-1")
	mmErr(t, "job.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "j-1")
	mmErr(t, "job.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "job.Delete(miss)", err, true)
}

// ======================= Resume =======================

func TestResumeRepoListAllCoverage(t *testing.T) {
	r := memory.NewResumeRepository()
	_, _ = r.Create(context.Background(), domain.Resume{ID: "res-1", UserID: "u-1", Title: "A"})
	_, _ = r.Create(context.Background(), domain.Resume{ID: "res-2", UserID: "u-2", Title: "B"})

	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "resume.ListAll", err, false)
	mmInt(t, "resume.ListAll.total", total, 2)
	mmInt(t, "resume.ListAll.len", len(all), 2)

	none, total2, err := r.ListAll(context.Background(), 5, 10)
	mmErr(t, "resume.ListAll(offset overflow)", err, false)
	mmInt(t, "resume.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "resume.ListAll(offset overflow).len", len(none), 0)
}

// ======================= Labour =======================

func TestLabourRepoAssignQuoteCoverage(t *testing.T) {
	r := memory.NewLabourOrderRepository()
	_, _ = r.CreateAssignment(context.Background(), domain.Assignment{ID: "asgn-1", OrderID: "lo-1", WorkerID: "w-1"})
	_, _ = r.CreateAssignment(context.Background(), domain.Assignment{ID: "asgn-2", OrderID: "lo-1", WorkerID: "w-2"})
	_, _ = r.CreateAssignment(context.Background(), domain.Assignment{ID: "asgn-3", OrderID: "lo-2", WorkerID: "w-1"})

	byOrder, err := r.ListAssignmentsByOrder(context.Background(), "lo-1")
	mmErr(t, "labour.ListAssignmentsByOrder", err, false)
	mmInt(t, "labour.ListAssignmentsByOrder.len", len(byOrder), 2)
	byOrderNone, err := r.ListAssignmentsByOrder(context.Background(), "nobody")
	mmErr(t, "labour.ListAssignmentsByOrder(empty)", err, false)
	mmInt(t, "labour.ListAssignmentsByOrder(empty).len", len(byOrderNone), 0)

	byWorker, err := r.ListAssignmentsByWorker(context.Background(), "w-1")
	mmErr(t, "labour.ListAssignmentsByWorker", err, false)
	mmInt(t, "labour.ListAssignmentsByWorker.len", len(byWorker), 2)
	byWorkerNone, err := r.ListAssignmentsByWorker(context.Background(), "nobody")
	mmErr(t, "labour.ListAssignmentsByWorker(empty)", err, false)
	mmInt(t, "labour.ListAssignmentsByWorker(empty).len", len(byWorkerNone), 0)

	q, err := r.CreateQuote(context.Background(), domain.LabourQuote{ID: "q-1", OrderID: "lo-1", QuoterID: "u-2", AmountFen: 10000})
	mmErr(t, "labour.CreateQuote", err, false)
	mmStr(t, "labour.CreateQuote.ID", q.ID, "q-1")
	quotes, err := r.ListQuotes(context.Background(), "lo-1")
	mmErr(t, "labour.ListQuotes", err, false)
	mmInt(t, "labour.ListQuotes.len", len(quotes), 1)
	quotesNone, err := r.ListQuotes(context.Background(), "nobody")
	mmErr(t, "labour.ListQuotes(empty)", err, false)
	mmInt(t, "labour.ListQuotes(empty).len", len(quotesNone), 0)
}

// ======================= User =======================

func TestUserRepoCipherCoverage(t *testing.T) {
	c := mmCipher(t)
	r := memory.NewUserRepository(c)

	// Create 存的是密文 PhoneCipher；FindByID 解密回明文
	encPhone, err := c.Encrypt("13800000000")
	mmErr(t, "cipher.Encrypt", err, false)
	u, err := r.Create(context.Background(), domain.User{ID: "u-ciph-1", WechatOpenID: "wx-ciph-1", PhoneCipher: encPhone, Status: "active", Role: domain.RoleIndividual})
	mmErr(t, "user.Create(cipher)", err, false)
	if u.PhoneCipher == "13800000000" {
		t.Fatalf("user.Create(cipher): got plaintext PhoneCipher %q, want ciphertext", u.PhoneCipher)
	}
	f, err := r.FindByID(context.Background(), "u-ciph-1")
	mmErr(t, "user.FindByID(cipher)", err, false)
	mmStr(t, "user.FindByID(cipher).PhoneCipher", f.PhoneCipher, "13800000000")

	// UpdateProfile 手机号加密落库，FindByID 解密回明文
	err = r.UpdateProfile(context.Background(), "u-ciph-1", domain.UserProfile{Gender: "男", Region: "重庆", Phone: "13900000001"})
	mmErr(t, "user.UpdateProfile(phone)", err, false)
	f2, err := r.FindByID(context.Background(), "u-ciph-1")
	mmErr(t, "user.FindByID(after UpdateProfile)", err, false)
	mmStr(t, "user.UpdateProfile.PhoneCipher", f2.PhoneCipher, "13900000001")
	mmStr(t, "user.UpdateProfile.Gender", f2.Gender, "男")

	// 空 Phone 不修改手机号
	err = r.UpdateProfile(context.Background(), "u-ciph-1", domain.UserProfile{Bio: "hello"})
	mmErr(t, "user.UpdateProfile(empty phone)", err, false)
	f3, err := r.FindByID(context.Background(), "u-ciph-1")
	mmErr(t, "user.FindByID(empty phone)", err, false)
	mmStr(t, "user.UpdateProfile(empty phone).PhoneCipher", f3.PhoneCipher, "13900000001")
	mmStr(t, "user.UpdateProfile(empty phone).Bio", f3.Bio, "hello")
}

func TestUserRepoNilCipherProfileCoverage(t *testing.T) {
	r := memory.NewUserRepository(nil)
	_, _ = r.Create(context.Background(), domain.User{ID: "u-nil", PhoneCipher: "13800000000", Status: "active"})
	err := r.UpdateProfile(context.Background(), "u-nil", domain.UserProfile{Phone: "13700000000"})
	mmErr(t, "user.UpdateProfile(nil cipher)", err, false)
	f, err := r.FindByID(context.Background(), "u-nil")
	mmErr(t, "user.FindByID(nil cipher)", err, false)
	mmStr(t, "user.UpdateProfile(nil cipher).PhoneCipher", f.PhoneCipher, "13700000000")
}

func TestUserRepoMutationCoverage(t *testing.T) {
	r := memory.NewUserRepository(nil)
	_, _ = r.Create(context.Background(), domain.User{ID: "u-1", Name: "旧名", Status: "active"})

	err := r.UpdateAvatar(context.Background(), "u-1", "http://avatar")
	mmErr(t, "user.UpdateAvatar(hit)", err, false)
	f, _ := r.FindByID(context.Background(), "u-1")
	mmStr(t, "user.UpdateAvatar.AvatarURL", f.AvatarURL, "http://avatar")
	err = r.UpdateAvatar(context.Background(), "missing", "x")
	mmErr(t, "user.UpdateAvatar(miss)", err, true)

	err = r.UpdateName(context.Background(), "u-1", "新名")
	mmErr(t, "user.UpdateName(hit)", err, false)
	f, _ = r.FindByID(context.Background(), "u-1")
	mmStr(t, "user.UpdateName.Name", f.Name, "新名")
	err = r.UpdateName(context.Background(), "missing", "x")
	mmErr(t, "user.UpdateName(miss)", err, true)

	err = r.Delete(context.Background(), "u-1")
	mmErr(t, "user.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "u-1")
	mmErr(t, "user.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "user.Delete(miss)", err, true)
}

// ======================= Application =======================

func TestApplicationRepoFindByIDCoverage(t *testing.T) {
	r := memory.NewJobApplicationRepository()
	_, _ = r.Create(context.Background(), domain.JobApplication{ID: "app-1", JobID: "job-1", ApplicantID: "u-1", Status: domain.AppSubmitted})
	f, err := r.FindByID(context.Background(), "app-1")
	mmErr(t, "application.FindByID", err, false)
	mmStr(t, "application.FindByID.ID", f.ID, "app-1")
	_, err = r.FindByID(context.Background(), "missing")
	mmErr(t, "application.FindByID(miss)", err, true)
}

// ======================= Contract Template =======================

func TestContractTemplateRepoCoverage(t *testing.T) {
	r := memory.NewContractTemplateRepository()
	all, err := r.List(context.Background())
	mmErr(t, "contractTpl.List(default)", err, false)
	mmInt(t, "contractTpl.List(default).len", len(all), 2)
	_, err = r.Create(context.Background(), domain.ContractTemplate{ID: "tpl-custom", Name: "自定义", Status: "active"})
	mmErr(t, "contractTpl.Create", err, false)
	all2, err := r.List(context.Background())
	mmErr(t, "contractTpl.List(after create)", err, false)
	mmInt(t, "contractTpl.List(after create).len", len(all2), 3)
}

// ======================= Intent =======================

func TestIntentRepoCoverage(t *testing.T) {
	r := memory.NewIntentRepository()
	_, err := r.Create(context.Background(), domain.DemandIntent{ID: "int-1", DemandID: "d-1", IntentorID: "u-1", Status: "pending"})
	mmErr(t, "intent.Create", err, false)
	// 同 (demand, intentor) 已存在 pending 时重复登记报错
	_, err = r.Create(context.Background(), domain.DemandIntent{ID: "int-2", DemandID: "d-1", IntentorID: "u-1", Status: "pending"})
	mmErr(t, "intent.Create(duplicate pending)", err, true)

	// 先转出 pending 状态，再登记则不再触发去重
	u, err := r.UpdateStatus(context.Background(), "int-1", "contacted")
	mmErr(t, "intent.UpdateStatus", err, false)
	mmStr(t, "intent.UpdateStatus.Status", u.Status, "contacted")
	_, err = r.Create(context.Background(), domain.DemandIntent{ID: "int-3", DemandID: "d-1", IntentorID: "u-1", Status: "pending"})
	mmErr(t, "intent.Create(after status change)", err, false)

	byDemand, err := r.ListByDemand(context.Background(), "d-1")
	mmErr(t, "intent.ListByDemand", err, false)
	mmInt(t, "intent.ListByDemand.len", len(byDemand), 2)
	byIntentor, err := r.ListByIntentor(context.Background(), "u-1")
	mmErr(t, "intent.ListByIntentor", err, false)
	mmInt(t, "intent.ListByIntentor.len", len(byIntentor), 2)

	_, err = r.UpdateStatus(context.Background(), "missing", "x")
	mmErr(t, "intent.UpdateStatus(miss)", err, true)
}

// ======================= WorkOrder =======================

func TestWorkOrderRepoCoverage(t *testing.T) {
	r := memory.NewWorkOrderRepository()
	_, err := r.Create(context.Background(), domain.WorkOrder{ID: "wo-1", PublisherID: "pub-1", WorkerID: "w-1", Status: domain.WorkOrderPending})
	mmErr(t, "workOrder.Create", err, false)

	f, err := r.FindByID(context.Background(), "wo-1")
	mmErr(t, "workOrder.FindByID", err, false)
	mmStr(t, "workOrder.FindByID.ID", f.ID, "wo-1")
	_, err = r.FindByID(context.Background(), "missing")
	mmErr(t, "workOrder.FindByID(miss)", err, true)

	pubs, err := r.ListByPublisher(context.Background(), "pub-1")
	mmErr(t, "workOrder.ListByPublisher", err, false)
	mmInt(t, "workOrder.ListByPublisher.len", len(pubs), 1)
	wk, err := r.ListByWorker(context.Background(), "w-1")
	mmErr(t, "workOrder.ListByWorker", err, false)
	mmInt(t, "workOrder.ListByWorker.len", len(wk), 1)

	u, err := r.UpdateStatus(context.Background(), "wo-1", domain.WorkOrderPending, domain.WorkOrderOngoing)
	mmErr(t, "workOrder.UpdateStatus", err, false)
	mmStr(t, "workOrder.UpdateStatus.Status", string(u.Status), "ongoing")
	_, err = r.UpdateStatus(context.Background(), "missing", domain.WorkOrderPending, domain.WorkOrderOngoing)
	mmErr(t, "workOrder.UpdateStatus(miss)", err, true)

	u2, err := r.UpdatePhotos(context.Background(), "wo-1", []string{"p1.jpg"})
	mmErr(t, "workOrder.UpdatePhotos", err, false)
	mmInt(t, "workOrder.UpdatePhotos.ResultPhotos", len(u2.ResultPhotos), 1)
	_, err = r.UpdatePhotos(context.Background(), "missing", nil)
	mmErr(t, "workOrder.UpdatePhotos(miss)", err, true)

	u3, err := r.UpdateRework(context.Background(), "wo-1", "请整改")
	mmErr(t, "workOrder.UpdateRework", err, false)
	mmStr(t, "workOrder.UpdateRework.ReworkNote", u3.ReworkNote, "请整改")
	_, err = r.UpdateRework(context.Background(), "missing", "x")
	mmErr(t, "workOrder.UpdateRework(miss)", err, true)

	u4, err := r.UpdateCancel(context.Background(), "wo-1", "取消原因")
	mmErr(t, "workOrder.UpdateCancel", err, false)
	mmStr(t, "workOrder.UpdateCancel.CancelReason", u4.CancelReason, "取消原因")
	_, err = r.UpdateCancel(context.Background(), "missing", "x")
	mmErr(t, "workOrder.UpdateCancel(miss)", err, true)
}

// ======================= Certificate =======================

func TestCertRepoUpdateDeleteCoverage(t *testing.T) {
	r := memory.NewCertificateRepository()
	_, _ = r.Create(context.Background(), domain.Certificate{ID: "cert-1", UserID: "u-1", Status: "pending"})
	u, err := r.Update(context.Background(), domain.Certificate{ID: "cert-1", UserID: "u-1", Status: "approved", CertNumber: "CN-1"})
	mmErr(t, "cert.Update", err, false)
	mmStr(t, "cert.Update.Status", u.Status, "approved")
	_, err = r.Update(context.Background(), domain.Certificate{ID: "missing"})
	mmErr(t, "cert.Update(miss)", err, true)
	err = r.Delete(context.Background(), "cert-1")
	mmErr(t, "cert.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "cert-1")
	mmErr(t, "cert.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "cert.Delete(miss)", err, true)
}

// ======================= Course =======================

func TestCourseRepoCoverage(t *testing.T) {
	r := memory.NewCourseRepository()
	_, _ = r.Create(context.Background(), domain.TrainingCourse{ID: "crs-1", Title: "课程"})
	f, err := r.FindByID(context.Background(), "crs-1")
	mmErr(t, "course.FindByID", err, false)
	mmStr(t, "course.FindByID.Title", f.Title, "课程")
	_, err = r.FindByID(context.Background(), "missing")
	mmErr(t, "course.FindByID(miss)", err, true)
	u, err := r.Update(context.Background(), domain.TrainingCourse{ID: "crs-1", Title: "updated"})
	mmErr(t, "course.Update", err, false)
	mmStr(t, "course.Update.Title", u.Title, "updated")
	_, err = r.Update(context.Background(), domain.TrainingCourse{ID: "missing"})
	mmErr(t, "course.Update(miss)", err, true)
	err = r.Delete(context.Background(), "crs-1")
	mmErr(t, "course.Delete(hit)", err, false)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "course.Delete(miss)", err, true)
}

// ======================= Pilot =======================

func TestPilotRepoUpdateCoverage(t *testing.T) {
	r := memory.NewPilotRepository(nil)
	_, _ = r.Create(context.Background(), domain.CertifiedPilot{ID: "pilot-1", UserID: "u-1", RealName: "飞手"})
	u, err := r.Update(context.Background(), domain.CertifiedPilot{ID: "pilot-1", UserID: "u-1", RealName: "updated"})
	mmErr(t, "pilot.Update", err, false)
	mmStr(t, "pilot.Update.RealName", u.RealName, "updated")
	_, err = r.Update(context.Background(), domain.CertifiedPilot{ID: "missing"})
	mmErr(t, "pilot.Update(miss)", err, true)
}

// ======================= Product =======================

func TestProductRepoCoverage(t *testing.T) {
	r := memory.NewProductRepository()
	_, _ = r.Create(context.Background(), domain.DroneProduct{ID: "prod-1", SellerID: "u-1", Title: "M300", Views: 0})
	f, err := r.FindByID(context.Background(), "prod-1")
	mmErr(t, "product.FindByID", err, false)
	mmStr(t, "product.FindByID.Title", f.Title, "M300")
	_, err = r.FindByID(context.Background(), "missing")
	mmErr(t, "product.FindByID(miss)", err, true)

	u, err := r.Update(context.Background(), domain.DroneProduct{ID: "prod-1", SellerID: "u-1", Title: "M350"})
	mmErr(t, "product.Update", err, false)
	mmStr(t, "product.Update.Title", u.Title, "M350")
	_, err = r.Update(context.Background(), domain.DroneProduct{ID: "missing"})
	mmErr(t, "product.Update(miss)", err, true)

	err = r.IncrementViews(context.Background(), "prod-1")
	mmErr(t, "product.IncrementViews", err, false)
	f2, _ := r.FindByID(context.Background(), "prod-1")
	mmInt(t, "product.IncrementViews.Views", f2.Views, 1)
	err = r.IncrementViews(context.Background(), "missing")
	mmErr(t, "product.IncrementViews(miss)", err, true)

	err = r.Delete(context.Background(), "prod-1")
	mmErr(t, "product.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "prod-1")
	mmErr(t, "product.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "product.Delete(miss)", err, true)
}

// ======================= Service Listing =======================

func TestServiceListingRepoCoverage(t *testing.T) {
	r := memory.NewServiceListingRepository()
	_, err := r.Create(context.Background(), domain.ServiceListing{ID: "sl-1", ProviderID: "p-1", Title: "巡检服务", Status: "published"})
	mmErr(t, "serviceListing.Create", err, false)
	f, err := r.FindByID(context.Background(), "sl-1")
	mmErr(t, "serviceListing.FindByID", err, false)
	mmStr(t, "serviceListing.FindByID.Title", f.Title, "巡检服务")
	_, err = r.FindByID(context.Background(), "missing")
	mmErr(t, "serviceListing.FindByID(miss)", err, true)
	all, err := r.List(context.Background())
	mmErr(t, "serviceListing.List", err, false)
	mmInt(t, "serviceListing.List.len", len(all), 1)
	u, err := r.Update(context.Background(), domain.ServiceListing{ID: "sl-1", Title: "updated"})
	mmErr(t, "serviceListing.Update", err, false)
	mmStr(t, "serviceListing.Update.Title", u.Title, "updated")
	_, err = r.Update(context.Background(), domain.ServiceListing{ID: "missing"})
	mmErr(t, "serviceListing.Update(miss)", err, true)
	err = r.Delete(context.Background(), "sl-1")
	mmErr(t, "serviceListing.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "sl-1")
	mmErr(t, "serviceListing.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "serviceListing.Delete(miss)", err, true)
}

// ======================= Repair / Policy / Loan =======================

func TestRepairRepoListAllCoverage(t *testing.T) {
	r := memory.NewRepairRepository()
	_, _ = r.Create(context.Background(), domain.RepairOrder{ID: "rep-1", CustomerID: "u-1"})
	_, _ = r.Create(context.Background(), domain.RepairOrder{ID: "rep-2", CustomerID: "u-2"})
	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "repair.ListAll", err, false)
	mmInt(t, "repair.ListAll.total", total, 2)
	mmInt(t, "repair.ListAll.len", len(all), 2)
	none, total2, err := r.ListAll(context.Background(), 5, 10)
	mmErr(t, "repair.ListAll(offset overflow)", err, false)
	mmInt(t, "repair.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "repair.ListAll(offset overflow).len", len(none), 0)
}

func TestPolicyRepoListAllCoverage(t *testing.T) {
	r := memory.NewPolicyRepository()
	_, _ = r.Create(context.Background(), domain.InsurancePolicy{ID: "pol-1", UserID: "u-1"})
	_, _ = r.Create(context.Background(), domain.InsurancePolicy{ID: "pol-2", UserID: "u-2"})
	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "policy.ListAll", err, false)
	mmInt(t, "policy.ListAll.total", total, 2)
	mmInt(t, "policy.ListAll.len", len(all), 2)
	none, total2, err := r.ListAll(context.Background(), 5, 10)
	mmErr(t, "policy.ListAll(offset overflow)", err, false)
	mmInt(t, "policy.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "policy.ListAll(offset overflow).len", len(none), 0)
}

func TestLoanRepoListAllCoverage(t *testing.T) {
	r := memory.NewLoanRepository()
	_, _ = r.Create(context.Background(), domain.LoanApplication{ID: "loan-1", UserID: "u-1"})
	_, _ = r.Create(context.Background(), domain.LoanApplication{ID: "loan-2", UserID: "u-2"})
	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "loan.ListAll", err, false)
	mmInt(t, "loan.ListAll.total", total, 2)
	mmInt(t, "loan.ListAll.len", len(all), 2)
	none, total2, err := r.ListAll(context.Background(), 5, 10)
	mmErr(t, "loan.ListAll(offset overflow)", err, false)
	mmInt(t, "loan.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "loan.ListAll(offset overflow).len", len(none), 0)
}

// ======================= Message =======================

func TestMessageRepoCoverage(t *testing.T) {
	r := memory.NewMessageRepository()
	_, _ = r.Create(context.Background(), domain.Message{ID: "msg-1", ReceiverID: "u-1", SenderID: "sys"})
	_, _ = r.Create(context.Background(), domain.Message{ID: "msg-2", ReceiverID: "u-1", SenderID: "sys"})
	f, err := r.FindByID(context.Background(), "msg-1")
	mmErr(t, "message.FindByID", err, false)
	mmStr(t, "message.FindByID.ID", f.ID, "msg-1")
	_, err = r.FindByID(context.Background(), "missing")
	mmErr(t, "message.FindByID(miss)", err, true)

	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "message.ListAll", err, false)
	mmInt(t, "message.ListAll.total", total, 2)
	mmInt(t, "message.ListAll.len", len(all), 2)
	none, total2, err := r.ListAll(context.Background(), 5, 10)
	mmErr(t, "message.ListAll(offset overflow)", err, false)
	mmInt(t, "message.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "message.ListAll(offset overflow).len", len(none), 0)

	err = r.Delete(context.Background(), "msg-1")
	mmErr(t, "message.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "msg-1")
	mmErr(t, "message.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "message.Delete(miss)", err, true)
}

// ======================= Enrollment =======================

func TestEnrollmentRepoCoverage(t *testing.T) {
	r := memory.NewEnrollmentRepository()
	_, _ = r.Create(context.Background(), domain.Enrollment{ID: "enr-1", CourseID: "crs-1", UserID: "u-1"})
	_, _ = r.Create(context.Background(), domain.Enrollment{ID: "enr-2", CourseID: "crs-1", UserID: "u-2"})
	u, err := r.Update(context.Background(), domain.Enrollment{ID: "enr-1", CourseID: "crs-1", UserID: "u-1", Status: "approved"})
	mmErr(t, "enrollment.Update", err, false)
	mmStr(t, "enrollment.Update.Status", u.Status, "approved")
	_, err = r.Update(context.Background(), domain.Enrollment{ID: "missing"})
	mmErr(t, "enrollment.Update(miss)", err, true)
	f, err := r.FindByID(context.Background(), "enr-1")
	mmErr(t, "enrollment.FindByID", err, false)
	mmStr(t, "enrollment.FindByID.ID", f.ID, "enr-1")
	_, err = r.FindByID(context.Background(), "missing")
	mmErr(t, "enrollment.FindByID(miss)", err, true)
	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "enrollment.ListAll", err, false)
	mmInt(t, "enrollment.ListAll.total", total, 2)
	mmInt(t, "enrollment.ListAll.len", len(all), 2)
	none, total2, err := r.ListAll(context.Background(), 5, 10)
	mmErr(t, "enrollment.ListAll(offset overflow)", err, false)
	mmInt(t, "enrollment.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "enrollment.ListAll(offset overflow).len", len(none), 0)
}

// ======================= TradeOrder =======================

func TestTradeOrderRepoCoverage(t *testing.T) {
	r := memory.NewTradeOrderRepository()
	_, _ = r.Create(context.Background(), domain.TradeOrder{ID: "to-1", BuyerID: "u-1", SellerID: "u-2", Status: "paid"})
	_, _ = r.Create(context.Background(), domain.TradeOrder{ID: "to-2", BuyerID: "u-3", SellerID: "u-4", Status: "paid"})
	u, err := r.UpdateAftersale(context.Background(), domain.TradeOrder{
		ID: "to-1", Status: "aftersale", AftersaleType: "refund", AftersaleReason: "不想要了",
		AftersaleDesc: "desc", AftersaleAmountFen: 100, AftersaleStatus: "pending",
	})
	mmErr(t, "tradeOrder.UpdateAftersale", err, false)
	mmStr(t, "tradeOrder.UpdateAftersale.Status", u.Status, "aftersale")
	mmStr(t, "tradeOrder.UpdateAftersale.AftersaleType", u.AftersaleType, "refund")
	_, err = r.UpdateAftersale(context.Background(), domain.TradeOrder{ID: "missing"})
	mmErr(t, "tradeOrder.UpdateAftersale(miss)", err, true)

	all, total, err := r.ListAll(context.Background(), 0, 10)
	mmErr(t, "tradeOrder.ListAll", err, false)
	mmInt(t, "tradeOrder.ListAll.total", total, 2)
	mmInt(t, "tradeOrder.ListAll.len", len(all), 2)
	none, total2, err := r.ListAll(context.Background(), 5, 10)
	mmErr(t, "tradeOrder.ListAll(offset overflow)", err, false)
	mmInt(t, "tradeOrder.ListAll(offset overflow).total", total2, 2)
	mmInt(t, "tradeOrder.ListAll(offset overflow).len", len(none), 0)

	err = r.Delete(context.Background(), "to-1")
	mmErr(t, "tradeOrder.Delete(hit)", err, false)
	_, err = r.FindByID(context.Background(), "to-1")
	mmErr(t, "tradeOrder.FindByID(after delete)", err, true)
	err = r.Delete(context.Background(), "missing")
	mmErr(t, "tradeOrder.Delete(miss)", err, true)
}

// ======================= Escrow =======================

func TestEscrowRepoFreezeReleaseRefundCoverage(t *testing.T) {
	r := memory.NewEscrowRepository()
	tx := func(id, typ string, amt int64) domain.EscrowTransaction {
		return domain.EscrowTransaction{ID: id, TxType: typ, AmountFen: amt}
	}
	_, _ = r.Deposit(context.Background(), "u-1", 100000, tx("tx-1", "deposit", 100000))

	// Freeze 成功
	_, err := r.Freeze(context.Background(), "u-1", 30000, tx("tx-2", "freeze", 30000))
	mmErr(t, "escrow.Freeze", err, false)
	acct, _ := r.GetAccount(context.Background(), "u-1")
	mmInt64(t, "escrow.Freeze.BalanceFen", acct.BalanceFen, 70000)
	mmInt64(t, "escrow.Freeze.FrozenFen", acct.FrozenFen, 30000)
	// Freeze 余额不足
	_, err = r.Freeze(context.Background(), "u-1", 80000, tx("tx-3", "freeze", 80000))
	if !errors.Is(err, repository.ErrInsufficientBalance) {
		t.Fatalf("escrow.Freeze(insufficient): got=%v want=ErrInsufficientBalance", err)
	}

	// Release 成功
	_, err = r.Release(context.Background(), "u-1", "u-2", 30000, tx("tx-4", "release", 30000))
	mmErr(t, "escrow.Release", err, false)
	from, _ := r.GetAccount(context.Background(), "u-1")
	to, _ := r.GetAccount(context.Background(), "u-2")
	mmInt64(t, "escrow.Release.from.FrozenFen", from.FrozenFen, 0)
	mmInt64(t, "escrow.Release.to.BalanceFen", to.BalanceFen, 30000)
	// Release 冻结不足
	_, err = r.Release(context.Background(), "u-1", "u-2", 1000, tx("tx-5", "release", 1000))
	if !errors.Is(err, repository.ErrInsufficientFrozenBalance) {
		t.Fatalf("escrow.Release(insufficient): got=%v want=ErrInsufficientFrozenBalance", err)
	}

	// Refund 成功
	_, _ = r.Freeze(context.Background(), "u-1", 5000, tx("tx-6", "freeze", 5000))
	_, err = r.Refund(context.Background(), "u-1", 5000, tx("tx-7", "refund", 5000))
	mmErr(t, "escrow.Refund", err, false)
	acct2, _ := r.GetAccount(context.Background(), "u-1")
	mmInt64(t, "escrow.Refund.FrozenFen", acct2.FrozenFen, 0)
	// Refund 冻结不足
	_, err = r.Refund(context.Background(), "u-1", 1000, tx("tx-8", "refund", 1000))
	if !errors.Is(err, repository.ErrInsufficientFrozenBalance) {
		t.Fatalf("escrow.Refund(insufficient): got=%v want=ErrInsufficientFrozenBalance", err)
	}
}
