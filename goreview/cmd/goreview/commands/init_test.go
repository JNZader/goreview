package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject(t *testing.T) {
	// Create temp directory with test files
	dir := t.TempDir()

	// Create go.mod
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	// Create .git directory
	os.Mkdir(filepath.Join(dir, ".git"), 0755)

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject() error = %v", err)
	}

	if !info.HasGit {
		t.Error("HasGit should be true")
	}

	found := false
	for _, lang := range info.Languages {
		if lang == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect Go language")
	}
}

func TestDetectProjectJavaScript(t *testing.T) {
	dir := t.TempDir()

	// Create package.json
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test"}`), 0644)

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject() error = %v", err)
	}

	found := false
	for _, lang := range info.Languages {
		if lang == "javascript" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect JavaScript language")
	}
}

func TestDetectProjectCI(t *testing.T) {
	dir := t.TempDir()

	// Create GitHub workflows directory
	os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755)

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject() error = %v", err)
	}

	if !info.HasCI {
		t.Error("HasCI should be true for GitHub Actions")
	}
}

func TestProjectInfoSuggestDefaults(t *testing.T) {
	tests := []struct {
		name      string
		languages []string
		wantExcl  []string
	}{
		{
			name:      "go project",
			languages: []string{"go"},
			wantExcl:  []string{"vendor/**"},
		},
		{
			name:      "js project",
			languages: []string{"javascript"},
			wantExcl:  []string{"node_modules/**"},
		},
		{
			name:      "python project",
			languages: []string{"python"},
			wantExcl:  []string{"__pycache__/**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ProjectInfo{Languages: tt.languages}
			defaults := info.SuggestDefaults()

			excludes := defaults["exclude"].([]string)
			found := false
			for _, excl := range excludes {
				for _, want := range tt.wantExcl {
					if excl == want {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("Expected excludes %v, got %v", tt.wantExcl, excludes)
			}
		})
	}
}

func TestProjectInfoSuggestDefaultsProvider(t *testing.T) {
	info := &ProjectInfo{Languages: []string{"go"}}
	defaults := info.SuggestDefaults()

	if defaults["provider"] != "ollama" {
		t.Errorf("Default provider should be ollama, got %v", defaults["provider"])
	}
	if defaults["model"] != "codellama" {
		t.Errorf("Default model should be codellama, got %v", defaults["model"])
	}
	if defaults["preset"] != "standard" {
		t.Errorf("Default preset should be standard, got %v", defaults["preset"])
	}
}

func TestBuildYAMLConfig(t *testing.T) {
	config := map[string]interface{}{
		"provider": "ollama",
		"model":    "codellama",
		"preset":   "standard",
		"exclude":  []string{"vendor/**"},
	}

	yamlConfig := buildYAMLConfig(config)

	// Check structure
	if yamlConfig["version"] != "1.0" {
		t.Error("Version should be 1.0")
	}

	provider := yamlConfig["provider"].(map[string]interface{})
	if provider["name"] != "ollama" {
		t.Error("Provider name should be ollama")
	}
	if provider["model"] != "codellama" {
		t.Error("Model should be codellama")
	}
	if provider["base_url"] != "http://localhost:11434" {
		t.Error("Ollama base_url should be set")
	}

	rules := yamlConfig["rules"].(map[string]interface{})
	if rules["preset"] != "standard" {
		t.Error("Preset should be standard")
	}
}

func TestBuildYAMLConfigOpenAI(t *testing.T) {
	config := map[string]interface{}{
		"provider": "openai",
		"model":    "gpt-4",
		"preset":   "strict",
		"exclude":  []string{},
	}

	yamlConfig := buildYAMLConfig(config)

	provider := yamlConfig["provider"].(map[string]interface{})
	if provider["name"] != "openai" {
		t.Error("Provider name should be openai")
	}
	if provider["api_key"] != "${OPENAI_API_KEY}" {
		t.Error("OpenAI api_key placeholder should be set")
	}
}

func TestPrimaryLanguage(t *testing.T) {
	tests := []struct {
		languages []string
		want      string
	}{
		{[]string{"go", "javascript"}, "go"},
		{[]string{"python"}, "python"},
		{[]string{}, "unknown"},
	}

	for _, tt := range tests {
		info := &ProjectInfo{Languages: tt.languages}
		got := info.PrimaryLanguage()
		if got != tt.want {
			t.Errorf("PrimaryLanguage() = %q, want %q", got, tt.want)
		}
	}
}

func TestDetectFrameworks(t *testing.T) {
	dir := t.TempDir()

	// Create package.json with React
	packageJSON := `{
		"name": "test",
		"dependencies": {
			"react": "^18.0.0"
		}
	}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject() error = %v", err)
	}

	found := false
	for _, fw := range info.Frameworks {
		if fw == "react" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect React framework")
	}
}

func TestScanForLanguages(t *testing.T) {
	dir := t.TempDir()

	// Create source files without project files
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "app.py"), []byte("print('hello')"), 0644)

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject() error = %v", err)
	}

	if len(info.Languages) < 2 {
		t.Errorf("Should detect multiple languages, got %v", info.Languages)
	}
}
