# GoReview: Guia Completa de Uso

> De principiante a experto - Todo lo que necesitas saber para dominar GoReview

## Tabla de Contenidos

1. [Introduccion](#introduccion)
2. [Instalacion](#instalacion)
3. [Configuracion Inicial](#configuracion-inicial)
4. [Uso Basico](#uso-basico)
5. [Modos de Review](#modos-de-review)
6. [Personalidades del Reviewer](#personalidades-del-reviewer)
7. [Formatos de Salida](#formatos-de-salida)
8. [Sistema de Cache](#sistema-de-cache)
9. [Historial y Busqueda](#historial-y-busqueda)
10. [Generacion de Commits](#generacion-de-commits)
11. [Generacion de Changelog](#generacion-de-changelog)
12. [Exportacion a Obsidian](#exportacion-a-obsidian)
13. [Integracion con Claude Code](#integracion-con-claude-code)
14. [Servidor MCP](#servidor-mcp)
15. [Configuracion Avanzada](#configuracion-avanzada)
16. [Proveedores de IA](#proveedores-de-ia)
17. [Troubleshooting](#troubleshooting)

---

## Introduccion

GoReview es una herramienta de revision de codigo potenciada por IA que analiza tus cambios y proporciona feedback detallado sobre:

- **Seguridad**: Vulnerabilidades, inyecciones, exposicion de datos
- **Rendimiento**: Cuellos de botella, optimizaciones posibles
- **Codigo Limpio**: Mejores practicas, legibilidad, mantenibilidad
- **Documentacion**: Comentarios faltantes, documentacion de API
- **Tests**: Cobertura, casos edge, mejoras en tests

### Caracteristicas Principales

- Review de cambios staged, commits, branches o archivos especificos
- Multiples personalidades de reviewer (senior, strict, friendly, security-expert)
- Cache inteligente para evitar re-analisis
- Historial persistente con busqueda full-text
- Generacion automatica de mensajes de commit (Conventional Commits)
- Generacion de changelog
- Exportacion a Obsidian
- Integracion nativa con Claude Code via MCP
- Soporte para multiples proveedores de IA (Ollama, OpenAI, Anthropic, etc.)

---

## Instalacion

### Requisitos Previos

- Go 1.21 o superior
- Git
- Un proveedor de IA (Ollama recomendado para uso local)

### Opcion 1: Desde Codigo Fuente

```bash
# Clonar el repositorio
git clone https://github.com/JNZader/goreview.git
cd goreview/goreview

# Compilar e instalar
go install ./cmd/goreview

# Verificar instalacion
goreview --version
```

### Opcion 2: Go Install Directo

```bash
go install github.com/JNZader/goreview/goreview/cmd/goreview@latest
```

### Instalar Ollama (Recomendado)

GoReview funciona mejor con Ollama para uso local:

```bash
# Linux/macOS
curl -fsSL https://ollama.com/install.sh | sh

# Windows - Descargar desde https://ollama.com/download

# Descargar modelo recomendado
ollama pull qwen2.5-coder:14b

# Verificar que Ollama esta corriendo
ollama list
```

### Verificar Instalacion

```bash
# Verificar GoReview
goreview --help

# Verificar conexion con Ollama
goreview review --staged
```

---

## Configuracion Inicial

### Crear Archivo de Configuracion

GoReview busca configuracion en este orden:
1. `.goreview.yaml` en el directorio actual
2. `.goreview.yaml` en el directorio home
3. Variables de entorno

```bash
# Crear configuracion en tu proyecto
touch .goreview.yaml
```

### Configuracion Minima

```yaml
# .goreview.yaml
provider:
  name: ollama
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434
```

### Configuracion Recomendada para Comenzar

```yaml
# .goreview.yaml
provider:
  name: ollama
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434
  timeout: 5m
  max_tokens: 4096
  temperature: 0.1

git:
  base_branch: main

review:
  mode: staged
  min_severity: warning
  max_issues: 50

output:
  format: markdown
  include_code: true
  color: true

cache:
  enabled: true
  ttl: 24h
```

### Variables de Entorno

Puedes configurar GoReview via variables de entorno:

```bash
# Proveedor
export GOREVIEW_PROVIDER_NAME=ollama
export GOREVIEW_PROVIDER_MODEL=qwen2.5-coder:14b
export GOREVIEW_PROVIDER_BASE_URL=http://localhost:11434

# Para OpenAI
export GOREVIEW_PROVIDER_NAME=openai
export GOREVIEW_PROVIDER_API_KEY=sk-...

# Para Anthropic
export GOREVIEW_PROVIDER_NAME=anthropic
export GOREVIEW_PROVIDER_API_KEY=sk-ant-...
```

---

## Uso Basico

### Tu Primer Review

```bash
# 1. Hacer algunos cambios en tu codigo
echo "func bad() { panic(\"oops\") }" >> main.go

# 2. Agregar al staging
git add main.go

# 3. Ejecutar review
goreview review --staged
```

### Salida Tipica

```
## Code Review Results

### main.go

#### :red_circle: [error-handling] Uso de panic en lugar de error handling

**Location:** Line 1
**Severity:** Error

**Issue:** El uso de `panic` para manejo de errores no es idiomatico en Go.

**Suggestion:** Retornar un error en lugar de hacer panic.

**Fixed Code:**
func bad() error {
    return errors.New("oops")
}

---

**Summary:** 1 issue found | Score: 60/100
```

### Comandos Esenciales

```bash
# Review de cambios staged (mas comun)
goreview review --staged

# Review del ultimo commit
goreview review --commit HEAD

# Review de un commit especifico
goreview review --commit abc1234

# Review de una rama completa
goreview review --branch feature/my-feature

# Review de archivos especificos
goreview review path/to/file.go another/file.go
```

---

## Modos de Review

GoReview tiene diferentes modos de analisis que puedes combinar:

### Modo Security

Enfocado en vulnerabilidades de seguridad:

```bash
goreview review --staged --mode security
```

Detecta:
- Inyeccion SQL
- XSS (Cross-Site Scripting)
- Command Injection
- Path Traversal
- Credenciales hardcodeadas
- Uso inseguro de crypto
- SSRF, CSRF, XXE

### Modo Performance

Enfocado en optimizaciones de rendimiento:

```bash
goreview review --staged --mode perf
```

Detecta:
- Allocaciones innecesarias
- Loops ineficientes
- Queries N+1
- Memory leaks
- Goroutine leaks
- Operaciones bloqueantes

### Modo Clean Code

Enfocado en calidad y mantenibilidad:

```bash
goreview review --staged --mode clean
```

Detecta:
- Funciones muy largas
- Complejidad ciclomatica alta
- Nombres poco descriptivos
- Codigo duplicado
- Violaciones de SOLID
- Code smells

### Modo Docs

Enfocado en documentacion:

```bash
goreview review --staged --mode docs
```

Detecta:
- Funciones publicas sin documentar
- Parametros sin explicacion
- Tipos exportados sin docs
- README desactualizado
- Ejemplos faltantes

### Modo Tests

Enfocado en testing:

```bash
goreview review --staged --mode tests
```

Detecta:
- Falta de tests
- Tests sin assertions
- Casos edge no cubiertos
- Tests fragiles
- Mocks incorrectos

### Combinar Modos

Puedes combinar varios modos:

```bash
# Security + Performance
goreview review --staged --mode security,perf

# Todos los modos
goreview review --staged --mode security,perf,clean,docs,tests
```

---

## Personalidades del Reviewer

Cada personalidad tiene un enfoque diferente:

### Senior (Default)

Equilibrado, enfocado en mejores practicas:

```bash
goreview review --staged --personality senior
```

Caracteristicas:
- Feedback constructivo
- Enfocado en el "por que"
- Sugiere alternativas
- Considera el contexto

### Strict

Muy exigente, no deja pasar nada:

```bash
goreview review --staged --personality strict
```

Caracteristicas:
- Reporta todo
- Severidades mas altas
- Ideal para codigo critico
- Sin excepciones

### Friendly

Amigable, ideal para juniors:

```bash
goreview review --staged --personality friendly
```

Caracteristicas:
- Tono educativo
- Explica conceptos
- Reconoce lo bueno
- Mas paciente

### Security Expert

Especialista en seguridad:

```bash
goreview review --staged --personality security-expert
```

Caracteristicas:
- Paranoia saludable
- Ve amenazas potenciales
- Referencias a CVEs
- OWASP Top 10

### Ejemplo Comparativo

El mismo codigo revisado con diferentes personalidades:

```go
func GetUser(id string) *User {
    query := "SELECT * FROM users WHERE id = " + id
    // ...
}
```

**Senior:**
> "Esta concatenacion de SQL es vulnerable a inyeccion. Considera usar prepared statements."

**Strict:**
> "CRITICO: SQL Injection (CWE-89). Este codigo NO debe llegar a produccion bajo ninguna circunstancia."

**Friendly:**
> "Hey! Veo que estas construyendo una query SQL. Hay una forma mas segura de hacerlo usando prepared statements. Te explico por que es importante..."

**Security Expert:**
> "SQL Injection detectado. Vector de ataque: manipulacion del parametro 'id'. Impacto: acceso no autorizado a datos, posible RCE via SQLi avanzado. Remediacion: usar db.Query con placeholders."

---

## Formatos de Salida

### Markdown (Default)

```bash
goreview review --staged --format markdown
```

Ideal para: Lectura en terminal, copiar a PRs

### JSON

```bash
goreview review --staged --format json
```

Ideal para: Integraciones, CI/CD, procesamiento automatico

```json
{
  "files": [
    {
      "file": "main.go",
      "response": {
        "issues": [
          {
            "type": "security",
            "severity": "critical",
            "message": "SQL Injection vulnerability",
            "location": {"start_line": 10, "end_line": 10},
            "suggestion": "Use prepared statements"
          }
        ],
        "score": 40
      }
    }
  ],
  "total_issues": 1,
  "duration": "2.5s"
}
```

### GitHub

```bash
goreview review --staged --format github
```

Ideal para: GitHub Actions, anotaciones en PRs

### SARIF

```bash
goreview review --staged --format sarif
```

Ideal para: Integracion con GitHub Security, herramientas de analisis

### Texto Plano

```bash
goreview review --staged --format text
```

Ideal para: Logs, procesamiento con grep/awk

### Opciones Adicionales

```bash
# Sin colores (para pipes)
goreview review --staged --no-color

# Modo silencioso (solo errores)
goreview review --staged --quiet

# Modo verbose (mas detalles)
goreview review --staged --verbose

# Incluir codigo en el output
goreview review --staged --include-code
```

---

## Sistema de Cache

GoReview cachea resultados para evitar re-analizar codigo sin cambios.

### Como Funciona

1. Calcula hash del contenido + configuracion
2. Si existe en cache y no expiro, retorna resultado cacheado
3. Si no, ejecuta review y guarda en cache

### Configuracion

```yaml
cache:
  enabled: true
  dir: ~/.cache/goreview
  ttl: 24h
  max_size_mb: 100
  max_entries: 1000
```

### Comandos de Cache

```bash
# Ver estadisticas del cache
goreview cache stats

# Limpiar cache completo
goreview cache clear

# Limpiar entradas expiradas
goreview cache gc
```

### Forzar Re-analisis

```bash
# Ignorar cache para este review
goreview review --staged --no-cache
```

### Ejemplo de Stats

```
Cache Statistics:
  Location: /home/user/.cache/goreview
  Entries: 142
  Size: 15.3 MB
  Hit Rate: 78.5%
  Oldest Entry: 2024-01-10
  Newest Entry: 2024-01-15
```

---

## Historial y Busqueda

GoReview mantiene un historial de todos los reviews.

### Ver Historial

```bash
# Ultimos 10 reviews
goreview history

# Ultimos 20 reviews
goreview history --limit 20
```

### Buscar en Historial

```bash
# Buscar por texto
goreview search "SQL injection"

# Buscar por severidad
goreview search --severity critical

# Buscar por archivo
goreview search --file "auth/"

# Buscar por tipo de issue
goreview search --type security

# Combinar filtros
goreview search "password" --severity error --file "*.go"
```

### Formato de Busqueda

```bash
# Resultados en JSON
goreview search "memory leak" --format json

# Limitar resultados
goreview search "error" --limit 5
```

### Ver Review Especifico

```bash
# Ver detalles de un review por ID
goreview history show abc123
```

---

## Generacion de Commits

GoReview puede generar mensajes de commit siguiendo Conventional Commits.

### Uso Basico

```bash
# Analizar staged changes y generar mensaje
goreview commit
```

### Ejemplo de Output

```
Analyzing staged changes...

Suggested commit message:

feat(auth): implement JWT token validation

- Add token parsing and validation logic
- Implement refresh token rotation
- Add unit tests for token service

Apply this commit message? [y/N]
```

### Opciones

```bash
# Forzar tipo de commit
goreview commit --type fix

# Forzar scope
goreview commit --scope database

# Marcar como breaking change
goreview commit --breaking

# Solo mostrar mensaje, no commitear
goreview commit --dry-run

# Aplicar automaticamente
goreview commit --apply
```

### Tipos de Commit Soportados

| Tipo | Descripcion |
|------|-------------|
| feat | Nueva funcionalidad |
| fix | Correccion de bug |
| docs | Documentacion |
| style | Formato, sin cambios de codigo |
| refactor | Refactorizacion |
| test | Tests |
| chore | Mantenimiento |
| perf | Mejora de rendimiento |
| ci | Cambios de CI/CD |
| build | Sistema de build |

---

## Generacion de Changelog

Genera changelogs automaticos desde los commits.

### Uso Basico

```bash
# Changelog desde el ultimo tag
goreview changelog

# Changelog entre versiones
goreview changelog --from v1.0.0 --to v1.1.0
```

### Formatos

```bash
# Markdown (default)
goreview changelog --format markdown

# JSON
goreview changelog --format json
```

### Ejemplo de Output

```markdown
# Changelog

## [Unreleased]

### Features
- **auth:** implement JWT token validation (#123)
- **api:** add rate limiting middleware (#124)

### Bug Fixes
- **database:** fix connection pool exhaustion (#125)
- **cache:** resolve memory leak in LRU cache (#126)

### Documentation
- **readme:** update installation instructions

### Breaking Changes
- **api:** remove deprecated v1 endpoints
```

### Guardar a Archivo

```bash
# Guardar a CHANGELOG.md
goreview changelog > CHANGELOG.md

# Actualizar CHANGELOG existente
goreview changelog --from v1.0.0 >> CHANGELOG.md
```

---

## Exportacion a Obsidian

Exporta reviews a tu vault de Obsidian con formato rico.

### Configuracion

```yaml
# .goreview.yaml
export:
  obsidian:
    enabled: true
    vault_path: ~/Documents/Obsidian/MyVault
    folder_name: GoReview
    include_tags: true
    include_callouts: true
    include_links: true
    link_to_previous: true
    custom_tags:
      - trabajo
      - backend
```

### Exportar Review

```bash
# Exportar automaticamente despues de review
goreview review --staged --export-obsidian

# Exportar con vault especifico
goreview review --staged --export-obsidian --obsidian-vault ~/MyVault
```

### Exportar desde JSON

```bash
# Guardar review como JSON
goreview review --staged --format json > report.json

# Exportar a Obsidian
goreview export obsidian --from report.json

# Con opciones
goreview export obsidian --from report.json --tags sprint-42 --no-callouts
```

### Estructura en Obsidian

```
MyVault/
  GoReview/
    mi-proyecto/
      review-001-2024-01-15.md
      review-002-2024-01-16.md
      review-003-2024-01-17.md
```

### Formato del Archivo

```markdown
---
date: 2024-01-15T10:30:00Z
project: mi-proyecto
branch: feature/auth
commit: abc123d
files_reviewed: 5
total_issues: 3
severity:
  critical: 0
  error: 1
  warning: 2
  info: 0
average_score: 75
tags:
  - goreview
  - code-review
  - security
---

# Code Review: mi-proyecto

#goreview #code-review #security

## Summary

| Metric | Value |
|--------|-------|
| Files Reviewed | 5 |
| Total Issues | 3 |
| Average Score | 75/100 |

## Issues

### src/auth/handler.go

> [!warning] [security] Potential timing attack in password comparison
> **Location:** Line 45
> **Suggestion:** Use constant-time comparison

---

## Related Reviews

- [[review-001-2024-01-14]]
- [[review-002-2024-01-13]]

---
*Generated by GoReview*
```

---

## Integracion con Claude Code

GoReview se integra nativamente con Claude Code de dos formas:

### Opcion 1: Plugin de Claude Code

El plugin proporciona comandos slash directos.

#### Instalacion del Plugin

```bash
# Desde el directorio del plugin
cd goreview/claude-code-plugin
npm install
npm run build

# Registrar con Claude Code
claude plugins add ./dist
```

#### Comandos Disponibles

```
/review              - Review de cambios staged
/review-commit       - Review del ultimo commit
/review-branch       - Review de una rama
/review-security     - Review enfocado en seguridad
/review-perf         - Review enfocado en rendimiento
/commit              - Generar mensaje de commit
/changelog           - Generar changelog
/goreview-stats      - Ver estadisticas
/goreview-search     - Buscar en historial
```

#### Ejemplos de Uso en Claude Code

```
> /review
Claude: Analizando cambios staged...

[Muestra resultados del review]

> /review-security
Claude: Ejecutando analisis de seguridad...

[Muestra vulnerabilidades encontradas]

> /commit
Claude: Generando mensaje de commit...

Sugerencia: feat(auth): add two-factor authentication support

Deseas aplicar este commit?
```

### Opcion 2: Servidor MCP

El servidor MCP permite que Claude Code use GoReview como herramienta.

#### Configurar MCP Server

Agregar a tu configuracion de Claude Code (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "goreview": {
      "command": "goreview",
      "args": ["mcp", "serve"],
      "env": {}
    }
  }
}
```

#### Herramientas MCP Disponibles

| Tool | Descripcion |
|------|-------------|
| `goreview_review` | Ejecutar code review |
| `goreview_commit` | Generar mensaje de commit |
| `goreview_fix` | Aplicar fixes automaticos |
| `goreview_search` | Buscar en historial |
| `goreview_stats` | Obtener estadisticas |
| `goreview_changelog` | Generar changelog |
| `goreview_doc` | Generar documentacion |

#### Ejemplo de Uso via MCP

Simplemente habla con Claude:

```
Tu: Revisa mis cambios staged y enfocate en seguridad

Claude: [Usa goreview_review internamente]

He revisado tus cambios. Encontre 2 issues:

1. **SQL Injection** en `user_repo.go:45`
   - El query usa concatenacion de strings
   - Recomendacion: Usar prepared statements

2. **Weak Hashing** en `auth.go:23`
   - MD5 no es seguro para passwords
   - Recomendacion: Usar bcrypt o argon2

Quieres que genere los fixes?
```

```
Tu: Si, genera los fixes

Claude: [Usa goreview_fix internamente]

He aplicado los siguientes cambios:
- user_repo.go: Cambiado a prepared statements
- auth.go: Reemplazado MD5 por bcrypt

Los cambios estan en staging. Quieres que genere el commit?
```

```
Tu: Genera el commit

Claude: [Usa goreview_commit internamente]

Commit sugerido:

fix(security): resolve SQL injection and weak hashing vulnerabilities

- Replace string concatenation with prepared statements in user_repo
- Replace MD5 with bcrypt for password hashing in auth

BREAKING CHANGE: Password hashes need to be regenerated

Aplico el commit?
```

---

## Servidor MCP

### Iniciar Servidor MCP

```bash
# Iniciar servidor (escucha en stdin/stdout)
goreview mcp serve

# Con logging
goreview mcp serve --verbose
```

### Protocolo

El servidor MCP usa JSON-RPC 2.0 sobre stdio.

#### Initialize

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
```

#### List Tools

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
```

#### Call Tool

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "goreview_review",
    "arguments": {
      "target": "staged",
      "mode": "security",
      "personality": "strict"
    }
  }
}
```

### Esquemas de Herramientas

#### goreview_review

```json
{
  "name": "goreview_review",
  "description": "Analyze code changes and identify issues",
  "inputSchema": {
    "type": "object",
    "properties": {
      "target": {
        "type": "string",
        "description": "What to review: 'staged', 'HEAD', commit SHA, or branch",
        "default": "staged"
      },
      "mode": {
        "type": "string",
        "enum": ["security", "perf", "clean", "docs", "tests"]
      },
      "personality": {
        "type": "string",
        "enum": ["senior", "strict", "friendly", "security-expert"]
      },
      "files": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Specific files to review"
      },
      "trace": {
        "type": "boolean",
        "description": "Enable root cause tracing",
        "default": false
      }
    }
  }
}
```

#### goreview_commit

```json
{
  "name": "goreview_commit",
  "description": "Generate commit message following Conventional Commits",
  "inputSchema": {
    "type": "object",
    "properties": {
      "type": {
        "type": "string",
        "enum": ["feat", "fix", "docs", "style", "refactor", "test", "chore"]
      },
      "scope": {"type": "string"},
      "breaking": {"type": "boolean", "default": false}
    }
  }
}
```

---

## Configuracion Avanzada

### Archivo de Configuracion Completo

```yaml
# .goreview.yaml - Configuracion completa

# Proveedor de IA
provider:
  name: ollama                    # ollama, openai, anthropic, azure, gemini
  model: qwen2.5-coder:14b       # Modelo a usar
  base_url: http://localhost:11434
  api_key: ""                    # Para proveedores cloud
  timeout: 5m
  max_tokens: 4096
  temperature: 0.1               # 0.0-1.0, menor = mas consistente
  rate_limit_rps: 0              # Requests por segundo (0 = sin limite)

# Configuracion de Git
git:
  repo_path: "."
  base_branch: main
  ignore_patterns:
    # Documentacion
    - "*.md"
    - "*.txt"
    - "LICENSE"
    # Generados
    - "*.pb.go"
    - "*_generated.go"
    - "*.gen.go"
    # Lock files
    - "go.sum"
    - "package-lock.json"
    - "yarn.lock"
    # Build
    - "dist/*"
    - "build/*"
    - "node_modules/*"
    - "vendor/*"

# Configuracion de Review
review:
  mode: staged                   # staged, commit, branch, files
  min_severity: warning          # info, warning, error, critical
  max_issues: 50                 # Limite de issues a reportar
  max_concurrency: 0             # 0 = auto-detectar CPUs
  personality: senior            # senior, strict, friendly, security-expert

# Configuracion de Output
output:
  format: markdown               # markdown, json, github, sarif, text
  include_code: true
  color: true
  verbose: false
  quiet: false

# Sistema de Cache
cache:
  enabled: true
  dir: ~/.cache/goreview
  ttl: 24h
  max_size_mb: 100
  max_entries: 1000

# Reglas personalizadas
rules:
  preset: standard               # minimal, standard, strict
  custom:
    - name: no-fmt-print
      pattern: "fmt\\.Print"
      message: "Use structured logging instead of fmt.Print"
      severity: warning
    - name: no-panic
      pattern: "panic\\("
      message: "Avoid using panic, return errors instead"
      severity: error

# Sistema de Memoria (experimental)
memory:
  enabled: false
  dir: ~/.cache/goreview/memory
  working:
    capacity: 100
    ttl: 15m
  session:
    max_sessions: 10
    session_ttl: 1h
  long_term:
    enabled: false
    max_size_mb: 500
    gc_interval: 5m
  hebbian:
    enabled: false
    learning_rate: 0.1
    decay_rate: 0.01
    min_strength: 0.1

# Exportacion
export:
  obsidian:
    enabled: false
    vault_path: ""               # Requerido si enabled
    folder_name: GoReview
    include_tags: true
    include_callouts: true
    include_links: true
    link_to_previous: true
    custom_tags: []
    template_file: ""            # Template personalizado
```

### Reglas Personalizadas

Puedes definir reglas custom con expresiones regulares:

```yaml
rules:
  preset: standard
  custom:
    # Detectar fmt.Print (usar logging estructurado)
    - name: no-fmt-print
      pattern: "fmt\\.(Print|Printf|Println)"
      message: "Use structured logging (log/slog) instead of fmt.Print"
      severity: warning

    # Detectar TODO sin owner
    - name: todo-needs-owner
      pattern: "TODO[^(]"
      message: "TODOs should have an owner: TODO(username)"
      severity: info

    # Detectar hardcoded localhost
    - name: no-localhost
      pattern: "localhost:\\d+"
      message: "Avoid hardcoded localhost, use configuration"
      severity: warning

    # Detectar panic fuera de init
    - name: no-panic
      pattern: "\\bpanic\\("
      message: "Avoid panic, return errors instead"
      severity: error
      exclude_patterns:
        - "_test.go"
        - "init()"
```

### Ignorar Archivos

Multiples formas de ignorar archivos:

```yaml
# En .goreview.yaml
git:
  ignore_patterns:
    - "*.generated.go"
    - "vendor/*"
    - "mocks/*"
```

```bash
# Via flag
goreview review --staged --ignore "*.pb.go" --ignore "vendor/*"
```

```go
// En el codigo (comentario magic)
//goreview:ignore
func legacyCode() {
    // Este codigo no sera revisado
}

//goreview:ignore-next-line
dangerousCall() // Esta linea no sera revisada
```

### Configuracion por Proyecto

Puedes tener configuraciones diferentes por proyecto:

```
mi-proyecto/
  .goreview.yaml          # Config del proyecto
  src/
    .goreview.yaml        # Config especifica para src/
  tests/
    .goreview.yaml        # Config mas relajada para tests
```

---

## Proveedores de IA

### Ollama (Recomendado para Local)

```yaml
provider:
  name: ollama
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434
```

Modelos recomendados:
- `qwen2.5-coder:14b` - Mejor balance calidad/velocidad
- `codellama:13b` - Bueno para codigo
- `deepseek-coder:6.7b` - Rapido, menor calidad
- `qwen2.5-coder:32b` - Mejor calidad, mas lento

### OpenAI

```yaml
provider:
  name: openai
  model: gpt-4-turbo
  api_key: sk-...
```

O via variable de entorno:
```bash
export OPENAI_API_KEY=sk-...
```

Modelos:
- `gpt-4-turbo` - Mejor calidad
- `gpt-4` - Muy bueno
- `gpt-3.5-turbo` - Rapido, menor calidad

### Anthropic (Claude)

```yaml
provider:
  name: anthropic
  model: claude-3-opus-20240229
  api_key: sk-ant-...
```

Modelos:
- `claude-3-opus-20240229` - Mejor calidad
- `claude-3-sonnet-20240229` - Buen balance
- `claude-3-haiku-20240307` - Rapido

### Azure OpenAI

```yaml
provider:
  name: azure
  model: gpt-4
  base_url: https://your-resource.openai.azure.com
  api_key: ...
  api_version: "2024-02-15-preview"
```

### Google Gemini

```yaml
provider:
  name: gemini
  model: gemini-pro
  api_key: ...
```

### OpenRouter

```yaml
provider:
  name: openrouter
  model: anthropic/claude-3-opus
  api_key: sk-or-...
  base_url: https://openrouter.ai/api/v1
```

---

## Troubleshooting

### Errores Comunes

#### "Connection refused" con Ollama

```bash
# Verificar que Ollama esta corriendo
ollama list

# Si no esta corriendo
ollama serve

# Verificar puerto
curl http://localhost:11434/api/tags
```

#### "Model not found"

```bash
# Descargar el modelo
ollama pull qwen2.5-coder:14b

# Listar modelos disponibles
ollama list
```

#### "Rate limit exceeded" (OpenAI/Anthropic)

```yaml
# Agregar rate limiting
provider:
  rate_limit_rps: 1  # 1 request por segundo
```

#### Review muy lento

```yaml
# Reducir max_tokens
provider:
  max_tokens: 2048

# Usar modelo mas rapido
provider:
  model: deepseek-coder:6.7b

# Aumentar concurrencia
review:
  max_concurrency: 4
```

#### Cache no funciona

```bash
# Verificar permisos del directorio
ls -la ~/.cache/goreview

# Limpiar cache corrupto
goreview cache clear

# Verificar stats
goreview cache stats
```

#### "No staged changes"

```bash
# Verificar que hay cambios staged
git status

# Agregar cambios
git add .

# O usar modo diferente
goreview review --commit HEAD
```

### Debug Mode

```bash
# Habilitar debug
export GOREVIEW_DEBUG=1
goreview review --staged --verbose

# Ver requests a la IA
export GOREVIEW_LOG_REQUESTS=1
```

### Logs

```bash
# Ver logs
tail -f ~/.cache/goreview/goreview.log

# Log level
export GOREVIEW_LOG_LEVEL=debug
```

### Reportar Bugs

Si encuentras un bug:

1. Ejecuta con `--verbose`
2. Captura el output completo
3. Incluye tu `.goreview.yaml` (sin API keys)
4. Abre un issue en GitHub

---

## Tips y Trucos

### Alias Utiles

```bash
# ~/.bashrc o ~/.zshrc

alias gr="goreview review --staged"
alias grs="goreview review --staged --mode security"
alias grp="goreview review --staged --mode perf"
alias grc="goreview commit"
alias grl="goreview changelog"
```

### Git Hooks

Pre-commit hook para review automatico:

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "Running GoReview..."
goreview review --staged --quiet --min-severity error

if [ $? -ne 0 ]; then
    echo "GoReview found critical issues. Fix them before committing."
    exit 1
fi
```

### CI/CD Integration

GitHub Actions:

```yaml
# .github/workflows/review.yml
name: Code Review

on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install GoReview
        run: go install github.com/JNZader/goreview/cmd/goreview@latest

      - name: Run Review
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          goreview review --branch ${{ github.head_ref }} \
            --format github \
            --mode security,perf
```

### Integracion con Editores

VSCode task:

```json
// .vscode/tasks.json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "GoReview: Review Staged",
      "type": "shell",
      "command": "goreview review --staged",
      "problemMatcher": []
    },
    {
      "label": "GoReview: Review Current File",
      "type": "shell",
      "command": "goreview review ${file}",
      "problemMatcher": []
    }
  ]
}
```

---

## Resumen de Comandos

| Comando | Descripcion |
|---------|-------------|
| `goreview review --staged` | Review cambios staged |
| `goreview review --commit HEAD` | Review ultimo commit |
| `goreview review --branch feature` | Review rama |
| `goreview review file.go` | Review archivo |
| `goreview commit` | Generar mensaje commit |
| `goreview changelog` | Generar changelog |
| `goreview search "query"` | Buscar en historial |
| `goreview history` | Ver historial |
| `goreview cache stats` | Stats del cache |
| `goreview cache clear` | Limpiar cache |
| `goreview export obsidian` | Exportar a Obsidian |
| `goreview mcp serve` | Iniciar servidor MCP |

---

## Recursos Adicionales

- **Repositorio**: https://github.com/JNZader/goreview
- **Issues**: https://github.com/JNZader/goreview/issues
- **Documentacion de Features**: [FEATURES.md](./FEATURES.md)

---

*Ultima actualizacion: Diciembre 2024*
