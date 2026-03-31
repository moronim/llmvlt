package cmd

import (
	"fmt"
	"os"

	"github.com/moronim/aikeys/preset"
	"github.com/moronim/aikeys/store"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault",
	Long: `Initialize a new encrypted vault, optionally with a provider preset.

Examples:
  aikeys init
  aikeys init --preset openai-stack
  aikeys init --preset full-llm-stack`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("preset", "", "provider preset to scaffold (e.g. openai-stack, full-llm-stack)")
}

func runInit(cmd *cobra.Command, args []string) error {
	storePath := getStorePath()

	// Check if store already exists
	if _, err := os.Stat(storePath); err == nil {
		return fmt.Errorf("vault already exists at %s — use 'aikeys set' to add secrets", storePath)
	}

	// Get password
	password, err := promptPasswordConfirm()
	if err != nil {
		return err
	}

	// Create empty vault
	v := store.NewVault()

	// If preset specified, scaffold empty keys
	presetName, _ := cmd.Flags().GetString("preset")
	if presetName != "" {
		p, err := preset.Get(presetName)
		if err != nil {
			return fmt.Errorf("unknown preset %q — run 'aikeys presets' to see available presets", presetName)
		}

		for _, s := range p.AllSecrets() {
			v.Set(s.Key, "")
		}

		fmt.Fprintf(os.Stderr, "✓ Initialized vault with preset: %s\n", p.Name)
		fmt.Fprintf(os.Stderr, "  %d secrets scaffolded. Fill them in with:\n", len(p.AllSecrets()))
		for _, s := range p.AllSecrets() {
			marker := "(required)"
			if !s.Required {
				marker = "(optional)"
			}
			fmt.Fprintf(os.Stderr, "    aikeys set %s  %s\n", s.Key, marker)
		}
	} else {
		fmt.Fprintln(os.Stderr, "✓ Initialized empty vault")
		fmt.Fprintln(os.Stderr, "  Add secrets with: aikeys set KEY value")
		fmt.Fprintln(os.Stderr, "  Or init with a preset: aikeys init --preset openai-stack")
	}

	// Save
	return store.Save(storePath, password, v)
}
