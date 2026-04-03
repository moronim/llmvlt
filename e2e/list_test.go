package e2e

import (
	"testing"
)

func TestListEmptyVault(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	stdout, _, err := llmvlt(t, storePath, "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	assertContains(t, stdout, "empty", "stdout")
}

func TestListWithSecrets(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"ALPHA_KEY": "value-a",
		"BETA_KEY":  "value-b",
	})

	stdout, _, err := llmvlt(t, storePath, "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	assertContains(t, stdout, "ALPHA_KEY", "list output")
	assertContains(t, stdout, "BETA_KEY", "list output")
	// Values should never appear in list output
	assertNotContains(t, stdout, "value-a", "list output")
	assertNotContains(t, stdout, "value-b", "list output")
}

func TestListShowsEmptyMarker(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"FILLED_KEY": "has-value",
		"EMPTY_KEY":  "",
	})

	stdout, _, err := llmvlt(t, storePath, "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	// Filled keys get ✓, empty keys get ⬚
	assertContains(t, stdout, "✓", "list output")
	assertContains(t, stdout, "⬚", "list output")
}

func TestListShowsProviderLabels(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY":    "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
		"ANTHROPIC_API_KEY": "sk-ant-abcdefghijklmnopqrstuvwxyz012345678",
		"MY_CUSTOM_KEY":     "custom-value",
	})

	stdout, _, err := llmvlt(t, storePath, "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	// Known keys show provider name in parentheses
	assertContains(t, stdout, "openai-stack", "list output")
	assertContains(t, stdout, "anthropic-stack", "list output")
	// Custom keys should NOT show a provider
	// MY_CUSTOM_KEY should appear without parenthesized provider
}

func TestListAlias(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": "value",
	})

	// "ls" is an alias for "list"
	stdout, _, err := llmvlt(t, storePath, "ls")
	if err != nil {
		t.Fatalf("ls alias failed: %v", err)
	}
	assertContains(t, stdout, "KEY", "ls output")
}

func TestListAfterPresetInit(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "wandb-stack")

	stdout, _, err := llmvlt(t, storePath, "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	// All scaffolded keys should be listed as empty
	assertContains(t, stdout, "WANDB_API_KEY", "list output")
	assertContains(t, stdout, "⬚", "list output")
}

func TestListWrongPassword(t *testing.T) {
	storePath := createVault(t, map[string]string{"KEY": "val"})
	_, stderr, err := llmvltWithPassword(t, storePath, "wrong", "list")
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
	assertContains(t, stderr, "wrong password", "stderr")
}
