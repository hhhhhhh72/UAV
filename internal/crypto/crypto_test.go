package crypto_test

import (
	"testing"

	"drone-platform/internal/crypto"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=" // 32 bytes base64
	c, err := crypto.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	plain := "13812345678"
	enc, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if enc == plain || enc == "" {
		t.Fatal("encrypted text should differ from plaintext")
	}

	dec, err := c.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != plain {
		t.Fatalf("roundtrip failed: got %q want %q", dec, plain)
	}

	// Same plaintext should produce different ciphertexts (random nonce).
	enc2, _ := c.Encrypt(plain)
	if enc == enc2 {
		t.Fatal("two encryptions should produce different ciphertexts")
	}
}

func TestMaskPhone(t *testing.T) {
	if m := crypto.MaskPhone("13812345678"); m != "138****5678" {
		t.Fatalf("mask phone: got %q", m)
	}
	if m := crypto.MaskPhone("12345"); m != "*****" {
		t.Fatalf("short phone: got %q", m)
	}
}

func TestMaskCreditCode(t *testing.T) {
	m := crypto.MaskCreditCode("91310000MA1FL0AN2C")
	if len(m) != len("91310000MA1FL0AN2C") {
		t.Fatalf("mask length mismatch: %d", len(m))
	}
}

func TestMaskIDCard(t *testing.T) {
	m := crypto.MaskIDCard("510123199001011234")
	if len(m) != 18 {
		t.Fatalf("mask length mismatch: %d", len(m))
	}
}

func TestEmptyEncrypt(t *testing.T) {
	key := "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI="
	c, _ := crypto.NewCipher(key)
	enc, err := c.Encrypt("")
	if err != nil || enc != "" {
		t.Fatal("empty string should encrypt to empty")
	}
	dec, err := c.Decrypt("")
	if err != nil || dec != "" {
		t.Fatal("empty string should decrypt to empty")
	}
}
