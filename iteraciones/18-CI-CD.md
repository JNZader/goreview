# Iteracion 18: CI/CD

## Objetivos

- GitHub Actions para CI
- Tests automatizados
- Build y lint en PRs
- Release automatico

## Tiempo Estimado: 6 horas

---

## Commit 18.1: Crear workflow de CI para GoReview

**Mensaje de commit:**
```
ci: add goreview ci workflow

- Run tests on push/PR
- Lint with golangci-lint
- Build verification
- Coverage reporting
```

### `.github/workflows/goreview-ci.yml`

```yaml
# =============================================================================
# GoReview CLI - CI Workflow
# =============================================================================

name: GoReview CI

on:
  push:
    branches: [main, develop]
    paths:
      - 'goreview/**'
      - '.github/workflows/goreview-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'goreview/**'
      - '.github/workflows/goreview-ci.yml'

defaults:
  run:
    working-directory: goreview

jobs:
  # ---------------------------------------------------------------------------
  # Lint Job
  # ---------------------------------------------------------------------------
  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache-dependency-path: goreview/go.sum

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          working-directory: goreview
          args: --timeout=5m

  # ---------------------------------------------------------------------------
  # Test Job
  # ---------------------------------------------------------------------------
  test:
    name: Test
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache-dependency-path: goreview/go.sum

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: goreview/coverage.out
          flags: goreview
          fail_ci_if_error: false

  # ---------------------------------------------------------------------------
  # Build Job
  # ---------------------------------------------------------------------------
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]

    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache-dependency-path: goreview/go.sum

      - name: Build binary
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          BINARY_NAME=goreview
          if [ "$GOOS" = "windows" ]; then
            BINARY_NAME="${BINARY_NAME}.exe"
          fi

          go build -ldflags="-w -s" -o "bin/${GOOS}-${GOARCH}/${BINARY_NAME}" ./cmd/goreview

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: goreview-${{ matrix.goos }}-${{ matrix.goarch }}
          path: goreview/bin/${{ matrix.goos }}-${{ matrix.goarch }}/
          retention-days: 7
```

---

## Commit 18.2: Crear workflow de CI para GitHub App

**Mensaje de commit:**
```
ci: add github-app ci workflow

- Run tests on push/PR
- TypeScript type checking
- ESLint
- Build verification
```

### `.github/workflows/github-app-ci.yml`

```yaml
# =============================================================================
# GitHub App - CI Workflow
# =============================================================================

name: GitHub App CI

on:
  push:
    branches: [main, develop]
    paths:
      - 'github-app/**'
      - '.github/workflows/github-app-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'github-app/**'
      - '.github/workflows/github-app-ci.yml'

defaults:
  run:
    working-directory: github-app

jobs:
  # ---------------------------------------------------------------------------
  # Lint and Type Check
  # ---------------------------------------------------------------------------
  lint:
    name: Lint & Type Check
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: github-app/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: Run ESLint
        run: npm run lint

      - name: Type check
        run: npm run typecheck

  # ---------------------------------------------------------------------------
  # Test Job
  # ---------------------------------------------------------------------------
  test:
    name: Test
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: github-app/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: Run tests
        run: npm run test:coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: github-app/coverage/lcov.info
          flags: github-app
          fail_ci_if_error: false

  # ---------------------------------------------------------------------------
  # Build Job
  # ---------------------------------------------------------------------------
  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: github-app/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: github-app-build
          path: github-app/dist/
          retention-days: 7
```

---

## Commit 18.3: Crear workflow de Docker build

**Mensaje de commit:**
```
ci: add docker build workflow

- Build Docker images on push
- Push to GitHub Container Registry
- Tag with commit SHA and version
```

### `.github/workflows/docker-build.yml`

```yaml
# =============================================================================
# Docker Build Workflow
# =============================================================================

name: Docker Build

on:
  push:
    branches: [main]
    tags:
      - 'v*'
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  GOREVIEW_IMAGE: ghcr.io/${{ github.repository_owner }}/goreview
  GITHUB_APP_IMAGE: ghcr.io/${{ github.repository_owner }}/goreview-github-app

jobs:
  # ---------------------------------------------------------------------------
  # Build GoReview Image
  # ---------------------------------------------------------------------------
  build-goreview:
    name: Build GoReview
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.GOREVIEW_IMAGE }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: ./goreview
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ github.ref_name }}
            COMMIT=${{ github.sha }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}

  # ---------------------------------------------------------------------------
  # Build GitHub App Image
  # ---------------------------------------------------------------------------
  build-github-app:
    name: Build GitHub App
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.GITHUB_APP_IMAGE }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: ./github-app
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

---

## Commit 18.4: Crear workflow de release

**Mensaje de commit:**
```
ci: add release workflow

- Triggered on version tags
- Build binaries for all platforms
- Create GitHub release
- Generate changelog
```

### `.github/workflows/release.yml`

```yaml
# =============================================================================
# Release Workflow
# =============================================================================

name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  # ---------------------------------------------------------------------------
  # Build Release Binaries
  # ---------------------------------------------------------------------------
  build-binaries:
    name: Build Binaries
    runs-on: ubuntu-latest

    strategy:
      matrix:
        include:
          - goos: linux
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: darwin
            goarch: amd64
          - goos: darwin
            goarch: arm64
          - goos: windows
            goarch: amd64

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        working-directory: goreview
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          VERSION=${GITHUB_REF_NAME#v}
          COMMIT=${GITHUB_SHA::8}
          BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

          BINARY_NAME=goreview
          if [ "$GOOS" = "windows" ]; then
            BINARY_NAME="${BINARY_NAME}.exe"
          fi

          go build -ldflags="-w -s \
            -X main.Version=${VERSION} \
            -X main.Commit=${COMMIT} \
            -X main.BuildDate=${BUILD_DATE}" \
            -o "../dist/${BINARY_NAME}" ./cmd/goreview

      - name: Create archive
        run: |
          cd dist
          ARCHIVE_NAME="goreview-${{ github.ref_name }}-${{ matrix.goos }}-${{ matrix.goarch }}"

          if [ "${{ matrix.goos }}" = "windows" ]; then
            zip "${ARCHIVE_NAME}.zip" goreview.exe
          else
            tar -czf "${ARCHIVE_NAME}.tar.gz" goreview
          fi

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: goreview-${{ matrix.goos }}-${{ matrix.goarch }}
          path: dist/goreview-*

  # ---------------------------------------------------------------------------
  # Create Release
  # ---------------------------------------------------------------------------
  create-release:
    name: Create Release
    runs-on: ubuntu-latest
    needs: build-binaries

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts
          merge-multiple: true

      - name: Generate changelog
        id: changelog
        run: |
          # Get previous tag
          PREV_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")

          if [ -n "$PREV_TAG" ]; then
            echo "## Changes since ${PREV_TAG}" > CHANGELOG.md
            echo "" >> CHANGELOG.md
            git log --pretty=format:"- %s (%h)" ${PREV_TAG}..HEAD >> CHANGELOG.md
          else
            echo "## Initial Release" > CHANGELOG.md
          fi

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v1
        with:
          body_path: CHANGELOG.md
          files: artifacts/*
          draft: false
          prerelease: ${{ contains(github.ref_name, '-') }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  # ---------------------------------------------------------------------------
  # Build and Push Docker Images
  # ---------------------------------------------------------------------------
  docker-release:
    name: Docker Release
    runs-on: ubuntu-latest
    needs: build-binaries

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push GoReview
        uses: docker/build-push-action@v5
        with:
          context: ./goreview
          push: true
          tags: |
            ghcr.io/${{ github.repository_owner }}/goreview:${{ github.ref_name }}
            ghcr.io/${{ github.repository_owner }}/goreview:latest
          build-args: |
            VERSION=${{ github.ref_name }}

      - name: Build and push GitHub App
        uses: docker/build-push-action@v5
        with:
          context: ./github-app
          push: true
          tags: |
            ghcr.io/${{ github.repository_owner }}/goreview-github-app:${{ github.ref_name }}
            ghcr.io/${{ github.repository_owner }}/goreview-github-app:latest
```

---

## Commit 18.5: Agregar workflow de seguridad

**Mensaje de commit:**
```
ci: add security scanning workflow

- Dependency vulnerability scanning
- SAST with CodeQL
- Container scanning
```

### `.github/workflows/security.yml`

```yaml
# =============================================================================
# Security Scanning Workflow
# =============================================================================

name: Security

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    # Run weekly on Monday
    - cron: '0 0 * * 1'

jobs:
  # ---------------------------------------------------------------------------
  # Go Security Scan
  # ---------------------------------------------------------------------------
  go-security:
    name: Go Security
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run govulncheck
        working-directory: goreview
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: -no-fail -fmt sarif -out gosec.sarif ./goreview/...

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif

  # ---------------------------------------------------------------------------
  # Node Security Scan
  # ---------------------------------------------------------------------------
  node-security:
    name: Node Security
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        working-directory: github-app
        run: npm ci

      - name: Run npm audit
        working-directory: github-app
        run: npm audit --audit-level=high

  # ---------------------------------------------------------------------------
  # CodeQL Analysis
  # ---------------------------------------------------------------------------
  codeql:
    name: CodeQL
    runs-on: ubuntu-latest
    permissions:
      security-events: write

    strategy:
      matrix:
        language: [go, javascript]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform Analysis
        uses: github/codeql-action/analyze@v3

  # ---------------------------------------------------------------------------
  # Container Scan
  # ---------------------------------------------------------------------------
  container-scan:
    name: Container Scan
    runs-on: ubuntu-latest
    if: github.event_name != 'pull_request'

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build image
        run: docker build -t goreview:scan ./goreview

      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: 'goreview:scan'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif
```

---

## Resumen de la Iteracion 18

### Commits:
1. `ci: add goreview ci workflow`
2. `ci: add github-app ci workflow`
3. `ci: add docker build workflow`
4. `ci: add release workflow`
5. `ci: add security scanning workflow`

### Archivos:
```
.github/workflows/
├── goreview-ci.yml
├── github-app-ci.yml
├── docker-build.yml
├── release.yml
└── security.yml
```

---

## Siguiente Iteracion

Continua con: **[19-SEGURIDAD.md](19-SEGURIDAD.md)**
