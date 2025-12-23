# Iteracion 01: CLI Basico con Cobra

## Objetivos

Al completar esta iteracion tendras:
- Entry point funcional (`main.go`)
- Comando root configurado con Cobra
- Comando `version` funcionando
- Tests unitarios del comando version
- CLI compilable y ejecutable

## Prerequisitos

- Iteracion 00 completada
- Estructura de directorios creada
- `go.mod` inicializado

## Tiempo Estimado: 4 horas

---

## Commit 1.1: Crear entry point main.go

**Mensaje de commit:**
```
feat(cli): add main entry point

- Create main.go with minimal setup
- Import commands package
- Execute root command
```

**Archivos a crear:**

### 1. `goreview/cmd/goreview/main.go`

```go
// Package main is the entry point for the goreview CLI.
//
// This file is intentionally minimal - all logic lives in the commands package.
// The main function only initializes and executes the root command.
package main

import (
	"os"

	"github.com/TU-USUARIO/ai-toolkit/goreview/cmd/goreview/commands"
)

func main() {
	// Execute the root command
	// If there's an error, Cobra will print it and we exit with code 1
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Verificacion:**
```bash
cd goreview

# Esto fallara porque commands.Execute() no existe aun
# pero valida que la sintaxis es correcta
go build ./cmd/goreview/... 2>&1 | head -5
# Esperado: error sobre commands.Execute no definido
```

**Explicacion didactica:**

El entry point en Go es siempre `func main()` en `package main`. Mantenemos este archivo minimal por varias razones:

1. **Testabilidad**: La logica en `commands/` se puede testear sin ejecutar `main()`
2. **Separacion de concerns**: `main.go` solo orquesta, no implementa
3. **Convencion Go**: El patron `cmd/<app>/main.go` es estandar

El `os.Exit(1)` es importante: indica al sistema operativo que hubo un error (util para scripts y CI).

---

## Commit 1.2: Crear comando root

**Mensaje de commit:**
```
feat(cli): add root command with cobra

- Create root command with description
- Add global flags (config, verbose, quiet)
- Add persistent pre-run for initialization
- Setup cobra command hierarchy
```

**Archivos a crear:**

### 1. `goreview/cmd/goreview/commands/root.go`

```go
// Package commands contains all CLI commands for goreview.
//
// This package uses the Cobra library for CLI management.
// Each command is defined in its own file and registered in init().
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile holds the path to the config file (from --config flag)
	cfgFile string

	// verbose enables detailed output
	verbose bool

	// quiet suppresses all output except errors
	quiet bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goreview",
	Short: "AI-powered code review tool",
	Long: `GoReview is a CLI tool that uses AI to review your code changes.

It analyzes diffs, identifies potential issues, and provides actionable feedback
on bugs, security vulnerabilities, performance problems, and best practices.

Examples:
  # Review staged changes
  goreview review

  # Review a specific commit
  goreview review --commit HEAD~1

  # Review changes compared to a branch
  goreview review --base main

  # Generate a commit message
  goreview commit

  # Show current configuration
  goreview config show`,

	// SilenceUsage prevents printing usage on errors
	// We want clean error messages, not the full help text
	SilenceUsage: true,

	// SilenceErrors lets us handle errors ourselves
	SilenceErrors: true,

	// PersistentPreRunE runs before any command (including subcommands)
	// Use this for initialization that all commands need
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flags are available to this command and all subcommands
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .goreview.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except errors")

	// Bind flags to viper for config file support
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
}

// initializeConfig reads in config file and ENV variables if set.
func initializeConfig() error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory and home directory
		viper.SetConfigName(".goreview")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	// Read environment variables that match
	// GOREVIEW_PROVIDER_NAME -> provider.name
	viper.SetEnvPrefix("GOREVIEW")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is not an error - we have defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	if verbose && !quiet {
		if viper.ConfigFileUsed() != "" {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	}

	return nil
}

// isVerbose returns true if verbose mode is enabled
func isVerbose() bool {
	return verbose && !quiet
}

// isQuiet returns true if quiet mode is enabled
func isQuiet() bool {
	return quiet
}
```

**Verificacion:**
```bash
cd goreview

# Ahora deberia compilar (aunque sin hacer nada util)
go build -o build/goreview ./cmd/goreview

# Ejecutar - mostrara help
./build/goreview

# Verificar flags
./build/goreview --help
```

**Explicacion didactica:**

Cobra organiza CLIs en una jerarquia de comandos:

```
rootCmd (goreview)
├── reviewCmd (goreview review)
├── commitCmd (goreview commit)
├── configCmd (goreview config)
│   └── showCmd (goreview config show)
└── versionCmd (goreview version)
```

**Conceptos clave:**

1. **PersistentFlags**: Disponibles en el comando y TODOS sus subcomandos
2. **Flags**: Solo disponibles en ESE comando
3. **PersistentPreRunE**: Se ejecuta ANTES de cualquier comando (bueno para init)
4. **SilenceUsage/Errors**: Control fino de output de errores

**Viper integration:**

Viper lee configuracion de multiples fuentes (en orden de prioridad):
1. Flags de CLI (`--config`)
2. Variables de entorno (`GOREVIEW_*`)
3. Archivo de configuracion (`.goreview.yaml`)
4. Valores por defecto

---

## Commit 1.3: Agregar comando version

**Mensaje de commit:**
```
feat(cli): add version command

- Create version command showing build info
- Add version, commit, and build date variables
- Support JSON output format
- Add --short flag for minimal output
```

**Archivos a crear:**

### 1. `goreview/cmd/goreview/commands/version.go`

```go
package commands

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information - these are set at build time via ldflags
// See Makefile for how these are injected
var (
	// Version is the semantic version (e.g., "1.0.0")
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is the date the binary was built
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Print detailed version information about the goreview binary.

This includes the version number, git commit hash, build date,
and Go runtime information.

Examples:
  # Print full version info
  goreview version

  # Print only version number
  goreview version --short

  # Print version as JSON
  goreview version --json`,

	// No arguments expected
	Args: cobra.NoArgs,

	// RunE returns an error (better than Run which panics)
	RunE: runVersion,
}

// Flags for version command
var (
	versionShort bool
	versionJSON  bool
)

func init() {
	// Register version command under root
	rootCmd.AddCommand(versionCmd)

	// Local flags for this command only
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "print only version number")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "output as JSON")
}

// VersionInfo holds all version information
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// runVersion implements the version command logic
func runVersion(cmd *cobra.Command, args []string) error {
	info := VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	// Short output - just version number
	if versionShort {
		fmt.Println(info.Version)
		return nil
	}

	// JSON output
	if versionJSON {
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal version info: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Default: human-readable output
	fmt.Printf("goreview version %s\n", info.Version)
	fmt.Printf("  Commit:     %s\n", info.Commit)
	fmt.Printf("  Built:      %s\n", info.BuildDate)
	fmt.Printf("  Go version: %s\n", info.GoVersion)
	fmt.Printf("  OS/Arch:    %s/%s\n", info.OS, info.Arch)

	return nil
}

// GetVersionInfo returns the current version info (useful for other packages)
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}
```

**Verificacion:**
```bash
cd goreview

# Recompilar
go build -o build/goreview ./cmd/goreview

# Probar version
./build/goreview version

# Output esperado:
# goreview version dev
#   Commit:     unknown
#   Built:      unknown
#   Go version: go1.23.0
#   OS/Arch:    linux/amd64

# Probar con ldflags (simula CI)
go build -ldflags "-X 'github.com/TU-USUARIO/ai-toolkit/goreview/cmd/goreview/commands.Version=1.0.0' -X 'github.com/TU-USUARIO/ai-toolkit/goreview/cmd/goreview/commands.Commit=abc123'" -o build/goreview ./cmd/goreview

./build/goreview version
# Ahora muestra Version=1.0.0, Commit=abc123

# Probar flags
./build/goreview version --short
# Output: 1.0.0

./build/goreview version --json
# Output: JSON formateado
```

**Explicacion didactica:**

**LDFLAGS para inyectar version:**

Go permite sobrescribir variables en tiempo de compilacion:

```bash
go build -ldflags "-X 'package.Variable=value'" ...
```

Esto es util porque:
1. No necesitas modificar codigo para cada release
2. CI/CD puede inyectar el commit hash automaticamente
3. El binario "sabe" de que version es

**Patron Command con RunE:**

Usamos `RunE` (que retorna error) en lugar de `Run` (que no retorna nada):
- Errores se manejan limpiamente
- Podemos testear el retorno
- Cobra muestra el error al usuario

**runtime package:**

`runtime.Version()`, `runtime.GOOS`, `runtime.GOARCH` dan informacion del runtime de Go. Util para debugging y reportes de bugs.

---

## Commit 1.4: Agregar tests del comando version

**Mensaje de commit:**
```
test(cli): add version command tests

- Test default output format
- Test short flag
- Test JSON flag
- Test GetVersionInfo function
```

**Archivos a crear:**

### 1. `goreview/cmd/goreview/commands/version_test.go`

```go
package commands

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
)

// TestVersionCommand tests the version command output
func TestVersionCommand(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate

	// Set test values
	Version = "1.2.3"
	Commit = "abc123def"
	BuildDate = "2024-01-15T10:00:00Z"

	// Restore after test
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	}()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:    "default output",
			args:    []string{},
			wantErr: false,
			contains: []string{
				"goreview version 1.2.3",
				"Commit:     abc123def",
				"Built:      2024-01-15T10:00:00Z",
				runtime.Version(),
			},
		},
		{
			name:     "short flag",
			args:     []string{"--short"},
			wantErr:  false,
			contains: []string{"1.2.3"},
		},
		{
			name:    "json flag",
			args:    []string{"--json"},
			wantErr: false,
			contains: []string{
				`"version": "1.2.3"`,
				`"commit": "abc123def"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			versionShort = false
			versionJSON = false

			// Create a new command for testing
			cmd := versionCmd
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Execute
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For non-error cases, we need to capture stdout differently
			// Since fmt.Printf goes to os.Stdout, not cmd.OutOrStdout()
			// We'll test the function directly instead
		})
	}
}

// TestGetVersionInfo tests the GetVersionInfo function
func TestGetVersionInfo(t *testing.T) {
	// Save original
	origVersion := Version
	Version = "test-version"
	defer func() { Version = origVersion }()

	info := GetVersionInfo()

	if info.Version != "test-version" {
		t.Errorf("GetVersionInfo().Version = %v, want %v", info.Version, "test-version")
	}

	if info.GoVersion != runtime.Version() {
		t.Errorf("GetVersionInfo().GoVersion = %v, want %v", info.GoVersion, runtime.Version())
	}

	if info.OS != runtime.GOOS {
		t.Errorf("GetVersionInfo().OS = %v, want %v", info.OS, runtime.GOOS)
	}

	if info.Arch != runtime.GOARCH {
		t.Errorf("GetVersionInfo().Arch = %v, want %v", info.Arch, runtime.GOARCH)
	}
}

// TestVersionInfoJSON tests JSON marshaling of VersionInfo
func TestVersionInfoJSON(t *testing.T) {
	info := VersionInfo{
		Version:   "1.0.0",
		Commit:    "abc123",
		BuildDate: "2024-01-15",
		GoVersion: "go1.23.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal VersionInfo: %v", err)
	}

	// Unmarshal back
	var decoded VersionInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal VersionInfo: %v", err)
	}

	if decoded.Version != info.Version {
		t.Errorf("Version mismatch: got %v, want %v", decoded.Version, info.Version)
	}

	// Check JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{"version", "commit", "build_date", "go_version", "os", "arch"}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing field: %s", field)
		}
	}
}

// TestVersionCommandArgs tests that version command rejects arguments
func TestVersionCommandArgs(t *testing.T) {
	cmd := versionCmd
	cmd.SetArgs([]string{"unexpected-arg"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for unexpected argument, got nil")
	}
}
```

**Verificacion:**
```bash
cd goreview

# Ejecutar tests
go test -v ./cmd/goreview/commands/...

# Output esperado:
# === RUN   TestVersionCommand
# === RUN   TestVersionCommand/default_output
# === RUN   TestVersionCommand/short_flag
# === RUN   TestVersionCommand/json_flag
# --- PASS: TestVersionCommand
# === RUN   TestGetVersionInfo
# --- PASS: TestGetVersionInfo
# === RUN   TestVersionInfoJSON
# --- PASS: TestVersionInfoJSON
# PASS

# Con coverage
go test -v -cover ./cmd/goreview/commands/...
```

**Explicacion didactica:**

**Patrones de testing en Go:**

1. **Table-driven tests**: Definimos casos en un slice y los iteramos
   ```go
   tests := []struct{...}{...}
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {...})
   }
   ```

2. **Subtests con t.Run()**: Cada caso es un subtest independiente
   - Se pueden ejecutar individualmente: `go test -run TestVersionCommand/short_flag`
   - Mejor reporte de errores

3. **Setup/Teardown con defer**:
   ```go
   orig := SomeVar
   SomeVar = "test value"
   defer func() { SomeVar = orig }()
   ```

4. **bytes.Buffer para capturar output**: Permite verificar lo que se imprime

**Por que testear version?**

Parece trivial, pero:
- Verifica que el build funciona
- Documenta el comportamiento esperado
- Base para tests mas complejos
- CI falla rapido si algo basico esta roto

---

## Commit 1.5: Agregar golangci-lint config

**Mensaje de commit:**
```
chore(lint): add golangci-lint configuration

- Configure linters for code quality
- Set timeout and exclusions
- Add issue severity levels
```

**Archivos a crear:**

### 1. `goreview/.golangci.yml`

```yaml
# golangci-lint configuration
# https://golangci-lint.run/usage/configuration/

run:
  # Timeout for analysis
  timeout: 5m

  # Include test files
  tests: true

  # Skip vendor, third_party, etc.
  skip-dirs:
    - vendor
    - third_party
    - testdata

  # Skip generated files
  skip-files:
    - ".*_generated\\.go$"
    - ".*\\.pb\\.go$"

linters:
  # Disable all linters by default
  disable-all: true

  # Enable specific linters
  enable:
    # Bugs
    - errcheck      # Check for unchecked errors
    - gosec         # Security issues
    - govet         # Suspicious constructs
    - staticcheck   # Static analysis

    # Style
    - gofmt         # Check formatting
    - goimports     # Check imports
    - misspell      # Spelling mistakes

    # Complexity
    - gocyclo       # Cyclomatic complexity
    - funlen        # Function length

    # Performance
    - prealloc      # Slice preallocation

    # Best practices
    - revive        # Replacement for golint
    - unconvert     # Unnecessary type conversions
    - unparam       # Unused function parameters

linters-settings:
  errcheck:
    # Check type assertions
    check-type-assertions: true
    # Check blank identifier assignments
    check-blank: true

  govet:
    # Report shadowed variables
    check-shadowing: true

  gocyclo:
    # Minimum complexity to report
    min-complexity: 15

  funlen:
    # Maximum function lines
    lines: 100
    # Maximum function statements
    statements: 50

  goimports:
    # Put local imports after 3rd party
    local-prefixes: github.com/TU-USUARIO/ai-toolkit

  revive:
    rules:
      - name: exported
        severity: warning
      - name: var-naming
        severity: warning
      - name: package-comments
        severity: warning

  gosec:
    # Exclude rules
    excludes:
      - G104 # Audit errors not checked (we use errcheck for this)

issues:
  # Maximum issues per linter
  max-issues-per-linter: 50

  # Maximum same issues
  max-same-issues: 10

  # Don't skip any linters for new code
  new: false

  # Exclude some issues in tests
  exclude-rules:
    - path: _test\.go
      linters:
        - funlen
        - gocyclo

    # Allow fmt.Printf in main
    - path: main\.go
      linters:
        - forbidigo

severity:
  # Default severity
  default-severity: error

  # Severity overrides
  rules:
    - linters:
        - revive
      severity: warning
```

**Verificacion:**
```bash
cd goreview

# Instalar golangci-lint si no esta instalado
go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest

# Ejecutar linter
golangci-lint run

# Output esperado: sin errores (o warnings menores)
# Si hay errores, corregalos antes de continuar
```

**Explicacion didactica:**

**golangci-lint** es un meta-linter que ejecuta multiples linters en paralelo:

1. **errcheck**: Detecta errores no manejados
   ```go
   // MAL
   file.Close()

   // BIEN
   if err := file.Close(); err != nil {
       log.Printf("failed to close file: %v", err)
   }
   ```

2. **gosec**: Encuentra vulnerabilidades de seguridad
   - SQL injection
   - Hardcoded credentials
   - Weak crypto

3. **govet**: Detecta errores sutiles
   - Printf con argumentos incorrectos
   - Mutex copiado incorrectamente

4. **gocyclo**: Mide complejidad ciclomatica
   - Funciones muy complejas son dificiles de testear
   - Limite de 15 es razonable

5. **revive**: Reemplaza al deprecado `golint`
   - Convenciones de nombres
   - Documentacion de exports

---

## Resumen de la Iteracion 01

### Commits realizados:
1. `feat(cli): add main entry point`
2. `feat(cli): add root command with cobra`
3. `feat(cli): add version command`
4. `test(cli): add version command tests`
5. `chore(lint): add golangci-lint configuration`

### Archivos creados:
```
goreview/
├── cmd/goreview/
│   ├── main.go
│   └── commands/
│       ├── root.go
│       ├── version.go
│       └── version_test.go
└── .golangci.yml
```

### Verificacion final:
```bash
cd goreview

# Compilar
make build

# Verificar version
./build/goreview version

# Ejecutar tests
make test

# Ejecutar linter
make lint

# Todo debe pasar sin errores
```

### Resultado esperado:
```
$ ./build/goreview version
goreview version dev
  Commit:     unknown
  Built:      unknown
  Go version: go1.23.0
  OS/Arch:    linux/amd64

$ ./build/goreview --help
GoReview is a CLI tool that uses AI to review your code changes.
...
```

---

## Siguiente Iteracion

Continua con: **[02-SISTEMA-CONFIG.md](02-SISTEMA-CONFIG.md)**

En la siguiente iteracion crearemos:
- Estructuras de configuracion
- Loader con Viper
- Valores por defecto
- Validacion de config
- Comando `config show`
