# Guias de Iteraciones - AI-Toolkit

## Introduccion

Esta carpeta contiene guias EXTREMADAMENTE DETALLADAS para reconstruir AI-Toolkit desde cero. Cada iteracion incluye:

- **Commits especificos** con mensajes exactos (Conventional Commits)
- **Codigo COMPLETO** de cada archivo (no snippets parciales)
- **Comandos de verificacion** para validar cada paso
- **Explicaciones didacticas** del "por que" de cada decision
- **Workflow GitFlow** con ramas, PRs y merges

## Antes de Empezar

**Lee primero:** [GITFLOW-SOLO-DEV.md](GITFLOW-SOLO-DEV.md) - Flujo de trabajo optimizado para solo developers.

**Template de referencia:** [TEMPLATE-SECCION-GITFLOW.md](TEMPLATE-SECCION-GITFLOW.md) - Como aplicar GitFlow a cada iteracion.

## Roadmap Completo

```
SEMANA 1: FUNDACION (40 horas)
================================================
Iteracion 00: Inicializacion del Proyecto     [4 commits]  ~2h
Iteracion 01: CLI Basico con Cobra            [5 commits]  ~4h
Iteracion 02: Sistema de Configuracion        [6 commits]  ~6h
Iteracion 03: Integracion con Git             [7 commits]  ~8h
Iteracion 04: Providers de IA                 [9 commits]  ~10h
Iteracion 05: Motor de Review                 [5 commits]  ~8h
Iteracion 06: Sistema de Cache                [6 commits]  ~4h

SEMANA 2: FUNCIONALIDADES (35 horas)
================================================
Iteracion 07: Generacion de Reportes          [5 commits]  ~4h
Iteracion 08: Sistema de Reglas               [8 commits]  ~6h
Iteracion 09: Comando Review                  [6 commits]  ~6h
Iteracion 10: Comando Commit                  [5 commits]  ~4h
Iteracion 11: Comando Doc                     [4 commits]  ~3h
Iteracion 12: Comando Init                    [4 commits]  ~3h
Iteracion 13: GitHub App Setup                [6 commits]  ~4h
Iteracion 14: GitHub App Config               [4 commits]  ~3h

SEMANA 3: INTEGRACION Y DEPLOYMENT (25 horas)
================================================
Iteracion 15: GitHub App Services             [5 commits]  ~5h
Iteracion 16: GitHub App Webhooks             [6 commits]  ~5h
Iteracion 17: Docker Setup                    [5 commits]  ~4h
Iteracion 18: CI/CD Pipeline                  [6 commits]  ~4h
Iteracion 19: Security Hardening              [5 commits]  ~4h
Iteracion 20: Performance Optimization        [5 commits]  ~3h

SEMANA 4: MEMORIA COGNITIVA (8 horas) [FEATURE FLAG]
================================================
Iteracion 21: Sistema de Memoria              [9 commits]  ~8h

TOTAL: 125 commits | ~110 horas
================================================
```

## Diagrama de Dependencias

```
                    00-INICIALIZACION
                           |
                    01-CLI-BASICO
                           |
                    02-SISTEMA-CONFIG
                           |
            +--------------+--------------+
            |              |              |
    03-GIT-INTEG    04-PROVIDERS    08-RULES
            |              |              |
            +--------------+--------------+
                           |
                    05-MOTOR-REVIEW
                           |
            +--------------+--------------+
            |              |              |
    06-CACHE        07-REPORTES    09-CMD-REVIEW
            |              |              |
            +--------------+--------------+
                           |
            +--------------+--------------+--------------+
            |              |              |              |
    10-CMD-COMMIT  11-CMD-DOC   12-CMD-INIT   13-GITHUB-SETUP
            |              |              |              |
            +--------------+--------------+--------------+
                           |
            +--------------+--------------+
            |              |              |
    14-GITHUB-CFG   15-GITHUB-SVC   16-WEBHOOKS
            |              |              |
            +--------------+--------------+
                           |
            +--------------+--------------+
            |              |              |
    17-DOCKER       18-CI-CD        19-SECURITY
            |              |              |
            +--------------+--------------+
                           |
                    20-PERFORMANCE
                           |
                    21-MEMORIA [Feature Flag]
```

## Checkpoints de Verificacion

### Checkpoint 1: CLI Basico (Iteracion 02)
```bash
cd goreview
go build -o build/goreview ./cmd/goreview
./build/goreview version
./build/goreview config show
go test ./... -v
```
**Resultado esperado:** CLI compila, muestra version y config

### Checkpoint 2: Review Funcional (Iteracion 09)
```bash
# Crear archivo de prueba
echo "package main\nfunc main() { println(x) }" > /tmp/test.go
git add /tmp/test.go
./build/goreview review --staged
```
**Resultado esperado:** Review analiza codigo y muestra issues

### Checkpoint 3: CLI Completo (Iteracion 12)
```bash
./build/goreview review --format json --output report.json
./build/goreview commit --dry-run
./build/goreview doc --staged
./build/goreview init-project --help
```
**Resultado esperado:** Todos los comandos funcionan

### Checkpoint 4: GitHub App (Iteracion 16)
```bash
cd integrations/github-app
npm run build
npm run test
npm start &
curl http://localhost:3000/health
```
**Resultado esperado:** Server responde 200 OK

### Checkpoint 5: Deployment (Iteracion 20)
```bash
docker compose up -d
docker compose ps
curl http://localhost:3000/health
make test
```
**Resultado esperado:** Todo funciona en Docker

## Convenciones de Commits

Usamos **Conventional Commits** estricto:

```
<tipo>(<scope>): <descripcion>

[cuerpo opcional]

[footer opcional]
```

### Tipos Permitidos
- `feat`: Nueva funcionalidad
- `fix`: Correccion de bug
- `docs`: Solo documentacion
- `style`: Formato, sin cambios de logica
- `refactor`: Refactorizacion sin cambios funcionales
- `perf`: Mejoras de rendimiento
- `test`: Agregar o corregir tests
- `chore`: Mantenimiento, configs, deps
- `ci`: Cambios en CI/CD
- `security`: Mejoras de seguridad

### Scopes Comunes
- `cli`: Comandos de CLI
- `config`: Sistema de configuracion
- `git`: Integracion Git
- `providers`: Providers de IA
- `review`: Motor de review
- `cache`: Sistema de cache
- `report`: Generacion de reportes
- `rules`: Sistema de reglas
- `github-app`: GitHub App
- `docker`: Docker configs
- `ci`: CI/CD pipelines
- `memory`: Sistema de memoria cognitiva

## Como Usar Estas Guias

### Opcion 1: Construccion Paso a Paso
```bash
# 1. Leer iteracion completa primero
# 2. Ejecutar cada commit en orden
# 3. Verificar con comandos dados
# 4. Pasar a siguiente commit
```

### Opcion 2: Por Modulos
```bash
# Si solo necesitas el CLI:
# Seguir iteraciones 00-12

# Si solo necesitas GitHub App:
# Seguir iteraciones 00-02, luego 13-16
```

### Opcion 3: Referencia
```bash
# Consultar iteraciones especificas
# cuando necesites implementar algo similar
```

## Estructura de Cada Iteracion

```markdown
# Iteracion X: Nombre

## Objetivos
- Que logras al completar esta iteracion

## Prerequisitos
- Que debe estar completado antes

## Commits

### Commit X.1: Titulo
**Mensaje:**
```
tipo(scope): descripcion
```

**Archivos:**
- `path/archivo.go` - Descripcion

**Codigo:**
[Codigo COMPLETO del archivo]

**Verificacion:**
```bash
comandos para verificar
```

**Explicacion:**
Por que hacemos esto...
```

## Indice de Iteraciones

| # | Nombre | Commits | Tiempo | Estado |
|---|--------|---------|--------|--------|
| 00 | [Inicializacion](00-INICIALIZACION.md) | 4 | 2h | Core |
| 01 | [CLI Basico](01-CLI-BASICO.md) | 5 | 4h | Core |
| 02 | [Sistema Config](02-SISTEMA-CONFIG.md) | 6 | 6h | Core |
| 03 | [Git Integration](03-GIT-INTEGRACION.md) | 7 | 8h | Core |
| 04 | [Providers IA](04-PROVIDERS-IA.md) | 9 | 10h | Core |
| 05 | [Motor Review](05-MOTOR-REVIEW.md) | 5 | 8h | Core |
| 06 | [Cache System](06-SISTEMA-CACHE.md) | 6 | 4h | Core |
| 07 | [Reportes](07-GENERACION-REPORTES.md) | 5 | 4h | Feature |
| 08 | [Reglas](08-SISTEMA-REGLAS.md) | 8 | 6h | Feature |
| 09 | [Cmd Review](09-COMANDO-REVIEW.md) | 6 | 6h | Feature |
| 10 | [Cmd Commit](10-COMANDO-COMMIT.md) | 5 | 4h | Feature |
| 11 | [Cmd Doc](11-COMANDO-DOC.md) | 4 | 3h | Feature |
| 12 | [Cmd Init](12-COMANDO-INIT.md) | 4 | 3h | Feature |
| 13 | [GitHub Setup](13-GITHUB-APP-SETUP.md) | 6 | 4h | Integration |
| 14 | [GitHub Config](14-GITHUB-APP-CONFIG.md) | 4 | 3h | Integration |
| 15 | [GitHub Services](15-GITHUB-APP-SERVICES.md) | 5 | 5h | Integration |
| 16 | [Webhooks](16-GITHUB-APP-WEBHOOKS.md) | 6 | 5h | Integration |
| 17 | [Docker](17-DOCKER-SETUP.md) | 5 | 4h | Deploy |
| 18 | [CI/CD](18-CI-CD.md) | 6 | 4h | Deploy |
| 19 | [Security](19-SEGURIDAD.md) | 5 | 4h | Polish |
| 20 | [Performance](20-PERFORMANCE.md) | 5 | 3h | Polish |
| 21 | [Sistema Memoria](21-SISTEMA-MEMORIA.md) | 9 | 8h | Feature Flag |

## Prerequisitos Generales

### Software Requerido
```bash
# Go 1.23+
go version  # go1.23.0 o superior

# Node.js 20+
node --version  # v20.x

# Docker
docker --version  # 24.x+
docker compose version  # v2.x

# Git
git --version  # 2.40+

# Make
make --version

# Ollama (para LLM local)
ollama --version
ollama pull qwen2.5-coder:14b
```

### Conocimientos Requeridos
- Go basico (structs, interfaces, goroutines)
- TypeScript basico (async/await, types)
- Git (commits, branches, diffs)
- Terminal/Bash
- Docker basico
- HTTP/REST basico

## Tips para el Exito

1. **Lee la iteracion completa** antes de empezar
2. **Sigue el orden exacto** de commits
3. **Verifica cada paso** antes de continuar
4. **No saltes commits** - cada uno construye sobre el anterior
5. **Entiende el "por que"** - no solo copies codigo
6. **Haz los tests** - te ahorran tiempo debugeando
7. **Consulta el codigo real** en `ai-toolkit/` si tienes dudas

## Problemas Comunes

### Error: go.mod not found
```bash
# Asegurate de estar en goreview/
cd goreview
go mod init github.com/TU-USUARIO/ai-toolkit/goreview
```

### Error: package not found
```bash
# Ejecuta go mod tidy despues de agregar imports
go mod tidy
```

### Error: Ollama connection refused
```bash
# Inicia Ollama primero
ollama serve &
# Espera unos segundos
ollama pull qwen2.5-coder:14b
```

### Error: Permission denied
```bash
# Da permisos de ejecucion
chmod +x build/goreview
```

## Siguiente Paso

Empieza con: **[00-INICIALIZACION.md](00-INICIALIZACION.md)**
