# Propuestas de Nuevas Features para GoReview

Documento generado a partir del analisis de:
- [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills)
- [obsidian-claude-pkm](https://github.com/ballred/obsidian-claude-pkm/)
- [Continuous-Claude](https://github.com/parcadei/Continuous-Claude)

---

## Alta Prioridad

### 1. `goreview changelog`

| | |
|---|---|
| **Origen** | [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) - Changelog Generator |
| **Complejidad** | Media |

**Descripcion:** Genera CHANGELOG.md automatico analizando commits de git.

**Funcionalidad:**
- Parsear commits siguiendo Conventional Commits
- Agrupar por tipo (feat, fix, refactor, etc.)
- Generar markdown estructurado por version
- Detectar breaking changes automaticamente

**Ejemplo de uso:**
```bash
goreview changelog                    # Desde ultimo tag
goreview changelog --from=v1.0.0      # Desde version especifica
goreview changelog --unreleased       # Solo cambios sin release
```

---

### 2. Modos de Revision Especializados

| | |
|---|---|
| **Origen** | [obsidian-claude-pkm](https://github.com/ballred/obsidian-claude-pkm/) - Agentes especializados |
| **Complejidad** | Media-Alta |

**Descripcion:** Ofrecer modos de revision enfocados en aspectos especificos.

**Modos propuestos:**
| Modo | Enfoque | Flag |
|------|---------|------|
| Security | Vulnerabilidades OWASP, secrets, injections | `--mode=security` |
| Performance | N+1 queries, complejidad, memory leaks | `--mode=perf` |
| Clean Code | SOLID, DRY, naming, code smells | `--mode=clean` |
| Docs | Comentarios faltantes, JSDoc/GoDoc | `--mode=docs` |
| Tests | Cobertura, edge cases, mocking | `--mode=tests` |

**Ejemplo de uso:**
```bash
goreview review --staged --mode=security
goreview review --mode=perf,clean      # Multiples modos
```

---

### 3. `goreview stats` - Dashboard de Metricas

| | |
|---|---|
| **Origen** | [obsidian-claude-pkm](https://github.com/ballred/obsidian-claude-pkm/) - Dashboard terminal, [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) - CSV Data Summarizer |
| **Complejidad** | Media |

**Descripcion:** Mostrar estadisticas del repositorio en terminal.

**Metricas propuestas:**
- Complejidad ciclomatica por archivo
- Archivos con mas issues historicos
- Distribucion de severidades
- Tendencia de calidad (mejorando/empeorando)
- Cobertura de tests (si disponible)

**Ejemplo de uso:**
```bash
goreview stats                         # Resumen general
goreview stats --detailed              # Por archivo
goreview stats --format=json           # Para CI/CD
```

---

### 4. Artifact Index - Base de Datos de Reviews

| | |
|---|---|
| **Origen** | [Continuous-Claude](https://github.com/parcadei/Continuous-Claude) - Artifact Index con SQLite + FTS5 |
| **Complejidad** | Media-Alta |

**Descripcion:** Base de datos SQLite con busqueda full-text de reviews historicos.

**Funcionalidad:**
- Indexar todos los reviews realizados
- Busqueda por texto, archivo, severidad, fecha
- Detectar issues recurrentes en el mismo archivo/desarrollador
- Encontrar soluciones a problemas similares del pasado

**Ejemplo de uso:**
```bash
goreview search "memory leak"          # Buscar en reviews pasados
goreview search --file=auth.go         # Reviews de un archivo
goreview search --author=john          # Reviews por autor
goreview history ./src/api/            # Historial de un directorio
```

**Esquema propuesto:**
```sql
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY,
    commit_hash TEXT,
    file_path TEXT,
    issue_type TEXT,
    severity TEXT,
    message TEXT,
    suggestion TEXT,
    created_at DATETIME,
    resolved BOOLEAN
);
CREATE VIRTUAL TABLE reviews_fts USING fts5(message, suggestion);
```

---

### 5. Sistema de Handoffs entre Reviews

| | |
|---|---|
| **Origen** | [Continuous-Claude](https://github.com/parcadei/Continuous-Claude) - Sistema de Handoffs |
| **Complejidad** | Media |

**Descripcion:** Preservar contexto entre multiples rondas de review del mismo PR.

**Funcionalidad:**
- Al re-revisar un PR, cargar estado anterior automaticamente
- Mostrar: "X issues detectados, Y corregidos, Z persisten"
- No re-analizar codigo ya aprobado
- Tracking de progreso del PR

**Ejemplo output en GitHub:**
```markdown
## Review Round 3

### Progreso desde ultima revision:
- 5/8 issues corregidos
- 3 issues pendientes (2 critical, 1 warning)
- 2 nuevos archivos para revisar

### Issues persistentes:
1. [CRITICAL] SQL injection en auth.go:45 (desde Round 1)
2. [CRITICAL] Missing error handling en api.go:123 (desde Round 2)
3. [WARNING] Variable shadowing en utils.go:67 (desde Round 1)
```

---

## Media Prioridad

### 6. Root Cause Tracing

| | |
|---|---|
| **Origen** | [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) - root-cause-tracing |
| **Complejidad** | Alta |

**Descripcion:** En lugar de solo reportar el bug, rastrear hasta la causa raiz.

**Ejemplo:**
```
ANTES:
  Error: Variable 'user' puede ser null en linea 45

DESPUES:
  Error: Variable 'user' puede ser null en linea 45
  Causa raiz: El fetch en linea 23 no maneja el caso 404
  Propagacion: linea 23 -> linea 38 -> linea 45
  Sugerencia: Agregar validacion en linea 23
```

---

### 7. RAG con Documentacion Externa

| | |
|---|---|
| **Origen** | [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) - Skill Seekers, article-extractor |
| **Complejidad** | Alta |

**Descripcion:** Enriquecer el contexto de review con documentacion oficial.

**Funcionalidad:**
- Auto-detectar frameworks/librerias usadas
- Fetch de best practices desde docs oficiales
- Cachear documentacion para uso offline
- Configurar fuentes personalizadas

**Ejemplo config (.goreview.yaml):**
```yaml
rag:
  external_sources:
    - url: https://go.dev/doc/effective_go
      type: style_guide
    - url: https://owasp.org/www-project-top-ten/
      type: security
```

---

### 8. Personalidades de Revisor

| | |
|---|---|
| **Origen** | [obsidian-claude-pkm](https://github.com/ballred/obsidian-claude-pkm/) - Estilos de output |
| **Complejidad** | Baja |

**Descripcion:** Diferentes "personalidades" para el tono de los reviews.

**Personalidades:**
| Nombre | Estilo |
|--------|--------|
| `senior` | Mentoring, explica el "por que" |
| `strict` | Directo, sin rodeos, exigente |
| `friendly` | Sugerencias amables, positivo |
| `security-expert` | Paranoia saludable, peor caso |

**Ejemplo de uso:**
```bash
goreview review --personality=senior
```

---

### 9. Conexion de Issues/PRs Relacionados

| | |
|---|---|
| **Origen** | [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) - Tapestry |
| **Complejidad** | Media |

**Descripcion:** Detectar y vincular issues/PRs que tocan el mismo codigo.

**Funcionalidad:**
- Al revisar un PR, mostrar PRs anteriores que modificaron los mismos archivos
- Alertar si hay issues abiertos relacionados
- Sugerir PRs para revisar en conjunto

---

### 10. Razonamiento Historico por Commit

| | |
|---|---|
| **Origen** | [Continuous-Claude](https://github.com/parcadei/Continuous-Claude) - Sistema de Razonamiento Historico |
| **Complejidad** | Media |

**Descripcion:** Guardar analisis detallado por cada commit revisado.

**Estructura:**
```
.git/goreview/
└── commits/
    └── a1b2c3d/
        ├── analysis.md      # Resumen del analisis
        ├── issues.json      # Issues detectados
        └── context.json     # Contexto usado
```

**Ejemplo de uso:**
```bash
goreview recall "authentication"       # Buscar analisis pasados
goreview recall --commit=a1b2c3d       # Ver analisis de commit especifico
goreview history --file=auth.go        # Historial de un archivo
```

---

### 11. StatusLine para GitHub App

| | |
|---|---|
| **Origen** | [Continuous-Claude](https://github.com/parcadei/Continuous-Claude) - StatusLine Inteligente |
| **Complejidad** | Baja |

**Descripcion:** Mostrar estado del PR con advertencias escalonadas.

**Indicadores:**
- Porcentaje de issues resueltos
- Tiempo desde ultima revision
- Advertencias por inactividad

**Ejemplo en PR:**
```
GoReview Status: 3/8 issues resolved (37%)
Last review: 2 days ago
Warning: 2 critical issues pending > 48h
```

---

### 12. Workflow TDD Obligatorio

| | |
|---|---|
| **Origen** | [Continuous-Claude](https://github.com/parcadei/Continuous-Claude) - Workflow TDD |
| **Complejidad** | Media |

**Descripcion:** Flag que bloquea aprobacion si no hay tests.

**Reglas:**
- Codigo nuevo de produccion requiere tests correspondientes
- La cobertura no puede bajar del umbral configurado
- Detectar archivos de produccion sin archivo de test asociado

**Ejemplo config (.goreview.yaml):**
```yaml
rules:
  require_tests:
    enabled: true
    min_coverage: 80
    block_on_coverage_drop: true
    exclude:
      - "**/*_generated.go"
      - "**/mocks/**"
```

**Ejemplo de uso:**
```bash
goreview review --require-tests        # Activar verificacion
```

---

## Baja Prioridad (Nice to Have)

### 13. `goreview plan`

| | |
|---|---|
| **Origen** | [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) - Review Implementing |
| **Complejidad** | Media |

**Descripcion:** Revisar planes/diseños antes de escribir codigo.

**Ejemplo de uso:**
```bash
goreview plan ./docs/RFC-001.md
```

---

### 14. Reglas Jerarquicas

| | |
|---|---|
| **Origen** | [obsidian-claude-pkm](https://github.com/ballred/obsidian-claude-pkm/) - Sistema de objetivos jerarquicos |
| **Complejidad** | Media |

**Descripcion:** Herencia de reglas: organizacion -> proyecto -> equipo -> archivo.

**Ejemplo config:**
```yaml
rules:
  inherit_from:
    - https://company.com/rules.yaml    # Organizacion
    - ./team-rules.yaml                  # Equipo
  override:
    max_complexity: 15                   # Proyecto
```

---

### 15. Integracion con Sistemas de Conocimiento

| | |
|---|---|
| **Origen** | [awesome-claude-skills](https://github.com/ComposioHQ/awesome-claude-skills) - NotebookLM Integration |
| **Complejidad** | Alta |

**Descripcion:** Conectar con bases de conocimiento del equipo.

**Posibles integraciones:**
- Notion
- Confluence
- Obsidian vault
- Documentacion interna

---

## Matriz de Priorizacion

| # | Feature | Origen | Impacto | Esfuerzo | Prioridad |
|---|---------|--------|---------|----------|-----------|
| 1 | Modos especializados | obsidian-claude-pkm | Alto | Medio | 1 |
| 2 | `goreview changelog` | awesome-claude-skills | Alto | Bajo | 2 |
| 3 | Sistema de Handoffs | Continuous-Claude | Alto | Medio | 3 |
| 4 | Artifact Index (SQLite) | Continuous-Claude | Alto | Medio | 4 |
| 5 | `goreview stats` | obsidian-claude-pkm | Medio | Medio | 5 |
| 6 | Personalidades | obsidian-claude-pkm | Medio | Bajo | 6 |
| 7 | StatusLine GitHub | Continuous-Claude | Medio | Bajo | 7 |
| 8 | Workflow TDD | Continuous-Claude | Medio | Medio | 8 |
| 9 | Razonamiento historico | Continuous-Claude | Medio | Medio | 9 |
| 10 | Root Cause Tracing | awesome-claude-skills | Alto | Alto | 10 |
| 11 | RAG externo | awesome-claude-skills | Alto | Alto | 11 |

---

## Resumen por Repositorio de Origen

### De awesome-claude-skills (4 features)
- `goreview changelog`
- Root Cause Tracing
- RAG con Documentacion Externa
- Conexion Issues/PRs (Tapestry)

### De obsidian-claude-pkm (4 features)
- Modos de Revision Especializados
- `goreview stats`
- Personalidades de Revisor
- Reglas Jerarquicas

### De Continuous-Claude (5 features)
- Artifact Index (SQLite + FTS5)
- Sistema de Handoffs
- Razonamiento Historico por Commit
- StatusLine para GitHub App
- Workflow TDD Obligatorio

---

## Proximos Pasos

1. [ ] Decidir cual feature implementar primero
2. [ ] Crear issues en GitHub para cada feature aprobada
3. [ ] Definir MVP para cada feature
4. [ ] Estimar iteraciones necesarias
