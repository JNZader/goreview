# Iteracion 00: Inicializacion del Proyecto

## Objetivos

Al completar esta iteracion tendras:
- Estructura base del proyecto creada
- Modulo Go inicializado
- Configuracion de editor y gitignore
- Makefile funcional para builds
- README inicial del proyecto

## Prerequisitos

- Go 1.24+ instalado
- Git instalado
- Editor de texto (VSCode recomendado)
- GitHub CLI (`gh`) instalado y autenticado

---

## Workflow GitFlow

Esta iteracion tiene **4 commits atomicos**. Cada commit = 1 rama feature = 1 PR = 1 merge.

### Ramas de Esta Iteracion

| Commit | Rama | Tipo |
|--------|------|------|
| 0.1 | `feature/00-01-project-structure` | chore |
| 0.2 | `feature/00-02-gitignore-editor` | chore |
| 0.3 | `feature/00-03-makefile` | chore |
| 0.4 | `feature/00-04-readme` | docs |

### Flujo para Cada Commit

```bash
# 1. Preparar rama (ANTES de cada commit)
git checkout develop && git pull origin develop
git checkout -b feature/00-XX-slug

# 2. Implementar cambios (segun esta guia)
# ... crear archivos ...

# 3. Commit con mensaje exacto
git add .
git commit -m "mensaje de la guia"

# 4. Push y crear PR
git push -u origin feature/00-XX-slug
gh pr create --base develop --fill

# 5. Merge y limpiar
gh pr merge --squash --delete-branch
git checkout develop && git pull
```

> **Guia completa:** [GITFLOW-SOLO-DEV.md](GITFLOW-SOLO-DEV.md)

---

## Setup Inicial del Repositorio

Antes de comenzar con los commits, necesitas inicializar el repositorio Git y configurar GitFlow:

```bash
# 1. Clonar repositorio existente
git clone https://github.com/JNZader/goreview.git ai-toolkit
cd ai-toolkit

# 2. Crear rama develop (rama principal de desarrollo en GitFlow)
git checkout -b develop
git push -u origin develop
```

**Verificacion:**
```bash
# Verificar que estas en develop
git branch

# Verificar remotes
git remote -v
```

Ahora estas listo para comenzar con el Commit 0.1.

---

## Commit 0.1: Crear estructura de directorios e inicializar modulo Go

**Rama:** `feature/00-01-project-structure`

```bash
git checkout develop && git pull origin develop
git checkout -b feature/00-01-project-structure
```

**Mensaje de commit:**
```
chore: create initial project structure and initialize Go module

- Create goreview directory for Go CLI
- Create integrations/github-app for Node.js app
- Create goreview-rules for YAML rules
- Create docs directory for documentation
- Initialize go.mod with module path
```

**Archivos a crear:**

Primero, crea la estructura de directorios:

```bash
# Ya estas en el directorio goreview clonado

# Crear estructura de goreview (CLI en Go)
mkdir -p goreview/cmd/goreview/commands
mkdir -p goreview/internal/config
mkdir -p goreview/internal/git
mkdir -p goreview/internal/providers
mkdir -p goreview/internal/review
mkdir -p goreview/internal/cache
mkdir -p goreview/internal/report
mkdir -p goreview/internal/rules
mkdir -p goreview/build

# Crear estructura de GitHub App (Node.js)
mkdir -p integrations/github-app/src/services
mkdir -p integrations/github-app/src/middleware
mkdir -p integrations/github-app/src/handlers
mkdir -p integrations/github-app/src/types
mkdir -p integrations/github-app/tests
mkdir -p integrations/github-app/config

# Crear estructura de reglas
mkdir -p goreview-rules/rules
mkdir -p goreview-rules/languages
mkdir -p goreview-rules/presets

# Crear estructura de docs
mkdir -p docs

# Crear estructura de GitHub Actions
mkdir -p .github/workflows

# Crear directorio de scripts
mkdir -p scripts
```

**Verificacion:**
```bash
# Verificar estructura creada
find . -type d | head -30
```

**Explicacion didactica:**

Esta estructura sigue las convenciones de Go para proyectos:

- `cmd/` - Entry points de aplicaciones
- `internal/` - Paquetes privados (no importables desde afuera)
- `build/` - Artefactos de compilacion

La separacion en subdirectorios (`config/`, `git/`, `providers/`, etc.) sigue el principio de **Single Responsibility** - cada paquete tiene una responsabilidad clara.

---

### Paso 2: Inicializar modulo Go

Ahora, inicializa el modulo Go dentro del directorio `goreview`:

**Comando para crear:**
```bash
cd goreview

# Inicializar modulo
go mod init github.com/JNZader/goreview/goreview

# Volver a la raiz
cd ..
```

**Archivo resultante: `goreview/go.mod`**

```go
module github.com/JNZader/goreview/goreview

go 1.24.0
```

**Verificacion:**
```bash
cat goreview/go.mod
```

**Explicacion didactica:**

El archivo `go.mod` es el corazon de los modulos de Go. Define:

1. **Module path** (`github.com/JNZader/goreview/goreview`):
   - Identificador unico del modulo
   - Usado para imports internos y externos
   - Convencion: usar URL de repositorio

2. **Go version** (`go 1.24.0`):
   - Version minima requerida
   - Habilita features del lenguaje de esa version

**Nota sobre dependencias:**

Las dependencias (cobra, viper, yaml, etc.) se agregaran automaticamente cuando escribas codigo que las importe. Go las descarga y agrega al `go.mod` de forma automatica. Esto evita tener dependencias "huerfanas" que no se usan

---

**Finalizar commit:**
```bash
# Commit
git add .
git commit -m "chore: create initial project structure and initialize Go module

- Create goreview directory for Go CLI
- Create integrations/github-app for Node.js app
- Create goreview-rules for YAML rules
- Create docs directory for documentation
- Initialize go.mod with module path"

# Push y PR
git push -u origin feature/00-01-project-structure
gh pr create --base develop --title "chore: create initial project structure and initialize Go module" --body "Estructura base del proyecto y modulo Go inicializado"

# Merge
gh pr merge --squash --delete-branch

# Volver a develop
git checkout develop && git pull origin develop
```

---

## Commit 0.2: Agregar .gitignore y configuracion de editor

**Rama:** `feature/00-02-gitignore-editor`

```bash
git checkout -b feature/00-02-gitignore-editor
```

**Mensaje de commit:**
```
chore: add gitignore and editor configuration

- Add comprehensive .gitignore for Go and Node.js
- Add .editorconfig for consistent formatting
- Add VSCode settings for Go development
```

**Archivos a crear:**

### 1. `.gitignore` (en raiz del proyecto)

```gitignore
# ===================
# Go
# ===================

# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
goreview/build/
goreview/bin/

# Test binary
*.test

# Output of go coverage
*.out
coverage.html

# Go workspace file
go.work

# Dependency directories
vendor/

# ===================
# Node.js
# ===================

node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.pnpm-debug.log*

# Build output
integrations/github-app/dist/
integrations/github-app/build/

# ===================
# IDE
# ===================

.idea/
*.swp
*.swo
*~
.vscode/*
!.vscode/settings.json
!.vscode/tasks.json
!.vscode/launch.json
!.vscode/extensions.json

# ===================
# OS
# ===================

.DS_Store
Thumbs.db
ehthumbs.db

# ===================
# Environment
# ===================

.env
.env.local
.env.*.local
*.pem
*.key

# ===================
# Cache
# ===================

.cache/
.goreview/
*.cache

# ===================
# Logs
# ===================

logs/
*.log

# ===================
# Temporary
# ===================

tmp/
temp/
*.tmp
*.temp

# ===================
# Coverage
# ===================

coverage/
.nyc_output/
*.lcov
```

### 2. `.editorconfig` (en raiz del proyecto)

```editorconfig
# EditorConfig helps maintain consistent coding styles
# https://editorconfig.org

root = true

# Default settings for all files
[*]
charset = utf-8
end_of_line = lf
indent_style = tab
indent_size = 4
insert_final_newline = true
trim_trailing_whitespace = true

# Go files
[*.go]
indent_style = tab
indent_size = 4

# YAML files
[*.{yml,yaml}]
indent_style = space
indent_size = 2

# JSON files
[*.json]
indent_style = space
indent_size = 2

# Markdown files
[*.md]
trim_trailing_whitespace = false
indent_style = space
indent_size = 2

# TypeScript/JavaScript
[*.{ts,tsx,js,jsx}]
indent_style = space
indent_size = 2

# Makefile
[Makefile]
indent_style = tab

# Shell scripts
[*.sh]
indent_style = space
indent_size = 2

# Docker
[Dockerfile*]
indent_style = space
indent_size = 2
```

### 3. `.vscode/settings.json`

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "highlight",
    "coveredHighlightColor": "rgba(64,128,128,0.2)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.2)"
  },
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  },
  "[go]": {
    "editor.insertSpaces": false,
    "editor.tabSize": 4,
    "editor.defaultFormatter": "golang.go"
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode",
    "editor.tabSize": 2
  },
  "[yaml]": {
    "editor.tabSize": 2,
    "editor.insertSpaces": true
  },
  "files.associations": {
    "*.yaml": "yaml",
    "*.yml": "yaml"
  },
  "files.exclude": {
    "**/node_modules": true,
    "**/build": true,
    "**/dist": true
  }
}
```

### 4. `.vscode/extensions.json`

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-typescript-next",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint",
    "editorconfig.editorconfig",
    "redhat.vscode-yaml",
    "ms-azuretools.vscode-docker"
  ]
}
```

**Verificacion:**
```bash
# Verificar archivos creados
ls -la .gitignore .editorconfig
ls -la .vscode/
```

**Explicacion didactica:**

1. **.gitignore** previene que archivos innecesarios entren al repo:
   - Binarios compilados (no versionables)
   - `node_modules/` (se regeneran con `npm install`)
   - `.env` (contiene secretos, NUNCA versionar)
   - Cache y logs (temporales)

2. **.editorconfig** garantiza consistencia entre editores:
   - Go usa tabs (convencion del lenguaje)
   - YAML/JSON usan 2 espacios
   - Todos los archivos terminan con newline

3. **VSCode settings** optimizan el desarrollo:
   - `golangci-lint` para linting avanzado
   - `goimports` formatea y organiza imports
   - Tests con `-race` detectan race conditions
   - Coverage visual en el editor

**Finalizar commit:**
```bash
git add .
git commit -m "chore: add gitignore and editor configuration

- Add comprehensive .gitignore for Go and Node.js
- Add .editorconfig for consistent formatting
- Add VSCode settings for Go development"

git push -u origin feature/00-02-gitignore-editor
gh pr create --base develop --fill
gh pr merge --squash --delete-branch
git checkout develop && git pull origin develop
```

---

## Commit 0.3: Agregar Makefile

**Rama:** `feature/00-03-makefile`

```bash
git checkout -b feature/00-03-makefile
```

**Mensaje de commit:**
```
chore: add Makefile with build targets

- Add build, test, lint, clean targets
- Add development helpers (run, dev)
- Add installation target
- Add help documentation
```

**Archivos a crear:**

### 1. `goreview/Makefile`

```makefile
# ===========================================
# GoReview Makefile
# ===========================================

# Variables
BINARY_NAME=goreview
BUILD_DIR=build
CMD_DIR=cmd/goreview
MAIN_FILE=$(CMD_DIR)/main.go

# Version info (se sobreescribe en CI)
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go flags
GOFLAGS=-trimpath
LDFLAGS=-s -w \
	-X 'github.com/JNZader/goreview/goreview/cmd/goreview/commands.Version=$(VERSION)' \
	-X 'github.com/JNZader/goreview/goreview/cmd/goreview/commands.Commit=$(COMMIT)' \
	-X 'github.com/JNZader/goreview/goreview/cmd/goreview/commands.BuildDate=$(BUILD_DATE)'

# Default target
.DEFAULT_GOAL := help

# ===========================================
# Build Targets
# ===========================================

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "All binaries built in $(BUILD_DIR)/"

.PHONY: install
install: build ## Install to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# ===========================================
# Test Targets
# ===========================================

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short: ## Run tests (short mode)
	go test -v -short ./...

.PHONY: coverage
coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

# ===========================================
# Lint Targets
# ===========================================

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix --timeout=5m

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# ===========================================
# Development Targets
# ===========================================

.PHONY: run
run: build ## Build and run
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

.PHONY: dev
dev: ## Run with hot reload (requires air)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with:"; \
		echo "  go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		$(MAKE) run; \
	fi

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

.PHONY: deps-update
deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

.PHONY: generate
generate: ## Run go generate
	go generate ./...

# ===========================================
# Clean Targets
# ===========================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache -testcache

.PHONY: clean-all
clean-all: clean ## Clean everything including deps
	go clean -modcache

# ===========================================
# Docker Targets
# ===========================================

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):$(VERSION) -f Dockerfile.production .

.PHONY: docker-run
docker-run: ## Run in Docker
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# ===========================================
# Release Targets
# ===========================================

.PHONY: release-dry
release-dry: ## Dry run of release
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed"; \
	fi

# ===========================================
# Help
# ===========================================

.PHONY: help
help: ## Show this help
	@echo "GoReview Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
```

### 2. `Makefile` (en raiz del proyecto)

```makefile
# ===========================================
# AI-Toolkit Root Makefile
# ===========================================

.DEFAULT_GOAL := help

# ===========================================
# Setup
# ===========================================

.PHONY: setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@echo ""
	@echo "1. Installing Go dependencies..."
	cd goreview && go mod download
	@echo ""
	@echo "2. Installing Node.js dependencies..."
	cd integrations/github-app && npm install
	@echo ""
	@echo "3. Installing development tools..."
	go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	@echo ""
	@echo "4. Pulling Ollama model..."
	ollama pull qwen2.5-coder:14b || echo "Ollama not available, skipping..."
	@echo ""
	@echo "Setup complete!"

# ===========================================
# Build
# ===========================================

.PHONY: build
build: ## Build all components
	@echo "Building GoReview CLI..."
	cd goreview && $(MAKE) build
	@echo ""
	@echo "Building GitHub App..."
	cd integrations/github-app && npm run build

.PHONY: build-goreview
build-goreview: ## Build only GoReview CLI
	cd goreview && $(MAKE) build

.PHONY: build-github-app
build-github-app: ## Build only GitHub App
	cd integrations/github-app && npm run build

# ===========================================
# Test
# ===========================================

.PHONY: test
test: ## Run all tests
	@echo "Testing GoReview CLI..."
	cd goreview && $(MAKE) test
	@echo ""
	@echo "Testing GitHub App..."
	cd integrations/github-app && npm test

.PHONY: test-goreview
test-goreview: ## Test only GoReview CLI
	cd goreview && $(MAKE) test

.PHONY: test-github-app
test-github-app: ## Test only GitHub App
	cd integrations/github-app && npm test

# ===========================================
# Lint
# ===========================================

.PHONY: lint
lint: ## Lint all code
	@echo "Linting GoReview CLI..."
	cd goreview && $(MAKE) lint
	@echo ""
	@echo "Linting GitHub App..."
	cd integrations/github-app && npm run lint

# ===========================================
# Docker
# ===========================================

.PHONY: docker-up
docker-up: ## Start all services with Docker Compose
	docker compose up -d

.PHONY: docker-down
docker-down: ## Stop all services
	docker compose down

.PHONY: docker-logs
docker-logs: ## Show logs
	docker compose logs -f

.PHONY: docker-build
docker-build: ## Build Docker images
	docker compose build

# ===========================================
# Development
# ===========================================

.PHONY: dev
dev: ## Start development mode
	@echo "Starting development mode..."
	@echo "GoReview: make -C goreview dev"
	@echo "GitHub App: cd integrations/github-app && npm run dev"

.PHONY: ollama-pull
ollama-pull: ## Pull Ollama model
	ollama pull qwen2.5-coder:14b

.PHONY: ollama-serve
ollama-serve: ## Start Ollama server
	ollama serve

# ===========================================
# Clean
# ===========================================

.PHONY: clean
clean: ## Clean all build artifacts
	cd goreview && $(MAKE) clean
	cd integrations/github-app && rm -rf dist node_modules
	rm -rf .cache logs

# ===========================================
# Help
# ===========================================

.PHONY: help
help: ## Show this help
	@echo "AI-Toolkit Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
```

**Verificacion:**
```bash
# Ver ayuda del Makefile
cd goreview
make help

# Volver a raiz y ver ayuda
cd ..
make help
```

**Explicacion didactica:**

El Makefile automatiza tareas repetitivas:

1. **LDFLAGS** inyecta informacion de version en el binario:
   - `Version`: Tag de version (ej: v1.0.0)
   - `Commit`: Hash corto del commit
   - `BuildDate`: Fecha de compilacion

2. **Targets principales**:
   - `build`: Compila el binario
   - `test`: Ejecuta tests con race detector
   - `lint`: Analiza codigo con golangci-lint
   - `clean`: Limpia artefactos

3. **Convencion de help**:
   - Cada target tiene `## Descripcion`
   - `make help` muestra todos los targets

4. **Cross-compilation** (`build-all`):
   - Go permite compilar para cualquier OS/arch
   - `GOOS=linux GOARCH=amd64` compila para Linux 64-bit

**Finalizar commit:**
```bash
git add .
git commit -m "chore: add Makefile with build targets

- Add build, test, lint, clean targets
- Add development helpers (run, dev)
- Add installation target
- Add help documentation"

git push -u origin feature/00-03-makefile
gh pr create --base develop --fill
gh pr merge --squash --delete-branch
git checkout develop && git pull origin develop
```

---

## Commit 0.4: Agregar README del proyecto

**Rama:** `feature/00-04-readme`

```bash
git checkout -b feature/00-04-readme
```

**Mensaje de commit:**
```
docs: add project README

- Add project overview and features
- Add installation instructions
- Add usage examples
- Add development setup guide
- Add contributing guidelines
```

**Archivos a crear:**

### 1. `README.md` (en raiz del proyecto)

```markdown
# AI-Toolkit

Suite de herramientas de IA para automatizar code review.

## Componentes

- **GoReview CLI**: Herramienta de linea de comandos para code review con IA
- **GitHub App**: Integracion con GitHub para reviews automaticos en PRs
- **Rules System**: Sistema de reglas YAML configurables

## Caracteristicas

- Review de codigo con IA (Ollama, OpenAI)
- Generacion de mensajes de commit
- Generacion de documentacion/changelog
- Cache LRU para optimizar requests
- Multiples formatos de output (Markdown, JSON, SARIF)
- Procesamiento paralelo de archivos
- Integracion con GitHub PRs

## Instalacion

### Prerequisitos

- Go 1.24+
- Node.js 20+ (para GitHub App)
- Docker (opcional)
- Ollama (para LLM local)

### Desde binarios

```bash
# Descargar ultima version
curl -sSL https://github.com/JNZader/goreview/releases/latest/download/goreview-linux-amd64 -o goreview
chmod +x goreview
sudo mv goreview /usr/local/bin/
```

### Desde codigo fuente

```bash
git clone https://github.com/JNZader/goreview.git
cd goreview/goreview
make build
./build/goreview version
```

## Uso

### Review de cambios staged

```bash
goreview review
```

### Review de un commit especifico

```bash
goreview review --commit HEAD~1
```

### Review comparando con rama

```bash
goreview review --base main
```

### Generar mensaje de commit

```bash
goreview commit
```

### Generar documentacion

```bash
goreview doc --output CHANGELOG.md
```

## Configuracion

Crear archivo `.goreview.yaml` en la raiz del proyecto:

```yaml
provider:
  name: ollama
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434
  timeout: 5m

review:
  mode: staged
  min_severity: warning
  max_concurrency: 5

output:
  format: markdown
  color: true

cache:
  enabled: true
  ttl: 24h
```

## Desarrollo

```bash
# Setup del entorno
make setup

# Compilar
make build

# Ejecutar tests
make test

# Linting
make lint

# Modo desarrollo con hot reload
make dev
```

## GitHub App

Para configurar la GitHub App:

1. Crear GitHub App en Settings > Developer settings
2. Configurar webhook URL
3. Copiar `.env.example` a `.env` y agregar credenciales
4. Ejecutar `docker compose up`

Ver [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) para mas detalles.

## Arquitectura

```
goreview/ (repo)
├── goreview/           # CLI en Go
│   ├── cmd/            # Entry points
│   └── internal/       # Paquetes internos
├── integrations/       # Integraciones
│   └── github-app/     # GitHub App en Node.js
├── goreview-rules/     # Reglas YAML
└── docs/               # Documentacion
```

## Contribuir

1. Fork el repositorio
2. Crear rama feature (`git checkout -b feature/amazing-feature`)
3. Commit cambios (`git commit -m 'feat: add amazing feature'`)
4. Push a la rama (`git push origin feature/amazing-feature`)
5. Abrir Pull Request

## Licencia

MIT License - ver [LICENSE](LICENSE) para detalles.
```

### 2. `goreview/README.md`

```markdown
# GoReview CLI

Herramienta de linea de comandos para code review con IA.

## Compilar

```bash
make build
```

## Ejecutar

```bash
./build/goreview review
```

## Tests

```bash
make test
```

## Comandos

- `review`: Analiza codigo y genera feedback
- `commit`: Genera mensaje de commit con IA
- `doc`: Genera documentacion de cambios
- `config`: Muestra configuracion actual
- `init-project`: Inicializa proyecto con config
- `version`: Muestra version

## Flags globales

- `--config, -c`: Archivo de configuracion
- `--verbose, -v`: Output detallado
- `--quiet, -q`: Solo errores
```

**Verificacion:**
```bash
# Ver READMEs
cat README.md
cat goreview/README.md
```

**Finalizar commit:**
```bash
git add .
git commit -m "docs: add project README

- Add project overview and features
- Add installation instructions
- Add usage examples
- Add development setup guide
- Add contributing guidelines"

git push -u origin feature/00-04-readme
gh pr create --base develop --fill
gh pr merge --squash --delete-branch
git checkout develop && git pull origin develop
```

---

## Resumen de la Iteracion 00

### Commits y Ramas (GitFlow):

| # | Rama | Commit |
|---|------|--------|
| 0.1 | `feature/00-01-project-structure` | `chore: create initial project structure and initialize Go module` |
| 0.2 | `feature/00-02-gitignore-editor` | `chore: add gitignore and editor configuration` |
| 0.3 | `feature/00-03-makefile` | `chore: add Makefile with build targets` |
| 0.4 | `feature/00-04-readme` | `docs: add project README` |

### Archivos creados:
```
goreview/ (repo)
├── .gitignore
├── .editorconfig
├── .vscode/
│   ├── settings.json
│   └── extensions.json
├── Makefile
├── README.md
├── goreview/
│   ├── go.mod
│   ├── Makefile
│   ├── README.md
│   ├── cmd/goreview/commands/
│   ├── internal/
│   │   ├── config/
│   │   ├── git/
│   │   ├── providers/
│   │   ├── review/
│   │   ├── cache/
│   │   ├── report/
│   │   └── rules/
│   └── build/
├── integrations/github-app/
│   └── src/services/
├── goreview-rules/
│   ├── rules/
│   ├── languages/
│   └── presets/
├── docs/
├── scripts/
└── .github/workflows/
```

### Verificacion final:
```bash
# Desde la raiz del proyecto (goreview clonado)

# Verificar estructura
find . -type f -name "*.go" -o -name "*.mod" -o -name "Makefile" -o -name "*.md" | head -20

# Verificar go.mod
cd goreview
go mod verify
cat go.mod

# Verificar Makefile
make help
```

---

## Siguiente Iteracion

Continua con: **[01-CLI-BASICO.md](01-CLI-BASICO.md)**

En la siguiente iteracion crearemos:
- Entry point `main.go`
- Comando root con Cobra
- Comando version
- Tests unitarios
