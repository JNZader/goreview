# GoReview Background Watcher

El Background Watcher es un agente que monitorea continuamente tus cambios de codigo y reporta issues sin interrumpir tu flujo de trabajo.

---

## Tabla de Contenidos

1. [Que es el Background Watcher?](#que-es-el-background-watcher)
2. [Como Activarlo](#como-activarlo)
3. [Como Funciona](#como-funciona)
4. [Configuracion](#configuracion)
5. [Interpretando los Mensajes](#interpretando-los-mensajes)
6. [Mejores Practicas](#mejores-practicas)

---

## Que es el Background Watcher?

El Background Watcher es un subagent de Claude Code que:

- **Corre en segundo plano** mientras trabajas
- **Monitorea cambios** en tu codigo staged
- **Ejecuta reviews** automaticos y ligeros
- **Reporta solo lo importante** sin interrumpir
- **Usa recursos minimos** (modelo Haiku)

### Beneficios

| Beneficio | Descripcion |
|-----------|-------------|
| **Feedback continuo** | Sabes si hay problemas antes de commit |
| **No invasivo** | Solo interrumpe para criticos |
| **Eficiente** | Usa modelo ligero y cache |
| **Preventivo** | Detecta issues temprano |

---

## Como Activarlo

### Metodo 1: Invocacion Natural

```
Activa el goreview-watcher para monitorear mis cambios
```

### Metodo 2: Mencion Directa

```
@goreview-watcher monitorea este proyecto
```

### Metodo 3: Via Plugin

Si tienes el plugin instalado, el watcher esta disponible automaticamente.

### Verificar que esta Activo

```
/tasks
```

Deberia mostrar:
```
goreview-watcher (background) - Running
```

### Detener el Watcher

```
Para el goreview-watcher
```

O:
```
/tasks kill goreview-watcher
```

---

## Como Funciona

### Ciclo de Monitoreo

```
┌─────────────────────────────────────────────────────────┐
│                    Background Watcher                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  1. Check for changes                                   │
│     └─> git diff --staged --name-only                   │
│                                                         │
│  2. If changes detected:                                │
│     └─> goreview review --staged --preset=minimal       │
│                                                         │
│  3. Filter results:                                     │
│     ├─> Critical/Error → Report immediately             │
│     └─> Warning/Info → Batch for later                  │
│                                                         │
│  4. Wait 30 seconds                                     │
│     └─> Repeat from step 1                              │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Logica de Reporte

| Severidad | Accion |
|-----------|--------|
| **Critical** | Reporta inmediatamente via `systemMessage` |
| **Error** | Reporta inmediatamente via `systemMessage` |
| **Warning** | Acumula y reporta periodicamente |
| **Info** | Acumula, no reporta a menos que se pida |

### Uso de Recursos

- **Modelo**: Haiku (rapido y economico)
- **Polling**: Cada 30 segundos
- **Cache**: Usa resultados cacheados cuando es posible
- **Scope**: Solo cambios staged (no todo el proyecto)

---

## Configuracion

### Archivo del Agente

El watcher se define en:
```
claude-code-plugin/agents/goreview-watcher/AGENT.md
```

### Parametros Configurables

```yaml
---
name: goreview-watcher
description: Background code review watcher
tools: Read, Bash(goreview:*, git:*)
model: haiku           # Modelo a usar
permissionMode: bypassPermissions
runInBackground: true  # Corre en background
---
```

### Personalizar el Watcher

Crea tu propia version en `.claude/agents/my-watcher/AGENT.md`:

```yaml
---
name: my-watcher
description: Custom code watcher with stricter rules
tools: Read, Bash(goreview:*, git:*)
model: haiku
runInBackground: true
---

# Custom Watcher

Monitor code with strict security focus.

## Check Command
goreview review --staged --mode=security --preset=strict --format json

## Report Thresholds
- Critical: Immediate
- Error: Immediate
- Warning: After 3 occurrences
- Info: Never

## Polling Interval
Check every 15 seconds for faster feedback.
```

### Cambiar Intervalo de Polling

Edita el agente y modifica la logica de espera:

```markdown
## Polling Strategy
- Check for changes every 15 seconds (more frequent)
- Only review if files have changed
```

### Cambiar Preset

Por defecto usa `minimal` para velocidad. Para mas cobertura:

```bash
goreview review --staged --preset=standard --format json
```

---

## Interpretando los Mensajes

### Mensaje de Issue Critico

```
⚠️ GoReview: 1 critical issue in auth.go:45 - possible SQL injection. Run /security-scan for details.
```

**Componentes:**
- `⚠️` - Indicador de alerta
- `1 critical issue` - Cantidad y severidad
- `auth.go:45` - Ubicacion
- `possible SQL injection` - Descripcion breve
- `Run /security-scan` - Accion sugerida

### Mensaje de Multiples Issues

```
⚠️ GoReview: 3 issues found (1 critical, 2 errors). Run /review for details.
```

### Mensaje de Estado Limpio

El watcher no reporta si no hay issues criticos/errores. Silencio = todo bien.

### Mensaje Periodico de Warnings

```
📋 GoReview Summary: 5 warnings accumulated in last hour. Run /stats for details.
```

---

## Mejores Practicas

### 1. Activar al Inicio de Sesion

```
Inicia el goreview-watcher al comenzar a trabajar
```

Esto asegura monitoreo continuo desde el principio.

### 2. No Ignorar Alertas

Cuando el watcher interrumpe, hay una razon. Los issues criticos deben atenderse:

```
El watcher reporto un issue critico, muestrame los detalles
```

### 3. Combinar con Review Manual

El watcher usa `preset=minimal`. Antes de commit, ejecuta un review completo:

```
/review
```

### 4. Revisar Warnings Acumulados

Periodicamente revisa los warnings:

```
/stats --period today
```

### 5. Ajustar Sensibilidad

Si recibes muchas alertas, ajusta el preset o modo:

```yaml
# En el agente
goreview review --staged --preset=minimal --mode=security
```

Solo alertara por issues de seguridad.

---

## Integracion con Flujo de Trabajo

### Desarrollo Normal

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Escribir   │ ──> │   Watcher    │ ──> │   Feedback   │
│    Codigo    │     │  Monitorea   │     │   Continuo   │
└──────────────┘     └──────────────┘     └──────────────┘
       │                                         │
       │         ┌──────────────┐                │
       └───────> │   git add    │ <──────────────┘
                 └──────────────┘
                        │
                 ┌──────────────┐
                 │   /review    │  <- Review completo
                 └──────────────┘
                        │
                 ┌──────────────┐
                 │  /commit-ai  │
                 └──────────────┘
```

### Con TDD

```
Codigo → Watcher detecta falta de tests → Alerta → Agregar tests → Watcher confirma
```

### Pre-Push

```
Watcher activo → Cambios listos → /review full → Fix issues → Push
```

---

## Troubleshooting

### El watcher no inicia

1. Verificar que el plugin esta instalado:
   ```
   /plugin list
   ```

2. Verificar que goreview funciona:
   ```bash
   goreview version
   ```

### Demasiadas alertas

Ajustar el preset a `minimal` o filtrar por modo:

```yaml
goreview review --staged --preset=minimal --mode=security
```

### No recibo alertas

1. Verificar que el watcher esta corriendo:
   ```
   /tasks
   ```

2. Verificar que hay cambios staged:
   ```bash
   git status
   ```

3. Los issues info/warning no generan alertas inmediatas

### Alto consumo de recursos

1. Aumentar intervalo de polling
2. Usar preset=minimal
3. Limitar scope a archivos especificos

---

## Comparativa con Otras Opciones

| Metodo | Automatico | Background | Recursos | Cobertura |
|--------|-----------|------------|----------|-----------|
| **Watcher** | Si | Si | Bajo | Minimal |
| `/review` | No | No | Medio | Full |
| **PostToolUse Hook** | Si | No | Bajo | Minimal |
| **MCP Tools** | No | No | Variable | Full |

El watcher es ideal para feedback continuo sin esfuerzo. Complementalo con reviews manuales antes de commits importantes.
