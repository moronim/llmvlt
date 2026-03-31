package preset

import (
	"fmt"
	"sort"
)

// SecretDef defines a single secret within a preset.
type SecretDef struct {
	Key          string `yaml:"key"`
	Description  string `yaml:"description"`
	Required     bool   `yaml:"required"`
	Pattern      string `yaml:"pattern"`
	PatternHint  string `yaml:"hint"`
	RotationDays int    `yaml:"rotation_days"`
}

// Preset defines a collection of secrets for a provider or stack.
type Preset struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Docs        string      `yaml:"docs"`
	Secrets     []SecretDef `yaml:"secrets"`
	Includes    []string    `yaml:"includes"`
}

// AllSecrets returns all secrets including those from included presets.
func (p *Preset) AllSecrets() []SecretDef {
	var all []SecretDef
	for _, inc := range p.Includes {
		if sub, err := Get(inc); err == nil {
			all = append(all, sub.AllSecrets()...)
		}
	}
	all = append(all, p.Secrets...)
	return all
}

// registry holds all loaded presets.
var registry = map[string]*Preset{}

// Register adds a preset to the global registry.
func Register(p *Preset) {
	registry[p.Name] = p
}

// Get retrieves a preset by name.
func Get(name string) (*Preset, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("preset %q not found", name)
	}
	return p, nil
}

// All returns all registered presets, sorted by name.
func All() []*Preset {
	var out []*Preset
	for _, p := range registry {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ProviderForKey returns the provider name for a known key, or "".
func ProviderForKey(key string) string {
	for _, p := range registry {
		// Skip composite presets (those with only includes)
		if len(p.Includes) > 0 && len(p.Secrets) == 0 {
			continue
		}
		for _, s := range p.Secrets {
			if s.Key == key {
				return p.Name
			}
		}
	}
	return ""
}

// SecretDefForKey finds the SecretDef for a given key across all presets.
func SecretDefForKey(key string) *SecretDef {
	for _, p := range registry {
		for _, s := range p.Secrets {
			if s.Key == key {
				return &s
			}
		}
	}
	return nil
}
