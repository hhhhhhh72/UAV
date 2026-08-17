package memory

import (
	"context"
	"testing"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
)

const testCipherKey = "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=" // 32 字节 base64

// TestCompRegPIIEncryptionAtRest: 审查 HIGH 修复——报名实名信息（id_card/phone）静态加密。
// 库内存储密文，CreateReg 返回值与 ListRegs 还原明文（与 pilotRepo 语义一致）。
func TestCompRegPIIEncryptionAtRest(t *testing.T) {
	cipher, err := crypto.NewCipher(testCipherKey)
	if err != nil {
		t.Fatal(err)
	}
	r := NewCompetitionRepository(cipher).(*compRepo)
	reg, err := r.CreateReg(context.Background(), domain.CompetitionReg{
		ID: "creg-enc-1", CompetitionID: "comp-enc-1", UserID: "u-1",
		Name: "张三", Phone: "13800000000", IDCard: "500101199001011234", Status: "submitted",
	})
	if err != nil {
		t.Fatal(err)
	}
	// 返回给调用方的是明文
	if reg.IDCard != "500101199001011234" || reg.Phone != "13800000000" {
		t.Fatalf("returned reg=%+v, want plaintext PII", reg)
	}
	// 库内是密文
	stored := r.regs[0]
	if stored.IDCard == "500101199001011234" || stored.Phone == "13800000000" {
		t.Fatalf("PII stored in plaintext: %+v", stored)
	}
	if dec, err := cipher.Decrypt(stored.IDCard); err != nil || dec != "500101199001011234" {
		t.Fatalf("decrypt stored id_card: %q err=%v", dec, err)
	}
	// 读取还原明文
	regs, err := r.ListRegs(context.Background(), "comp-enc-1")
	if err != nil || len(regs) != 1 {
		t.Fatalf("list regs: %d, %v", len(regs), err)
	}
	if regs[0].IDCard != "500101199001011234" || regs[0].Phone != "13800000000" {
		t.Fatalf("ListRegs reg=%+v, want plaintext PII", regs[0])
	}
}

// TestCompRegNilCipherKeepsPlaintext: 无 ENCRYPTION_KEY 的开发环境/测试传 nil，保持明文。
func TestCompRegNilCipherKeepsPlaintext(t *testing.T) {
	r := NewCompetitionRepository(nil).(*compRepo)
	reg, err := r.CreateReg(context.Background(), domain.CompetitionReg{
		ID: "creg-enc-2", CompetitionID: "comp-enc-2", UserID: "u-1", IDCard: "500101199001011234", Status: "submitted",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reg.IDCard != "500101199001011234" || r.regs[0].IDCard != "500101199001011234" {
		t.Fatalf("nil cipher should keep plaintext, got %+v / %+v", reg, r.regs[0])
	}
}
