package rag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// FrameworkDetector detects frameworks and libraries in a project.
type FrameworkDetector struct {
	projectRoot string
}

// NewFrameworkDetector creates a new framework detector.
func NewFrameworkDetector(projectRoot string) *FrameworkDetector {
	return &FrameworkDetector{projectRoot: projectRoot}
}

// Detect scans the project and returns detected frameworks.
func (d *FrameworkDetector) Detect() []DetectedFramework {
	var frameworks []DetectedFramework

	// Check for Go modules
	if goMod, err := d.readFile("go.mod"); err == nil {
		frameworks = append(frameworks, d.detectGoFrameworks(goMod)...)
	}

	// Check for Node.js
	if packageJSON, err := d.readFile("package.json"); err == nil {
		frameworks = append(frameworks, d.detectNodeFrameworks(packageJSON)...)
	}

	// Check for Python
	if requirements, err := d.readFile("requirements.txt"); err == nil {
		frameworks = append(frameworks, d.detectPythonFrameworks(requirements)...)
	}
	if pyproject, err := d.readFile("pyproject.toml"); err == nil {
		frameworks = append(frameworks, d.detectPythonFrameworks(pyproject)...)
	}

	// Check for Java
	if pom, err := d.readFile("pom.xml"); err == nil {
		frameworks = append(frameworks, d.detectJavaFrameworks(pom)...)
	}

	// Check for Rust
	if cargo, err := d.readFile("Cargo.toml"); err == nil {
		frameworks = append(frameworks, d.detectRustFrameworks(cargo)...)
	}

	return frameworks
}

func (d *FrameworkDetector) readFile(name string) (string, error) {
	cleanPath := filepath.Clean(filepath.Join(d.projectRoot, name))
	data, err := os.ReadFile(cleanPath) // #nosec G304 - path built from project root + known filenames
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (d *FrameworkDetector) detectGoFrameworks(goMod string) []DetectedFramework {
	var frameworks []DetectedFramework

	goFrameworks := map[string]DetectedFramework{
		"github.com/gin-gonic/gin": {
			Name:     "Gin",
			Language: "go",
			DocsURL:  "https://gin-gonic.com/docs/",
		},
		"github.com/labstack/echo": {
			Name:     "Echo",
			Language: "go",
			DocsURL:  "https://echo.labstack.com/docs",
		},
		"github.com/gofiber/fiber": {
			Name:     "Fiber",
			Language: "go",
			DocsURL:  "https://docs.gofiber.io/",
		},
		"github.com/gorilla/mux": {
			Name:     "Gorilla Mux",
			Language: "go",
			DocsURL:  "https://pkg.go.dev/github.com/gorilla/mux",
		},
		"gorm.io/gorm": {
			Name:     "GORM",
			Language: "go",
			DocsURL:  "https://gorm.io/docs/",
		},
		"github.com/spf13/cobra": {
			Name:     "Cobra",
			Language: "go",
			DocsURL:  "https://cobra.dev/",
		},
		"github.com/spf13/viper": {
			Name:     "Viper",
			Language: "go",
			DocsURL:  "https://pkg.go.dev/github.com/spf13/viper",
		},
	}

	for dep, fw := range goFrameworks {
		if strings.Contains(goMod, dep) {
			fw.Confidence = 0.95
			frameworks = append(frameworks, fw)
		}
	}

	return frameworks
}

func (d *FrameworkDetector) detectNodeFrameworks(packageJSON string) []DetectedFramework {
	var frameworks []DetectedFramework

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal([]byte(packageJSON), &pkg); err != nil {
		return frameworks
	}

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	nodeFrameworks := map[string]DetectedFramework{
		"react": {
			Name:     "React",
			Language: "javascript",
			DocsURL:  "https://react.dev/reference/react",
		},
		"vue": {
			Name:     "Vue.js",
			Language: "javascript",
			DocsURL:  "https://vuejs.org/guide/introduction.html",
		},
		"@angular/core": {
			Name:     "Angular",
			Language: "typescript",
			DocsURL:  "https://angular.io/docs",
		},
		"next": {
			Name:     "Next.js",
			Language: "javascript",
			DocsURL:  "https://nextjs.org/docs",
		},
		"express": {
			Name:     "Express",
			Language: "javascript",
			DocsURL:  "https://expressjs.com/en/guide/routing.html",
		},
		"nestjs": {
			Name:     "NestJS",
			Language: "typescript",
			DocsURL:  "https://docs.nestjs.com/",
		},
		"typescript": {
			Name:     "TypeScript",
			Language: "typescript",
			DocsURL:  "https://www.typescriptlang.org/docs/",
		},
		"jest": {
			Name:     "Jest",
			Language: "javascript",
			DocsURL:  "https://jestjs.io/docs/getting-started",
		},
	}

	for dep, fw := range nodeFrameworks {
		if version, ok := allDeps[dep]; ok {
			fw.Version = version
			fw.Confidence = 0.95
			frameworks = append(frameworks, fw)
		}
	}

	return frameworks
}

func (d *FrameworkDetector) detectPythonFrameworks(content string) []DetectedFramework {
	var frameworks []DetectedFramework

	pythonFrameworks := map[string]DetectedFramework{
		"django": {
			Name:     "Django",
			Language: "python",
			DocsURL:  "https://docs.djangoproject.com/",
		},
		"flask": {
			Name:     "Flask",
			Language: "python",
			DocsURL:  "https://flask.palletsprojects.com/",
		},
		"fastapi": {
			Name:     "FastAPI",
			Language: "python",
			DocsURL:  "https://fastapi.tiangolo.com/",
		},
		"sqlalchemy": {
			Name:     "SQLAlchemy",
			Language: "python",
			DocsURL:  "https://docs.sqlalchemy.org/",
		},
		"pytest": {
			Name:     "pytest",
			Language: "python",
			DocsURL:  "https://docs.pytest.org/",
		},
		"pandas": {
			Name:     "pandas",
			Language: "python",
			DocsURL:  "https://pandas.pydata.org/docs/",
		},
	}

	contentLower := strings.ToLower(content)
	for dep, fw := range pythonFrameworks {
		if strings.Contains(contentLower, dep) {
			fw.Confidence = 0.85
			frameworks = append(frameworks, fw)
		}
	}

	return frameworks
}

func (d *FrameworkDetector) detectJavaFrameworks(pom string) []DetectedFramework {
	var frameworks []DetectedFramework

	javaFrameworks := map[string]DetectedFramework{
		"spring-boot": {
			Name:     "Spring Boot",
			Language: "java",
			DocsURL:  "https://docs.spring.io/spring-boot/docs/current/reference/html/",
		},
		"spring-framework": {
			Name:     "Spring Framework",
			Language: "java",
			DocsURL:  "https://docs.spring.io/spring-framework/reference/",
		},
		"junit": {
			Name:     "JUnit",
			Language: "java",
			DocsURL:  "https://junit.org/junit5/docs/current/user-guide/",
		},
		"hibernate": {
			Name:     "Hibernate",
			Language: "java",
			DocsURL:  "https://hibernate.org/orm/documentation/",
		},
	}

	for dep, fw := range javaFrameworks {
		if strings.Contains(pom, dep) {
			fw.Confidence = 0.9
			frameworks = append(frameworks, fw)
		}
	}

	return frameworks
}

func (d *FrameworkDetector) detectRustFrameworks(cargo string) []DetectedFramework {
	var frameworks []DetectedFramework

	rustFrameworks := map[string]DetectedFramework{
		"actix-web": {
			Name:     "Actix Web",
			Language: "rust",
			DocsURL:  "https://actix.rs/docs/",
		},
		"tokio": {
			Name:     "Tokio",
			Language: "rust",
			DocsURL:  "https://tokio.rs/tokio/tutorial",
		},
		"serde": {
			Name:     "Serde",
			Language: "rust",
			DocsURL:  "https://serde.rs/",
		},
		"diesel": {
			Name:     "Diesel",
			Language: "rust",
			DocsURL:  "https://diesel.rs/guides/getting-started",
		},
	}

	for dep, fw := range rustFrameworks {
		if strings.Contains(cargo, dep) {
			fw.Confidence = 0.9
			frameworks = append(frameworks, fw)
		}
	}

	return frameworks
}

// GetRelevantDocs returns documentation URLs relevant to a specific language.
func GetRelevantDocs(language string) []Source {
	docs := map[string][]Source{
		"go": {
			{URL: "https://go.dev/doc/effective_go", Type: SourceTypeStyleGuide, Name: "Effective Go", Enabled: true},
			{URL: "https://go.dev/doc/faq", Type: SourceTypeBestPractice, Name: "Go FAQ", Enabled: true},
		},
		"javascript": {
			{URL: "https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide", Type: SourceTypeStyleGuide, Name: "MDN JavaScript Guide", Enabled: true},
		},
		"typescript": {
			{URL: "https://www.typescriptlang.org/docs/handbook/intro.html", Type: SourceTypeStyleGuide, Name: "TypeScript Handbook", Enabled: true},
		},
		"python": {
			{URL: "https://peps.python.org/pep-0008/", Type: SourceTypeStyleGuide, Name: "PEP 8", Enabled: true},
		},
		"java": {
			{URL: "https://google.github.io/styleguide/javaguide.html", Type: SourceTypeStyleGuide, Name: "Google Java Style", Enabled: true},
		},
		"rust": {
			{URL: "https://doc.rust-lang.org/book/", Type: SourceTypeStyleGuide, Name: "The Rust Book", Enabled: true},
		},
	}

	// Common security docs for all languages
	securityDocs := []Source{
		{URL: "https://owasp.org/www-project-top-ten/", Type: SourceTypeSecurity, Name: "OWASP Top 10", Enabled: true},
	}

	result := docs[language]
	result = append(result, securityDocs...)

	return result
}
