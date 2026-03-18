package services

import (
	"encoding/base64"
	"strings"
	"testing"
)

func makeKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := makeKey(t)
	plain := "super-secret-value"

	encrypted, err := encryptAESGCM(plain, key)
	if err != nil {
		t.Fatalf("encryptAESGCM: %v", err)
	}
	if encrypted == "" {
		t.Fatal("encrypted is empty")
	}
	if encrypted == plain {
		t.Fatal("encrypted equals plaintext — no encryption occurred")
	}

	decrypted, err := decryptAESGCM(encrypted, key)
	if err != nil {
		t.Fatalf("decryptAESGCM: %v", err)
	}
	if decrypted != plain {
		t.Errorf("decrypted = %q, want %q", decrypted, plain)
	}
}

func TestEncryptAESGCM_WrongKeyLen(t *testing.T) {
	_, err := encryptAESGCM("hello", []byte("short"))
	if err == nil {
		t.Fatal("encryptAESGCM with short key: expected error, got nil")
	}
}

func TestDecryptAESGCM_WrongKeyLen(t *testing.T) {
	_, err := decryptAESGCM("anything", []byte("short"))
	if err == nil {
		t.Fatal("decryptAESGCM with short key: expected error, got nil")
	}
}

func TestDecryptAESGCM_InvalidBase64(t *testing.T) {
	key := makeKey(t)
	_, err := decryptAESGCM("not-valid-base64!!!", key)
	if err == nil {
		t.Fatal("decryptAESGCM invalid base64: expected error, got nil")
	}
}

func TestDecryptAESGCM_TooShort(t *testing.T) {
	key := makeKey(t)
	// Create a valid base64 string but with too few bytes.
	short := base64.StdEncoding.EncodeToString([]byte{1, 2})
	_, err := decryptAESGCM(short, key)
	if err == nil {
		t.Fatal("decryptAESGCM too short: expected error, got nil")
	}
}

func TestDecryptAESGCM_Tampered(t *testing.T) {
	key := makeKey(t)
	enc, _ := encryptAESGCM("original", key)
	// Flip a byte in the ciphertext.
	decoded, _ := base64.StdEncoding.DecodeString(enc)
	decoded[len(decoded)-1] ^= 0xFF
	tampered := base64.StdEncoding.EncodeToString(decoded)

	_, err := decryptAESGCM(tampered, key)
	if err == nil {
		t.Fatal("decryptAESGCM tampered: expected error, got nil")
	}
}

func TestEncryptDecrypt_EmptyString(t *testing.T) {
	key := makeKey(t)
	enc, err := encryptAESGCM("", key)
	if err != nil {
		t.Fatalf("encryptAESGCM empty: %v", err)
	}
	dec, err := decryptAESGCM(enc, key)
	if err != nil {
		t.Fatalf("decryptAESGCM empty: %v", err)
	}
	if dec != "" {
		t.Errorf("decrypted = %q, want empty", dec)
	}
}

func TestLoadEncryptionKey_NotSet(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")
	_, err := loadEncryptionKey()
	if err == nil {
		t.Fatal("loadEncryptionKey empty env: expected error, got nil")
	}
}

func TestLoadEncryptionKey_32ByteRaw(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "abcdefghijklmnopqrstuvwxyz123456") // exactly 32 bytes
	key, err := loadEncryptionKey()
	if err != nil {
		t.Fatalf("loadEncryptionKey 32-byte raw: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("key len = %d, want 32", len(key))
	}
}

func TestLoadEncryptionKey_Base64Encoded(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	t.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(raw))
	key, err := loadEncryptionKey()
	if err != nil {
		t.Fatalf("loadEncryptionKey base64: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("key len = %d, want 32", len(key))
	}
}

func TestLoadEncryptionKey_Invalid(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", strings.Repeat("x", 16)) // 16 bytes, not 32
	_, err := loadEncryptionKey()
	if err == nil {
		t.Fatal("loadEncryptionKey invalid: expected error, got nil")
	}
}
