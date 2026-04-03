package e2e

import (
	"testing"
)

func TestGetExistingKey(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"API_KEY": "sk-test-value-123",
	})

	stdout, _, err := llmvlt(t, storePath, "get", "API_KEY")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if stdout != "sk-test-value-123" {
		t.Errorf("get output = %q, want %q", stdout, "sk-test-value-123")
	}
}

func TestGetMissingKey(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"API_KEY": "value",
	})

	_, stderr, err := llmvlt(t, storePath, "get", "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	assertContains(t, stderr, "not found", "stderr")
}

func TestGetEmptyValueKey(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"SCAFFOLDED_KEY": "",
	})

	_, stderr, err := llmvlt(t, storePath, "get", "SCAFFOLDED_KEY")
	if err == nil {
		t.Fatal("expected error for empty-value key")
	}
	assertContains(t, stderr, "no value", "stderr")
}

func TestGetWrongPassword(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": "value",
	})

	_, stderr, err := llmvltWithPassword(t, storePath, "wrong-password", "get", "KEY")
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
	assertContains(t, stderr, "wrong password", "stderr")
}

func TestGetNonexistentVault(t *testing.T) {
	_, stderr, err := llmvlt(t, "/tmp/nonexistent_vault_xyz.store", "get", "KEY")
	if err == nil {
		t.Fatal("expected error for nonexistent vault")
	}
	assertContains(t, stderr, "could not", "stderr")
}

func TestGetMultipleKeys(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY_A": "value-a",
		"KEY_B": "value-b",
		"KEY_C": "value-c",
	})

	for _, tc := range []struct {
		key  string
		want string
	}{
		{"KEY_A", "value-a"},
		{"KEY_B", "value-b"},
		{"KEY_C", "value-c"},
	} {
		stdout, _, err := llmvlt(t, storePath, "get", tc.key)
		if err != nil {
			t.Errorf("get %s failed: %v", tc.key, err)
			continue
		}
		if stdout != tc.want {
			t.Errorf("get %s = %q, want %q", tc.key, stdout, tc.want)
		}
	}
}

func TestGetAfterOverwrite(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")
	llmvlt(t, storePath, "set", "KEY", "original")
	llmvlt(t, storePath, "set", "KEY", "updated")

	stdout, _, err := llmvlt(t, storePath, "get", "KEY")
	if err != nil {
		t.Fatalf("get after overwrite failed: %v", err)
	}
	if stdout != "updated" {
		t.Errorf("get = %q, want %q", stdout, "updated")
	}
}

func TestGetSpecialCharactersInValue(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": `value with "quotes" and $pecial chars & more!`,
	})

	stdout, _, err := llmvlt(t, storePath, "get", "KEY")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if stdout != `value with "quotes" and $pecial chars & more!` {
		t.Errorf("special chars not preserved: %q", stdout)
	}
}
