package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// --- Test: llmvlt run -- python train.py ---

func TestRunPythonTrain(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
		"OPENAI_ORG_ID":  "org-testorg123",
		"WANDB_API_KEY":  "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")
	stdout, stderr, err := llmvlt(t, storePath,
		"run", "--", "python3", scriptPath)

	if err != nil {
		t.Fatalf("run failed: %v\nstderr: %s", err, stderr)
	}

	var result struct {
		Found   map[string]string `json:"found"`
		Missing []string          `json:"missing"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("could not parse train.py output: %v\nstdout: %s", err, stdout)
	}

	if len(result.Missing) > 0 {
		t.Errorf("train.py reported missing keys: %v", result.Missing)
	}
	if result.Found["OPENAI_API_KEY"] != "sk-proj-abcdefghijklmnopqrstuvwxyz012345678" {
		t.Errorf("OPENAI_API_KEY = %q, want test value", result.Found["OPENAI_API_KEY"])
	}
	if result.Found["OPENAI_ORG_ID"] != "org-testorg123" {
		t.Errorf("OPENAI_ORG_ID = %q, want org-testorg123", result.Found["OPENAI_ORG_ID"])
	}
	if result.Found["WANDB_API_KEY"] != "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2" {
		t.Errorf("WANDB_API_KEY not injected")
	}
}

// --- Test: llmvlt run -- jupyter notebook (simulated) ---

func TestRunJupyterNotebook(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY":    "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
		"ANTHROPIC_API_KEY": "sk-ant-abcdefghijklmnopqrstuvwxyz012345678",
		"HF_TOKEN":          "hf_testtoken123abc",
	})

	scriptPath, _ := filepath.Abs("testdata/fake_jupyter.py")
	stdout, stderr, err := llmvlt(t, storePath,
		"run", "--", "python3", scriptPath)

	if err != nil {
		t.Fatalf("run failed: %v\nstderr: %s", err, stderr)
	}

	var result struct {
		Found   map[string]string `json:"found"`
		Missing []string          `json:"missing"`
	}
	json.Unmarshal([]byte(stdout), &result)

	for _, key := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "HF_TOKEN"} {
		if _, ok := result.Found[key]; !ok {
			t.Errorf("jupyter env missing %s", key)
		}
	}
}

// --- Test: llmvlt run -- pytest tests/ (simulated) ---

func TestRunPytest(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/fake_pytest.py")
	stdout, stderr, err := llmvlt(t, storePath,
		"run", "--", "python3", scriptPath)

	if err != nil {
		t.Fatalf("pytest run failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	var result struct {
		TestsRun    int `json:"tests_run"`
		TestsPassed int `json:"tests_passed"`
		TestsFailed int `json:"tests_failed"`
		Results     []struct {
			Test   string `json:"test"`
			Passed bool   `json:"passed"`
			Error  string `json:"error,omitempty"`
		} `json:"results"`
	}
	json.Unmarshal([]byte(stdout), &result)

	if result.TestsFailed > 0 {
		for _, r := range result.Results {
			if !r.Passed {
				t.Errorf("test %q failed: %s", r.Test, r.Error)
			}
		}
	}
	if result.TestsRun != 3 {
		t.Errorf("expected 3 tests run, got %d", result.TestsRun)
	}
}

// --- Test: empty secrets are not injected ---

func TestRunSkipsEmptySecrets(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY":    "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
		"ANTHROPIC_API_KEY": "",
		"HF_TOKEN":          "",
	})

	scriptPath, _ := filepath.Abs("testdata/fake_jupyter.py")
	stdout, _, err := llmvlt(t, storePath,
		"run", "--", "python3", scriptPath)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var result struct {
		Found   map[string]string `json:"found"`
		Missing []string          `json:"missing"`
	}
	json.Unmarshal([]byte(stdout), &result)

	if _, ok := result.Found["OPENAI_API_KEY"]; !ok {
		t.Error("OPENAI_API_KEY should be injected")
	}
	if _, ok := result.Found["ANTHROPIC_API_KEY"]; ok {
		t.Error("empty ANTHROPIC_API_KEY should NOT be injected")
	}
	if _, ok := result.Found["HF_TOKEN"]; ok {
		t.Error("empty HF_TOKEN should NOT be injected")
	}
}

// --- Test: child exit code is propagated ---

func TestRunPropagatesExitCode(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"WANDB_API_KEY": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")
	_, _, err := llmvlt(t, storePath,
		"run", "--", "python3", scriptPath)

	if err == nil {
		t.Fatal("expected non-zero exit when train.py fails")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1", exitErr.ExitCode())
	}
}

// --- Test: --tag flag logs to history ---

func TestRunWithTag(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")
	_, stderr, err := llmvlt(t, storePath,
		"run", "--tag", "gpt4-baseline", "--", "python3", scriptPath)
	if err != nil {
		t.Fatalf("tagged run failed: %v\nstderr: %s", err, stderr)
	}

	stdout, _, err := llmvlt(t, storePath, "history")
	if err != nil {
		t.Fatalf("history command failed: %v", err)
	}
	assertContains(t, stdout, "gpt4-baseline", "history output")
}

// --- Test: secrets are NOT in the parent process environment ---

func TestSecretsNotInParentEnv(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")
	llmvlt(t, storePath, "run", "--", "python3", scriptPath)

	if val := os.Getenv("OPENAI_API_KEY"); val == "sk-proj-abcdefghijklmnopqrstuvwxyz012345678" {
		t.Error("OPENAI_API_KEY leaked into parent process environment")
	}
}

// --- Test: no command specified ---

func TestRunNoCommand(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	_, stderr, err := llmvlt(t, storePath, "run")
	if err == nil {
		t.Fatal("expected error when no command specified")
	}
	assertContains(t, stderr, "no command specified", "stderr")
}

// --- Test: run with nonexistent command ---

func TestRunNonexistentCommand(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	_, _, err := llmvlt(t, storePath,
		"run", "--", "nonexistent_command_xyz_123")
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}

// --- Test: multiple tagged runs appear in history ---

func TestRunMultipleTagsInHistory(t *testing.T) {
	storePath := createVault(t, map[string]string{
		"OPENAI_API_KEY": "sk-proj-abcdefghijklmnopqrstuvwxyz012345678",
	})

	scriptPath, _ := filepath.Abs("testdata/train.py")

	tags := []string{"run-alpha", "run-beta", "run-gamma"}
	for _, tag := range tags {
		llmvlt(t, storePath, "run", "--tag", tag, "--", "python3", scriptPath)
	}

	stdout, _, _ := llmvlt(t, storePath, "history")
	for _, tag := range tags {
		assertContains(t, stdout, tag, "history output")
	}
}
