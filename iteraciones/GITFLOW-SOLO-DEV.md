# GitFlow para Solo Developers

## Filosofia: Atomic Sequential Merges

Este flujo esta optimizado para **un solo developer** trabajando en un proyecto. La clave es:

```
1 Commit de Guia = 1 Feature Branch = 1 PR = 1 Merge
```

**Principios:**
- **WIP Limit = 1**: Solo una rama activa a la vez
- **Self-Merge**: Tu apruebas y mergeas tus propios PRs
- **Atomic Commits**: Cada commit es autocontenido y funcional
- **Sequential**: No hay PRs paralelos esperando

---

## Estructura de Ramas

```
main (produccion estable)
  │
  └── develop (integracion continua)
          │
          └── feature/XX-YY-slug (efimera, 1 por iteracion)
```

| Rama | Proposito | Duracion |
|------|-----------|----------|
| `main` | Produccion, releases taggeados | Permanente |
| `develop` | Integracion, siempre deployable | Permanente |
| `feature/XX-YY-slug` | Un commit de una guia | Horas |

---

## Nomenclatura de Ramas

```
feature/XX-YY-slug

XX = Numero de iteracion (00, 01, 02...)
YY = Numero de commit dentro de la iteracion (01, 02, 03...)
slug = Descripcion corta en kebab-case
```

**Ejemplos:**
```bash
feature/00-01-project-structure
feature/00-02-go-module
feature/01-01-main-entrypoint
feature/01-02-root-command
feature/03-05-diff-parser
```

---

## Flujo Completo por Commit

### Paso 1: Preparar Rama

```bash
# Asegurate de estar en develop actualizado
git checkout develop
git pull origin develop

# Crear rama feature para este commit especifico
git checkout -b feature/00-01-project-structure
```

### Paso 2: Implementar y Commit

```bash
# Hacer los cambios segun la guia
# ... crear archivos, codigo, etc ...

# Agregar archivos
git add .

# Commit con mensaje exacto de la guia
git commit -m "chore: create initial project structure

- Create goreview directory for Go CLI
- Create integrations/github-app for Node.js app
- Create goreview-rules for YAML rules
- Create docs directory for documentation"
```

### Paso 3: Push y Crear PR

```bash
# Push de la rama
git push -u origin feature/00-01-project-structure

# Crear PR con GitHub CLI
gh pr create \
  --base develop \
  --title "chore: create initial project structure" \
  --body "## Commit 0.1: Crear estructura de directorios

### Cambios
- Estructura base del proyecto creada
- Directorios para CLI Go, GitHub App, reglas

### Verificacion
\`\`\`bash
find . -type d | head -20
\`\`\`

### Parte de
Iteracion 00: Inicializacion del Proyecto"
```

### Paso 4: Verificar y Merge

```bash
# Ver estado del PR
gh pr status

# Si CI pasa (o no hay CI aun), merge
gh pr merge --squash --delete-branch

# O si prefieres mantener el commit individual
gh pr merge --merge --delete-branch
```

### Paso 5: Sincronizar y Continuar

```bash
# Volver a develop y actualizar
git checkout develop
git pull origin develop

# Listo para el siguiente commit
git checkout -b feature/00-02-go-module
```

---

## Script de Automatizacion

Crea este script como `scripts/atomic-commit.sh`:

```bash
#!/bin/bash
# Uso: ./scripts/atomic-commit.sh <iter> <commit> <slug> "<mensaje>"
# Ejemplo: ./scripts/atomic-commit.sh 00 01 project-structure "chore: create initial project structure"

ITER=$1
COMMIT=$2
SLUG=$3
MSG=$4

BRANCH="feature/${ITER}-${COMMIT}-${SLUG}"

# 1. Actualizar develop
git checkout develop
git pull origin develop

# 2. Crear rama
git checkout -b "$BRANCH"

echo ""
echo "==================================="
echo "Rama creada: $BRANCH"
echo "==================================="
echo ""
echo "Ahora:"
echo "1. Implementa los cambios segun la guia"
echo "2. git add ."
echo "3. git commit -m \"$MSG\""
echo "4. Ejecuta: ./scripts/push-and-merge.sh"
echo ""
```

Script `scripts/push-and-merge.sh`:

```bash
#!/bin/bash
# Ejecutar despues de hacer commit

BRANCH=$(git branch --show-current)

# Push
git push -u origin "$BRANCH"

# Crear PR y merge automatico
gh pr create --base develop --fill
gh pr merge --squash --delete-branch

# Volver a develop
git checkout develop
git pull origin develop

echo ""
echo "==================================="
echo "Commit mergeado exitosamente!"
echo "Listo para el siguiente commit"
echo "==================================="
```

---

## Flujo Visual

```
develop ─────●─────●─────●─────●─────●─────●───────>
             │     │     │     │     │     │
             │     │     │     │     │     └─ feature/00-05-readme
             │     │     │     │     └─ feature/00-04-makefile
             │     │     │     └─ feature/00-03-gitignore
             │     │     └─ feature/00-02-go-module
             │     └─ feature/00-01-project-structure
             │
         (inicio)

Cada feature branch:
1. Se crea desde develop
2. Tiene 1 solo commit
3. Se mergea inmediatamente
4. Se elimina despues del merge
```

---

## Commits de la Iteracion 00

| # | Rama | Mensaje |
|---|------|---------|
| 0.1 | `feature/00-01-project-structure` | `chore: create initial project structure` |
| 0.2 | `feature/00-02-go-module` | `chore(go): initialize go module` |
| 0.3 | `feature/00-03-gitignore-editor` | `chore: add gitignore and editor configuration` |
| 0.4 | `feature/00-04-makefile` | `chore: add Makefile with build targets` |
| 0.5 | `feature/00-05-readme` | `docs: add project README` |

---

## Commits de la Iteracion 01

| # | Rama | Mensaje |
|---|------|---------|
| 1.1 | `feature/01-01-main-entrypoint` | `feat(cli): add main entry point` |
| 1.2 | `feature/01-02-root-command` | `feat(cli): add root command with cobra` |
| 1.3 | `feature/01-03-version-command` | `feat(cli): add version command` |
| 1.4 | `feature/01-04-version-tests` | `test(cli): add version command tests` |
| 1.5 | `feature/01-05-golangci-config` | `chore(lint): add golangci-lint configuration` |

---

## Por Que Este Flujo?

### Ventajas para Solo Developer

1. **Sin conflictos de merge**: Solo una rama activa
2. **Historial limpio**: Cada PR es un commit logico
3. **Facil de revertir**: Un PR = un cambio atomico
4. **Sin overhead**: No esperas reviews de otros
5. **CI simple**: Solo necesitas validar develop

### Comparacion con Otros Flujos

| Aspecto | GitFlow Clasico | Este Flujo |
|---------|-----------------|------------|
| Ramas activas | Multiples | 1 |
| Duracion de feature | Dias/semanas | Horas |
| Merge conflicts | Frecuentes | Ninguno |
| PRs pendientes | Varios | 0-1 |
| Complejidad | Alta | Baja |

---

## Configuracion Inicial

### 1. Crear Repositorio

```bash
# En GitHub
gh repo create ai-toolkit --public --source=. --remote=origin

# O clonar existente
git clone https://github.com/TU-USUARIO/ai-toolkit.git
```

### 2. Configurar Ramas

```bash
# Crear develop si no existe
git checkout -b develop
git push -u origin develop

# Configurar develop como rama por defecto en GitHub
gh repo edit --default-branch develop
```

### 3. Configurar Branch Protection (Opcional)

```bash
# Proteger main (requiere PR para merge)
gh api repos/{owner}/{repo}/branches/main/protection \
  -X PUT \
  -H "Accept: application/vnd.github+json" \
  -f "required_status_checks=null" \
  -f "enforce_admins=false" \
  -f "required_pull_request_reviews=null" \
  -f "restrictions=null"
```

---

## GitHub Actions (CI Basico)

Crea `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  pull_request:
    branches: [develop, main]
  push:
    branches: [develop]

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        working-directory: ./goreview
        run: make build

      - name: Test
        working-directory: ./goreview
        run: make test

      - name: Lint
        working-directory: ./goreview
        run: |
          go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
          make lint
```

---

## Troubleshooting

### Error: "Branch already exists"

```bash
# Eliminar rama local
git branch -D feature/00-01-project-structure

# Eliminar rama remota
git push origin --delete feature/00-01-project-structure
```

### Error: "PR merge conflicts"

```bash
# Actualizar tu rama con develop
git checkout feature/00-01-project-structure
git rebase develop
git push --force-with-lease
```

### Error: "gh command not found"

```bash
# Instalar GitHub CLI
# Windows (winget)
winget install GitHub.cli

# macOS
brew install gh

# Linux
sudo apt install gh

# Autenticar
gh auth login
```

---

## Resumen del Flujo

```
Para cada commit de la guia:

1. git checkout develop && git pull
2. git checkout -b feature/XX-YY-slug
3. [hacer cambios]
4. git add . && git commit -m "mensaje"
5. git push -u origin feature/XX-YY-slug
6. gh pr create --base develop --fill
7. gh pr merge --squash --delete-branch
8. git checkout develop && git pull

Repetir para el siguiente commit.
```

---

## Siguiente Paso

Aplica este flujo a tu primera iteracion: **[00-INICIALIZACION.md](00-INICIALIZACION.md)**
