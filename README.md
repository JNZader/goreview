# GoReview - AI Code Review Toolkit

Suite de herramientas para automatizar code review usando inteligencia artificial.

## Code Quality

| Metric | Status |
|--------|--------|
| Bugs | 0 |
| Vulnerabilities | 0 |
| Code Smells | 0 |
| Security Hotspots | 0 |
| Duplications | < 3% |

> Analizado con SonarQube. Cognitive complexity mantenida por debajo de 15 en todas las funciones.

## Componentes

| Componente | Descripcion | Tecnologia |
|------------|-------------|------------|
| **[GoReview CLI](./goreview/)** | Herramienta de linea de comandos | Go |
| **[GitHub App](./github-app/)** | Integracion con GitHub PRs | Node.js/TypeScript |

## Caracteristicas Principales

### Core
- **Review de codigo con IA** - Detecta bugs, vulnerabilidades, problemas de rendimiento
- **Multiples proveedores** - Ollama (local), OpenAI, Gemini, Groq, Mistral
- **Generacion de commits** - Mensajes siguiendo Conventional Commits
- **Documentacion automatica** - Changelogs y documentacion de cambios
- **Integracion GitHub** - Reviews automaticos en Pull Requests
- **Sistema de reglas** - Configurable por severidad, categoria y lenguaje
- **Cache inteligente** - Evita re-analizar codigo sin cambios
- **Multiples formatos** - Markdown, JSON, SARIF

### Avanzadas
- **Auto-fix** - Comando `goreview fix` para aplicar correcciones automaticamente
- **Comentarios interactivos** - Responde a menciones @goreview en GitHub
- **Token Budgeting** - Gestion inteligente de tokens con chunking de codigo
- **AST Parsing** - Analisis de contexto de codigo (Go, JS/TS, Python, Java, Rust)
- **RAG para Style Guides** - Integra guias de estilo del proyecto en reviews
- **Worker Pool** - Procesamiento concurrente optimizado
- **Secret Masking** - Enmascaramiento automatico de secretos en logs
- **Cola persistente** - BullMQ + Redis para alta disponibilidad

## Inicio Rapido

### CLI

```bash
# Instalar
go install github.com/JNZader/goreview/goreview/cmd/goreview@latest

# O compilar desde fuente
cd goreview && make build

# Configurar proyecto
goreview init

# Revisar cambios
goreview review --staged

# Auto-fix issues
goreview fix --staged

# Generar commit
goreview commit
```

### GitHub App

```bash
# Configurar variables de entorno
cp github-app/.env.example github-app/.env
# Editar .env con credenciales de GitHub App

# Ejecutar con Docker
docker compose up -d

# O ejecutar localmente
cd github-app && npm install && npm run dev
```

## Estructura del Proyecto

```
iatoolkit/
├── goreview/               # CLI en Go
│   ├── cmd/goreview/       # Comandos (review, commit, doc, init, fix)
│   └── internal/           # Paquetes internos
│       ├── providers/      # Ollama, OpenAI, Gemini, etc.
│       ├── review/         # Motor de analisis con Worker Pool
│       ├── rules/          # Sistema de reglas
│       ├── cache/          # Cache LRU
│       ├── git/            # Integracion Git
│       ├── report/         # Generadores de reportes
│       ├── tokenizer/      # Token budgeting y chunking
│       ├── ast/            # AST parsing multi-lenguaje
│       ├── rag/            # RAG para style guides
│       └── logger/         # Logger con secret masking
│
├── github-app/             # GitHub App en Node.js
│   └── src/
│       ├── handlers/       # Webhooks (PRs, comments)
│       ├── services/       # AI providers, GitHub client
│       ├── queue/          # BullMQ + Redis
│       └── routes/         # Endpoints HTTP
│
├── scripts/                # Scripts de utilidad
├── iteraciones/            # Documentacion de desarrollo
├── docker-compose.yml      # Desarrollo local
└── docker-compose.prod.yml # Produccion
```

## Configuracion

### CLI (.goreview.yaml)

```yaml
version: "1.0"

provider:
  name: ollama                    # ollama, openai, gemini, groq, mistral
  model: qwen2.5-coder:14b
  base_url: http://localhost:11434

review:
  max_concurrency: 5
  min_severity: warning           # info, warning, error, critical

rules:
  preset: standard                # minimal, standard, strict

cache:
  enabled: true
  ttl: 24h
```

### Variables de Entorno

```bash
# Proveedores de IA
GOREVIEW_PROVIDER_NAME=ollama
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=...
GROQ_API_KEY=...

# GitHub App
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY_PATH=./private-key.pem
GITHUB_WEBHOOK_SECRET=your-secret

# Redis (opcional, para cola persistente)
REDIS_URL=redis://localhost:6379
```

## Desarrollo

### Requisitos

- Go 1.23+
- Node.js 20+
- Docker (opcional)
- Ollama (para LLM local)

### Comandos

```bash
# Root - Docker
make dev              # Iniciar todo en desarrollo
make build            # Compilar todo
make test             # Ejecutar tests

# GoReview CLI
cd goreview
make build            # Compilar
make test             # Tests
make lint             # Linting

# GitHub App
cd github-app
npm install           # Instalar dependencias
npm run dev           # Desarrollo con hot-reload
npm test              # Tests
npm run lint          # Linting
```

## Docker

```bash
# Desarrollo (con hot-reload)
docker compose up

# Produccion
docker compose -f docker-compose.prod.yml up -d

# Solo CLI
docker build -t goreview ./goreview
docker run -v $(pwd):/app goreview review --staged

# Solo GitHub App
docker build -t goreview-app ./github-app
docker run -p 3000:3000 --env-file .env goreview-app
```

## Proveedores de IA Soportados

| Proveedor | Tipo | Modelo Recomendado |
|-----------|------|-------------------|
| Ollama | Local | qwen2.5-coder:14b |
| OpenAI | Cloud | gpt-4 |
| Gemini | Cloud | gemini-pro |
| Groq | Cloud | llama-3.1-70b-versatile |
| Mistral | Cloud | codestral-latest |

### Configurar Ollama (Recomendado)

```bash
# Instalar
curl -fsSL https://ollama.com/install.sh | sh

# Descargar modelo
ollama pull qwen2.5-coder:14b

# Iniciar servidor
ollama serve
```

## CI/CD

El proyecto incluye GitHub Actions para:

- **Lint** - golangci-lint (Go), ESLint (TypeScript)
- **Test** - Tests unitarios con coverage
- **Build** - Compilacion multi-plataforma

## Contribuir

1. Fork el repositorio
2. Crear rama (`git checkout -b feature/nueva-funcionalidad`)
3. Commit (`git commit -m 'feat: agregar funcionalidad'`)
4. Push (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## Licencia

MIT License - ver [LICENSE](LICENSE)
