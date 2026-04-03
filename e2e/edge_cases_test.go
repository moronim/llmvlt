package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Unicode and special characters ---

func TestUnicodeSecretValue(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": "密码🔑パスワード",
	})

	stdout, _, err := llmvlt(t, storePath, "get", "KEY")
	if err != nil {
		t.Fatalf("get unicode value failed: %v", err)
	}
	if stdout != "密码🔑パスワード" {
		t.Errorf("unicode not preserved: %q", stdout)
	}
}

func TestUnicodeInShellInject(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": "café résumé naïve",
	})

	stdout, _, err := llmvlt(t, storePath, "inject", "--format", "shell")
	if err != nil {
		t.Fatalf("inject failed: %v", err)
	}
	assertContains(t, stdout, "café résumé naïve", "shell output")
}

func TestUnicodeInDotenvInject(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": "日本語テスト",
	})

	stdout, _, err := llmvlt(t, storePath, "inject", "--format", "dotenv")
	if err != nil {
		t.Fatalf("inject failed: %v", err)
	}
	assertContains(t, stdout, "日本語テスト", "dotenv output")
}

func TestUnicodeInJupyterInject(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": "emoji🎉test",
	})

	stdout, _, err := llmvlt(t, storePath, "inject", "--format", "jupyter")
	if err != nil {
		t.Fatalf("inject failed: %v", err)
	}
	assertContains(t, stdout, "emoji🎉test", "jupyter output")
}

// --- Shell metacharacters ---

func TestShellMetacharactersInValue(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"KEY": `$(whoami) && rm -rf / ; echo "pwned" | cat`,
	})

	stdout, _, err := llmvlt(t, storePath, "get", "KEY")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if stdout != `$(whoami) && rm -rf / ; echo "pwned" | cat` {
		t.Errorf("shell metacharacters not preserved: %q", stdout)
	}
}

func TestNewlinesInValue(t *testing.T) {
	// Newlines in values need careful handling
	storePath := createVault(t, map[string]string{
		"MULTILINE": "line1\nline2\nline3",
	})

	stdout, _, err := llmvlt(t, storePath, "get", "MULTILINE")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !strings.Contains(stdout, "line1") {
		t.Errorf("multiline value not preserved: %q", stdout)
	}
}

// --- Wrong password across commands ---

func TestWrongPasswordOnEveryCommand(t *testing.T) {
	storePath := createVault(t, map[string]string{"KEY": "val"})

	commands := [][]string{
		{"get", "KEY"},
		{"set", "KEY", "new"},
		{"list"},
		{"unset", "KEY"},
		{"check"},
		{"inject"},
		{"use", "openai"},
	}

	for _, args := range commands {
		_, stderr, err := llmvltWithPassword(t, storePath, "wrong-pw", args...)
		if err == nil {
			t.Errorf("command %v should fail with wrong password", args)
		}
		assertContains(t, stderr, "wrong password", "stderr for "+args[0])
	}
}

// --- Corrupted vault file ---

func TestCorruptedVaultFile(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "test.store")

	// Write garbage to the store file
	os.WriteFile(storePath, []byte("this is not an encrypted vault"), 0600)

	_, stderr, err := llmvlt(t, storePath, "list")
	if err == nil {
		t.Fatal("expected error for corrupted vault")
	}
	// Should not panic — should give a clear error
	assertContains(t, stderr, "wrong password", "stderr")
}

func TestTruncatedVaultFile(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "test.store")

	// Write a file that's too short to be a valid vault
	os.WriteFile(storePath, []byte{0x01, 0x02, 0x03}, 0600)

	_, _, err := llmvlt(t, storePath, "list")
	if err == nil {
		t.Fatal("expected error for truncated vault")
	}
}

func TestEmptyVaultFile(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "test.store")

	os.WriteFile(storePath, []byte{}, 0600)

	_, _, err := llmvlt(t, storePath, "list")
	if err == nil {
		t.Fatal("expected error for empty vault file")
	}
}

// --- Large values ---

func TestLargeSecretValue(t *testing.T) {
	// 100KB value
	largeValue := strings.Repeat("x", 100*1024)
	storePath := createVault(t, map[string]string{
		"LARGE_KEY": largeValue,
	})

	stdout, _, err := llmvlt(t, storePath, "get", "LARGE_KEY")
	if err != nil {
		t.Fatalf("get large value failed: %v", err)
	}
	if len(stdout) != 100*1024 {
		t.Errorf("large value length = %d, want %d", len(stdout), 100*1024)
	}
}

func TestManySecrets(t *testing.T) {
	secrets := map[string]string{}
	for i := 0; i < 100; i++ {
		key := "KEY_" + strings.Repeat("0", 3-len(itoa(i))) + itoa(i)
		secrets[key] = "value-" + itoa(i)
	}

	storePath := createVault(t, secrets)

	stdout, _, err := llmvlt(t, storePath, "list")
	if err != nil {
		t.Fatalf("list with 100 keys failed: %v", err)
	}
	assertContains(t, stdout, "KEY_000", "list output")
	assertContains(t, stdout, "KEY_099", "list output")

	// Verify all values retrievable
	for i := 0; i < 100; i++ {
		key := "KEY_" + strings.Repeat("0", 3-len(itoa(i))) + itoa(i)
		out, _, err := llmvlt(t, storePath, "get", key)
		if err != nil {
			t.Errorf("get %s failed: %v", key, err)
			continue
		}
		want := "value-" + itoa(i)
		if out != want {
			t.Errorf("get %s = %q, want %q", key, out, want)
		}
	}
}

// --- Full workflow integration ---

func TestFullWorkflow(t *testing.T) {
	storePath := createEmptyDir(t)

	// 1. Init with preset
	_, stderr, err := llmvlt(t, storePath, "init", "--preset", "openai-stack")
	if err != nil {
		t.Fatalf("init failed: %v\nstderr: %s", err, stderr)
	}

	// 2. Check — should show empty keys
	_, stderr, _ = llmvlt(t, storePath, "check")
	assertContains(t, stderr, "⬚", "check after init")

	// 3. Set valid keys
	llmvlt(t, storePath, "set", "OPENAI_API_KEY", "sk-proj-abcdefghijklmnopqrstuvwxyz012345678")
	llmvlt(t, storePath, "set", "OPENAI_ORG_ID", "org-myorg123")

	// 4. List — should show filled and empty keys
	stdout, _, _ := llmvlt(t, storePath, "list")
	assertContains(t, stdout, "✓", "list filled")
	assertContains(t, stdout, "⬚", "list empty") // PROJECT_ID still empty

	// 5. Get specific value
	stdout, _, _ = llmvlt(t, storePath, "get", "OPENAI_API_KEY")
	if stdout != "sk-proj-abcdefghijklmnopqrstuvwxyz012345678" {
		t.Errorf("get = %q", stdout)
	}

	// 6. Inject as dotenv
	stdout, _, _ = llmvlt(t, storePath, "inject", "--format", "dotenv")
	assertContains(t, stdout, "OPENAI_API_KEY=", "dotenv")
	assertContains(t, stdout, "OPENAI_ORG_ID=", "dotenv")
	assertNotContains(t, stdout, "OPENAI_PROJECT_ID", "dotenv empty key")

	// 7. Run a script
	scriptPath, _ := filepath.Abs("testdata/train.py")
	stdout, _, err = llmvlt(t, storePath, "run", "--tag", "integration-test", "--", "python3", scriptPath)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// 8. Check history
	stdout, _, _ = llmvlt(t, storePath, "history")
	assertContains(t, stdout, "integration-test", "history")

	// 9. Unset a key
	llmvlt(t, storePath, "unset", "OPENAI_ORG_ID")
	_, _, err = llmvlt(t, storePath, "get", "OPENAI_ORG_ID")
	if err == nil {
		t.Error("OPENAI_ORG_ID should be gone after unset")
	}

	// 10. Check again — should report remaining issues
	_, stderr, _ = llmvlt(t, storePath, "check")
	assertContains(t, stderr, "✓", "check after set")
}

// --- Helper ---

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
