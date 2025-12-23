# GoReview GitHub App

GitHub App para reviews automaticos de codigo con IA en Pull Requests.

## Caracteristicas

- **Reviews automaticos** - Analiza PRs cuando se abren o actualizan
- **Comentarios en linea** - Agrega sugerencias directamente en el codigo
- **Multiples proveedores** - Soporta Ollama (local) y OpenAI
- **Rate limiting** - Proteccion contra abuso
- **Cache** - Evita re-analizar codigo sin cambios
- **Verificacion de webhooks** - Seguridad con firma HMAC

## Requisitos

- Node.js 20+
- GitHub App configurada
- Proveedor de IA (Ollama u OpenAI)

## Instalacion

### 1. Crear GitHub App

1. Ir a GitHub Settings > Developer settings > GitHub Apps
2. Click "New GitHub App"
3. Configurar:
   - **Name**: GoReview (o el nombre que prefieras)
   - **Homepage URL**: URL de tu servidor
   - **Webhook URL**: `https://tu-servidor.com/webhook`
   - **Webhook secret**: Generar con `openssl rand -hex 32`

4. Permisos necesarios:
   - **Pull requests**: Read & Write
   - **Contents**: Read
   - **Checks**: Read & Write (opcional)

5. Eventos a suscribirse:
   - Pull request
   - Pull request review
   - Push (opcional)

6. Generar y descargar la clave privada

### 2. Configurar Variables de Entorno

```bash
cp .env.example .env
```

Editar `.env`:

```bash
# GitHub App
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
GITHUB_WEBHOOK_SECRET=tu-secreto-generado

# AI Provider
AI_PROVIDER=ollama
AI_MODEL=qwen2.5-coder:14b
OLLAMA_BASE_URL=http://localhost:11434

# O usar OpenAI
# AI_PROVIDER=openai
# OPENAI_API_KEY=sk-...
```

### 3. Ejecutar

#### Con Docker (Recomendado)

```bash
# Desde la raiz del proyecto
docker compose up github-app

# O solo la app
docker build -t goreview-app .
docker run -p 3000:3000 --env-file .env goreview-app
```

#### Localmente

```bash
# Instalar dependencias
npm install

# Desarrollo con hot-reload
npm run dev

# Produccion
npm run build
npm start
```

## Endpoints

| Endpoint | Metodo | Descripcion |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/health/ready` | GET | Readiness check |
| `/webhook` | POST | Webhook de GitHub |
| `/admin/stats` | GET | Estadisticas (protegido) |

## Estructura

```
src/
├── index.ts              # Entry point
├── config/               # Configuracion y validacion
├── handlers/             # Manejadores de eventos
│   ├── pullRequest.ts    # Eventos de PR
│   └── review.ts         # Logica de review
├── services/             # Servicios
│   ├── github.ts         # Cliente de GitHub
│   ├── ai.ts             # Cliente de IA
│   └── cache.ts          # Cache en memoria
├── middleware/           # Middleware Express
│   ├── errorHandler.ts   # Manejo de errores
│   ├── requestLogger.ts  # Logging de requests
│   └── rateLimit.ts      # Rate limiting
├── routes/               # Rutas HTTP
├── queue/                # Cola de procesamiento
└── utils/                # Utilidades
```

## Eventos Soportados

### Pull Request

- `opened` - PR abierto, inicia review
- `synchronize` - PR actualizado, re-analiza cambios
- `reopened` - PR reabierto

### Push (Opcional)

- Analiza commits directos a ramas protegidas

## Configuracion Avanzada

### Rate Limiting

```bash
RATE_LIMIT_RPS=10        # Requests por segundo
RATE_LIMIT_BURST=20      # Maximo concurrente
```

### Cache

```bash
CACHE_TTL=1h             # Tiempo de vida
CACHE_MAX_ENTRIES=1000   # Entradas maximas
```

### Limites de Review

```bash
REVIEW_MAX_FILES=50      # Archivos maximos por PR
REVIEW_MAX_DIFF_SIZE=500000  # Tamano maximo de diff (500KB)
REVIEW_TIMEOUT=5m        # Timeout de review
```

## Desarrollo

```bash
# Tests
npm test

# Tests con coverage
npm run test:coverage

# Linting
npm run lint

# Type checking
npm run typecheck
```

## Seguridad

- Los webhooks se verifican con firma HMAC-SHA256
- La clave privada nunca se expone en logs
- Rate limiting previene abuso
- Validacion de payloads con Zod

## Troubleshooting

### Error de verificacion de webhook

```
Error: Webhook signature verification failed
```

Verificar que `GITHUB_WEBHOOK_SECRET` coincide con el configurado en GitHub.

### Error de autenticacion

```
Error: Could not authenticate as GitHub App
```

Verificar:
1. `GITHUB_APP_ID` es correcto
2. `GITHUB_PRIVATE_KEY` incluye `-----BEGIN/END RSA PRIVATE KEY-----`
3. Los saltos de linea estan como `\n`

### Ollama no responde

```
Error: Connection refused
```

Verificar:
1. Ollama esta corriendo: `ollama serve`
2. El modelo esta descargado: `ollama pull qwen2.5-coder:14b`
3. `OLLAMA_BASE_URL` es accesible desde el contenedor
