package e2e

import (
	"testing"
)

// --- Valid keys ---

func TestSetValidOpenAIKey(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "openai-stack")

	_, stderr, err := llmvlt(t, storePath,
		"set", "OPENAI_API_KEY", "sk-proj-abcdefghijklmnopqrstuvwxyz012345678")

	if err != nil {
		t.Fatalf("set failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "format looks valid", "stderr")
	assertContains(t, stderr, "Secret OPENAI_API_KEY saved", "stderr")
}

func TestSetValidAnthropicKey(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "anthropic-stack")

	_, stderr, err := llmvlt(t, storePath,
		"set", "ANTHROPIC_API_KEY", "sk-ant-abcdefghijklmnopqrstuvwxyz012345678")

	if err != nil {
		t.Fatalf("set failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "format looks valid", "stderr")
}

func TestSetValidHuggingFaceToken(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "huggingface-stack")

	_, stderr, err := llmvlt(t, storePath,
		"set", "HF_TOKEN", "hf_abcDEF123xyz789")

	if err != nil {
		t.Fatalf("set failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "format looks valid", "stderr")
}

// --- Invalid keys blocked ---

func TestSetInvalidOpenAIKeyBlocked(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "openai-stack")

	_, stderr, err := llmvlt(t, storePath,
		"set", "OPENAI_API_KEY", "wrong-format")

	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	assertContains(t, stderr, "Invalid format", "stderr")
	assertContains(t, stderr, "--force", "stderr")
}

func TestSetInvalidAnthropicKeyBlocked(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "anthropic-stack")

	_, stderr, err := llmvlt(t, storePath,
		"set", "ANTHROPIC_API_KEY", "not-a-valid-key")

	if err == nil {
		t.Fatal("expected error for invalid Anthropic key")
	}
	assertContains(t, stderr, "Invalid format", "stderr")
}

func TestSetInvalidWandbKeyBlocked(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "wandb-stack")

	_, stderr, err := llmvlt(t, storePath,
		"set", "WANDB_API_KEY", "not-a-hex-string")

	if err == nil {
		t.Fatal("expected error for invalid W&B key")
	}
	assertContains(t, stderr, "Invalid format", "stderr")
}

// --- --force overrides validation ---

func TestSetForceOverridesValidation(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init", "--preset", "openai-stack")

	_, stderr, err := llmvlt(t, storePath,
		"set", "OPENAI_API_KEY", "wrong-format", "--force")

	if err != nil {
		t.Fatalf("set --force failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "stored despite format mismatch", "stderr")
	assertContains(t, stderr, "Secret OPENAI_API_KEY saved", "stderr")

	// Verify the value was actually stored
	stdout, _, _ := llmvlt(t, storePath, "get", "OPENAI_API_KEY")
	if stdout != "wrong-format" {
		t.Errorf("stored value = %q, want %q", stdout, "wrong-format")
	}
}

// --- Unknown keys (no validation) ---

func TestSetUnknownKeyNoValidation(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	_, stderr, err := llmvlt(t, storePath,
		"set", "MY_CUSTOM_KEY", "any-value-works")

	if err != nil {
		t.Fatalf("set unknown key failed: %v\nstderr: %s", err, stderr)
	}
	// Should NOT say "format looks valid" for unknown keys
	assertNotContains(t, stderr, "format looks valid", "stderr")
	assertContains(t, stderr, "Secret MY_CUSTOM_KEY saved", "stderr")
}

// --- Reading from stdin ---

func TestSetFromStdin(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	_, stderr, err := llmvltWithStdin(t, storePath,
		"my-secret-value\n",
		"set", "MY_KEY")

	if err != nil {
		t.Fatalf("set from stdin failed: %v\nstderr: %s", err, stderr)
	}
	assertContains(t, stderr, "Secret MY_KEY saved", "stderr")

	stdout, _, _ := llmvlt(t, storePath, "get", "MY_KEY")
	if stdout != "my-secret-value" {
		t.Errorf("stored value = %q, want %q", stdout, "my-secret-value")
	}
}

// --- Overwrite existing key ---

func TestSetOverwriteKey(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	llmvlt(t, storePath, "set", "KEY", "first-value")
	llmvlt(t, storePath, "set", "KEY", "second-value")

	stdout, _, _ := llmvlt(t, storePath, "get", "KEY")
	if stdout != "second-value" {
		t.Errorf("overwritten value = %q, want %q", stdout, "second-value")
	}
}

// --- Empty value ---

func TestSetEmptyValueFails(t *testing.T) {
	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	_, stderr, err := llmvltWithStdin(t, storePath, "", "set", "KEY")
	if err == nil {
		t.Fatal("expected error for empty value")
	}
	assertContains(t, stderr, "empty", "stderr")
}

// --- Wrong password ---

func TestSetWrongPassword(t *testing.T) {
	storePath := createVault(t, map[string]string{"KEY": "value"})

	_, stderr, err := llmvltWithPassword(t, storePath, "wrong-password",
		"set", "KEY", "new-value")

	if err == nil {
		t.Fatal("expected error with wrong password")
	}
	assertContains(t, stderr, "wrong password", "stderr")
}
