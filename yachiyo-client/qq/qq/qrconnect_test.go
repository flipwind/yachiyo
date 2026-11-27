package qq

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func TestDecryptSecretRoundTrip(t *testing.T) {
	keyB64, err := GenerateBindKey()
	if err != nil {
		t.Fatalf("GenerateBindKey failed: %v", err)
	}
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil || len(key) != 32 {
		t.Fatalf("expected 32-byte key, got len=%d err=%v", len(key), err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	wantSecret := "my-super-secret-app-secret-123"
	sealed := gcm.Seal(nil, nonce, []byte(wantSecret), nil)
	wire := append(nonce, sealed...)
	encB64 := base64.StdEncoding.EncodeToString(wire)

	got, err := DecryptSecret(encB64, keyB64)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}
	if got != wantSecret {
		t.Fatalf("DecryptSecret = %q, want %q", got, wantSecret)
	}
}

func TestBuildConnectURLAndRenderTerminalQR(t *testing.T) {
	u := BuildConnectURL("task 123", "yachiyo")
	if !strings.Contains(u, "task_id=task+123") || !strings.Contains(u, "source=yachiyo") || !strings.Contains(u, "_wv=2") {
		t.Fatalf("unexpected connect URL: %s", u)
	}

	art, err := RenderTerminalQR(u)
	if err != nil {
		t.Fatalf("RenderTerminalQR failed: %v", err)
	}
	if len(art) == 0 || !strings.Contains(art, "\n") {
		t.Fatalf("expected multi-line QR output, got %q", art)
	}
}
