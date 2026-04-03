package e2e

import (
	"testing"
)

func TestUnsetExistingKey(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY_A": "value-a",
		"KEY_B": "value-b",
	})

	_, stderr, err := llmvlt(t, storePath, "unset", "KEY_A")
	if err != nil {
		t.Fatalf("unset failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "removed", "stderr")

	// KEY_A should be gone
	_, _, err = llmvlt(t, storePath, "get", "KEY_A")
	if err == nil {
		t.Error("KEY_A should be gone after unset")
	}

	// KEY_B should still exist
	stdout, _, err := llmvlt(t, storePath, "get", "KEY_B")
	if err != nil {
		t.Errorf("KEY_B should still exist: %v", err)
	}
	if stdout != "value-b" {
		t.Errorf("KEY_B = %q, want value-b", stdout)
	}
}

func TestUnsetMissingKey(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": "value",
	})

	_, stderr, err := llmvlt(t, storePath, "unset", "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for unset of nonexistent key")
	}
	assertContains(t, stderr, "not found", "stderr")
}

func TestUnsetThenSetAgain(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	llmvlt(t, storePath, "set", "KEY", "original")
	llmvlt(t, storePath, "unset", "KEY")
	llmvlt(t, storePath, "set", "KEY", "new-value")

	stdout, _, err := llmvlt(t, storePath, "get", "KEY")
	if err != nil {
		t.Fatalf("get after re-set failed: %v", err)
	}
	if stdout != "new-value" {
		t.Errorf("get = %q, want new-value", stdout)
	}
}

func TestUnsetAllKeys(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"A": "1",
		"B": "2",
	})

	llmvlt(t, storePath, "unset", "A")
	llmvlt(t, storePath, "unset", "B")

	stdout, _, _ := llmvlt(t, storePath, "list")
	assertContains(t, stdout, "empty", "list output")
}

func TestUnsetWrongPassword(t *testing.T) {
	storePath := createVault(t, map[string]string{"KEY": "val"})
	_, stderr, err := llmvltWithPassword(t, storePath, "wrong", "unset", "KEY")
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
	assertContains(t, stderr, "wrong password", "stderr")
}
