package e2e

import (
	"os"
	"testing"
)

func TestInitEmptyVault(t *testing.T) {
	storePath := createEmptyDir(t)
	_, stderr, err := llmvlt(t, storePath, "init")

	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "Initialized empty vault", "stderr")

	// Vault file should exist
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		t.Error("vault file was not created")
	}
}

func TestInitWithPreset(t *testing.T) {
	storePath := createEmptyDir(t)
	_, stderr, err := llmvlt(t, storePath, "init", "--preset", "openai-stack")

	if err != nil {
		t.Fatalf("init with preset failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "openai-stack", "stderr")
	assertContains(t, stderr, "OPENAI_API_KEY", "stderr")
	assertContains(t, stderr, "(required)", "stderr")

	// Verify the scaffolded keys exist via list
	stdout, _, err := llmvlt(t, storePath, "list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	assertContains(t, stdout, "OPENAI_API_KEY", "list output")
	assertContains(t, stdout, "OPENAI_ORG_ID", "list output")
	assertContains(t, stdout, "OPENAI_PROJECT_ID", "list output")
}

func TestInitWithFullLlmStack(t *testing.T) {
	storePath := createEmptyDir(t)
	_, stderr, err := llmvlt(t, storePath, "init", "--preset", "full-llm-stack")

	if err != nil {
		t.Fatalf("init with full-llm-stack failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "full-llm-stack", "stderr")

	// Should have secrets from multiple providers
	stdout, _, _ := llmvlt(t, storePath, "list")
	for _, key := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "HF_TOKEN", "REPLICATE_API_TOKEN"} {
		assertContains(t, stdout, key, "list output")
	}
}

func TestInitDuplicateVault(t *testing.T) {
	storePath := createEmptyDir(t)

	// First init succeeds
	_, _, err := llmvlt(t, storePath, "init")
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// Second init should fail
	_, stderr, err := llmvlt(t, storePath, "init")
	if err == nil {
		t.Fatal("expected error on duplicate init")
	}
	assertContains(t, stderr, "already exists", "stderr")
}

func TestInitUnknownPreset(t *testing.T) {
	storePath := createEmptyDir(t)
	_, stderr, err := llmvlt(t, storePath, "init", "--preset", "nonexistent-stack")

	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
	assertContains(t, stderr, "unknown preset", "stderr")
}

func TestInitWithMlopsStack(t *testing.T) {
	storePath := createEmptyDir(t)
	_, stderr, err := llmvlt(t, storePath, "init", "--preset", "mlops-stack")

	if err != nil {
		t.Fatalf("init with mlops-stack failed: %v\nstderr: %s", err, stderr)
	}

	stdout, _, _ := llmvlt(t, storePath, "list")
	assertContains(t, stdout, "WANDB_API_KEY", "list output")
	assertContains(t, stdout, "LANGCHAIN_API_KEY", "list output")
}

func TestInitVaultFilePermissions(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	info, err := os.Stat(storePath)
	if err != nil {
		t.Fatalf("could not stat vault: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("vault permissions = %o, want 0600", info.Mode().Perm())
	}
}
