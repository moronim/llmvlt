package store

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Vault in-memory operations ---

func TestNewVault(t *testing.T) {
	v := NewVault()
	if v == nil {
		t.Fatal("NewVault returned nil")
	}
	if len(v.Keys()) != 0 {
		t.Errorf("new vault should be empty, got %d keys", len(v.Keys()))
	}
}

func TestSetAndGet(t *testing.T) {
	v := NewVault()
	v.Set("KEY", "value")

	got, ok := v.Get("KEY")
	if !ok {
		t.Fatal("Get returned not-found for key that was set")
	}
	if got != "value" {
		t.Errorf("Get = %q, want %q", got, "value")
	}
}

func TestGetMissing(t *testing.T) {
	v := NewVault()
	_, ok := v.Get("MISSING")
	if ok {
		t.Error("Get returned ok for missing key")
	}
}

func TestSetOverwrite(t *testing.T) {
	v := NewVault()
	v.Set("KEY", "first")
	v.Set("KEY", "second")

	got, _ := v.Get("KEY")
	if got != "second" {
		t.Errorf("overwritten value = %q, want %q", got, "second")
	}
}

func TestVersionIncrement(t *testing.T) {
	v := NewVault()
	v.Set("KEY", "v1")
	v.Set("KEY", "v2")
	v.Set("KEY", "v3")

	entry := v.data.Secrets["KEY"]
	if entry.Version != 3 {
		t.Errorf("version = %d, want 3", entry.Version)
	}
}

func TestKeys(t *testing.T) {
	v := NewVault()
	v.Set("A", "1")
	v.Set("B", "2")
	v.Set("C", "3")

	keys := v.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}

	found := map[string]bool{}
	for _, k := range keys {
		found[k] = true
	}
	for _, want := range []string{"A", "B", "C"} {
		if !found[want] {
			t.Errorf("Keys() missing %q", want)
		}
	}
}

func TestAll(t *testing.T) {
	v := NewVault()
	v.Set("X", "1")
	v.Set("Y", "2")

	all := v.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d entries, want 2", len(all))
	}
	if all["X"] != "1" || all["Y"] != "2" {
		t.Errorf("All() = %v, want {X:1 Y:2}", all)
	}
}

func TestUnset(t *testing.T) {
	v := NewVault()
	v.Set("KEY", "value")
	v.Unset("KEY")

	_, ok := v.Get("KEY")
	if ok {
		t.Error("Get returned ok after Unset")
	}
	if len(v.Keys()) != 0 {
		t.Errorf("Keys() should be empty after Unset, got %d", len(v.Keys()))
	}
}

func TestUnsetMissing(t *testing.T) {
	v := NewVault()
	v.Unset("NOPE") // should not panic
}

func TestSecretAgeDays(t *testing.T) {
	v := NewVault()

	// Missing key returns 0
	if age := v.SecretAgeDays("MISSING"); age != 0 {
		t.Errorf("SecretAgeDays for missing key = %d, want 0", age)
	}

	// Just-set key should be 0 days old
	v.Set("KEY", "value")
	if age := v.SecretAgeDays("KEY"); age != 0 {
		t.Errorf("SecretAgeDays for fresh key = %d, want 0", age)
	}
}

func TestActiveProvider(t *testing.T) {
	v := NewVault()

	if p := v.GetActiveProvider(); p != "" {
		t.Errorf("default active provider = %q, want empty", p)
	}

	v.SetActiveProvider("anthropic")
	if p := v.GetActiveProvider(); p != "anthropic" {
		t.Errorf("active provider = %q, want %q", p, "anthropic")
	}

	v.SetActiveProvider("all")
	if p := v.GetActiveProvider(); p != "" {
		t.Errorf("active provider after 'all' = %q, want empty", p)
	}
}

// --- Encryption round-trip ---

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("hello, world")
	password := "test-password"

	encrypted, err := encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if string(encrypted) == string(plaintext) {
		t.Fatal("encrypted data should not equal plaintext")
	}

	decrypted, err := decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	plaintext := []byte("secret data")

	encrypted, err := encrypt(plaintext, "correct-password")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	_, err = decrypt(encrypted, "wrong-password")
	if err == nil {
		t.Fatal("decrypt with wrong password should fail")
	}
}

func TestDecryptTooShort(t *testing.T) {
	_, err := decrypt([]byte("short"), "password")
	if err == nil {
		t.Fatal("decrypt of too-short data should fail")
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	plaintext := []byte("same input")
	password := "same-password"

	a, _ := encrypt(plaintext, password)
	b, _ := encrypt(plaintext, password)

	if string(a) == string(b) {
		t.Fatal("two encryptions of the same data should differ (random salt/nonce)")
	}
}

// --- Save/Load round-trip ---

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.store")
	password := "test-pw"

	v := NewVault()
	v.Set("API_KEY", "sk-abc123")
	v.Set("ORG_ID", "org-xyz")
	v.SetActiveProvider("openai")

	if err := Save(path, password, v); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// File should exist with restricted permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("vault file not found: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("vault file permissions = %o, want 0600", info.Mode().Perm())
	}

	loaded, err := Load(path, password)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	val, ok := loaded.Get("API_KEY")
	if !ok || val != "sk-abc123" {
		t.Errorf("loaded API_KEY = (%q, %v), want (sk-abc123, true)", val, ok)
	}

	val, ok = loaded.Get("ORG_ID")
	if !ok || val != "org-xyz" {
		t.Errorf("loaded ORG_ID = (%q, %v), want (org-xyz, true)", val, ok)
	}

	if p := loaded.GetActiveProvider(); p != "openai" {
		t.Errorf("loaded active provider = %q, want %q", p, "openai")
	}
}

func TestLoadWrongPassword(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.store")

	v := NewVault()
	v.Set("KEY", "value")
	Save(path, "correct", v)

	_, err := Load(path, "wrong")
	if err == nil {
		t.Fatal("Load with wrong password should fail")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/vault.store", "password")
	if err == nil {
		t.Fatal("Load of missing file should fail")
	}
}

func TestSaveLoadEmptyVault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.store")
	password := "pw"

	v := NewVault()
	if err := Save(path, password, v); err != nil {
		t.Fatalf("Save empty vault failed: %v", err)
	}

	loaded, err := Load(path, password)
	if err != nil {
		t.Fatalf("Load empty vault failed: %v", err)
	}

	if len(loaded.Keys()) != 0 {
		t.Errorf("loaded empty vault has %d keys, want 0", len(loaded.Keys()))
	}
}

func TestSaveLoadPreservesVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ver.store")
	password := "pw"

	v := NewVault()
	v.Set("K", "v1")
	v.Set("K", "v2")
	v.Set("K", "v3")
	Save(path, password, v)

	loaded, _ := Load(path, password)
	entry := loaded.data.Secrets["K"]
	if entry.Version != 3 {
		t.Errorf("loaded version = %d, want 3", entry.Version)
	}
}
