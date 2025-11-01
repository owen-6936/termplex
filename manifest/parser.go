package manifest

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadFromFile reads a manifest file from the given path and parses it.
func LoadFromFile(filePath string) (*Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	return &m, nil
}
