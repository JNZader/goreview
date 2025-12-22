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
