package e2e

import (
	"testing"
)

func TestUseSwitchProvider(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "openai-stack")

	_, stderr, err := llmvlt(t, storePath, "use", "openai")
	if err != nil {
		t.Fatalf("use failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "Switched to openai", "stderr")
}

func TestUseSwitchToAll(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "openai-stack")

	llmvlt(t, storePath, "use", "openai")
	_, stderr, err := llmvlt(t, storePath, "use", "all")
	if err != nil {
		t.Fatalf("use all failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "All providers active", "stderr")
}

func TestUseUnknownProvider(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	_, stderr, err := llmvlt(t, storePath, "use", "nonexistent-provider")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	assertContains(t, stderr, "unknown provider", "stderr")
}

func TestUseSwitchBetweenProviders(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "full-llm-stack")

	// Switch to anthropic
	_, stderr, err := llmvlt(t, storePath, "use", "anthropic")
	if err != nil {
		t.Fatalf("use anthropic failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "Switched to anthropic", "stderr")

	// Switch to openai
	_, stderr, err = llmvlt(t, storePath, "use", "openai")
	if err != nil {
		t.Fatalf("use openai failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "Switched to openai", "stderr")

	// Back to all
	_, stderr, err = llmvlt(t, storePath, "use", "all")
	if err != nil {
		t.Fatalf("use all failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "All providers active", "stderr")
}

func TestUseNoArgument(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	_, _, err := llmvlt(t, storePath, "use")
	if err == nil {
		t.Fatal("expected error when no provider specified")
	}
}

func TestUseWrongPassword(t *testing.T) {
	storePath := createVault(t, map[string]string{"KEY": "val"})
	_, stderr, err := llmvltWithPassword(t, storePath, "wrong", "use", "openai")
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
	assertContains(t, stderr, "wrong password", "stderr")
}
