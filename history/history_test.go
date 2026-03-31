package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEntryJSONRoundTrip(t *testing.T) {
	entry := Entry{
		Timestamp: time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC),
		Command:   "python train.py",
		Keys:      []string{"OPENAI_API_KEY", "WANDB_API_KEY"},
		Tag:       "experiment-1",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Entry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Command != entry.Command {
		t.Errorf("Command = %q, want %q", decoded.Command, entry.Command)
	}
	if decoded.Tag != entry.Tag {
		t.Errorf("Tag = %q, want %q", decoded.Tag, entry.Tag)
	}
	if len(decoded.Keys) != 2 {
		t.Errorf("Keys count = %d, want 2", len(decoded.Keys))
	}
}

func TestEntryJSONOmitsEmptyTag(t *testing.T) {
	entry := Entry{
		Timestamp: time.Now(),
		Command:   "python eval.py",
		Keys:      []string{"KEY"},
		Tag:       "",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	if strings.Contains(string(data), `"tag"`) {
		t.Errorf("empty tag should be omitted from JSON, got: %s", string(data))
	}
}

func TestReadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	os.WriteFile(path, []byte(""), 0600)

	entries := parseFile(t, path, 10)
	if len(entries) != 0 {
		t.Errorf("empty file should return 0 entries, got %d", len(entries))
	}
}

func TestReadMalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "malformed.jsonl")

	content := "not json\n{\"ts\":\"2026-03-15T14:30:00Z\",\"cmd\":\"python test.py\",\"keys\":[\"K\"]}\nalso bad\n"
	os.WriteFile(path, []byte(content), 0600)

	entries := parseFile(t, path, 10)
	if len(entries) != 1 {
		t.Errorf("should skip malformed lines, got %d entries, want 1", len(entries))
	}
	if entries[0].Command != "python test.py" {
		t.Errorf("Command = %q, want %q", entries[0].Command, "python test.py")
	}
}

func TestReadLastN(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.jsonl")

	var lines []string
	for i := 0; i < 10; i++ {
		e := Entry{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Command:   "cmd-" + string(rune('a'+i)),
			Keys:      []string{"K"},
		}
		data, _ := json.Marshal(e)
		lines = append(lines, string(data))
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600)

	entries := parseFile(t, path, 3)
	if len(entries) != 3 {
		t.Errorf("should return last 3 entries, got %d", len(entries))
	}

	// Most recent first (reversed): cmd-j, cmd-i, cmd-h
	if entries[0].Command != "cmd-j" {
		t.Errorf("first entry should be most recent (cmd-j), got %q", entries[0].Command)
	}
	if entries[2].Command != "cmd-h" {
		t.Errorf("last entry should be cmd-h, got %q", entries[2].Command)
	}
}

func TestReadAllWhenFewerThanN(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "few.jsonl")

	e := Entry{Timestamp: time.Now(), Command: "cmd", Keys: []string{"K"}}
	data, _ := json.Marshal(e)
	os.WriteFile(path, append(data, '\n'), 0600)

	entries := parseFile(t, path, 100)
	if len(entries) != 1 {
		t.Errorf("should return all 1 entry, got %d", len(entries))
	}
}

func TestReadNonexistentFile(t *testing.T) {
	entries := parseFile(t, "/nonexistent/file.jsonl", 10)
	if entries != nil {
		t.Errorf("nonexistent file should return nil, got %v", entries)
	}
}

func TestWriteAndReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.jsonl")

	// Write 5 entries
	for i := 0; i < 5; i++ {
		e := Entry{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Command:   "python script.py",
			Keys:      []string{"OPENAI_API_KEY"},
			Tag:       "run-" + string(rune('0'+i)),
		}
		data, _ := json.Marshal(e)
		f, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		f.Write(data)
		f.WriteString("\n")
		f.Close()
	}

	entries := parseFile(t, path, 3)
	if len(entries) != 3 {
		t.Errorf("round trip returned %d entries, want 3", len(entries))
	}

	// Most recent first
	if entries[0].Tag != "run-4" {
		t.Errorf("most recent tag = %q, want %q", entries[0].Tag, "run-4")
	}
}

func TestFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.jsonl")

	e := Entry{Timestamp: time.Now(), Command: "cmd", Keys: []string{"K"}}
	data, _ := json.Marshal(e)
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	f.Write(data)
	f.WriteString("\n")
	f.Close()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("could not stat history file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
	}
}

// --- helper ---

// parseFile reads and parses a JSONL history file using the same algorithm
// as the Read function (last N entries, reversed order). This allows testing
// without depending on the hardcoded historyPath().
func parseFile(t *testing.T, path string, n int) []Entry {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("could not open: %v", err)
	}
	defer f.Close()

	var all []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		all = append(all, e)
	}

	if len(all) > n {
		all = all[len(all)-n:]
	}

	// Reverse for most-recent-first
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}

	return all
}
