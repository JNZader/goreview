# Integracion con Claude Code

Esta guia explica como integrar GoReview con Claude Code para obtener reviews automaticos, commits inteligentes y herramientas de calidad de codigo.

---

## Tabla de Contenidos

1. [Metodos de Integracion](#metodos-de-integracion)
2. [MCP Server](#mcp-server)
3. [Plugin para Claude Code](#plugin-para-claude-code)
4. [Configuracion Manual](#configuracion-manual)
5. [Uso Avanzado](#uso-avanzado)

---

## Metodos de Integracion

GoReview se puede integrar con Claude Code de tres formas:

| Metodo | Complejidad | Caracteristicas |
|--------|-------------|-----------------|
| **MCP Server** | Baja | Tools disponibles para Claude |
| **Plugin** | Media | Commands, agents, hooks, skills |
| **Manual** | Alta | Configuracion personalizada |

---

## MCP Server

El metodo mas simple es agregar GoReview como servidor MCP.

### Instalacion Rapida

```bash
# Agregar GoReview como MCP server
claude mcp add --transport stdio goreview -- goreview mcp-serve
```

### Verificar Instalacion

```bash
# Ver servidores MCP activos
claude mcp list
```

### Herramientas Disponibles

Una vez instalado, Claude Code puede usar:

| Herramienta | Descripcion |
|-------------|-------------|
| `goreview_review` | Analizar cambios de codigo |
| `goreview_commit` | Generar mensaje de commit |
| `goreview_fix` | Auto-corregir issues |
| `goreview_search` | Buscar en historial |
| `goreview_stats` | Ver estadisticas |
| `goreview_changelog` | Generar changelog |
| `goreview_doc` | Generar documentacion |

### Ejemplo de Uso

En Claude Code, simplemente pide:

```
Revisa mis cambios staged con goreview
Genera un mensaje de commit para estos cambios
Busca issues de seguridad en el historial
```

### Configuracion Avanzada

Crea `.mcp.json` en tu proyecto:

```json
{
  "mcpServers": {
    "goreview": {
      "type": "stdio",
      "command": "goreview",
      "args": ["mcp-serve"],
      "env": {
        "GOREVIEW_PROVIDER_NAME": "ollama",
        "GOREVIEW_PROVIDER_MODEL": "qwen2.5-coder:14b"
      }
    }
  }
}
```

---

## Plugin para Claude Code

El plugin proporciona integracion completa con commands, agents y hooks.

### Instalacion

```bash
# Opcion 1: Desde marketplace (cuando este disponible)
/plugin marketplace add JNZader/goreview
/plugin install goreview

# Opcion 2: Instalacion local
cd /path/to/goreview/claude-code-plugin
/plugin install --local .
```

### Slash Commands

| Comando | Accion |
|---------|--------|
| `/review` | Revisar cambios staged |
| `/commit-ai` | Generar commit con IA |
| `/fix-issues` | Auto-corregir issues |
| `/changelog` | Generar changelog |
| `/stats` | Ver estadisticas |
| `/security-scan` | Scan de seguridad profundo |

### Subagents Especializados

El plugin incluye agentes expertos:

#### Security Reviewer
```
Usa el security-reviewer para auditar este codigo
```
- Enfocado en OWASP Top 10
- Root cause tracing
- Recomendaciones de remediacion

#### Performance Reviewer
```
Usa el perf-reviewer para encontrar cuellos de botella
```
- Deteccion de N+1 queries
- Analisis de complejidad
- Optimizaciones sugeridas

#### Test Reviewer
```
Usa el test-reviewer para verificar cobertura
```
- Analisis de cobertura
- Identificacion de edge cases
- Sugerencias de tests

#### Background Watcher
```
Activa el goreview-watcher en background
```
- Monitoreo continuo
- Alertas de issues criticos
- Reportes periodicos

### Hooks Automaticos

El plugin configura hooks para:

1. **PostToolUse (Edit/Write)**: Review automatico despues de editar codigo
2. **SessionStart**: Muestra salud del proyecto al iniciar
3. **Stop**: Guarda estado del review al terminar
4. **PreCompact**: Sincroniza con checkpoints

### Skills

Los skills se invocan automaticamente:

- **goreview-workflow**: Cuando revisas codigo o preparas commits
- **commit-standards**: Cuando escribes mensajes de commit

---

## Configuracion Manual

Si prefieres configurar manualmente:

### 1. Crear Hooks

Edita `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [{
          "type": "command",
          "command": "goreview review --staged --preset=minimal 2>/dev/null || true"
        }]
      }
    ]
  }
}
```

### 2. Crear Commands

Crea `.claude/.commands/review.md`:

```markdown
---
description: Review staged changes with GoReview
---

Run GoReview analysis:
\`\`\`bash
goreview review --staged --format markdown
\`\`\`
```

### 3. Crear Agents

Crea `.claude/agents/security/AGENT.md`:

```yaml
---
name: security
description: Security expert using GoReview
tools: Read, Bash(goreview:*)
model: sonnet
---

You are a security expert. Use:
goreview review --staged --mode=security --trace
```

### 4. Crear Skills

Crea `.claude/skills/goreview/SKILL.md`:

```yaml
---
name: goreview
description: Code review with GoReview
allowed-tools: Bash(goreview:*)
---

# GoReview Commands

- Review: `goreview review --staged`
- Commit: `goreview commit`
- Fix: `goreview fix --staged`
```

---

## Uso Avanzado

### Permisos con Wildcards

Permitir todas las herramientas de GoReview:

```json
{
  "permissions": {
    "allow": ["mcp__goreview__*"]
  }
}
```

### Integracion con Checkpoints

Guardar reviews con cada checkpoint:

```json
{
  "hooks": {
    "PreCompact": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": "goreview review --staged -o .claude/checkpoints/review-$(date +%s).json"
      }]
    }]
  }
}
```

### Background Agent Continuo

El `goreview-watcher` corre en background y reporta issues:

```
Activa el goreview-watcher para monitorear mis cambios
```

Solo interrumpe para issues criticos, acumula warnings.

### Combinacion con Claude in Chrome

Si usas Claude in Chrome, puedes:

1. Revisar PRs en GitHub
2. Analizar codigo en la web
3. Generar commits desde el browser

---

## Troubleshooting

### El MCP server no responde

```bash
# Verificar que goreview esta en PATH
which goreview

# Probar manualmente
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | goreview mcp-serve
```

### Los hooks no se ejecutan

```bash
# Verificar configuracion
/config hooks

# Ver logs de hooks
/hooks --verbose
```

### El plugin no se instala

```bash
# Verificar version de Claude Code
claude --version  # Debe ser >= 2.0.12

# Reinstalar
/plugin uninstall goreview
/plugin install goreview
```

---

## Recursos

- [Documentacion de Claude Code](https://code.claude.com/docs)
- [MCP Specification](https://modelcontextprotocol.io)
- [GoReview GitHub](https://github.com/JNZader/goreview)
