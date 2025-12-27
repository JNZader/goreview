# GoReview CLI

Herramienta de linea de comandos para code review con IA. Analiza cambios de codigo, identifica problemas potenciales y proporciona feedback accionable usando modelos de lenguaje.

## Code Quality

| Metric | Status |
|--------|--------|
| Bugs | 0 |
| Vulnerabilities | 0 |
| Code Smells | 0 |
| Security Hotspots | 0 |

> Cognitive complexity < 15 en todas las funciones. Analizado con SonarQube y golangci-lint.

## Caracteristicas

### Core
- **Review de codigo con IA**: Analiza diffs y detecta bugs, vulnerabilidades, problemas de rendimiento
- **Generacion de commits**: Mensajes siguiendo Conventional Commits
- **Generacion de changelog**: Changelog automatico desde commits
- **Multiples proveedores de IA**: Ollama (local), OpenAI, Gemini, Groq, Mistral
- **Sistema de cache**: Evita re-analizar codigo sin cambios
- **Reportes multiples**: Markdown, JSON, SARIF
- **Sistema de reglas**: Presets minimal, standard, strict

### Modos de Revision (`--mode`)
| Modo | Enfoque |
|------|---------|
| `security` | Vulnerabilidades OWASP, secrets, injections |
| `perf` | N+1 queries, complejidad, memory leaks |
| `clean` | SOLID, DRY, naming, code smells |
| `docs` | Comentarios faltantes, JSDoc/GoDoc |
| `tests` | Cobertura, edge cases, mocking |

### Personalidades (`--personality`)
| Personalidad | Estilo |
|--------------|--------|
| `senior` | Mentoring, explica el "por que" |
| `strict` | Directo, sin rodeos, exigente |
| `friendly` | Sugerencias amables, positivo |
| `security-expert` | Paranoia saludable, peor caso |

### Avanzadas
- **Root Cause Tracing**: `--trace` rastrea hasta la causa raiz
- **Workflow TDD**: `--require-tests` bloquea sin tests
- **Historial de Reviews**: SQLite + FTS5 para busqueda full-text
- **Auto-fix**: Aplica correcciones automaticamente
- **RAG**: Integra guias de estilo y documentacion externa
- **AST Parsing**: Contexto multi-lenguaje (Go, JS/TS, Python, Java, Rust)

### Integraciones
- **Claude Code**: Plugin completo con MCP server, agentes y hooks
- **Obsidian**: Exportar reviews a vault de Obsidian
- **SARIF**: Integracion con IDEs via Static Analysis Results Format

## Integracion con Claude Code

GoReview se integra nativamente con [Claude Code](https://claude.com/product/claude-code) para proporcionar code review automatizado dentro de tu flujo de trabajo con IA.

### Instalacion Rapida

```bash
# Agregar como servidor MCP
claude mcp add --transport stdio goreview -- goreview mcp-serve
```

### Caracteristicas

| Caracteristica | Descripcion |
|----------------|-------------|
| **MCP Server** | 7 herramientas disponibles para Claude |
| **Plugin** | Slash commands, agentes, skills, hooks |
| **Background Watcher** | Monitoreo continuo de cambios |
| **Checkpoint Sync** | Reviews sincronizados con checkpoints |

### Herramientas MCP Disponibles

```
goreview_review    - Analizar codigo
goreview_commit    - Generar commits
goreview_fix       - Auto-corregir issues
goreview_search    - Buscar en historial
goreview_stats     - Ver estadisticas
goreview_changelog - Generar changelog
goreview_doc       - Generar documentacion
```

### Uso en Claude Code

```
Revisa mis cambios staged con enfoque en seguridad
Genera un mensaje de commit para estos cambios
Busca issues de SQL injection en el historial
```

### Plugin para Claude Code

El plugin incluye:
- **6 Slash Commands**: `/review`, `/commit-ai`, `/fix-issues`, `/changelog`, `/stats`, `/security-scan`
- **5 Subagents**: security-reviewer, perf-reviewer, test-reviewer, fix-agent, goreview-watcher
- **2 Skills**: goreview-workflow, commit-standards
- **Hooks**: Auto-review, project health, checkpoint sync

```bash
# Instalar plugin
cd claude-code-plugin
/plugin install --local .
```

### Documentacion

- [MCP Server](docs/MCP_SERVER.md) - Guia completa del servidor MCP
- [Plugin Guide](docs/PLUGIN_GUIDE.md) - Guia del plugin para Claude Code
- [Background Watcher](docs/BACKGROUND_WATCHER.md) - Monitoreo continuo
- [Checkpoint Sync](docs/CHECKPOINT_SYNC.md) - Sincronizacion con checkpoints
- [Integracion Completa](docs/CLAUDE_CODE_INTEGRATION.md) - Guia de integracion

## Instalacion

### Desde codigo fuente

```bash
# Clonar repositorio
git clone https://github.com/JNZader/goreview.git
cd goreview

# Compilar
make build

# El binario estara en ./bin/goreview
```

### Con Go

```bash
go install github.com/JNZader/goreview/goreview/cmd/goreview@latest
```

## Inicio rapido

```bash
# Inicializar configuracion en tu proyecto
goreview init

# Revisar cambios staged
goreview review --staged

# Generar mensaje de commit
goreview commit
```

## Comandos

### `review` - Analizar codigo

Analiza cambios de codigo y genera feedback con issues categorizados por severidad.

```bash
# Revisar cambios staged
goreview review --staged

# Revisar un commit especifico
goreview review --commit abc123

# Comparar con una rama
goreview review --branch main

# Revisar archivos especificos
goreview review file1.go file2.go

# Exportar a JSON
goreview review --staged --format json -o report.json

# Exportar a SARIF (para IDEs)
goreview review --staged --format sarif -o report.sarif

# Review enfocado en seguridad
goreview review --staged --mode=security

# Multiples modos combinados
goreview review --staged --mode=security,perf

# Con personalidad de mentor
goreview review --staged --personality=senior

# Con verificacion de tests (TDD)
goreview review --staged --require-tests --min-coverage=80

# Con root cause tracing
goreview review --staged --trace
```

**Flags:**

| Flag | Descripcion |
|------|-------------|
| `--staged` | Revisar cambios en staging |
| `--commit <sha>` | Revisar commit especifico |
| `--branch <branch>` | Comparar con rama |
| `--format` | Formato de salida: markdown, json, sarif |
| `--output, -o` | Escribir a archivo |
| `--include` | Patrones de archivos a incluir |
| `--exclude` | Patrones de archivos a excluir |
| `--provider` | Proveedor de IA a usar |
| `--model` | Modelo a usar |
| `--concurrency` | Reviews paralelos (0=auto) |
| `--no-cache` | Desactivar cache |
| `--preset` | Preset de reglas: minimal, standard, strict |
| `--mode` | Modo de revision: security, perf, clean, docs, tests |
| `--personality` | Estilo de reviewer: senior, strict, friendly, security-expert |
| `--require-tests` | Fallar si no hay tests correspondientes |
| `--min-coverage` | Cobertura minima requerida (0=desactivado) |
| `--trace` | Activar root cause tracing |

### `commit` - Generar mensaje de commit

Genera mensajes de commit siguiendo el formato Conventional Commits.

```bash
# Generar y mostrar mensaje
goreview commit

# Generar y ejecutar commit
goreview commit --execute

# Enmendar ultimo commit
goreview commit --amend

# Forzar tipo y scope
goreview commit --type feat --scope api

# Marcar como breaking change
goreview commit --breaking
```

**Flags:**

| Flag | Descripcion |
|------|-------------|
| `--execute, -e` | Ejecutar git commit |
| `--amend` | Enmendar ultimo commit |
| `--type, -t` | Forzar tipo de commit |
| `--scope, -s` | Forzar scope |
| `--breaking` | Marcar como breaking change |
| `--body, -b` | Cuerpo adicional |
| `--dry-run` | Mostrar sin ejecutar |

### `doc` - Generar documentacion

Genera documentacion automatica de los cambios.

```bash
# Documentar cambios staged
goreview doc --staged

# Generar entrada de changelog
goreview doc --type changelog

# Generar documentacion de API
goreview doc --type api

# Escribir a archivo
goreview doc --staged -o CHANGELOG.md --append
```

**Flags:**

| Flag | Descripcion |
|------|-------------|
| `--staged` | Documentar cambios staged |
| `--commit` | Documentar commit especifico |
| `--type` | Tipo: changes, changelog, api, readme |
| `--style` | Estilo: markdown, jsdoc, godoc |
| `--output, -o` | Escribir a archivo |
| `--append` | Agregar al final del archivo |
| `--prepend` | Agregar al inicio del archivo |

### `init` - Inicializar proyecto

Configura GoReview en tu proyecto con deteccion automatica de lenguaje y framework.

```bash
# Wizard interactivo
goreview init

# Usar valores por defecto
goreview init --yes

# Especificar proveedor
goreview init --provider openai --model gpt-4
```

### `config` - Ver configuracion

```bash
# Mostrar configuracion actual
goreview config show

# Mostrar como JSON
goreview config show --json
```

### `version` - Mostrar version

```bash
goreview version
```

### `fix` - Auto-corregir issues

Aplica correcciones automaticas a los issues detectados.

```bash
# Corregir issues en cambios staged
goreview fix --staged

# Modo dry-run (ver sin aplicar)
goreview fix --staged --dry-run

# Corregir archivo especifico
goreview fix file.go
```

### `history` - Historial de reviews

Gestiona el historial de reviews realizados.

```bash
# Mostrar historial reciente
goreview history

# Buscar en historial
goreview history search "error handling"

# Limpiar historial antiguo
goreview history prune --days 30
```

### `recall` - Recordar contexto

Recupera informacion de reviews anteriores para contexto.

```bash
# Buscar reviews anteriores
goreview recall "authentication"

# Ver estadisticas
goreview recall --stats
```

### `stats` - Estadisticas

Muestra estadisticas del proyecto y reviews.

```bash
# Estadisticas generales
goreview stats

# Estadisticas por archivo
goreview stats --by-file

# Estadisticas por severidad
goreview stats --by-severity
```

### `changelog` - Generar changelog

Genera changelog automatico basado en commits.

```bash
# Generar changelog desde ultimo tag
goreview changelog

# Generar desde version especifica
goreview changelog --from v1.0.0

# Escribir a archivo
goreview changelog -o CHANGELOG.md
```

## Flags globales

| Flag | Descripcion |
|------|-------------|
| `--config, -c` | Ruta al archivo de configuracion |
| `--verbose, -v` | Output detallado |
| `--quiet, -q` | Solo mostrar errores |

## Configuracion

GoReview usa un archivo `.goreview.yaml` en la raiz del proyecto.

```yaml
version: "1.0"

provider:
  name: ollama                    # ollama, openai, gemini, groq, mistral, auto
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434
  timeout: 30s
  max_tokens: 4096
  temperature: 0.1

git:
  base_branch: main
  ignore_patterns:
    - "vendor/**"
    - "node_modules/**"
    - "*.min.js"

review:
  max_concurrency: 5              # 0 = auto (CPUs * 2, max 10)
  min_severity: warning           # info, warning, error, critical
  timeout: 5m

output:
  format: markdown
  include_code: true
  color: true

cache:
  enabled: true
  ttl: 24h
  max_entries: 100

rules:
  preset: standard                # minimal, standard, strict
```

### Variables de entorno

Todas las configuraciones pueden sobrescribirse con variables de entorno usando el prefijo `GOREVIEW_`:

```bash
export GOREVIEW_PROVIDER_NAME=openai
export GOREVIEW_PROVIDER_APIKEY=sk-...
export GOREVIEW_PROVIDER_MODEL=gpt-4
```

## Proveedores de IA

### Ollama (Local - Recomendado)

Ejecucion local, gratuita y privada.

```bash
# Instalar Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Descargar modelo
ollama pull qwen2.5-coder:14b

# Iniciar servidor
ollama serve
```

### OpenAI

```yaml
provider:
  name: openai
  model: gpt-4
  api_key: ${OPENAI_API_KEY}
```

### Google Gemini

```yaml
provider:
  name: gemini
  model: gemini-pro
  api_key: ${GEMINI_API_KEY}
```

### Groq

```yaml
provider:
  name: groq
  model: llama-3.1-70b-versatile
  api_key: ${GROQ_API_KEY}
```

### Mistral

```yaml
provider:
  name: mistral
  model: codestral-latest
  api_key: ${MISTRAL_API_KEY}
```

### Auto-deteccion

Con `name: auto`, GoReview detecta automaticamente el proveedor disponible:
1. Intenta Ollama en localhost:11434
2. Usa proveedores cloud segun API keys disponibles

## Niveles de severidad

| Nivel | Descripcion |
|-------|-------------|
| `critical` | Vulnerabilidades de seguridad, bugs criticos |
| `error` | Bugs, errores logicos |
| `warning` | Problemas de rendimiento, code smells |
| `info` | Sugerencias de mejora, estilo |

## Presets de reglas

| Preset | Descripcion |
|--------|-------------|
| `minimal` | Solo reglas criticas de seguridad |
| `standard` | Balance entre cobertura y ruido (recomendado) |
| `strict` | Maxima cobertura de calidad |

## Formatos de salida

### Markdown (default)

Formato legible para humanos con snippets de codigo.

### JSON

Formato estructurado para procesamiento programatico.

```json
{
  "summary": {
    "total_issues": 3,
    "by_severity": {"warning": 2, "error": 1}
  },
  "files": [...]
}
```

### SARIF

Static Analysis Results Interchange Format para integracion con IDEs y herramientas de CI.

## Desarrollo

### Requisitos

- Go 1.23+
- Make

### Comandos de desarrollo

```bash
# Compilar
make build

# Ejecutar tests
make test

# Tests con race detector
make test-race

# Lint
make lint

# Formatear codigo
make fmt

# Compilar para todas las plataformas
make build-all

# Limpiar artefactos
make clean
```

### Estructura del proyecto

```
goreview/
├── cmd/goreview/           # Punto de entrada y comandos
│   └── commands/           # Implementacion de comandos CLI
├── internal/
│   ├── ast/                # AST parsing multi-lenguaje
│   ├── cache/              # Sistema de cache LRU
│   ├── config/             # Carga y validacion de config
│   ├── git/                # Integracion con Git
│   ├── history/            # Historial y recall de reviews
│   ├── knowledge/          # Base de conocimiento
│   ├── logger/             # Logger con secret masking
│   ├── memory/             # Sistema de memoria cognitiva
│   ├── metrics/            # Metricas de rendimiento
│   ├── profiler/           # Profiling CPU/memoria
│   ├── providers/          # Proveedores de IA
│   ├── rag/                # RAG para style guides
│   ├── report/             # Generadores de reportes
│   ├── review/             # Motor de review
│   ├── rules/              # Sistema de reglas
│   ├── tokenizer/          # Token budgeting y chunking
│   └── worker/             # Pool de workers concurrentes
├── .golangci.yml           # Configuracion de linter
├── Makefile                # Comandos de build
└── Dockerfile              # Build de contenedor
```

## Docker

```bash
# Construir imagen
docker build -t goreview .

# Ejecutar
docker run -v $(pwd):/app goreview review --staged
```

## CI/CD

GoReview incluye workflows de GitHub Actions para:

- **Lint**: Validacion con golangci-lint
- **Test**: Tests con race detector y coverage
- **Build**: Compilacion multi-plataforma (Linux, macOS, Windows)

## Licencia

MIT License - Ver [LICENSE](LICENSE) para detalles.

## Contribuir

1. Fork el repositorio
2. Crear rama feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -m 'feat: agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request
