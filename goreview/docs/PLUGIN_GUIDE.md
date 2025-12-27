# GoReview Plugin para Claude Code

El plugin de GoReview proporciona integracion completa con Claude Code: slash commands, agentes especializados, skills automaticos y hooks de automatizacion.

---

## Tabla de Contenidos

1. [Instalacion](#instalacion)
2. [Slash Commands](#slash-commands)
3. [Subagents Especializados](#subagents-especializados)
4. [Skills Automaticos](#skills-automaticos)
5. [Hooks de Automatizacion](#hooks-de-automatizacion)
6. [Configuracion](#configuracion)
7. [Estructura del Plugin](#estructura-del-plugin)

---

## Instalacion

### Prerequisitos

- Claude Code v2.0.12 o superior
- GoReview CLI instalado y en PATH

### Metodo 1: Instalacion Local

```bash
# Navegar al directorio del plugin
cd /path/to/goreview/claude-code-plugin

# Instalar localmente
/plugin install --local .
```

### Metodo 2: Desde Marketplace (Futuro)

```bash
# Agregar marketplace
/plugin marketplace add JNZader/goreview

# Instalar plugin
/plugin install goreview
```

### Verificar Instalacion

```bash
# Ver plugins instalados
/plugin list

# Deberia mostrar: goreview (1.0.0) - Active
```

### Habilitar/Deshabilitar

```bash
# Deshabilitar temporalmente
/plugin disable goreview

# Habilitar de nuevo
/plugin enable goreview
```

---

## Slash Commands

El plugin agrega 6 comandos rapidos:

### /review

Revisar cambios staged con GoReview.

```
/review
```

**Opciones de modo:**
```
/review security    # Enfoque en seguridad
/review perf        # Enfoque en performance
/review clean       # Enfoque en codigo limpio
/review docs        # Enfoque en documentacion
/review tests       # Enfoque en tests
```

**Con root cause tracing:**
```
/review security --trace
```

---

### /commit-ai

Generar mensaje de commit con IA.

```
/commit-ai
```

**Ejecutar commit directamente:**
```
/commit-ai --execute
```

**Forzar tipo y scope:**
```
/commit-ai --type feat --scope api
```

---

### /fix-issues

Auto-corregir issues encontrados.

```
/fix-issues
```

**Preview sin aplicar:**
```
/fix-issues --dry-run
```

**Solo issues criticos:**
```
/fix-issues --severity critical,error
```

---

### /changelog

Generar changelog desde commits.

```
/changelog
```

**Desde version especifica:**
```
/changelog --from v1.0.0
```

**Guardar a archivo:**
```
/changelog -o CHANGELOG.md
```

---

### /stats

Ver estadisticas de reviews.

```
/stats
```

**Por periodo:**
```
/stats --period month
```

**Agrupado:**
```
/stats --by-severity
/stats --by-file
```

---

### /security-scan

Scan de seguridad profundo con root cause tracing.

```
/security-scan
```

Este comando ejecuta:
```bash
goreview review --staged --mode=security --personality=security-expert --trace
```

---

## Subagents Especializados

El plugin incluye 5 agentes expertos que puedes invocar naturalmente:

### Security Reviewer

**Invocacion:**
```
Usa el security-reviewer para auditar este codigo
```

**Especialidades:**
- OWASP Top 10 vulnerabilities
- Deteccion de secrets hardcodeados
- Analisis de injection (SQL, XSS, command)
- Root cause tracing
- Recomendaciones de remediacion

**Ejemplo de uso:**
```
@security-reviewer revisa los cambios en el modulo de autenticacion
```

**Checklist que verifica:**
- [ ] No hay secrets hardcodeados
- [ ] Input validation en datos de usuario
- [ ] Queries parametrizadas
- [ ] Output encoding
- [ ] Autenticacion correcta
- [ ] Autorizacion en endpoints
- [ ] Manejo seguro de sesiones

---

### Performance Reviewer

**Invocacion:**
```
Usa el perf-reviewer para encontrar cuellos de botella
```

**Especialidades:**
- Deteccion de N+1 queries
- Analisis de complejidad algoritmica
- Memory leaks
- Blocking I/O
- Oportunidades de caching

**Ejemplo de uso:**
```
@perf-reviewer analiza el rendimiento de las queries en UserService
```

**Checklist que verifica:**
- [ ] No hay N+1 queries
- [ ] Paginacion correcta
- [ ] Caching apropiado
- [ ] Estructuras de datos eficientes
- [ ] Allocations minimas en loops
- [ ] I/O asincrono donde corresponde

---

### Test Reviewer

**Invocacion:**
```
Usa el test-reviewer para verificar cobertura
```

**Especialidades:**
- Analisis de cobertura
- Calidad de tests
- Edge cases faltantes
- TDD enforcement

**Ejemplo de uso:**
```
@test-reviewer verifica que los tests cubran los nuevos endpoints
```

**Checklist que verifica:**
- [ ] Codigo nuevo tiene tests
- [ ] Tests cubren happy path
- [ ] Tests cubren error cases
- [ ] Tests cubren edge cases
- [ ] Tests son deterministicos
- [ ] Mocks apropiados

---

### Fix Agent

**Invocacion:**
```
Usa el fix-agent para corregir los issues automaticamente
```

**Especialidades:**
- Aplicar fixes de forma segura
- Verificar cambios
- Commits incrementales
- Rollback si es necesario

**Workflow:**
1. Identifica issues con review
2. Preview de fixes con --dry-run
3. Aplica fixes por severidad
4. Verifica con tests

**Ejemplo de uso:**
```
@fix-agent corrige todos los issues de seguridad criticos
```

---

### GoReview Watcher (Background)

**Invocacion:**
```
Activa el goreview-watcher para monitorear mis cambios
```

**Caracteristicas:**
- Corre en background
- Usa modelo Haiku (eficiente)
- Monitorea cada 30 segundos
- Solo interrumpe para criticos
- Acumula warnings

**Como funciona:**
1. Detecta cambios staged
2. Ejecuta review minimal
3. Filtra por severidad
4. Reporta via systemMessage

**Output ejemplo:**
```
⚠️ GoReview: 1 critical issue in auth.go:45 - possible SQL injection
```

---

## Skills Automaticos

Los skills se invocan automaticamente cuando el contexto coincide.

### GoReview Workflow

**Se activa cuando:**
- Revisas codigo
- Analizas cambios
- Preparas commits
- Verificas calidad

**Proporciona:**
- Comandos rapidos de review
- Modos disponibles
- Personalidades
- Ejemplos de uso combinado

---

### Commit Standards

**Se activa cuando:**
- Escribes mensajes de commit
- Preparas cambios para version control
- Mencionas "commit"

**Proporciona:**
- Formato Conventional Commits
- Tipos disponibles (feat, fix, docs...)
- Best practices
- Ejemplos

---

## Hooks de Automatizacion

### PostToolUse: Auto-Review

Despues de cada Edit o Write, ejecuta review automatico:

```json
{
  "matcher": "Edit|Write",
  "hooks": [{
    "command": "goreview review --staged --preset=minimal"
  }]
}
```

**Resultado:** Si hay issues criticos, muestra:
```
⚠️ GoReview: 2 critical/error issues found. Run /review for details.
```

---

### SessionStart: Project Health

Al iniciar sesion, muestra salud del proyecto:

```json
{
  "matcher": "startup",
  "hooks": [{
    "command": "goreview stats --format json"
  }]
}
```

**Resultado:**
```
📊 Project health: 85/100 score, 12 issues in last week
```

---

### Stop: Save Review State

Al terminar sesion, guarda estado del review:

```json
{
  "matcher": "*",
  "hooks": [{
    "command": "goreview review --staged -o .claude/last-review.json"
  }]
}
```

---

### PreToolUse: Commit Tip

Cuando detecta `git commit`, sugiere usar GoReview:

```
💡 Tip: Use /commit-ai for AI-generated commit messages
```

---

## Configuracion

### Opciones del Plugin

Edita tu configuracion de Claude Code:

```json
{
  "goreview.autoReview": true,
  "goreview.mode": "standard",
  "goreview.personality": "senior"
}
```

| Opcion | Tipo | Default | Descripcion |
|--------|------|---------|-------------|
| `autoReview` | boolean | `true` | Review automatico despues de edits |
| `mode` | string | `"standard"` | Preset por defecto |
| `personality` | string | `"senior"` | Personalidad del reviewer |

### Deshabilitar Hooks

Para deshabilitar hooks especificos:

```json
{
  "hooks": {
    "PostToolUse": []
  }
}
```

### Personalizar Agentes

Copia y modifica agentes en `.claude/agents/`:

```bash
cp -r claude-code-plugin/agents/security-reviewer .claude/agents/my-security/
```

Edita `AGENT.md` segun tus necesidades.

---

## Estructura del Plugin

```
claude-code-plugin/
├── .claude-plugin/
│   └── plugin.json           # Metadata y configuracion
│
├── commands/                  # Slash commands
│   ├── review.md             # /review
│   ├── commit-ai.md          # /commit-ai
│   ├── fix-issues.md         # /fix-issues
│   ├── changelog.md          # /changelog
│   ├── stats.md              # /stats
│   └── security-scan.md      # /security-scan
│
├── agents/                    # Subagents especializados
│   ├── security-reviewer/
│   │   └── AGENT.md          # Experto en seguridad
│   ├── perf-reviewer/
│   │   └── AGENT.md          # Experto en performance
│   ├── test-reviewer/
│   │   └── AGENT.md          # Experto en testing
│   ├── fix-agent/
│   │   └── AGENT.md          # Agente de auto-fix
│   └── goreview-watcher/
│       └── AGENT.md          # Background watcher
│
├── skills/                    # Skills auto-invocados
│   ├── goreview-workflow/
│   │   └── SKILL.md          # Workflow de review
│   └── commit-standards/
│       └── SKILL.md          # Estandares de commit
│
├── hooks/                     # Automatizacion
│   ├── hooks.json            # Hooks principales
│   └── checkpoint-sync.json  # Sincronizacion con checkpoints
│
├── scripts/                   # Scripts de utilidad
│   ├── quick-review.sh       # Review rapido
│   ├── security-scan.sh      # Scan de seguridad
│   └── pre-commit-check.sh   # Check pre-commit
│
└── README.md                  # Documentacion del plugin
```

---

## Ejemplos de Flujos de Trabajo

### Flujo de Desarrollo Normal

1. Escribir codigo
2. Ver feedback automatico del watcher
3. Ejecutar `/review` para detalles
4. Usar `@fix-agent` para correcciones
5. Ejecutar `/commit-ai` para commit

### Review de Seguridad Pre-Release

1. Activar watcher: `@goreview-watcher`
2. Scan profundo: `/security-scan`
3. Analisis detallado: `@security-reviewer`
4. Corregir: `@fix-agent --severity critical`
5. Verificar: `/review security`

### Code Review de PR

1. Checkout del branch
2. Review completo: `/review`
3. Performance: `@perf-reviewer`
4. Tests: `@test-reviewer`
5. Documentar: `goreview doc`

---

## Desinstalacion

```bash
# Desinstalar plugin
/plugin uninstall goreview

# Remover MCP server (si se agrego)
claude mcp remove goreview
```
