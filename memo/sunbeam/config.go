package sunbeam

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Preferences holds the memo token and URL.
type Preferences struct {
	MemoToken string `json:"memo_token"`
	MemoURL   string `json:"memo_url"`
}

// MemoExtension holds the preferences for the memos extension.
type MemoExtension struct {
	Origin      string      `json:"origin"`
	Preferences Preferences `json:"preferences"`
}

// Extensions contains all extensions, including memos.
type Extensions struct {
	Memos MemoExtension `json:"memos"`
}

// SunbeamConfig represents the full Sunbeam configuration.
type SunbeamConfig struct {
	Extensions Extensions `json:"extensions"`
}

// readSunbeamConfig reads the Sunbeam extensions configuration file and retrieves the memo token and URL.
func ReadSunbeamConfig(configPath string) (Preferences, error) {
	// Open the configuration file
	file, err := os.Open(configPath)
	if err != nil {
		return Preferences{}, fmt.Errorf("error opening configuration file: %v", err)
	}
	defer file.Close()

	// Read the file content
	data, err := io.ReadAll(file)
	if err != nil {
		return Preferences{}, fmt.Errorf("error reading configuration file: %v", err)
	}

	// Parse JSON content
	var config SunbeamConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Preferences{}, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Return preferences
	return config.Extensions.Memos.Preferences, nil
}
