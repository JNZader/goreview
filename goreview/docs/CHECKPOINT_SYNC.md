# Sincronizacion con Checkpoints de Claude Code

Esta guia explica como GoReview se integra con el sistema de checkpoints de Claude Code para mantener un historial de reviews sincronizado con el estado del codigo.

---

## Tabla de Contenidos

1. [Que son los Checkpoints?](#que-son-los-checkpoints)
2. [Como Funciona la Sincronizacion](#como-funciona-la-sincronizacion)
3. [Configuracion](#configuracion)
4. [Uso Practico](#uso-practico)
5. [Gestion de Archivos](#gestion-de-archivos)

---

## Que son los Checkpoints?

Los checkpoints son snapshots automaticos que Claude Code crea antes de cada edicion de archivo. Permiten:

- **Rewind**: Volver a un estado anterior del codigo
- **Undo granular**: Deshacer cambios especificos
- **Recuperacion**: Restaurar si algo sale mal

### Acceso a Checkpoints

```
Esc + Esc    # Abrir menu de rewind
/rewind      # Comando alternativo
```

### Opciones de Restauracion

1. **Solo conversacion**: Volver a un punto sin cambiar codigo
2. **Solo codigo**: Revertir archivos sin afectar conversacion
3. **Ambos**: Restaurar estado completo

---

## Como Funciona la Sincronizacion

### El Problema

Cuando haces `/rewind`, el codigo vuelve a un estado anterior, pero:
- Pierdes contexto de que issues habia en ese momento
- No sabes que reviews se hicieron
- Dificil entender por que se hicieron ciertos cambios

### La Solucion

GoReview guarda automaticamente el estado del review con cada checkpoint:

```
┌─────────────────────────────────────────────────────────┐
│                  Flujo de Checkpoints                   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Edit file                                              │
│      │                                                  │
│      ▼                                                  │
│  ┌─────────────┐    ┌─────────────────────────┐        │
│  │ Checkpoint  │───>│ PreCompact Hook          │        │
│  │   Created   │    │ goreview review --staged │        │
│  └─────────────┘    │ -o .claude/checkpoints/  │        │
│                     │    review-TIMESTAMP.json │        │
│                     └─────────────────────────┘        │
│                                                         │
│  Later: /rewind                                         │
│      │                                                  │
│      ▼                                                  │
│  ┌─────────────┐    ┌─────────────────────────┐        │
│  │   Restore   │───>│ Review file available   │        │
│  │    Code     │    │ in .claude/checkpoints/ │        │
│  └─────────────┘    └─────────────────────────┘        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Hooks Involucrados

#### PreCompact Hook

Se ejecuta antes de que Claude Code compacte el contexto:

```json
{
  "PreCompact": [{
    "matcher": "*",
    "hooks": [{
      "type": "command",
      "command": "goreview review --staged --format json -o .claude/checkpoints/review-$(date +%Y%m%d-%H%M%S).json"
    }]
  }]
}
```

#### SessionEnd Hook

Limpia reviews antiguos al terminar la sesion:

```json
{
  "SessionEnd": [{
    "matcher": "*",
    "hooks": [{
      "type": "command",
      "command": "find .claude/checkpoints -name 'review-*.json' -mtime +7 -delete"
    }]
  }]
}
```

---

## Configuracion

### Activar Sincronizacion

La sincronizacion viene incluida en el plugin. Para activarla manualmente:

1. Copia el archivo de hooks:
   ```bash
   cp claude-code-plugin/hooks/checkpoint-sync.json .claude/hooks/
   ```

2. O agrega a tu configuracion existente:
   ```json
   {
     "hooks": {
       "PreCompact": [{
         "matcher": "*",
         "hooks": [{
           "type": "command",
           "command": "mkdir -p .claude/checkpoints && goreview review --staged --format json -o .claude/checkpoints/review-$(date +%Y%m%d-%H%M%S).json 2>/dev/null || true"
         }]
       }]
     }
   }
   ```

### Ubicacion de Archivos

Los reviews se guardan en:
```
.claude/
└── checkpoints/
    ├── review-20241227-134500.json
    ├── review-20241227-141230.json
    └── review-20241227-153045.json
```

### Configurar Retencion

Por defecto, los archivos se eliminan despues de 7 dias. Para cambiar:

```json
{
  "SessionEnd": [{
    "hooks": [{
      "command": "find .claude/checkpoints -name 'review-*.json' -mtime +30 -delete"
    }]
  }]
}
```

- `+7` = mas de 7 dias
- `+30` = mas de 30 dias
- `+0` = mas de 0 dias (eliminar todo)

### Agregar a .gitignore

Los checkpoints son locales, no deben subirse al repo:

```gitignore
# Claude Code checkpoints
.claude/checkpoints/
```

---

## Uso Practico

### Escenario 1: Rewind y Entender

Hiciste varios cambios y quieres volver atras:

1. Ejecutar `/rewind`
2. Seleccionar el punto de restauracion
3. Ver el review asociado:
   ```bash
   cat .claude/checkpoints/review-20241227-134500.json | jq '.summary'
   ```

Ahora sabes que issues habia en ese momento.

### Escenario 2: Comparar Estados

Quieres ver como evoluciono la calidad:

```bash
# Review mas antiguo
cat .claude/checkpoints/review-20241227-100000.json | jq '.total_issues'

# Review mas reciente
cat .claude/checkpoints/review-20241227-160000.json | jq '.total_issues'
```

### Escenario 3: Recuperar Contexto

Despues de un `/rewind`, quieres saber por que hiciste ciertos cambios:

```bash
# Ver todos los reviews del dia
ls -la .claude/checkpoints/

# Ver issues especificos
cat .claude/checkpoints/review-*.json | jq '.files[].response.issues[] | select(.severity == "critical")'
```

### Escenario 4: Auditar Historial

Ver tendencia de issues en el tiempo:

```bash
for f in .claude/checkpoints/review-*.json; do
  echo "$(basename $f): $(cat $f | jq '.total_issues')"
done
```

---

## Gestion de Archivos

### Ver Checkpoints Guardados

```bash
ls -lh .claude/checkpoints/
```

### Ver Resumen de un Checkpoint

```bash
cat .claude/checkpoints/review-YYYYMMDD-HHMMSS.json | jq '{
  timestamp: .metadata.review_date,
  files: (.files | length),
  issues: .total_issues,
  score: .average_score
}'
```

### Limpiar Checkpoints Manualmente

```bash
# Eliminar todos
rm -rf .claude/checkpoints/*

# Eliminar mas antiguos que 3 dias
find .claude/checkpoints -name '*.json' -mtime +3 -delete
```

### Exportar para Analisis

```bash
# Combinar todos en un archivo
cat .claude/checkpoints/*.json | jq -s '.' > all-reviews.json

# Generar CSV de issues
cat .claude/checkpoints/*.json | jq -r '
  .files[].response.issues[] |
  [.severity, .type, .message] |
  @csv
' > issues-history.csv
```

---

## Formato de Archivo

Cada archivo de checkpoint contiene:

```json
{
  "metadata": {
    "project_name": "my-project",
    "branch": "feature/auth",
    "commit": "abc123",
    "review_date": "2024-12-27T13:45:00Z",
    "review_mode": "staged"
  },
  "summary": {
    "files_reviewed": 3,
    "total_issues": 5,
    "by_severity": {
      "critical": 0,
      "error": 2,
      "warning": 2,
      "info": 1
    },
    "average_score": 78
  },
  "files": [
    {
      "file": "src/auth/handler.go",
      "response": {
        "issues": [...],
        "score": 75
      }
    }
  ],
  "duration": "2.5s"
}
```

---

## Integracion con Git

### Pre-Push Hook

Usar el ultimo checkpoint como validacion:

```bash
#!/bin/bash
# .git/hooks/pre-push

LATEST=$(ls -t .claude/checkpoints/*.json 2>/dev/null | head -1)
if [ -n "$LATEST" ]; then
  CRITICAL=$(cat "$LATEST" | jq '.summary.by_severity.critical')
  if [ "$CRITICAL" -gt 0 ]; then
    echo "Error: $CRITICAL critical issues in last review"
    exit 1
  fi
fi
```

### Commit con Contexto

Incluir resumen del ultimo review en commit:

```bash
LATEST=$(ls -t .claude/checkpoints/*.json 2>/dev/null | head -1)
SUMMARY=$(cat "$LATEST" | jq -r '"Score: \(.summary.average_score)/100, Issues: \(.summary.total_issues)"')
git commit -m "feat: implement auth

Review: $SUMMARY"
```

---

## Troubleshooting

### Los checkpoints no se crean

1. Verificar que el hook esta configurado:
   ```
   /config hooks
   ```

2. Verificar que hay cambios staged:
   ```bash
   git status
   ```

3. Probar manualmente:
   ```bash
   mkdir -p .claude/checkpoints
   goreview review --staged --format json -o .claude/checkpoints/test.json
   ```

### Demasiados archivos

Reducir retencion en SessionEnd hook:

```json
{
  "command": "find .claude/checkpoints -name '*.json' -mtime +3 -delete"
}
```

### Archivos muy grandes

Usar formato compacto:

```bash
goreview review --staged --format json | gzip > .claude/checkpoints/review-$(date +%s).json.gz
```

### No tengo el directorio .claude

Crear manualmente:

```bash
mkdir -p .claude/checkpoints
```

---

## Comparativa

| Caracteristica | Checkpoints Solo | + GoReview Sync |
|----------------|------------------|-----------------|
| Restaurar codigo | Si | Si |
| Restaurar conversacion | Si | Si |
| Ver issues en ese punto | No | Si |
| Entender "por que" | No | Si |
| Auditar calidad | No | Si |
| Tendencias | No | Si |
