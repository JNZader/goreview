# GoReview MCP Server

El servidor MCP (Model Context Protocol) permite que Claude Code y otros clientes MCP usen GoReview como herramienta integrada.

---

## Tabla de Contenidos

1. [Que es MCP?](#que-es-mcp)
2. [Instalacion Rapida](#instalacion-rapida)
3. [Configuracion](#configuracion)
4. [Herramientas Disponibles](#herramientas-disponibles)
5. [Ejemplos de Uso](#ejemplos-de-uso)
6. [Troubleshooting](#troubleshooting)

---

## Que es MCP?

MCP (Model Context Protocol) es un protocolo abierto que permite a los modelos de IA conectarse con herramientas externas. Cuando agregas GoReview como servidor MCP, Claude Code puede:

- Revisar tu codigo automaticamente
- Generar commits inteligentes
- Buscar en el historial de reviews
- Y mas...

Todo sin salir de la conversacion con Claude.

---

## Instalacion Rapida

### Prerequisitos

1. **GoReview instalado** y en el PATH:
   ```bash
   # Verificar instalacion
   goreview version
   ```

2. **Claude Code v2.0.12+**:
   ```bash
   claude --version
   ```

### Agregar como MCP Server

```bash
# Comando unico para agregar GoReview
claude mcp add --transport stdio goreview -- goreview mcp-serve
```

### Verificar Instalacion

```bash
# Ver servidores MCP activos
claude mcp list

# Deberia mostrar:
# goreview (stdio) - Active
```

---

## Configuracion

### Metodo 1: Comando CLI (Recomendado)

```bash
# Agregar con scope local (solo este proyecto)
claude mcp add --transport stdio goreview --scope local -- goreview mcp-serve

# Agregar con scope de usuario (todos los proyectos)
claude mcp add --transport stdio goreview --scope user -- goreview mcp-serve
```

### Metodo 2: Archivo .mcp.json

Crea `.mcp.json` en la raiz de tu proyecto:

```json
{
  "mcpServers": {
    "goreview": {
      "type": "stdio",
      "command": "goreview",
      "args": ["mcp-serve"]
    }
  }
}
```

### Metodo 3: Con Variables de Entorno

Para configurar el proveedor de IA:

```json
{
  "mcpServers": {
    "goreview": {
      "type": "stdio",
      "command": "goreview",
      "args": ["mcp-serve"],
      "env": {
        "GOREVIEW_PROVIDER_NAME": "ollama",
        "GOREVIEW_PROVIDER_MODEL": "qwen2.5-coder:14b",
        "GOREVIEW_PROVIDER_BASE_URL": "http://localhost:11434"
      }
    }
  }
}
```

### Metodo 4: Path Absoluto

Si goreview no esta en el PATH:

```json
{
  "mcpServers": {
    "goreview": {
      "type": "stdio",
      "command": "/usr/local/bin/goreview",
      "args": ["mcp-serve"]
    }
  }
}
```

---

## Herramientas Disponibles

### 1. goreview_review

Analiza cambios de codigo e identifica issues.

**Parametros:**

| Parametro | Tipo | Default | Descripcion |
|-----------|------|---------|-------------|
| `target` | string | `"staged"` | Que revisar: staged, HEAD, SHA, branch |
| `mode` | string | - | Modo: security, perf, clean, docs, tests |
| `personality` | string | - | Estilo: senior, strict, friendly, security-expert |
| `files` | array | - | Archivos especificos a revisar |
| `trace` | boolean | `false` | Activar root cause tracing |

**Ejemplo en Claude Code:**

```
Revisa mis cambios staged enfocandote en seguridad
```

Claude usara:
```json
{
  "name": "goreview_review",
  "arguments": {
    "target": "staged",
    "mode": "security"
  }
}
```

---

### 2. goreview_commit

Genera mensajes de commit siguiendo Conventional Commits.

**Parametros:**

| Parametro | Tipo | Default | Descripcion |
|-----------|------|---------|-------------|
| `type` | string | - | Forzar tipo: feat, fix, docs, etc. |
| `scope` | string | - | Forzar scope |
| `breaking` | boolean | `false` | Marcar como breaking change |

**Ejemplo en Claude Code:**

```
Genera un mensaje de commit para estos cambios
```

---

### 3. goreview_fix

Aplica correcciones automaticas a issues encontrados.

**Parametros:**

| Parametro | Tipo | Default | Descripcion |
|-----------|------|---------|-------------|
| `target` | string | `"staged"` | Que corregir: staged o path |
| `severity` | array | - | Solo estas severidades |
| `types` | array | - | Solo estos tipos de issue |
| `dryRun` | boolean | `false` | Mostrar sin aplicar |

**Ejemplo en Claude Code:**

```
Corrige automaticamente los issues criticos en mis cambios
```

---

### 4. goreview_search

Busca en el historial de reviews.

**Parametros:**

| Parametro | Tipo | Default | Descripcion |
|-----------|------|---------|-------------|
| `query` | string | **required** | Texto a buscar |
| `severity` | string | - | Filtrar por severidad |
| `file` | string | - | Filtrar por archivo |
| `limit` | integer | `10` | Maximo de resultados |

**Ejemplo en Claude Code:**

```
Busca issues de SQL injection en reviews anteriores
```

---

### 5. goreview_stats

Obtiene estadisticas y metricas del proyecto.

**Parametros:**

| Parametro | Tipo | Default | Descripcion |
|-----------|------|---------|-------------|
| `period` | string | `"week"` | Periodo: today, week, month, all |
| `groupBy` | string | - | Agrupar por: file, severity, type, author |

**Ejemplo en Claude Code:**

```
Muestrame las estadisticas de reviews de esta semana
```

---

### 6. goreview_changelog

Genera changelog desde commits de git.

**Parametros:**

| Parametro | Tipo | Default | Descripcion |
|-----------|------|---------|-------------|
| `from` | string | - | Punto inicial (tag o commit) |
| `to` | string | `"HEAD"` | Punto final |
| `format` | string | `"markdown"` | Formato: markdown, json |

**Ejemplo en Claude Code:**

```
Genera el changelog desde la version v1.0.0
```

---

### 7. goreview_doc

Genera documentacion para cambios de codigo.

**Parametros:**

| Parametro | Tipo | Default | Descripcion |
|-----------|------|---------|-------------|
| `target` | string | `"staged"` | Que documentar |
| `type` | string | `"changes"` | Tipo: changes, changelog, api, readme |
| `style` | string | `"markdown"` | Estilo: markdown, jsdoc, godoc |

**Ejemplo en Claude Code:**

```
Genera documentacion para los cambios que hice
```

---

## Ejemplos de Uso

### Flujo de Trabajo Tipico

1. **Escribir codigo** normalmente

2. **Pedir review**:
   ```
   Revisa mis cambios staged
   ```

3. **Ver sugerencias** y aplicar fixes:
   ```
   Aplica las correcciones sugeridas
   ```

4. **Generar commit**:
   ```
   Genera un mensaje de commit
   ```

### Review de Seguridad Profundo

```
Haz un review de seguridad completo con root cause tracing
```

Claude ejecutara:
```json
{
  "name": "goreview_review",
  "arguments": {
    "target": "staged",
    "mode": "security",
    "personality": "security-expert",
    "trace": true
  }
}
```

### Buscar Patrones en Historial

```
Busca todos los issues de performance relacionados con database
```

### Estadisticas de Calidad

```
Dame un resumen de la calidad del codigo este mes, agrupado por archivo
```

---

## Permisos

### Permitir Todas las Herramientas GoReview

Edita tu configuracion de Claude Code:

```json
{
  "permissions": {
    "allow": ["mcp__goreview__*"]
  }
}
```

### Permitir Herramientas Especificas

```json
{
  "permissions": {
    "allow": [
      "mcp__goreview__review",
      "mcp__goreview__commit",
      "mcp__goreview__stats"
    ]
  }
}
```

---

## Troubleshooting

### El servidor no inicia

```bash
# Verificar que goreview funciona
goreview version

# Probar el servidor manualmente
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | goreview mcp-serve
```

Deberia responder:
```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","serverInfo":{"name":"goreview","version":"1.0.0"},"capabilities":{"tools":{}}}}
```

### Las herramientas no aparecen

```bash
# Listar herramientas
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | goreview mcp-serve
```

### Error "goreview not found"

1. Verificar que esta en el PATH:
   ```bash
   which goreview
   ```

2. O usar path absoluto en `.mcp.json`:
   ```json
   {
     "command": "/ruta/completa/a/goreview"
   }
   ```

### El review tarda mucho

El servidor ejecuta goreview internamente. Si tarda:

1. Verifica que el proveedor de IA esta corriendo (ej: Ollama)
2. Usa un modelo mas rapido
3. Reduce el scope del review

### Ver logs del servidor

El servidor escribe logs a stderr:

```bash
goreview mcp-serve 2>mcp-server.log
```

---

## Arquitectura

```
┌─────────────────┐         ┌──────────────────┐
│   Claude Code   │  stdin  │  GoReview MCP    │
│                 │ ──────> │     Server       │
│  (MCP Client)   │         │                  │
│                 │ <────── │  JSON-RPC 2.0    │
└─────────────────┘  stdout └──────────────────┘
                                    │
                                    │ exec
                                    ▼
                            ┌──────────────────┐
                            │  GoReview CLI    │
                            │                  │
                            │  review, commit  │
                            │  fix, search...  │
                            └──────────────────┘
```

El servidor MCP:
1. Recibe peticiones JSON-RPC de Claude Code
2. Traduce a comandos GoReview CLI
3. Ejecuta y captura output
4. Retorna resultados formateados

---

## Referencia Tecnica

### Protocolo

- **Transporte**: stdio (stdin/stdout)
- **Formato**: JSON-RPC 2.0
- **Version MCP**: 2024-11-05

### Metodos Soportados

| Metodo | Descripcion |
|--------|-------------|
| `initialize` | Handshake inicial |
| `initialized` | Confirmacion de cliente |
| `tools/list` | Listar herramientas |
| `tools/call` | Ejecutar herramienta |
| `ping` | Health check |

### Codigo Fuente

- `internal/mcp/server.go` - Servidor JSON-RPC
- `internal/mcp/tools.go` - Definicion de herramientas
- `cmd/goreview/commands/mcp.go` - Comando CLI
