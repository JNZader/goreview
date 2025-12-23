# Template: Seccion GitFlow para Iteraciones

Este template muestra como agregar la seccion de GitFlow a cada guia de iteracion.

---

## Seccion a Agregar (Copiar y Personalizar)

Agregar esta seccion **al inicio de cada iteracion**, justo despues de "Prerequisitos":

```markdown
## Workflow GitFlow

Esta iteracion se divide en commits atomicos. Cada commit tiene su propia rama feature.

### Ramas de Esta Iteracion

| Commit | Rama | Comando |
|--------|------|---------|
| X.1 | `feature/XX-01-slug` | `git checkout -b feature/XX-01-slug` |
| X.2 | `feature/XX-02-slug` | `git checkout -b feature/XX-02-slug` |
| X.3 | `feature/XX-03-slug` | `git checkout -b feature/XX-03-slug` |

### Flujo por Commit

\`\`\`bash
# 1. Preparar rama
git checkout develop && git pull origin develop
git checkout -b feature/XX-YY-slug

# 2. Implementar cambios (segun la guia)
# ... crear archivos ...

# 3. Commit
git add .
git commit -m "tipo(scope): mensaje del commit"

# 4. Push y PR
git push -u origin feature/XX-YY-slug
gh pr create --base develop --fill

# 5. Merge y limpiar
gh pr merge --squash --delete-branch
git checkout develop && git pull
\`\`\`

> **Ver guia completa:** [GITFLOW-SOLO-DEV.md](GITFLOW-SOLO-DEV.md)

---
```

---

## Ejemplo Aplicado: Iteracion 00

```markdown
## Workflow GitFlow

Esta iteracion se divide en 5 commits atomicos. Cada commit tiene su propia rama feature.

### Ramas de Esta Iteracion

| Commit | Rama | Comando |
|--------|------|---------|
| 0.1 | `feature/00-01-project-structure` | `git checkout -b feature/00-01-project-structure` |
| 0.2 | `feature/00-02-go-module` | `git checkout -b feature/00-02-go-module` |
| 0.3 | `feature/00-03-gitignore-editor` | `git checkout -b feature/00-03-gitignore-editor` |
| 0.4 | `feature/00-04-makefile` | `git checkout -b feature/00-04-makefile` |
| 0.5 | `feature/00-05-readme` | `git checkout -b feature/00-05-readme` |

### Flujo por Commit

\`\`\`bash
# Ejemplo para Commit 0.1
git checkout develop && git pull origin develop
git checkout -b feature/00-01-project-structure

# ... crear estructura de directorios ...

git add .
git commit -m "chore: create initial project structure

- Create goreview directory for Go CLI
- Create integrations/github-app for Node.js app
- Create goreview-rules for YAML rules
- Create docs directory for documentation"

git push -u origin feature/00-01-project-structure
gh pr create --base develop --fill
gh pr merge --squash --delete-branch
git checkout develop && git pull
\`\`\`

> **Ver guia completa:** [GITFLOW-SOLO-DEV.md](GITFLOW-SOLO-DEV.md)
```

---

## Ejemplo Aplicado: Iteracion 01

```markdown
## Workflow GitFlow

Esta iteracion se divide en 5 commits atomicos.

### Ramas de Esta Iteracion

| Commit | Rama | Comando |
|--------|------|---------|
| 1.1 | `feature/01-01-main-entrypoint` | `git checkout -b feature/01-01-main-entrypoint` |
| 1.2 | `feature/01-02-root-command` | `git checkout -b feature/01-02-root-command` |
| 1.3 | `feature/01-03-version-command` | `git checkout -b feature/01-03-version-command` |
| 1.4 | `feature/01-04-version-tests` | `git checkout -b feature/01-04-version-tests` |
| 1.5 | `feature/01-05-golangci-config` | `git checkout -b feature/01-05-golangci-config` |

### Flujo por Commit

\`\`\`bash
# Ejemplo para Commit 1.1
git checkout develop && git pull origin develop
git checkout -b feature/01-01-main-entrypoint

# ... crear cmd/goreview/main.go ...

git add .
git commit -m "feat(cli): add main entry point

- Create main.go with minimal setup
- Import commands package
- Execute root command"

git push -u origin feature/01-01-main-entrypoint
gh pr create --base develop --fill
gh pr merge --squash --delete-branch
git checkout develop && git pull
\`\`\`
```

---

## Slugs Sugeridos por Iteracion

| Iter | Commits | Slugs Sugeridos |
|------|---------|-----------------|
| 00 | 5 | `project-structure`, `go-module`, `gitignore-editor`, `makefile`, `readme` |
| 01 | 5 | `main-entrypoint`, `root-command`, `version-command`, `version-tests`, `golangci-config` |
| 02 | 6 | `config-struct`, `config-loader`, `config-defaults`, `config-validation`, `config-command`, `config-tests` |
| 03 | 7 | `git-client`, `diff-types`, `diff-parser`, `hunk-parser`, `staged-changes`, `commit-range`, `git-tests` |
| 04 | 6 | `provider-interface`, `ollama-provider`, `openai-provider`, `provider-factory`, `response-parser`, `provider-tests` |
| 05 | 5 | `review-engine`, `file-analyzer`, `issue-collector`, `parallel-processing`, `engine-tests` |
| 06 | 6 | `cache-interface`, `lru-cache`, `file-cache`, `cache-key`, `cache-config`, `cache-tests` |

---

## Donde Insertar la Seccion

En cada archivo de iteracion:

```markdown
# Iteracion XX: Nombre

## Objetivos
...

## Prerequisitos
...

## Workflow GitFlow        <-- INSERTAR AQUI
...

## Tiempo Estimado: Xh     <-- (opcional, remover si no quieres estimados)

---

## Commit X.1: Titulo
...
```

---

## Tip: Agregar a Todas las Iteraciones

Script para agregar la seccion a todas las guias:

```bash
#!/bin/bash
# Mostrar donde insertar en cada archivo

for file in [0-9][0-9]-*.md; do
  echo "=== $file ==="
  grep -n "## Prerequisitos" "$file" || echo "No encontrado"
  echo ""
done
```

Luego editar manualmente cada archivo insertando la seccion correspondiente despues de "## Prerequisitos".
