package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/moronim/llmvlt/store"
)

const testPassword = "test-password-123"

var (
	binaryOnce sync.Once
	binaryPath string
	buildErr   error
)

// getBinary builds the llmvlt binary once per test run and caches the path.
func getBinary(t *testing.T) string {
	t.Helper()
	binaryOnce.Do(func() {
		dir, err := os.MkdirTemp("", "llmvlt-test-bin")
		if err != nil {
			buildErr = err
			return
		}
		binaryPath = filepath.Join(dir, "llmvlt")
		cmd := exec.Command("go", "build", "-o", binaryPath, "..")
		cmd.Stderr = os.Stderr
		buildErr = cmd.Run()
	})
	if buildErr != nil {
		t.Fatalf("could not build binary: %v", buildErr)
	}
	return binaryPath
}

// createVault creates a vault file with the given secrets.
func createVault(t *testing.T, secrets map[string]string) string {
	t.Helper()
	storePath := filepath.Join(t.TempDir(), "test.store")
	v := store.NewVault()
	for k, val := range secrets {
		v.Set(k, val)
	}
	if err := store.Save(storePath, testPassword, v); err != nil {
		t.Fatalf("could not create test vault: %v", err)
	}
	return storePath
}

// createEmptyDir returns a path to a non-existent store file in a temp dir.
func createEmptyDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.store")
}

// llmvlt runs the llmvlt binary with the given store, password, and args.
// Returns stdout, stderr, and any error.
func llmvlt(t *testing.T, storePath string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	bin := getBinary(t)
	fullArgs := append([]string{"-s", storePath, "-p", testPassword}, args...)
	cmd := exec.Command(bin, fullArgs...)
	cmd.Env = minimalEnv()

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// llmvltWithStdin runs llmvlt with stdin piped from the given string.
func llmvltWithStdin(t *testing.T, storePath, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	bin := getBinary(t)
	fullArgs := append([]string{"-s", storePath, "-p", testPassword}, args...)
	cmd := exec.Command(bin, fullArgs...)
	cmd.Env = minimalEnv()
	cmd.Stdin = strings.NewReader(stdin)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// llmvltWithPassword runs llmvlt with a custom password (for wrong-password tests).
func llmvltWithPassword(t *testing.T, storePath, password string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	bin := getBinary(t)
	fullArgs := append([]string{"-s", storePath, "-p", password}, args...)
	cmd := exec.Command(bin, fullArgs...)
	cmd.Env = minimalEnv()

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// minimalEnv returns a clean env without inheriting secrets from the parent.
func minimalEnv() []string {
	env := []string{}
	for _, e := range os.Environ() {
		// Keep PATH, HOME, TMPDIR, and Go-related vars only
		if strings.HasPrefix(e, "PATH=") ||
			strings.HasPrefix(e, "HOME=") ||
			strings.HasPrefix(e, "TMPDIR=") ||
			strings.HasPrefix(e, "GOPATH=") ||
			strings.HasPrefix(e, "GOROOT=") ||
			strings.HasPrefix(e, "GOCACHE=") ||
			strings.HasPrefix(e, "GOMODCACHE=") {
			env = append(env, e)
		}
	}
	return env
}

// assertContains checks that stdout or stderr contains the expected substring.
func assertContains(t *testing.T, output, expected, label string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("%s should contain %q, got:\n%s", label, expected, output)
	}
}

// assertNotContains checks that output does NOT contain the given substring.
func assertNotContains(t *testing.T, output, unexpected, label string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("%s should NOT contain %q, got:\n%s", label, unexpected, output)
	}
}
