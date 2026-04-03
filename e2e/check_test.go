package e2e

import (
	"testing"
)

func TestCheckAllGood(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	_, stderr, err := llmvlt(t, storePath, "check")
	if err != nil {
		t.Fatalf("check failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "✓", "stderr")
	assertContains(t, stderr, "All secrets look good", "stderr")
}

func TestCheckEmptyValue(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "",
	})

	_, stderr, err := llmvlt(t, storePath, "check")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	assertContains(t, stderr, "⬚", "stderr")
	assertContains(t, stderr, "empty", "stderr")
	assertContains(t, stderr, "1 issue", "stderr")
}

func TestCheckInvalidFormat(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "openai-stack")

	// Force-set an invalid value
	llmvlt(t, storePath, "set", "OPENAI_API_KEY", "wrong-format", "--force")

	_, stderr, err := llmvlt(t, storePath, "check")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	assertContains(t, stderr, "⚠", "stderr")
	assertContains(t, stderr, "issue", "stderr")
}

func TestCheckMixedIssues(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "openai-stack")

	// One valid, one forced-invalid, one empty
	llmvlt(t, storePath, "set", "OPENAI_API_KEY", "sk-proj-abcdefghijklmnopqrstuvwxyz012345678")
	llmvlt(t, storePath, "set", "OPENAI_ORG_ID", "bad-org-format", "--force")
	// OPENAI_PROJECT_ID stays empty from scaffold

	_, stderr, err := llmvlt(t, storePath, "check")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	// Should see both ✓ and issues
	assertContains(t, stderr, "✓", "stderr")
	assertContains(t, stderr, "issue", "stderr")
}

func TestCheckEmptyVault(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	stdout, _, err := llmvlt(t, storePath, "check")
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	assertContains(t, stdout, "empty", "stdout")
}

func TestCheckWrongPassword(t *testing.T) {
	storePath := createVault(t, map[string]string{"KEY": "val"})
	_, stderr, err := llmvltWithPassword(t, storePath, "wrong", "check")
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
	assertContains(t, stderr, "wrong password", "stderr")
}
