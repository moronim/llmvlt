package preset

import "testing"

func TestBuiltinPresetsRegistered(t *testing.T) {
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
		if _, err := Get(name); err != nil {
			t.Errorf("expected preset %q to be registered, got error: %v", name, err)
		}
	}
}

func TestGetUnknownPreset(t *testing.T) {
	_, err := Get("nonexistent-preset")
	if err == nil {
		t.Error("Get of unknown preset should return error")
	}
}

func TestAllReturnsSorted(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("All() returned no presets")
	}
	for i := 1; i < len(all); i++ {
		if all[i].Name < all[i-1].Name {
			t.Errorf("All() not sorted: %q came after %q", all[i].Name, all[i-1].Name)
		}
	}
}

func TestAllSecretsSimplePreset(t *testing.T) {
	p, _ := Get("openai-stack")
	secrets := p.AllSecrets()
	if len(secrets) != 3 {
		t.Errorf("openai-stack AllSecrets() = %d, want 3", len(secrets))
	}

	keys := map[string]bool{}
	for _, s := range secrets {
		keys[s.Key] = true
	}
	for _, want := range []string{"OPENAI_API_KEY", "OPENAI_ORG_ID", "OPENAI_PROJECT_ID"} {
		if !keys[want] {
			t.Errorf("openai-stack missing key %q", want)
		}
	}
}

func TestAllSecretsCompositePreset(t *testing.T) {
	p, _ := Get("full-llm-stack")
	secrets := p.AllSecrets()

	// full-llm-stack includes 8 provider presets, should have all their secrets
	if len(secrets) < 8 {
		t.Errorf("full-llm-stack AllSecrets() = %d, want at least 8", len(secrets))
	}

	// Spot-check a few keys from different providers
	keys := map[string]bool{}
	for _, s := range secrets {
		keys[s.Key] = true
	}
	checks := []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "HF_TOKEN", "REPLICATE_API_TOKEN"}
	for _, want := range checks {
		if !keys[want] {
			t.Errorf("full-llm-stack missing key %q", want)
		}
	}
}

func TestAllSecretsMlopsComposite(t *testing.T) {
	p, _ := Get("mlops-stack")
	secrets := p.AllSecrets()

	keys := map[string]bool{}
	for _, s := range secrets {
		keys[s.Key] = true
	}
	if !keys["WANDB_API_KEY"] {
		t.Error("mlops-stack missing WANDB_API_KEY")
	}
	if !keys["LANGCHAIN_API_KEY"] {
		t.Error("mlops-stack missing LANGCHAIN_API_KEY")
	}
}

func TestProviderForKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"OPENAI_API_KEY", "openai-stack"},
		{"ANTHROPIC_API_KEY", "anthropic-stack"},
		{"HF_TOKEN", "huggingface-stack"},
		{"WANDB_API_KEY", "wandb-stack"},
		{"UNKNOWN_KEY", ""},
	}

	for _, tt := range tests {
		got := ProviderForKey(tt.key)
		if got != tt.want {
			t.Errorf("ProviderForKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestSecretDefForKey(t *testing.T) {
	def := SecretDefForKey("OPENAI_API_KEY")
	if def == nil {
		t.Fatal("SecretDefForKey(OPENAI_API_KEY) returned nil")
	}
	if !def.Required {
		t.Error("OPENAI_API_KEY should be required")
	}
	if def.Pattern == "" {
		t.Error("OPENAI_API_KEY should have a validation pattern")
	}
	if def.RotationDays != 90 {
		t.Errorf("OPENAI_API_KEY rotation_days = %d, want 90", def.RotationDays)
	}
}

func TestSecretDefForKeyUnknown(t *testing.T) {
	def := SecretDefForKey("TOTALLY_UNKNOWN")
	if def != nil {
		t.Errorf("SecretDefForKey for unknown key should be nil, got %+v", def)
	}
}

func TestSecretDefForKeyOptional(t *testing.T) {
	def := SecretDefForKey("HF_HOME")
	if def == nil {
		t.Fatal("SecretDefForKey(HF_HOME) returned nil")
	}
	if def.Required {
		t.Error("HF_HOME should not be required")
	}
}

func TestRegisterCustomPreset(t *testing.T) {
	Register(&Preset{
		Name:        "test-custom",
		Description: "test preset",
		Secrets: []SecretDef{
			{Key: "CUSTOM_KEY", Required: true},
		},
	})

	p, err := Get("test-custom")
	if err != nil {
		t.Fatalf("custom preset not found: %v", err)
	}
	if len(p.Secrets) != 1 {
		t.Errorf("custom preset has %d secrets, want 1", len(p.Secrets))
	}

	// Clean up
	delete(registry, "test-custom")
}
