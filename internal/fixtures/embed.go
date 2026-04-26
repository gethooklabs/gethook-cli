package fixtures

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed stripe/*.json github/*.json shopify/*.json slack/*.json
var FS embed.FS

// Providers returns all available provider names.
func Providers() []string {
	entries, _ := FS.ReadDir(".")
	providers := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			providers = append(providers, e.Name())
		}
	}
	return providers
}

// EventTypes returns all event types for a provider.
func EventTypes(provider string) ([]string, error) {
	entries, err := FS.ReadDir(provider)
	if err != nil {
		return nil, fmt.Errorf("unknown provider %q", provider)
	}
	types := make([]string, 0, len(entries))
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".json")
		types = append(types, name)
	}
	return types, nil
}

// Load returns the fixture payload for provider/eventType.
// Merges any override fields from extraJSON on top of the base fixture.
func Load(provider, eventType string, overrides map[string]interface{}) ([]byte, error) {
	path := fmt.Sprintf("%s/%s.json", provider, eventType)
	data, err := FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no fixture for %s/%s — run `gethook trigger --list` to see available types", provider, eventType)
	}

	if len(overrides) == 0 {
		return data, nil
	}

	// Merge overrides into the base fixture.
	var base map[string]interface{}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("parse fixture: %w", err)
	}
	for k, v := range overrides {
		base[k] = v
	}
	return json.Marshal(base)
}
