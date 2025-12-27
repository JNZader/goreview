# GoReview - Documentacion Completa de Caracteristicas

Esta documentacion detalla todas las caracteristicas de GoReview, explicando como funcionan internamente y como usarlas.

---

## Tabla de Contenidos

1. [Arquitectura General](#arquitectura-general)
2. [Comandos CLI](#comandos-cli)
3. [Sistema de Review](#sistema-de-review)
4. [Modos de Revision](#modos-de-revision)
5. [Personalidades del Reviewer](#personalidades-del-reviewer)
6. [Sistema de Proveedores de IA](#sistema-de-proveedores-de-ia)
7. [Sistema de Cache](#sistema-de-cache)
8. [Sistema de Memoria Cognitiva](#sistema-de-memoria-cognitiva)
9. [Sistema RAG](#sistema-rag)
10. [AST Parsing y Analisis de Codigo](#ast-parsing-y-analisis-de-codigo)
11. [Sistema de Reglas](#sistema-de-reglas)
12. [Sistema de Historial](#sistema-de-historial)
13. [Formatos de Reporte](#formatos-de-reporte)
14. [Sistema de Export](#sistema-de-export)
15. [Integracion Git](#integracion-git)
16. [Worker Pool y Concurrencia](#worker-pool-y-concurrencia)
17. [Tokenizer y Gestion de Tokens](#tokenizer-y-gestion-de-tokens)
18. [Configuracion](#configuracion)

---

## Arquitectura General

GoReview sigue una arquitectura modular con 19 paquetes internos:

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CLI Commands                               │
│  (review, commit, changelog, fix, stats, history, export, mcp...)   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Review Engine                               │
│            (Orquesta todo el proceso de review)                      │
└─────────────────────────────────────────────────────────────────────┘
          │              │              │              │
          ▼              ▼              ▼              ▼
    ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
    │   Git    │  │  Cache   │  │ Provider │  │  Rules   │
    │Integration│  │  System  │  │  System  │  │  System  │
    └──────────┘  └──────────┘  └──────────┘  └──────────┘
          │              │              │              │
          ▼              ▼              ▼              ▼
    ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
    │   AST    │  │  Memory  │  │   RAG    │  │ History  │
    │  Parser  │  │  System  │  │  System  │  │  System  │
    └──────────┘  └──────────┘  └──────────┘  └──────────┘
                                    │
                                    ▼
                          ┌──────────────────┐
                          │  Report/Export   │
                          │  (MD, JSON, SARIF│
                          │   Obsidian)      │
                          └──────────────────┘
```

### Flujo de Datos Principal

```
Input (CLI)
    ↓
Config (YAML + ENV + flags)
    ↓
Git (obtener diff)
    ↓
Rules (cargar reglas aplicables)
    ↓
Memory Check (memoria working/session)
    ↓
Review Engine (Worker Pool)
    ├→ Por cada archivo:
    │   ├→ Cache Check
    │   ├→ AST Parse (extraer contexto)
    │   ├→ RAG Fetch (conocimiento externo)
    │   ├→ Provider Call (review con IA)
    │   └→ Cache Store
    └→ Procesamiento paralelo
    ↓
History Storage (SQLite)
    ↓
Long-Term Memory (BadgerDB + Hebbian)
    ↓
Export (Obsidian, SARIF)
    ↓
Report (Markdown, JSON, SARIF)
    ↓
Output (consola, archivo)
```

---

## Comandos CLI

### `review` - Comando Principal

Analiza cambios de codigo y genera feedback con issues categorizados.

**Ubicacion:** `cmd/goreview/commands/review.go`

**Funcionamiento interno:**

1. Carga configuracion (YAML → ENV → flags)
2. Obtiene diff segun modo (staged/commit/branch/files)
3. Parsea diff en archivos individuales
4. Carga reglas aplicables (preset + custom)
5. Verifica cache para cada archivo
6. Si no hay cache: envia a provider de IA
7. Agrega resultados al historial
8. Genera reporte en formato solicitado

**Modos de operacion:**

| Modo | Flag | Descripcion |
|------|------|-------------|
| Staged | `--staged` | Cambios en staging area |
| Commit | `--commit <sha>` | Commit especifico |
| Branch | `--branch <name>` | Comparar con rama |
| Files | `file1 file2...` | Archivos especificos |

**Uso:**

```bash
# Review basico
goreview review --staged

# Review con modo especifico
goreview review --staged --mode=security

# Review con personalidad
goreview review --staged --personality=senior

# Multiples modos
goreview review --staged --mode=security,perf

# Con root cause tracing
goreview review --staged --trace

# Con verificacion TDD
goreview review --staged --require-tests --min-coverage=80

# Exportar a JSON
goreview review --staged --format json -o report.json

# Exportar a SARIF (para IDEs)
goreview review --staged --format sarif -o report.sarif
```

---

### `commit` - Generar Mensajes de Commit

Genera mensajes siguiendo Conventional Commits analizando los cambios staged.

**Ubicacion:** `cmd/goreview/commands/commit.go`, `commit_interactive.go`

**Funcionamiento interno:**

1. Obtiene diff de cambios staged
2. Analiza con IA para determinar:
   - Tipo (feat, fix, docs, style, refactor, test, chore)
   - Scope (modulo afectado)
   - Descripcion concisa
   - Body con detalles
   - Breaking changes
3. Formatea segun Conventional Commits
4. Opcionalmente ejecuta `git commit`

**Uso:**

```bash
# Generar mensaje (solo mostrar)
goreview commit

# Generar y ejecutar commit
goreview commit --execute

# Enmendar ultimo commit
goreview commit --amend

# Forzar tipo y scope
goreview commit --type feat --scope api

# Marcar como breaking change
goreview commit --breaking

# Con cuerpo adicional
goreview commit --body "Detalles adicionales"

# Dry-run (mostrar sin ejecutar)
goreview commit --dry-run
```

**Formato de salida:**

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]

BREAKING CHANGE: <description>
```

---

### `changelog` - Generar Changelog

Genera changelog automatico desde commits de git.

**Ubicacion:** `cmd/goreview/commands/changelog.go`

**Funcionamiento interno:**

1. Lee commits desde punto inicial (tag/commit)
2. Parsea mensajes Conventional Commits
3. Agrupa por tipo (Features, Fixes, etc.)
4. Genera markdown con formato Keep a Changelog
5. Incluye links a commits/PRs si es posible

**Uso:**

```bash
# Desde ultimo tag
goreview changelog

# Desde version especifica
goreview changelog --from v1.0.0

# Hasta version especifica
goreview changelog --from v1.0.0 --to v2.0.0

# Guardar a archivo
goreview changelog -o CHANGELOG.md

# Formato JSON
goreview changelog --format json
```

**Formato de salida:**

```markdown
# Changelog

## [Unreleased]

### Added
- feat(api): add user authentication endpoint

### Fixed
- fix(db): resolve connection pool leak

### Changed
- refactor(core): improve error handling
```

---

### `doc` - Generar Documentacion

Genera documentacion automatica de cambios de codigo.

**Ubicacion:** `cmd/goreview/commands/doc.go`, `doc_templates.go`

**Tipos de documentacion:**

| Tipo | Descripcion |
|------|-------------|
| `changes` | Descripcion de cambios realizados |
| `changelog` | Entrada de changelog |
| `api` | Documentacion de API |
| `readme` | Seccion de README |

**Uso:**

```bash
# Documentar cambios staged
goreview doc --staged

# Tipo especifico
goreview doc --staged --type changelog

# Estilo especifico
goreview doc --staged --style godoc

# Guardar a archivo
goreview doc --staged -o CHANGES.md

# Agregar al final
goreview doc --staged -o CHANGELOG.md --append

# Agregar al inicio
goreview doc --staged -o CHANGELOG.md --prepend
```

---

### `fix` - Auto-corregir Issues

Aplica correcciones automaticas a issues detectados.

**Ubicacion:** `cmd/goreview/commands/fix.go`

**Funcionamiento interno:**

1. Ejecuta review para detectar issues
2. Filtra issues que tienen `FixedCode` sugerido
3. Aplica patches a los archivos
4. Verifica sintaxis despues de aplicar
5. Reporta cambios realizados

**Uso:**

```bash
# Corregir cambios staged
goreview fix --staged

# Modo dry-run (preview)
goreview fix --staged --dry-run

# Solo severidades especificas
goreview fix --staged --severity critical,error

# Archivo especifico
goreview fix file.go
```

---

### `history` - Ver Historial

Gestiona el historial de reviews realizados.

**Ubicacion:** `cmd/goreview/commands/history.go`

**Uso:**

```bash
# Ver historial reciente
goreview history

# Buscar en historial
goreview history search "error handling"

# Filtrar por severidad
goreview history --severity critical

# Filtrar por archivo
goreview history --file auth.go

# Limpiar historial antiguo
goreview history prune --days 30
```

---

### `recall` - Recordar Contexto

Recupera informacion de reviews anteriores.

**Ubicacion:** `cmd/goreview/commands/recall.go`

**Funcionamiento interno:**

1. Busca en base de datos SQLite con FTS5
2. Aplica filtros (archivo, severidad, fecha)
3. Calcula relevancia de resultados
4. Retorna matches ordenados por relevancia

**Uso:**

```bash
# Buscar reviews anteriores
goreview recall "authentication"

# Ver estadisticas
goreview recall --stats

# Filtrar por archivo
goreview recall "sql injection" --file db.go

# Limitar resultados
goreview recall "memory leak" --limit 5
```

---

### `stats` - Estadisticas

Muestra estadisticas del proyecto y reviews.

**Ubicacion:** `cmd/goreview/commands/stats.go`

**Uso:**

```bash
# Estadisticas generales
goreview stats

# Por periodo
goreview stats --period month

# Agrupado por archivo
goreview stats --by-file

# Agrupado por severidad
goreview stats --by-severity

# Agrupado por tipo
goreview stats --by-type

# Formato JSON
goreview stats --format json
```

**Metricas disponibles:**

- Total de reviews
- Issues por severidad
- Issues por tipo
- Archivos mas problematicos
- Tendencias (mejorando/empeorando)
- Score promedio del proyecto

---

### `init` - Inicializar Proyecto

Configura GoReview en tu proyecto.

**Ubicacion:** `cmd/goreview/commands/init.go`, `init_detect.go`, `init_wizard.go`

**Funcionamiento interno:**

1. Detecta lenguaje/framework del proyecto
2. Detecta proveedor de IA disponible (Ollama local)
3. Ejecuta wizard interactivo
4. Genera `.goreview.yaml` configurado

**Uso:**

```bash
# Wizard interactivo
goreview init

# Usar valores por defecto
goreview init --yes

# Especificar proveedor
goreview init --provider openai --model gpt-4
```

---

### `config` - Ver Configuracion

Muestra la configuracion actual.

**Ubicacion:** `cmd/goreview/commands/config.go`

**Uso:**

```bash
# Mostrar configuracion
goreview config show

# Formato JSON
goreview config show --json

# Ver ubicacion de archivos
goreview config paths
```

---

### `export` - Exportar Reviews

Exporta reviews a sistemas externos.

**Ubicacion:** `cmd/goreview/commands/export.go`

**Uso:**

```bash
# Exportar a Obsidian
goreview export obsidian --from report.json

# Desde stdin
cat report.json | goreview export obsidian

# Con opciones
goreview export obsidian --from report.json \
  --vault ~/MyVault \
  --folder CodeReviews \
  --tags sprint-42
```

---

### `mcp-serve` - Servidor MCP

Inicia GoReview como servidor MCP para Claude Code.

**Ubicacion:** `cmd/goreview/commands/mcp.go`

**Uso:**

```bash
# Iniciar servidor
goreview mcp-serve

# Agregar a Claude Code
claude mcp add --transport stdio goreview -- goreview mcp-serve
```

Ver [MCP_SERVER.md](MCP_SERVER.md) para documentacion completa.

---

## Sistema de Review

### Motor de Review

**Ubicacion:** `internal/review/engine.go`

El motor de review es el componente central que orquesta todo el proceso:

```go
type Engine struct {
    cfg        *config.Config
    provider   providers.Provider
    cache      cache.Cache
    ruleLoader *rules.Loader
    gitRepo    git.Repository
    workerPool *worker.Pool
    metrics    *Metrics
}
```

### Proceso de Review

```
1. PrepareReview()
   ├─ Cargar configuracion
   ├─ Inicializar provider
   ├─ Cargar reglas
   └─ Preparar worker pool

2. GetChanges()
   ├─ Obtener diff de git
   ├─ Parsear en FileDiff[]
   └─ Filtrar archivos ignorados

3. ReviewFiles() [paralelo]
   ├─ Para cada archivo:
   │   ├─ Verificar cache
   │   ├─ Si cache miss:
   │   │   ├─ Extraer contexto AST
   │   │   ├─ Obtener conocimiento RAG
   │   │   ├─ Construir prompt
   │   │   ├─ Llamar provider IA
   │   │   └─ Guardar en cache
   │   └─ Agregar a resultados
   └─ Esperar todos los workers

4. AggregateResults()
   ├─ Combinar issues
   ├─ Calcular scores
   └─ Generar resumen

5. StoreHistory()
   ├─ Guardar en SQLite
   └─ Actualizar memoria

6. GenerateReport()
   └─ Formato solicitado
```

### Estructura de Issue

```go
type Issue struct {
    ID         string        // Identificador unico
    Type       string        // bug, security, performance, style, maintenance
    Severity   string        // info, warning, error, critical
    Message    string        // Descripcion del problema
    Suggestion string        // Sugerencia de solucion
    Location   Location      // Archivo, linea, columna
    RuleID     string        // ID de la regla que lo detecto
    FixedCode  string        // Codigo corregido (para auto-fix)
    RootCause  *RootCause    // Causa raiz (si --trace)
}

type RootCause struct {
    Description     string   // Explicacion de la causa raiz
    OriginFile      string   // Archivo donde se origina
    OriginLine      int      // Linea de origen
    PropagationPath []string // Camino de propagacion
    RelatedIssues   []string // Issues relacionados
    Recommendation  string   // Recomendacion de solucion
}
```

---

## Modos de Revision

**Ubicacion:** `internal/providers/modes.go`

Los modos enfocan el review en aspectos especificos del codigo.

### Modo `security`

Enfocado en vulnerabilidades y seguridad.

**Verifica:**
- OWASP Top 10
- SQL Injection, XSS, Command Injection
- Secrets hardcodeados
- Autenticacion/Autorizacion debil
- Configuracion insegura
- Dependencias vulnerables

**Uso:**
```bash
goreview review --staged --mode=security
```

### Modo `perf`

Enfocado en rendimiento y eficiencia.

**Verifica:**
- N+1 queries
- Falta de indices en queries
- Memory leaks
- Allocations innecesarias en loops
- Complejidad algoritmica excesiva
- Blocking I/O innecesario
- Falta de caching

**Uso:**
```bash
goreview review --staged --mode=perf
```

### Modo `clean`

Enfocado en codigo limpio y mantenible.

**Verifica:**
- Principios SOLID
- DRY (Don't Repeat Yourself)
- Naming conventions
- Code smells
- Funciones muy largas
- Clases con muchas responsabilidades
- Acoplamiento excesivo

**Uso:**
```bash
goreview review --staged --mode=clean
```

### Modo `docs`

Enfocado en documentacion.

**Verifica:**
- Funciones sin documentar
- Parametros sin descripcion
- Returns sin explicar
- JSDoc/GoDoc/docstrings faltantes
- Comentarios desactualizados
- README incompleto

**Uso:**
```bash
goreview review --staged --mode=docs
```

### Modo `tests`

Enfocado en testing y cobertura.

**Verifica:**
- Cobertura de tests
- Edge cases no testeados
- Mocking apropiado
- Tests deterministicos
- Assertions correctas
- Test isolation

**Uso:**
```bash
goreview review --staged --mode=tests
```

### Combinando Modos

Los modos son combinables para reviews mas completos:

```bash
# Seguridad + Performance
goreview review --staged --mode=security,perf

# Todos los modos
goreview review --staged --mode=security,perf,clean,docs,tests
```

---

## Personalidades del Reviewer

**Ubicacion:** `internal/providers/personalities.go`

Las personalidades cambian el estilo y tono del feedback.

### `default`

Balanceado, profesional, constructivo.

```
Issue: Variable name 'x' is not descriptive
Suggestion: Consider renaming to 'userCount' for clarity
```

### `senior`

Estilo de mentoring, explica el "por que".

```
Issue: Variable name 'x' is not descriptive

Why this matters: Descriptive names are crucial for code maintainability.
When another developer (or future you) reads this code, they should
understand the purpose without tracing through the logic.

Best practice: Use names that reveal intent. 'userCount' tells us this
tracks users, while 'x' requires mental mapping.

Suggestion: Rename to 'userCount' or 'activeUserCount' if that's more specific.
```

### `strict`

Directo, exigente, sin rodeos.

```
Issue: Unacceptable variable naming. 'x' violates naming conventions.
Fix immediately: Rename to descriptive name like 'userCount'.
This code will not pass review until fixed.
```

### `friendly`

Amable, positivo, alentador.

```
Hey! I noticed the variable 'x' could use a more descriptive name.

It's a small thing, but renaming it to something like 'userCount' would
make the code even clearer. You're doing great work overall!
```

### `security-expert`

Paranoia saludable, peor caso.

```
SECURITY CONCERN: Ambiguous variable naming

Risk assessment: While 'x' seems harmless, unclear variable names can
lead to security issues when developers misunderstand data flow.

In a security context, if 'x' contains user input that gets used in
a query later, the unclear name increases risk of injection vulnerabilities
being overlooked during code review.

Remediation: Rename to clearly indicate data source and sensitivity level.
Example: 'unsanitizedUserCount' or 'validatedUserCount'
```

**Uso:**

```bash
goreview review --staged --personality=senior
goreview review --staged --personality=strict
goreview review --staged --personality=friendly
goreview review --staged --personality=security-expert
```

---

## Sistema de Proveedores de IA

**Ubicacion:** `internal/providers/`

GoReview soporta multiples proveedores de IA con un sistema de fallback.

### Arquitectura de Proveedores

```go
type Provider interface {
    Review(ctx context.Context, req ReviewRequest) (*ReviewResponse, error)
    Name() string
    Model() string
    Close() error
}
```

### Proveedores Disponibles

#### Ollama (Local)

**Archivo:** `internal/providers/ollama.go`

- Ejecucion 100% local
- Sin costo, sin limites
- Privacidad total
- Requiere instalacion de Ollama

**Modelos recomendados:**
- `qwen2.5-coder:14b` (mejor balance)
- `codellama:13b` (alternativa)
- `deepseek-coder:6.7b` (mas ligero)

**Configuracion:**

```yaml
provider:
  name: ollama
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434
  timeout: 60s
```

#### OpenAI

**Archivo:** `internal/providers/openai.go`

- Alta calidad
- Requiere API key
- Costo por token

**Modelos:**
- `gpt-4-turbo` (mejor calidad)
- `gpt-4` (alta calidad)
- `gpt-3.5-turbo` (economico)

**Configuracion:**

```yaml
provider:
  name: openai
  model: gpt-4-turbo
  api_key: ${OPENAI_API_KEY}
```

#### Google Gemini

**Archivo:** `internal/providers/gemini.go`

- Tier gratuito generoso
- Alta calidad
- Ventana de contexto grande

**Modelos:**
- `gemini-pro` (estandar)
- `gemini-1.5-flash` (rapido, 1M tokens)

**Configuracion:**

```yaml
provider:
  name: gemini
  model: gemini-pro
  api_key: ${GEMINI_API_KEY}
```

#### Groq

**Archivo:** `internal/providers/groq.go`

- Inferencia ultra-rapida
- Tier gratuito disponible
- Ideal para iteracion rapida

**Modelos:**
- `llama-3.1-70b-versatile`
- `mixtral-8x7b-32768`

**Configuracion:**

```yaml
provider:
  name: groq
  model: llama-3.1-70b-versatile
  api_key: ${GROQ_API_KEY}
```

#### Mistral

**Archivo:** `internal/providers/mistral.go`

- Especializado en codigo
- Buena relacion calidad/precio

**Modelos:**
- `codestral-latest` (especializado codigo)
- `mistral-large-latest`

**Configuracion:**

```yaml
provider:
  name: mistral
  model: codestral-latest
  api_key: ${MISTRAL_API_KEY}
```

### Sistema de Fallback

**Archivo:** `internal/providers/fallback.go`

Encadena multiples proveedores con failover automatico:

```yaml
provider:
  name: fallback
  fallback:
    providers:
      - name: ollama
        model: qwen2.5-coder:14b
      - name: gemini
        model: gemini-pro
      - name: groq
        model: llama-3.1-70b-versatile
```

**Orden de prioridad por defecto:**
1. Gemini (gratis, alta calidad)
2. Groq (gratis, muy rapido)
3. Mistral (economico)
4. OpenAI (fallback final)

### Auto-deteccion

**Archivo:** `internal/providers/factory.go`

Con `name: auto`, GoReview detecta automaticamente:

1. Intenta conectar a Ollama en localhost:11434
2. Si falla, usa fallback con proveedores cloud

```yaml
provider:
  name: auto  # Detecta automaticamente
```

### Rate Limiting

**Archivo:** `internal/providers/ratelimit.go`

Control de tasa de peticiones por proveedor:

```yaml
provider:
  rate_limit_rps: 10  # Requests per second
```

### Retry con Backoff

**Archivo:** `internal/providers/retry.go`

Reintentos automaticos con backoff exponencial:

- Maximo 3 reintentos
- Backoff: 1s, 2s, 4s
- Errores retryables: timeout, rate limit, server error

---

## Sistema de Cache

**Ubicacion:** `internal/cache/`

El sistema de cache evita re-analizar codigo que no ha cambiado.

### Arquitectura

```go
type Cache interface {
    Get(key string) (*CacheEntry, bool)
    Set(key string, entry *CacheEntry) error
    Delete(key string) error
    Clear() error
    Stats() CacheStats
}
```

### Generacion de Keys

**Archivo:** `internal/cache/cache.go`

La key se genera hasheando:
- Contenido del diff
- Lenguaje del archivo
- Path del archivo
- Reglas aplicadas
- Version de GoReview

```go
func ComputeKey(req ReviewRequest) string {
    data := fmt.Sprintf("%s|%s|%s|%v|%s",
        req.Diff,
        req.Language,
        req.FilePath,
        req.Rules,
        Version,
    )
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

### Cache LRU (In-Memory)

**Archivo:** `internal/cache/lru.go`

Cache en memoria con eviccion LRU:

```yaml
cache:
  enabled: true
  max_entries: 1000  # Maximo de entradas
  ttl: 24h           # Time-to-live
```

**Funcionamiento:**
- Acceso O(1) con map + doubly-linked list
- Eviccion del menos recientemente usado
- TTL por entrada

### Cache de Archivos

**Archivo:** `internal/cache/file.go`

Cache persistente en disco:

```yaml
cache:
  enabled: true
  dir: ~/.goreview/cache
  max_size_mb: 100
  ttl: 168h  # 1 semana
```

**Formato de archivo:**
```
~/.goreview/cache/
├── ab/
│   └── abcd1234...json
├── cd/
│   └── cdef5678...json
└── ...
```

### Estadisticas de Cache

```go
type CacheStats struct {
    Hits       int64
    Misses     int64
    Entries    int64
    SizeBytes  int64
    HitRate    float64
}
```

**Ver estadisticas:**

```bash
goreview stats --cache
```

---

## Sistema de Memoria Cognitiva

**Ubicacion:** `internal/memory/`

Sistema de memoria multi-nivel inspirado en cognicion humana.

### Niveles de Memoria

```
┌─────────────────────────────────────────────────────────────┐
│                     Working Memory                           │
│  (Corto plazo, in-memory, LRU, TTL corto)                   │
│  Capacidad: ~100 items, TTL: 1 hora                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Session Memory                           │
│  (Por sesion, multiples sesiones concurrentes)              │
│  Persiste durante la sesion de trabajo                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Long-Term Memory                          │
│  (BadgerDB, persistente, busqueda semantica)                │
│  Garbage collection automatico                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Hebbian Learning                           │
│  (Asociaciones entre entradas, fortalecimiento/decaimiento) │
└─────────────────────────────────────────────────────────────┘
```

### Working Memory

**Archivo:** `internal/memory/working.go`

Memoria de corto plazo para contexto inmediato:

```go
type WorkingMemory struct {
    entries  map[string]*MemoryEntry
    order    *list.List  // LRU ordering
    capacity int
    ttl      time.Duration
}
```

**Configuracion:**

```yaml
memory:
  enabled: true
  working:
    capacity: 100
    ttl: 1h
```

### Session Memory

**Archivo:** `internal/memory/session.go`

Memoria persistente por sesion de trabajo:

```go
type SessionMemory struct {
    sessions    map[string]*Session
    maxSessions int
    sessionTTL  time.Duration
}
```

**Configuracion:**

```yaml
memory:
  session:
    max_sessions: 10
    session_ttl: 24h
```

### Long-Term Memory

**Archivo:** `internal/memory/longterm.go`

Almacenamiento persistente con BadgerDB:

```go
type LongTermMemory struct {
    db          *badger.DB
    maxSize     int64
    gcInterval  time.Duration
}
```

**Caracteristicas:**
- Persistencia en disco
- Busqueda semantica con embeddings
- Garbage collection automatico
- Compactacion periodica

**Configuracion:**

```yaml
memory:
  long_term:
    enabled: true
    max_size_mb: 500
    gc_interval: 1h
```

### Hebbian Learning

**Archivo:** `internal/memory/hebbian.go`

Sistema de aprendizaje asociativo:

```go
type HebbianNetwork struct {
    connections map[string]map[string]float64  // Fortaleza de asociacion
    learningRate float64
    decayRate    float64
    minStrength  float64
}
```

**Funcionamiento:**
- Cuando dos entradas se activan juntas, su conexion se fortalece
- Las conexiones decaen con el tiempo si no se usan
- Conexiones debiles se eliminan (pruning)

**Configuracion:**

```yaml
memory:
  hebbian:
    learning_rate: 0.1   # Cuanto aumenta por co-activacion
    decay_rate: 0.01     # Decaimiento por tick
    min_strength: 0.05   # Umbral de eliminacion
```

### Estructura de Entrada

```go
type MemoryEntry struct {
    ID        string
    Content   string
    Type      string            // "issue", "pattern", "context"
    Tags      []string
    Metadata  map[string]string
    Embedding []float64         // Vector para busqueda semantica
    CreatedAt time.Time
    AccessedAt time.Time
    AccessCount int
}
```

### Busqueda en Memoria

```go
type MemoryQuery struct {
    Text       string    // Busqueda por texto
    Semantic   bool      // Usar busqueda semantica
    Tags       []string  // Filtrar por tags
    Types      []string  // Filtrar por tipo
    Limit      int       // Maximo resultados
    MinScore   float64   // Score minimo de relevancia
}
```

---

## Sistema RAG

**Ubicacion:** `internal/rag/`

Retrieval-Augmented Generation para enriquecer reviews con conocimiento externo.

### Arquitectura

```
┌──────────────────┐     ┌──────────────────┐
│   Code Review    │ ──▶ │  RAG Detector    │
│     Request      │     │ (Framework/Lang) │
└──────────────────┘     └──────────────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │  Source Fetcher  │
                         │  (HTTP/Local)    │
                         └──────────────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    ▼             ▼             ▼
             ┌──────────┐  ┌──────────┐  ┌──────────┐
             │  Style   │  │ Security │  │Framework │
             │  Guide   │  │  Guide   │  │   Docs   │
             └──────────┘  └──────────┘  └──────────┘
                    │             │             │
                    └─────────────┼─────────────┘
                                  ▼
                         ┌──────────────────┐
                         │  Context Builder │
                         │  (Token Budget)  │
                         └──────────────────┘
                                  │
                                  ▼
                         ┌──────────────────┐
                         │    AI Review     │
                         │   with Context   │
                         └──────────────────┘
```

### Tipos de Fuentes

```go
type SourceType string

const (
    StyleGuide    SourceType = "style_guide"
    Security      SourceType = "security"
    BestPractice  SourceType = "best_practice"
    API           SourceType = "api"
    Framework     SourceType = "framework"
)
```

### Configuracion de Fuentes

```yaml
rag:
  enabled: true
  cache_dir: ~/.goreview/rag
  default_cache_ttl: 24h
  max_cache_size: 100  # MB
  auto_detect: true    # Detectar frameworks automaticamente

  sources:
    - type: style_guide
      url: https://example.com/style-guide.md
      cache_ttl: 168h  # 1 semana

    - type: security
      url: https://owasp.org/www-project-code-review-guide/
      cache_ttl: 168h

    - type: framework
      url: https://docs.example.com/api
      languages: ["go", "typescript"]
      cache_ttl: 24h
```

### Auto-deteccion de Frameworks

**Archivo:** `internal/rag/detector.go`

Detecta automaticamente frameworks usados:

```go
type Detector struct {
    patterns map[string][]DetectionPattern
}

type DetectionPattern struct {
    FilePattern string   // e.g., "package.json"
    Contains    []string // e.g., ["\"react\"", "\"next\""]
    Framework   string   // e.g., "react", "nextjs"
}
```

**Frameworks detectados:**
- Go: Gin, Echo, Fiber, Chi
- JavaScript: React, Vue, Angular, Next.js, Express
- Python: Django, Flask, FastAPI
- Java: Spring Boot
- Rust: Actix, Rocket

### Fetcher con Cache

**Archivo:** `internal/rag/fetcher.go`

```go
type Fetcher struct {
    cacheDir   string
    httpClient *http.Client
    cacheTTL   time.Duration
}

func (f *Fetcher) Fetch(source Source) (*Document, error) {
    // 1. Check cache
    if cached := f.getFromCache(source); cached != nil {
        return cached, nil
    }

    // 2. Fetch from source
    doc, err := f.fetchFromURL(source.URL)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache
    f.storeInCache(source, doc)

    return doc, nil
}
```

### Integracion con Review

El conocimiento RAG se inyecta en el prompt:

```
[System prompt base]

## Additional Context from Style Guide:
<style_guide content>

## Security Guidelines:
<security content>

## Framework Best Practices:
<framework content>

[User's code diff]
```

---

## AST Parsing y Analisis de Codigo

**Ubicacion:** `internal/ast/`

Sistema de parsing multi-lenguaje para extraer contexto estructural.

### Parser Multi-lenguaje

**Archivo:** `internal/ast/parser.go`

```go
type Parser struct {
    languageParsers map[string]LanguageParser
}

type LanguageParser interface {
    Parse(content string) (*ParseResult, error)
    Language() string
}

type ParseResult struct {
    Imports     []Import
    Functions   []Function
    Classes     []Class
    Interfaces  []Interface
    Variables   []Variable
    Constants   []Constant
}
```

### Lenguajes Soportados

#### Go

```go
type GoParser struct{}

// Detecta:
// - package declarations
// - import statements
// - func declarations (with receivers)
// - type declarations (struct, interface)
// - const/var blocks
```

#### JavaScript/TypeScript

```go
type JSParser struct{}

// Detecta:
// - import/require statements
// - function declarations
// - arrow functions
// - class declarations
// - const/let/var declarations
// - export statements
```

#### Python

```go
type PythonParser struct{}

// Detecta:
// - import statements
// - from...import statements
// - def declarations
// - class declarations
// - decorators (@)
// - docstrings
```

#### Java

```go
type JavaParser struct{}

// Detecta:
// - package declarations
// - import statements
// - class declarations
// - interface declarations
// - method declarations
// - annotations (@)
```

#### Rust

```go
type RustParser struct{}

// Detecta:
// - use statements
// - mod declarations
// - fn declarations
// - struct declarations
// - impl blocks
// - trait declarations
```

### Estructuras Extraidas

```go
type Function struct {
    Name       string
    StartLine  int
    EndLine    int
    Exported   bool       // Publico/privado
    Receiver   string     // Para metodos
    Parameters []Parameter
    Returns    []string
    DocComment string
}

type Class struct {
    Name       string
    StartLine  int
    EndLine    int
    Exported   bool
    Extends    string
    Implements []string
    Methods    []Function
    Fields     []Variable
    DocComment string
}

type Import struct {
    Path      string
    Alias     string
    Line      int
    IsDefault bool
}
```

### Context Builder

**Archivo:** `internal/ast/context_builder.go`

Construye contexto estructural para el review:

```go
type ContextBuilder struct {
    parser *Parser
}

func (cb *ContextBuilder) BuildContext(diff FileDiff) (*CodeContext, error) {
    // 1. Parse el archivo completo
    result, err := cb.parser.Parse(diff.Content)

    // 2. Identificar funciones/clases modificadas
    changedFuncs := cb.findChangedFunctions(diff.Hunks, result.Functions)

    // 3. Extraer contexto circundante
    context := cb.extractSurroundingContext(changedFuncs, result)

    return context, nil
}
```

### Uso en Review

El contexto AST enriquece el prompt:

```
## File Structure:
- Package: main
- Imports: fmt, net/http, encoding/json
- Functions modified:
  - HandleUser (line 45-78)
  - ValidateInput (line 80-95)

## Function Context:
```go
// HandleUser handles user requests
func HandleUser(w http.ResponseWriter, r *http.Request) {
    // ... modified code ...
}
```

## Related Code:
```go
// ValidateInput is called by HandleUser
func ValidateInput(input string) error {
    // ... implementation ...
}
```
```

---

## Sistema de Reglas

**Ubicacion:** `internal/rules/`

Sistema de reglas YAML con herencia y presets.

### Estructura de Regla

```go
type Rule struct {
    ID          string   `yaml:"id"`
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Category    string   `yaml:"category"`    // security, performance, style, etc.
    Severity    string   `yaml:"severity"`    // info, warning, error, critical
    Languages   []string `yaml:"languages"`   // Lenguajes aplicables
    Patterns    []string `yaml:"patterns"`    // Patrones de archivo
    Enabled     bool     `yaml:"enabled"`
    Message     string   `yaml:"message"`     // Mensaje al detectar
    Suggestion  string   `yaml:"suggestion"`  // Sugerencia de fix
}
```

### Categorias de Reglas

| Categoria | Descripcion |
|-----------|-------------|
| `security` | Vulnerabilidades, autenticacion, secrets |
| `performance` | N+1, memory leaks, algoritmos |
| `best_practice` | Patrones recomendados |
| `style` | Naming, formato, consistencia |
| `bug` | Errores logicos, null checks |
| `maintenance` | Complejidad, duplicacion |

### Presets

**Archivo:** `internal/rules/defaults/`

#### `minimal`

Solo reglas criticas:
- SQL Injection
- XSS
- Command Injection
- Hardcoded secrets
- Null pointer dereference

#### `standard` (default)

Balance entre cobertura y ruido:
- Todo de `minimal`
- Memory leaks
- N+1 queries
- Error handling
- Basic code smells

#### `strict`

Maxima cobertura:
- Todo de `standard`
- Style enforcement
- Documentation requirements
- Complexity limits
- Test coverage

### Archivo de Reglas YAML

```yaml
# .goreview/rules/custom.yaml
rules:
  - id: custom-no-print
    name: No Print Statements
    description: Print statements should not be in production code
    category: style
    severity: warning
    languages: [go, python, javascript]
    patterns: ["*.go", "*.py", "*.js"]
    enabled: true
    message: "Found print statement in production code"
    suggestion: "Use proper logging instead"

  - id: custom-max-params
    name: Maximum Parameters
    description: Functions should not have more than 5 parameters
    category: maintenance
    severity: warning
    languages: [go, javascript, typescript]
    enabled: true
    message: "Function has too many parameters"
    suggestion: "Consider using a configuration object"
```

### Herencia de Reglas

**Archivo:** `internal/rules/inherit.go`

```yaml
rules:
  preset: standard
  inherit_from:
    - https://company.com/coding-standards/rules.yaml
    - ./team-rules.yaml

  # Override reglas heredadas
  override:
    - id: security-sql-injection
      severity: critical  # Aumentar severidad
    - id: style-max-line-length
      enabled: false      # Deshabilitar

  # Reglas adicionales
  enabled:
    - custom-no-print
    - custom-max-params

  disabled:
    - style-trailing-whitespace
```

### Filtrado de Reglas

**Archivo:** `internal/rules/filter.go`

```go
type RuleFilter struct {
    IDs        []string   // Filtrar por ID
    Categories []string   // Filtrar por categoria
    Severities []string   // Filtrar por severidad
    Languages  []string   // Filtrar por lenguaje
}

func (l *Loader) Filter(filter RuleFilter) []Rule {
    // Aplica filtros y retorna reglas coincidentes
}
```

---

## Sistema de Historial

**Ubicacion:** `internal/history/`

Almacenamiento y busqueda de reviews historicos.

### Almacenamiento por Commit

```
.git/goreview/
└── commits/
    ├── abc123/
    │   ├── analysis.json
    │   └── issues.json
    ├── def456/
    │   ├── analysis.json
    │   └── issues.json
    └── ...
```

### Estructura de Analisis

```go
type CommitAnalysis struct {
    CommitHash  string    `json:"commit_hash"`
    Author      string    `json:"author"`
    Date        time.Time `json:"date"`
    Branch      string    `json:"branch"`
    Message     string    `json:"message"`

    FilesReviewed int     `json:"files_reviewed"`
    TotalIssues   int     `json:"total_issues"`

    Issues      []Issue   `json:"issues"`
    Score       int       `json:"score"`
    Duration    string    `json:"duration"`
}
```

### Base de Datos SQLite

**Esquema:**

```sql
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY,
    commit_hash TEXT,
    file_path TEXT,
    issue_type TEXT,
    severity TEXT,
    message TEXT,
    line_number INTEGER,
    created_at DATETIME,
    resolved BOOLEAN DEFAULT FALSE
);

CREATE VIRTUAL TABLE reviews_fts USING fts5(
    message,
    content='reviews',
    content_rowid='id'
);

CREATE INDEX idx_reviews_commit ON reviews(commit_hash);
CREATE INDEX idx_reviews_file ON reviews(file_path);
CREATE INDEX idx_reviews_severity ON reviews(severity);
```

### Busqueda Full-Text (FTS5)

```go
func (h *History) Search(query string, opts SearchOptions) ([]RecallResult, error) {
    sql := `
        SELECT r.*,
               bm25(reviews_fts) as relevance
        FROM reviews r
        JOIN reviews_fts ON r.id = reviews_fts.rowid
        WHERE reviews_fts MATCH ?
        ORDER BY relevance
        LIMIT ?
    `
    // ...
}
```

### Estadisticas Agregadas

```go
type IssueStats struct {
    TotalIssues    int
    BySeverity     map[string]int
    ByType         map[string]int
    ByFile         map[string]int
    ByAuthor       map[string]int
    TrendDirection string  // "improving", "stable", "worsening"
}

type RecurringIssue struct {
    Pattern     string
    Occurrences int
    Files       []string
    FirstSeen   time.Time
    LastSeen    time.Time
}
```

---

## Formatos de Reporte

**Ubicacion:** `internal/report/`

### Markdown

**Archivo:** `internal/report/markdown.go`

Formato legible para humanos:

```markdown
# Code Review Report

## Summary

| Metric | Value |
|--------|-------|
| Files Reviewed | 5 |
| Total Issues | 12 |
| Critical | 1 |
| Errors | 3 |
| Warnings | 6 |
| Info | 2 |
| Score | 72/100 |

## Issues

### src/auth/handler.go

#### [CRITICAL] SQL Injection Vulnerability (Line 45)

**Message:** User input is directly concatenated into SQL query

**Suggestion:** Use parameterized queries instead

```go
// Before
query := "SELECT * FROM users WHERE id = " + userID

// After
query := "SELECT * FROM users WHERE id = ?"
rows, err := db.Query(query, userID)
```

---

### src/api/routes.go

#### [WARNING] Missing Error Handling (Line 78)

...
```

### JSON

**Archivo:** `internal/report/json.go`

Formato estructurado para procesamiento:

```json
{
  "metadata": {
    "project_name": "my-project",
    "branch": "feature/auth",
    "commit": "abc123",
    "review_date": "2024-12-27T10:30:00Z",
    "review_mode": "staged",
    "preset": "standard"
  },
  "summary": {
    "files_reviewed": 5,
    "total_issues": 12,
    "by_severity": {
      "critical": 1,
      "error": 3,
      "warning": 6,
      "info": 2
    },
    "average_score": 72
  },
  "files": [
    {
      "path": "src/auth/handler.go",
      "language": "go",
      "issues": [
        {
          "id": "issue-001",
          "type": "security",
          "severity": "critical",
          "message": "SQL Injection Vulnerability",
          "suggestion": "Use parameterized queries",
          "location": {
            "file": "src/auth/handler.go",
            "line": 45,
            "column": 12
          },
          "rule_id": "security-sql-injection",
          "fixed_code": "query := \"SELECT * FROM users WHERE id = ?\"\nrows, err := db.Query(query, userID)"
        }
      ],
      "score": 65
    }
  ],
  "duration": "2.5s"
}
```

### SARIF

**Archivo:** `internal/report/sarif.go`

Static Analysis Results Interchange Format para IDEs:

```json
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "GoReview",
          "version": "1.0.0",
          "informationUri": "https://github.com/JNZader/goreview",
          "rules": [
            {
              "id": "security-sql-injection",
              "name": "SQL Injection",
              "shortDescription": {
                "text": "SQL Injection Vulnerability"
              },
              "defaultConfiguration": {
                "level": "error"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "security-sql-injection",
          "level": "error",
          "message": {
            "text": "User input is directly concatenated into SQL query"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "src/auth/handler.go"
                },
                "region": {
                  "startLine": 45,
                  "startColumn": 12
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

---

## Sistema de Export

**Ubicacion:** `internal/export/`

### Obsidian Export

**Archivos:** `internal/export/obsidian.go`, `obsidian_template.go`

Exporta reviews a vault de Obsidian con formato optimizado.

#### Caracteristicas

- **Frontmatter YAML** con metadatos completos
- **Tags** para organizacion (#goreview, #security, etc.)
- **Callouts** de Obsidian para severidades
- **Wiki links** a reviews relacionados
- **Organizacion** por proyecto: `GoReview/{proyecto}/review-001.md`

#### Formato de Salida

```markdown
---
date: 2024-12-27T10:30:00Z
project: my-api-service
branch: feature/auth
commit: abc123
files_reviewed: 5
total_issues: 7
severity:
  critical: 1
  error: 2
  warning: 3
  info: 1
average_score: 72
tags:
  - goreview
  - security
  - bug
aliases:
  - review-2024-12-27
related:
  - "[[review-001-2024-12-25]]"
  - "[[review-002-2024-12-26]]"
---

# Code Review: my-api-service

#goreview #security #bug

## Summary

| Metric | Value |
|--------|-------|
| Files Reviewed | 5 |
| Total Issues | 7 |
| Average Score | 72/100 |

## Issues

### src/auth/handler.go

> [!danger] SQL injection vulnerability
> **Location:** Line 45
> **Type:** security
>
> User input is directly concatenated into SQL query
>
> **Suggestion:** Use parameterized queries

> [!warning] Missing error handling
> **Location:** Line 78
> **Type:** bug
>
> Error from database query is ignored

---

## Related Reviews

- [[review-001-2024-12-25]]
- [[review-002-2024-12-26]]
```

#### Configuracion

```yaml
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
      - team-backend
      - sprint-42
    template_file: ~/.goreview/obsidian-template.md
```

#### Uso

```bash
# Durante review
goreview review --staged --export-obsidian

# Con vault especifico
goreview review --staged --export-obsidian --obsidian-vault ~/MyVault

# Comando separado
goreview export obsidian --from report.json

# Desde stdin
cat report.json | goreview export obsidian
```

---

## Integracion Git

**Ubicacion:** `internal/git/`

### Repository Interface

```go
type Repository interface {
    GetStagedDiff() ([]FileDiff, error)
    GetCommitDiff(sha string) ([]FileDiff, error)
    GetBranchDiff(branch string) ([]FileDiff, error)
    GetFileDiff(files []string) ([]FileDiff, error)
    GetRepoRoot() (string, error)
    GetCurrentBranch() (string, error)
    GetLastCommitSHA() (string, error)
}
```

### Estructura de Diff

```go
type FileDiff struct {
    Path       string
    OldPath    string   // Para renombrados
    Status     string   // added, modified, deleted, renamed
    Language   string   // Detectado por extension
    IsBinary   bool
    Hunks      []Hunk
    Stats      DiffStats
}

type Hunk struct {
    OldStart int
    OldLines int
    NewStart int
    NewLines int
    Content  string
    Context  []string
}

type DiffStats struct {
    Additions int
    Deletions int
}
```

### Parser de Diffs

**Archivo:** `internal/git/parser.go`

```go
type DiffParser struct {
    languageMap map[string]string  // extension -> language
}

func (p *DiffParser) Parse(diffOutput string) ([]FileDiff, error) {
    // Parsea formato unified diff
    // Detecta lenguaje por extension
    // Extrae hunks con contexto
    // Detecta archivos binarios
}
```

### Deteccion de Lenguaje

```go
var languageExtensions = map[string]string{
    ".go":    "go",
    ".js":    "javascript",
    ".ts":    "typescript",
    ".tsx":   "typescript",
    ".jsx":   "javascript",
    ".py":    "python",
    ".java":  "java",
    ".rs":    "rust",
    ".rb":    "ruby",
    ".php":   "php",
    ".c":     "c",
    ".cpp":   "cpp",
    ".h":     "c",
    ".hpp":   "cpp",
    ".cs":    "csharp",
    ".swift": "swift",
    ".kt":    "kotlin",
    ".scala": "scala",
    // ...
}
```

---

## Worker Pool y Concurrencia

**Ubicacion:** `internal/worker/`

### Arquitectura del Pool

```go
type Pool struct {
    workers    int
    taskQueue  chan Task
    results    chan Result
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

type Task interface {
    Execute(ctx context.Context) error
    ID() string
}

type Result struct {
    TaskID string
    Data   interface{}
    Error  error
}
```

### Funcionamiento

```
                    ┌───────────────────────┐
                    │      Task Queue       │
                    │  (buffered channel)   │
                    └───────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        ┌──────────┐    ┌──────────┐    ┌──────────┐
        │ Worker 1 │    │ Worker 2 │    │ Worker N │
        │(goroutine)│   │(goroutine)│   │(goroutine)│
        └──────────┘    └──────────┘    └──────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
                    ┌───────────────────────┐
                    │    Results Channel    │
                    │  (streaming results)  │
                    └───────────────────────┘
```

### Configuracion

```yaml
review:
  max_concurrency: 5  # 0 = auto (GOMAXPROCS * 2, max 10)
```

### Uso en Review

```go
func (e *Engine) ReviewFiles(files []FileDiff) (*ReviewResult, error) {
    pool := worker.NewPool(e.cfg.Review.MaxConcurrency)
    pool.Start(e.ctx)

    // Submit tasks
    for _, file := range files {
        task := &ReviewTask{
            File:     file,
            Provider: e.provider,
            Rules:    e.rules,
        }
        pool.Submit(task)
    }

    // Collect results
    results := pool.CollectAll()

    return e.aggregateResults(results), nil
}
```

---

## Tokenizer y Gestion de Tokens

**Ubicacion:** `internal/tokenizer/`

### Estimacion de Tokens

```go
type Tokenizer struct {
    modelRatios map[string]float64
}

func (t *Tokenizer) EstimateTokens(text string, model string) int {
    // 1. Obtener ratio para el modelo
    ratio := t.getRatio(model)

    // 2. Contar caracteres
    charCount := len(text)

    // 3. Ajustar por contenido de codigo
    if isCode(text) {
        ratio *= 1.2  // Codigo usa mas tokens
    }

    // 4. Ajustar por whitespace
    wsRatio := countWhitespace(text) / float64(charCount)
    ratio *= (1 - wsRatio * 0.3)

    return int(float64(charCount) * ratio)
}
```

### Ratios por Modelo

```go
var modelRatios = map[string]float64{
    "gpt-4":           0.25,  // ~4 chars per token
    "gpt-3.5":         0.25,
    "claude-3":        0.28,
    "gemini":          0.26,
    "llama":           0.30,
    "mistral":         0.28,
    "qwen":            0.30,
    "codellama":       0.32,  // Codigo es mas denso
}
```

### Limites de Tokens por Modelo

```go
var modelLimits = map[string]int{
    "gpt-4-turbo":      128000,
    "gpt-4":            8192,
    "gpt-3.5-turbo":    16384,
    "claude-3-opus":    200000,
    "claude-3-sonnet":  200000,
    "gemini-pro":       32768,
    "gemini-1.5-flash": 1048576,  // 1M tokens!
    "llama-3.1-70b":    32768,
    "qwen2.5-coder":    32768,
    "codestral":        32768,
}
```

### Budget Management

```go
type TokenBudget struct {
    MaxTokens       int
    ResponseReserve int     // Reservar para respuesta
    ContextReserve  int     // Reservar para contexto RAG
}

func (b *TokenBudget) AvailableForPrompt() int {
    return b.MaxTokens - b.ResponseReserve - b.ContextReserve
}

func (b *TokenBudget) CanFit(text string, model string) bool {
    estimated := estimateTokens(text, model)
    return estimated <= b.AvailableForPrompt()
}
```

---

## Configuracion

**Ubicacion:** `internal/config/`

### Archivo de Configuracion

**Archivo:** `.goreview.yaml`

```yaml
# Version del schema de configuracion
version: "1.0"

# Proveedor de IA
provider:
  name: ollama                    # ollama, openai, gemini, groq, mistral, auto, fallback
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434
  api_key: ${OPENAI_API_KEY}      # Soporta variables de entorno
  timeout: 60s
  max_tokens: 4096
  temperature: 0.1
  rate_limit_rps: 10

# Integracion Git
git:
  repo_path: .
  base_branch: main
  ignore_patterns:
    - "vendor/**"
    - "node_modules/**"
    - "*.min.js"
    - "*.generated.*"
    - "**/*.pb.go"

# Configuracion de Review
review:
  mode: staged                    # staged, commit, branch, files
  min_severity: warning           # info, warning, error, critical
  max_issues: 100
  max_concurrency: 5              # 0 = auto
  timeout: 5m
  context: ""                     # Contexto adicional para prompts
  personality: default            # default, senior, strict, friendly, security-expert
  modes: []                       # security, perf, clean, docs, tests (combinables)
  root_cause_tracing: false
  require_tests: false
  min_coverage: 0

# Configuracion de Output
output:
  format: markdown                # markdown, json, sarif
  file: ""                        # Archivo de salida (vacio = stdout)
  include_code: true
  color: true
  verbose: false
  quiet: false

# Sistema de Cache
cache:
  enabled: true
  dir: ~/.goreview/cache
  ttl: 24h
  max_size_mb: 100
  max_entries: 1000

# Sistema de Reglas
rules:
  preset: standard                # minimal, standard, strict
  rules_dir: .goreview/rules
  enabled: []                     # IDs de reglas a habilitar
  disabled: []                    # IDs de reglas a deshabilitar
  inherit_from: []                # URLs o paths de reglas a heredar
  override: []                    # Overrides de propiedades de reglas

# Sistema de Memoria
memory:
  enabled: true
  dir: ~/.goreview/memory
  working:
    capacity: 100
    ttl: 1h
  session:
    max_sessions: 10
    session_ttl: 24h
  long_term:
    enabled: true
    max_size_mb: 500
    gc_interval: 1h
  hebbian:
    learning_rate: 0.1
    decay_rate: 0.01
    min_strength: 0.05

# Sistema RAG
rag:
  enabled: true
  cache_dir: ~/.goreview/rag
  default_cache_ttl: 24h
  max_cache_size: 100
  auto_detect: true
  sources: []

# Sistema de Export
export:
  obsidian:
    enabled: false
    vault_path: ""
    folder_name: GoReview
    include_tags: true
    include_callouts: true
    include_links: true
    link_to_previous: true
    custom_tags: []
    template_file: ""
```

### Variables de Entorno

Todas las configuraciones pueden sobrescribirse con variables de entorno:

```bash
# Formato: GOREVIEW_<SECCION>_<PROPIEDAD>
export GOREVIEW_PROVIDER_NAME=openai
export GOREVIEW_PROVIDER_MODEL=gpt-4
export GOREVIEW_PROVIDER_APIKEY=sk-...

export GOREVIEW_REVIEW_PERSONALITY=senior
export GOREVIEW_REVIEW_MODES=security,perf

export GOREVIEW_CACHE_ENABLED=true
export GOREVIEW_CACHE_TTL=48h
```

### Orden de Precedencia

1. **Flags CLI** (mayor prioridad)
2. **Variables de entorno**
3. **Archivo `.goreview.yaml` local**
4. **Archivo `~/.goreview.yaml` global**
5. **Valores por defecto** (menor prioridad)

### Loader de Configuracion

**Archivo:** `internal/config/loader.go`

```go
type Loader struct {
    v *viper.Viper
}

func (l *Loader) Load() (*Config, error) {
    // 1. Set defaults
    l.setDefaults()

    // 2. Read config file
    l.v.SetConfigName(".goreview")
    l.v.AddConfigPath(".")
    l.v.AddConfigPath("$HOME")
    l.v.ReadInConfig()

    // 3. Bind environment variables
    l.v.SetEnvPrefix("GOREVIEW")
    l.v.AutomaticEnv()

    // 4. Unmarshal
    var cfg Config
    l.v.Unmarshal(&cfg)

    // 5. Validate
    return l.validate(&cfg)
}
```

---

## Apendice: Estructura Completa del Proyecto

```
goreview/
├── cmd/goreview/
│   ├── main.go                    # Punto de entrada
│   └── commands/
│       ├── root.go                # Comando raiz
│       ├── review.go              # Comando review
│       ├── commit.go              # Comando commit
│       ├── commit_interactive.go  # Commit interactivo
│       ├── doc.go                 # Comando doc
│       ├── doc_templates.go       # Templates de doc
│       ├── changelog.go           # Comando changelog
│       ├── config.go              # Comando config
│       ├── init.go                # Comando init
│       ├── init_detect.go         # Deteccion de proyecto
│       ├── init_wizard.go         # Wizard de init
│       ├── export.go              # Comando export
│       ├── history.go             # Comando history
│       ├── recall.go              # Comando recall
│       ├── search.go              # Comando search
│       ├── stats.go               # Comando stats
│       ├── plan.go                # Comando plan
│       ├── fix.go                 # Comando fix
│       ├── mcp.go                 # Comando mcp-serve
│       ├── version.go             # Comando version
│       ├── constants.go           # Constantes
│       ├── output.go              # Utilidades de output
│       └── progress.go            # Display de progreso
│
├── internal/
│   ├── ast/
│   │   ├── parser.go              # Parser multi-lenguaje
│   │   └── context_builder.go     # Constructor de contexto
│   │
│   ├── cache/
│   │   ├── cache.go               # Interface de cache
│   │   ├── lru.go                 # Cache LRU in-memory
│   │   └── file.go                # Cache en disco
│   │
│   ├── config/
│   │   ├── config.go              # Estructuras de config
│   │   ├── defaults.go            # Valores por defecto
│   │   ├── loader.go              # Carga de config
│   │   └── validate.go            # Validacion
│   │
│   ├── export/
│   │   ├── types.go               # Tipos de export
│   │   ├── obsidian.go            # Exporter Obsidian
│   │   └── obsidian_template.go   # Template Obsidian
│   │
│   ├── git/
│   │   ├── repository.go          # Interface Repository
│   │   ├── types.go               # Tipos de Git
│   │   ├── parser.go              # Parser de diffs
│   │   └── parser_optimized.go    # Parser optimizado
│   │
│   ├── history/
│   │   ├── history.go             # Gestion de historial
│   │   ├── storage.go             # Almacenamiento
│   │   ├── search.go              # Busqueda FTS5
│   │   └── stats.go               # Estadisticas
│   │
│   ├── knowledge/
│   │   └── fetcher.go             # Fetcher de conocimiento
│   │
│   ├── logger/
│   │   └── logger.go              # Sistema de logging
│   │
│   ├── mcp/
│   │   ├── server.go              # Servidor MCP
│   │   └── tools.go               # Herramientas MCP
│   │
│   ├── memory/
│   │   ├── types.go               # Tipos de memoria
│   │   ├── working.go             # Working memory
│   │   ├── session.go             # Session memory
│   │   ├── longterm.go            # Long-term memory
│   │   ├── hebbian.go             # Hebbian learning
│   │   └── embedding.go           # Embeddings
│   │
│   ├── metrics/
│   │   └── metrics.go             # Metricas globales
│   │
│   ├── profiler/
│   │   └── profiler.go            # Profiling
│   │
│   ├── providers/
│   │   ├── provider.go            # Interface Provider
│   │   ├── factory.go             # Factory de providers
│   │   ├── ollama.go              # Provider Ollama
│   │   ├── openai.go              # Provider OpenAI
│   │   ├── gemini.go              # Provider Gemini
│   │   ├── groq.go                # Provider Groq
│   │   ├── mistral.go             # Provider Mistral
│   │   ├── fallback.go            # Provider Fallback
│   │   ├── personalities.go       # Personalidades
│   │   ├── modes.go               # Modos de review
│   │   ├── ratelimit.go           # Rate limiting
│   │   ├── retry.go               # Retry logic
│   │   ├── validation.go          # Validacion
│   │   └── constants.go           # Constantes
│   │
│   ├── rag/
│   │   ├── types.go               # Tipos RAG
│   │   ├── fetcher.go             # Fetcher de docs
│   │   ├── detector.go            # Detector de frameworks
│   │   └── styleguide.go          # Style guides
│   │
│   ├── report/
│   │   ├── report.go              # Interface Report
│   │   ├── markdown.go            # Reporte Markdown
│   │   ├── json.go                # Reporte JSON
│   │   └── sarif.go               # Reporte SARIF
│   │
│   ├── review/
│   │   ├── engine.go              # Motor de review
│   │   ├── engine_metrics.go      # Metricas del engine
│   │   └── types.go               # Tipos de review
│   │
│   ├── rules/
│   │   ├── types.go               # Tipos de reglas
│   │   ├── loader.go              # Carga de reglas
│   │   ├── filter.go              # Filtrado de reglas
│   │   ├── inherit.go             # Herencia de reglas
│   │   └── defaults/
│   │       └── base.yaml          # Reglas por defecto
│   │
│   ├── tokenizer/
│   │   ├── tokenizer.go           # Estimacion de tokens
│   │   └── budget.go              # Budget management
│   │
│   └── worker/
│       └── pool.go                # Worker pool
│
├── claude-code-plugin/            # Plugin para Claude Code
│   ├── .claude-plugin/
│   │   └── plugin.json
│   ├── commands/
│   ├── agents/
│   ├── skills/
│   ├── hooks/
│   └── scripts/
│
├── docs/
│   ├── FEATURES.md               # Esta documentacion
│   ├── MCP_SERVER.md             # Guia MCP
│   ├── PLUGIN_GUIDE.md           # Guia del plugin
│   ├── BACKGROUND_WATCHER.md     # Guia del watcher
│   ├── CHECKPOINT_SYNC.md        # Guia de checkpoints
│   └── CLAUDE_CODE_INTEGRATION.md
│
├── .goreview.yaml                # Config por defecto
├── .golangci.yml                 # Config del linter
├── Makefile                      # Build commands
├── Dockerfile                    # Docker build
└── README.md                     # README principal
```

---

## Estadisticas del Proyecto

| Metrica | Valor |
|---------|-------|
| Paquetes internos | 19 |
| Archivos Go (impl) | 66+ |
| Comandos CLI | 16 |
| Proveedores IA | 7 |
| Lenguajes AST | 6 |
| Personalidades | 5 |
| Modos de review | 6 |
| Formatos de reporte | 3 |

---

*Documentacion generada para GoReview v1.0.0*
