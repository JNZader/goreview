package commands

import (
	"os"
	"path/filepath"
	"strings"
)

// ProjectInfo contains detected project information.
type ProjectInfo struct {
	Languages   []string
	ProjectType string
	HasGit      bool
	HasCI       bool
	Frameworks  []string
}

// DetectProject analyzes the current directory for project info.
func DetectProject(dir string) (*ProjectInfo, error) {
	info := &ProjectInfo{}

	// Check for git
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		info.HasGit = true
	}

	// Check for CI
	ciFiles := []string{
		".github/workflows",
		".gitlab-ci.yml",
		".circleci",
		"Jenkinsfile",
	}
	for _, ci := range ciFiles {
		if _, err := os.Stat(filepath.Join(dir, ci)); err == nil {
			info.HasCI = true
			break
		}
	}

	// Detect languages and project types
	info.detectLanguages(dir)
	info.detectFrameworks(dir)

	return info, nil
}

func (p *ProjectInfo) detectLanguages(dir string) {
	languageFiles := map[string]string{
		"go.mod":           "go",
		"package.json":     "javascript",
		"requirements.txt": "python",
		"Cargo.toml":       "rust",
		"pom.xml":          "java",
		"build.gradle":     "java",
		"Gemfile":          "ruby",
		"composer.json":    "php",
	}

	for file, lang := range languageFiles {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			p.Languages = append(p.Languages, lang)
			p.ProjectType = file
		}
	}

	// If no specific files found, scan for source files
	if len(p.Languages) == 0 {
		p.scanForLanguages(dir)
	}
}

func (p *ProjectInfo) scanForLanguages(dir string) {
	extToLang := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".py":   "python",
		".rs":   "rust",
		".java": "java",
		".rb":   "ruby",
		".php":  "php",
	}

	seen := make(map[string]bool)

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden and vendor directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if lang, ok := extToLang[ext]; ok {
			if !seen[lang] {
				p.Languages = append(p.Languages, lang)
				seen[lang] = true
			}
		}

		return nil
	})
}

func (p *ProjectInfo) detectFrameworks(dir string) {
	// Check package.json for JS frameworks
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		content := string(data)
		frameworks := []string{"react", "vue", "angular", "express", "next", "nest"}
		for _, fw := range frameworks {
			if strings.Contains(content, "\""+fw) {
				p.Frameworks = append(p.Frameworks, fw)
			}
		}
	}

	// Check go.mod for Go frameworks
	if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		content := string(data)
		if strings.Contains(content, "gin-gonic") {
			p.Frameworks = append(p.Frameworks, "gin")
		}
		if strings.Contains(content, "echo") {
			p.Frameworks = append(p.Frameworks, "echo")
		}
		if strings.Contains(content, "fiber") {
			p.Frameworks = append(p.Frameworks, "fiber")
		}
	}
}

// SuggestDefaults returns suggested configuration based on project info.
func (p *ProjectInfo) SuggestDefaults() map[string]interface{} {
	excludes := []string{}

	// Language-specific excludes
	for _, lang := range p.Languages {
		switch lang {
		case "javascript", "typescript":
			excludes = append(excludes,
				"node_modules/**", "dist/**", "build/**", "*.min.js")
		case "go":
			excludes = append(excludes,
				"vendor/**", "*_test.go")
		case "python":
			excludes = append(excludes,
				"__pycache__/**", "venv/**", ".venv/**", "*.pyc")
		}
	}

	return map[string]interface{}{
		"provider": "ollama",
		"model":    "codellama",
		"preset":   "standard",
		"exclude":  excludes,
	}
}

// PrimaryLanguage returns the main language of the project.
func (p *ProjectInfo) PrimaryLanguage() string {
	if len(p.Languages) > 0 {
		return p.Languages[0]
	}
	return "unknown"
}
