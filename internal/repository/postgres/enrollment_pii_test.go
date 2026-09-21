package postgres_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
)

// TestPG_EnrollmentPIIEncrypted 培训报名的实名信息（手机号 / 身份证号）必须**加密入库**。
//
// 背景（2026-09-21 彻查）：生产实测 training_enrollments.id_card=500202100766642255、
// phone=19823864146 是**明文**，而同一个仓里 certified_pilots 与 competition_registrations
// 早已加密 —— 只有这条链路漏了。这里把三件事一次钉死：
//
//	① 写进去的是密文（绕开仓储直接查原始列）
//	② 调用方拿到的仍是明文（不能把密文回传给业务层）
//	③ 换错密钥读时**置空**而不是回传密文（与 compRepo.decRegPII 同一约定）
func TestPG_EnrollmentPIIEncrypted(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("k"), 32))
	other := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("z"), 32))
	c, err := crypto.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	wrong, err := crypto.NewCipher(other)
	if err != nil {
		t.Fatal(err)
	}
	repo := store.NewEnrollmentRepository(c)

	const phone, idCard = "13800000000", "500202100766642255"
	created, err := repo.Create(ctx, domain.Enrollment{
		ID: ug("enr"), CourseID: ug("course"), UserID: ug("user"),
		Name: "张三", Phone: phone, IDCard: idCard, Status: "enrolled",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// ② 返回值必须是明文
	if created.Phone != phone || created.IDCard != idCard {
		t.Fatalf("写入后返回的应是明文，实得 phone=%q id_card=%q", created.Phone, created.IDCard)
	}
	// ① 库里必须是密文
	var rawPhone, rawCard string
	if err := store.Pool().QueryRow(ctx,
		"SELECT phone, id_card FROM training_enrollments WHERE id=$1", created.ID).
		Scan(&rawPhone, &rawCard); err != nil {
		t.Fatalf("raw query: %v", err)
	}
	if rawPhone == phone || rawCard == idCard {
		t.Fatalf("明文入库：phone=%q id_card=%q", rawPhone, rawCard)
	}
	// ③ 正常读回是明文
	got, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Phone != phone || got.IDCard != idCard {
		t.Fatalf("读回应为明文，实得 phone=%q id_card=%q", got.Phone, got.IDCard)
	}
	// ④ 换错密钥 → 置空，绝不回传密文
	bad, err := store.NewEnrollmentRepository(wrong).FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find with wrong key: %v", err)
	}
	if bad.Phone != "" || bad.IDCard != "" {
		t.Fatalf("密钥不对时应置空，不得回传密文：phone=%q id_card=%q", bad.Phone, bad.IDCard)
	}
}
