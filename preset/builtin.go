package preset

// init registers all built-in presets. These are defined in Go rather than
// loading YAML at runtime to keep the binary self-contained with zero
// external file dependencies. Custom user presets can still be loaded from YAML.
func init() {
	Register(&Preset{
		Name:        "openai-stack",
		Description: "OpenAI API credentials",
		Docs:        "https://platform.openai.com/api-keys",
		Secrets: []SecretDef{
			{
				Key:          "OPENAI_API_KEY",
				Description:  "Your OpenAI API key",
				Required:     true,
				Pattern:      `^sk-(proj-)?[a-zA-Z0-9_-]{30,}$`,
				PatternHint:  "Should start with 'sk-' or 'sk-proj-'",
				RotationDays: 90,
			},
			{
				Key:         "OPENAI_ORG_ID",
				Description: "OpenAI organization ID",
				Required:    false,
				Pattern:     `^org-[a-zA-Z0-9]+$`,
				PatternHint: "Should start with 'org-'",
			},
			{
				Key:         "OPENAI_PROJECT_ID",
				Description: "OpenAI project ID (for scoped keys)",
				Required:    false,
				Pattern:     `^proj_[a-zA-Z0-9]+$`,
				PatternHint: "Should start with 'proj_'",
			},
		},
	})

	Register(&Preset{
		Name:        "anthropic-stack",
		Description: "Anthropic (Claude) API credentials",
		Docs:        "https://console.anthropic.com/settings/keys",
		Secrets: []SecretDef{
			{
				Key:          "ANTHROPIC_API_KEY",
				Description:  "Your Anthropic API key",
				Required:     true,
				Pattern:      `^sk-ant-[a-zA-Z0-9_-]{30,}$`,
				PatternHint:  "Should start with 'sk-ant-'",
				RotationDays: 90,
			},
		},
	})

	Register(&Preset{
		Name:        "huggingface-stack",
		Description: "Hugging Face credentials",
		Docs:        "https://huggingface.co/settings/tokens",
		Secrets: []SecretDef{
			{
				Key:          "HF_TOKEN",
				Description:  "Hugging Face access token",
				Required:     true,
				Pattern:      `^hf_[a-zA-Z0-9]+$`,
				PatternHint:  "Should start with 'hf_'",
				RotationDays: 180,
			},
			{
				Key:         "HF_HOME",
				Description: "Hugging Face cache directory (optional)",
				Required:    false,
			},
		},
	})

	Register(&Preset{
		Name:        "replicate-stack",
		Description: "Replicate API credentials",
		Docs:        "https://replicate.com/account/api-tokens",
		Secrets: []SecretDef{
			{
				Key:          "REPLICATE_API_TOKEN",
				Description:  "Replicate API token",
				Required:     true,
				Pattern:      `^r8_[a-zA-Z0-9]+$`,
				PatternHint:  "Should start with 'r8_'",
				RotationDays: 90,
			},
		},
	})

	Register(&Preset{
		Name:        "wandb-stack",
		Description: "Weights & Biases credentials",
		Docs:        "https://wandb.ai/authorize",
		Secrets: []SecretDef{
			{
				Key:          "WANDB_API_KEY",
				Description:  "W&B API key",
				Required:     true,
				Pattern:      `^[a-f0-9]{40}$`,
				PatternHint:  "Should be a 40-character hex string",
				RotationDays: 180,
			},
			{
				Key:         "WANDB_PROJECT",
				Description: "Default W&B project name",
				Required:    false,
			},
			{
				Key:         "WANDB_ENTITY",
				Description: "W&B team/user entity",
				Required:    false,
			},
		},
	})

	Register(&Preset{
		Name:        "langchain-stack",
		Description: "LangChain / LangSmith credentials",
		Docs:        "https://smith.langchain.com/settings",
		Secrets: []SecretDef{
			{
				Key:          "LANGCHAIN_API_KEY",
				Description:  "LangSmith API key",
				Required:     true,
				Pattern:      `^lsv2_[a-zA-Z0-9_]+$`,
				PatternHint:  "Should start with 'lsv2_'",
				RotationDays: 90,
			},
			{
				Key:         "LANGCHAIN_TRACING_V2",
				Description: "Enable LangSmith tracing (set to 'true')",
				Required:    false,
				Pattern:     `^(true|false)$`,
				PatternHint: "Should be 'true' or 'false'",
			},
			{
				Key:         "LANGCHAIN_ENDPOINT",
				Description: "LangSmith API endpoint",
				Required:    false,
			},
		},
	})

	Register(&Preset{
		Name:        "together-stack",
		Description: "Together AI credentials",
		Docs:        "https://api.together.xyz/settings/api-keys",
		Secrets: []SecretDef{
			{
				Key:          "TOGETHER_API_KEY",
				Description:  "Together AI API key",
				Required:     true,
				RotationDays: 90,
			},
		},
	})

	Register(&Preset{
		Name:        "mistral-stack",
		Description: "Mistral AI credentials",
		Docs:        "https://console.mistral.ai/api-keys",
		Secrets: []SecretDef{
			{
				Key:          "MISTRAL_API_KEY",
				Description:  "Mistral AI API key",
				Required:     true,
				RotationDays: 90,
			},
		},
	})

	Register(&Preset{
		Name:        "google-ai-stack",
		Description: "Google AI (Gemini) credentials",
		Docs:        "https://aistudio.google.com/app/apikey",
		Secrets: []SecretDef{
			{
				Key:          "GOOGLE_API_KEY",
				Description:  "Google AI API key for Gemini",
				Required:     true,
				RotationDays: 90,
			},
		},
	})

	Register(&Preset{
		Name:        "cohere-stack",
		Description: "Cohere API credentials",
		Docs:        "https://dashboard.cohere.com/api-keys",
		Secrets: []SecretDef{
			{
				Key:          "COHERE_API_KEY",
				Description:  "Cohere API key",
				Required:     true,
				RotationDays: 90,
			},
		},
	})

	// Composite presets
	Register(&Preset{
		Name:        "full-llm-stack",
		Description: "All major LLM provider credentials combined",
		Includes: []string{
			"openai-stack",
			"anthropic-stack",
			"huggingface-stack",
			"replicate-stack",
			"together-stack",
			"mistral-stack",
			"google-ai-stack",
			"cohere-stack",
		},
	})

	Register(&Preset{
		Name:        "mlops-stack",
		Description: "ML experiment tracking & ops credentials",
		Includes: []string{
			"wandb-stack",
			"langchain-stack",
		},
	})
}
