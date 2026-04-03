package e2e

import (
	"testing"
)

func TestPresetsList(t *testing.T) {
	storePath := createEmptyDir(t)
	// presets command doesn't need a vault, but the binary requires -s
	stdout, _, err := llmvlt(t, storePath, "presets")
	if err != nil {
		t.Fatalf("presets failed: %v", err)
	}

	expected := []string{
		"openai-stack",
		"anthropic-stack",
		"huggingface-stack",
		"replicate-stack",
		"wandb-stack",
		"langchain-stack",
		"together-stack",
		"mistral-stack",
		"google-ai-stack",
		"cohere-stack",
		"full-llm-stack",
		"mlops-stack",
	}
	for _, name := range expected {
		assertContains(t, stdout, name, "presets output")
	}
}

func TestPresetsDetail(t *testing.T) {
	storePath := createEmptyDir(t)
	stdout, _, err := llmvlt(t, storePath, "presets", "--detail")
	if err != nil {
		t.Fatalf("presets --detail failed: %v", err)
	}

	// Detail mode shows individual secret names
	assertContains(t, stdout, "OPENAI_API_KEY", "presets detail")
	assertContains(t, stdout, "ANTHROPIC_API_KEY", "presets detail")
	assertContains(t, stdout, "HF_TOKEN", "presets detail")
	assertContains(t, stdout, "WANDB_API_KEY", "presets detail")
	assertContains(t, stdout, "(optional)", "presets detail")
}

func TestPresetsShowsDescriptions(t *testing.T) {
	storePath := createEmptyDir(t)
	stdout, _, err := llmvlt(t, storePath, "presets")
	if err != nil {
		t.Fatalf("presets failed: %v", err)
	}

	// Each preset has a description after the name
	assertContains(t, stdout, "OpenAI", "presets description")
	assertContains(t, stdout, "Anthropic", "presets description")
	assertContains(t, stdout, "Hugging Face", "presets description")
}
