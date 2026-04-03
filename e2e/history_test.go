package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryEmpty(t *testing.T) {
	// History is stored in a global file (~/.llmvlt/history.jsonl).
	// To test "empty history", temporarily rename the existing file.
	home, _ := os.UserHomeDir()
	histPath := filepath.Join(home, ".llmvlt", "history.jsonl")
	backupPath := histPath + ".bak"

	// Backup existing history if it exists
	if _, err := os.Stat(histPath); err == nil {
		os.Rename(histPath, backupPath)
		defer os.Rename(backupPath, histPath)
	}

	storePath := createEmptyDir(t)
	llmvlt(t, storePath, "init")

	stdout, _, err := llmvlt(t, storePath, "history")
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	assertContains(t, stdout, "No history", "stdout")
}

func TestHistoryAfterRun(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")
	llmvlt(t, storePath, "run", "--", "python3", scriptPath)

	stdout, _, err := llmvlt(t, storePath, "history")
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	assertContains(t, stdout, "python3", "history output")
	assertContains(t, stdout, "OPENAI_API_KEY", "history output")
}

func TestHistoryWithTags(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")
	llmvlt(t, storePath, "run", "--tag", "experiment-1", "--", "python3", scriptPath)
	llmvlt(t, storePath, "run", "--tag", "experiment-2", "--", "python3", scriptPath)

	stdout, _, err := llmvlt(t, storePath, "history")
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	assertContains(t, stdout, "experiment-1", "history output")
	assertContains(t, stdout, "experiment-2", "history output")
}

func TestHistoryLastN(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")

	// Run 5 times with different tags
	for i := 0; i < 5; i++ {
		tag := string(rune('A' + i))
		llmvlt(t, storePath, "run", "--tag", "run-"+tag, "--", "python3", scriptPath)
	}

	// Request only last 2
	stdout, _, err := llmvlt(t, storePath, "history", "--last", "2")
	if err != nil {
		t.Fatalf("history --last failed: %v", err)
	}

	// Should have the 2 most recent
	assertContains(t, stdout, "run-E", "history output")
	assertContains(t, stdout, "run-D", "history output")
	// Should NOT have older entries
	assertNotContains(t, stdout, "run-A", "history output")
	assertNotContains(t, stdout, "run-B", "history output")
}

func TestHistoryMostRecentFirst(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")
	llmvlt(t, storePath, "run", "--tag", "first-run", "--", "python3", scriptPath)
	llmvlt(t, storePath, "run", "--tag", "second-run", "--", "python3", scriptPath)

	stdout, _, _ := llmvlt(t, storePath, "history")

	// "second-run" should appear before "first-run" in the output
	secondIdx := indexOf(stdout, "second-run")
	firstIdx := indexOf(stdout, "first-run")
	if secondIdx < 0 || firstIdx < 0 {
		t.Fatalf("both tags should appear in history output:\n%s", stdout)
	}
	if secondIdx > firstIdx {
		t.Error("most recent entry should appear first")
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
