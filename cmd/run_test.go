package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSandboxValues(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single value fs",
			input: "fs",
			want:  []string{"fs"},
		},
		{
			name:  "single value net",
			input: "net",
			want:  []string{"net"},
		},
		{
			name:  "single value default",
			input: "default",
			want:  []string{"default"},
		},
		{
			name:  "multiple values",
			input: "fs,net",
			want:  []string{"fs", "net"},
		},
		{
			name:  "all values",
			input: "default,fs,net",
			want:  []string{"default", "fs", "net"},
		},
		{
			name:  "with spaces",
			input: " fs , net ",
			want:  []string{"fs", "net"},
		},
		{
			name:  "duplicate values deduplicated",
			input: "fs,fs,net",
			want:  []string{"fs", "net"},
		},
		{
			name:    "invalid value",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "valid and invalid mixed",
			input:   "fs,invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSandboxValues(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSandboxValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !sliceEqual(got, tt.want) {
				t.Errorf("parseSandboxValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransformSandboxValues(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty input",
			input: []string{},
			want:  nil,
		},
		{
			name:  "fs only",
			input: []string{"fs"},
			want:  []string{"fs"},
		},
		{
			name:  "net only",
			input: []string{"net"},
			want:  []string{"net"},
		},
		{
			name:  "default expands to include fs",
			input: []string{"default"},
			want:  []string{"fs", "default"},
		},
		{
			name:  "default with net",
			input: []string{"default", "net"},
			want:  []string{"fs", "default", "net"},
		},
		{
			name:  "default with fs already present",
			input: []string{"fs", "default"},
			want:  []string{"fs", "default"},
		},
		{
			name:  "all values with default",
			input: []string{"default", "fs", "net"},
			want:  []string{"fs", "default", "net"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformSandboxValues(tt.input)
			if !sliceEqual(got, tt.want) {
				t.Errorf("transformSandboxValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteSandboxSettings(t *testing.T) {
	tests := []struct {
		name           string
		sandboxValues  []string
		existingConfig map[string]any
		wantConfig     map[string]any
	}{
		{
			name:          "empty sandbox values does nothing",
			sandboxValues: []string{},
			wantConfig:    nil, // no file should be created
		},
		{
			name:          "fs mode sets allowedDirectories",
			sandboxValues: []string{"fs"},
			wantConfig: map[string]any{
				"sandbox": map[string]any{
					"enabled":            true,
					"allowedDirectories": []any{}, // will be checked separately for path
				},
			},
		},
		{
			name:          "net mode sets allowedDomains",
			sandboxValues: []string{"net"},
			wantConfig: map[string]any{
				"sandbox": map[string]any{
					"enabled":        true,
					"allowedDomains": []any{},
				},
			},
		},
		{
			name:          "fs and net mode sets both",
			sandboxValues: []string{"fs", "net"},
			wantConfig: map[string]any{
				"sandbox": map[string]any{
					"enabled":            true,
					"allowedDirectories": []any{},
					"allowedDomains":     []any{},
				},
			},
		},
		{
			name:          "preserves existing config",
			sandboxValues: []string{"fs"},
			existingConfig: map[string]any{
				"otherSetting": "preserved",
			},
			wantConfig: map[string]any{
				"otherSetting": "preserved",
				"sandbox": map[string]any{
					"enabled":            true,
					"allowedDirectories": []any{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Write existing config if specified
			if tt.existingConfig != nil {
				claudeDir := filepath.Join(tmpDir, ".claude")
				if err := os.MkdirAll(claudeDir, 0755); err != nil {
					t.Fatal(err)
				}
				data, _ := json.Marshal(tt.existingConfig)
				if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
			}

			// Run the function
			err := writeSandboxSettings(tmpDir, tt.sandboxValues)
			if err != nil {
				t.Fatalf("writeSandboxSettings() error = %v", err)
			}

			// Check the result
			settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")

			if tt.wantConfig == nil {
				// File should not exist
				if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
					t.Error("expected settings file to not exist")
				}
				return
			}

			// Read and parse the file
			data, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("failed to read settings file: %v", err)
			}

			var got map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to parse settings file: %v", err)
			}

			// Check that sandbox.enabled is true
			sandbox, ok := got["sandbox"].(map[string]any)
			if !ok {
				t.Fatal("sandbox key missing or not an object")
			}
			if sandbox["enabled"] != true {
				t.Error("sandbox.enabled should be true")
			}

			// Check for allowedDirectories if fs mode
			for _, v := range tt.sandboxValues {
				if v == "fs" {
					dirs, ok := sandbox["allowedDirectories"].([]any)
					if !ok {
						t.Error("allowedDirectories missing or not an array")
					} else if len(dirs) != 1 {
						t.Errorf("allowedDirectories should have 1 entry, got %d", len(dirs))
					}
				}
				if v == "net" {
					_, ok := sandbox["allowedDomains"].([]any)
					if !ok {
						t.Error("allowedDomains missing or not an array")
					}
				}
			}

			// Check that existing config is preserved
			if tt.existingConfig != nil {
				if got["otherSetting"] != "preserved" {
					t.Error("existing config was not preserved")
				}
			}
		})
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
