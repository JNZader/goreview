# Iteracion 20: Performance y Optimizaciones

## Objetivos

- Optimizaciones de memoria en Go
- Mejoras de concurrencia
- Benchmarking y profiling
- Metricas de rendimiento

## Tiempo Estimado: 6 horas

## Prerequisitos

- Iteracion 19 completada (Seguridad)
- Conocimiento basico de profiling en Go
- Familiaridad con pprof

---

## Por que Performance?

La optimizacion de rendimiento es crucial para una herramienta CLI que:
1. **Procesa archivos grandes**: Reviews de PRs con miles de lineas
2. **Hace llamadas de red**: Latencia de API de LLM
3. **Maneja concurrencia**: Multiples archivos en paralelo
4. **Usa memoria**: Cache de resultados y contexto

### Principios de Optimizacion

1. **Medir primero**: No optimizar sin datos
2. **Identificar cuellos de botella**: Usar profiling
3. **Optimizar lo que importa**: El 80/20 aplica
4. **Mantener legibilidad**: No sacrificar claridad por microsegundos

---

## Commit 20.1: Agregar benchmarks basicos

**Mensaje de commit:**
```
perf(goreview): add basic benchmarks

- Benchmark for diff parsing
- Benchmark for cache operations
- Benchmark for rule matching
- Memory allocation tracking
```

### Por que benchmarks?

Los benchmarks de Go nos permiten:
- Medir el rendimiento de forma reproducible
- Detectar regresiones automaticamente
- Comparar diferentes implementaciones
- Identificar allocaciones de memoria

### `goreview/internal/git/diff_benchmark_test.go`

```go
// =============================================================================
// Diff Parser Benchmarks
// Mide el rendimiento del parsing de diffs
// =============================================================================

package git

import (
	"strings"
	"testing"
)

// generateDiff crea un diff sintetico para benchmarking
func generateDiff(files, hunksPerFile, linesPerHunk int) string {
	var sb strings.Builder

	for f := 0; f < files; f++ {
		sb.WriteString("diff --git a/file")
		sb.WriteString(string(rune('0' + f)))
		sb.WriteString(".go b/file")
		sb.WriteString(string(rune('0' + f)))
		sb.WriteString(".go\n")
		sb.WriteString("index abc123..def456 100644\n")
		sb.WriteString("--- a/file")
		sb.WriteString(string(rune('0' + f)))
		sb.WriteString(".go\n")
		sb.WriteString("+++ b/file")
		sb.WriteString(string(rune('0' + f)))
		sb.WriteString(".go\n")

		for h := 0; h < hunksPerFile; h++ {
			startLine := h*linesPerHunk + 1
			sb.WriteString("@@ -")
			sb.WriteString(string(rune('0' + startLine)))
			sb.WriteString(",")
			sb.WriteString(string(rune('0' + linesPerHunk)))
			sb.WriteString(" +")
			sb.WriteString(string(rune('0' + startLine)))
			sb.WriteString(",")
			sb.WriteString(string(rune('0' + linesPerHunk)))
			sb.WriteString(" @@\n")

			for l := 0; l < linesPerHunk; l++ {
				if l%3 == 0 {
					sb.WriteString("+func added")
					sb.WriteString(string(rune('0' + l)))
					sb.WriteString("() {}\n")
				} else if l%3 == 1 {
					sb.WriteString("-func removed")
					sb.WriteString(string(rune('0' + l)))
					sb.WriteString("() {}\n")
				} else {
					sb.WriteString(" func unchanged")
					sb.WriteString(string(rune('0' + l)))
					sb.WriteString("() {}\n")
				}
			}
		}
	}

	return sb.String()
}

// BenchmarkParseDiff_Small mide parsing de diff pequeno
// 1 archivo, 1 hunk, 10 lineas
func BenchmarkParseDiff_Small(b *testing.B) {
	diff := generateDiff(1, 1, 10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiff_Medium mide parsing de diff mediano
// 5 archivos, 3 hunks, 20 lineas
func BenchmarkParseDiff_Medium(b *testing.B) {
	diff := generateDiff(5, 3, 20)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiff_Large mide parsing de diff grande
// 20 archivos, 10 hunks, 50 lineas
func BenchmarkParseDiff_Large(b *testing.B) {
	diff := generateDiff(20, 10, 50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiff_Allocs rastrea allocaciones
func BenchmarkParseDiff_Allocs(b *testing.B) {
	diff := generateDiff(5, 3, 20)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}
```

### `goreview/internal/cache/cache_benchmark_test.go`

```go
// =============================================================================
// Cache Benchmarks
// Mide el rendimiento de operaciones de cache
// =============================================================================

package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// BenchmarkCache_Get mide lectura de cache
func BenchmarkCache_Get(b *testing.B) {
	c := NewLRUCache(1000, time.Hour)

	// Pre-poblar cache
	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("key-%d", i)
		c.Set(key, fmt.Sprintf("value-%d", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%500)
		c.Get(key)
	}
}

// BenchmarkCache_Set mide escritura de cache
func BenchmarkCache_Set(b *testing.B) {
	c := NewLRUCache(1000, time.Hour)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		c.Set(key, fmt.Sprintf("value-%d", i))
	}
}

// BenchmarkCache_Concurrent mide acceso concurrente
func BenchmarkCache_Concurrent(b *testing.B) {
	c := NewLRUCache(1000, time.Hour)

	// Pre-poblar
	for i := 0; i < 500; i++ {
		c.Set(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%500)
			if i%2 == 0 {
				c.Get(key)
			} else {
				c.Set(key, "new-value")
			}
			i++
		}
	})
}

// BenchmarkCache_WithEviction mide con eviccion activa
func BenchmarkCache_WithEviction(b *testing.B) {
	// Cache pequeno para forzar eviccion
	c := NewLRUCache(100, time.Hour)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		c.Set(key, fmt.Sprintf("value-%d", i))
	}
}

// BenchmarkFileCache_Save mide persistencia
func BenchmarkFileCache_Save(b *testing.B) {
	ctx := context.Background()
	fc, err := NewFileCache(b.TempDir(), time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	entry := &CacheEntry{
		Key:       "test-key",
		Value:     "test-value-with-some-content-to-simulate-real-data",
		CreatedAt: time.Now(),
		TTL:       time.Hour,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		entry.Key = fmt.Sprintf("key-%d", i)
		if err := fc.Save(ctx, entry); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFileCache_Load mide carga desde disco
func BenchmarkFileCache_Load(b *testing.B) {
	ctx := context.Background()
	fc, err := NewFileCache(b.TempDir(), time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-guardar entradas
	for i := 0; i < 100; i++ {
		entry := &CacheEntry{
			Key:       fmt.Sprintf("key-%d", i),
			Value:     "test-value",
			CreatedAt: time.Now(),
			TTL:       time.Hour,
		}
		fc.Save(ctx, entry)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		fc.Load(ctx, key)
	}
}
```

### `goreview/internal/rules/rules_benchmark_test.go`

```go
// =============================================================================
// Rules Engine Benchmarks
// Mide el rendimiento del motor de reglas
// =============================================================================

package rules

import (
	"testing"
)

// createTestRules genera reglas para benchmark
func createTestRules(count int) []Rule {
	rules := make([]Rule, count)

	for i := 0; i < count; i++ {
		rules[i] = Rule{
			ID:       fmt.Sprintf("rule-%d", i),
			Name:     fmt.Sprintf("Test Rule %d", i),
			Severity: Severity(i % 3),
			Pattern:  fmt.Sprintf("pattern%d", i),
			Enabled:  true,
			Languages: []string{"go"},
		}
	}

	return rules
}

// BenchmarkRuleEngine_Match mide matching de reglas
func BenchmarkRuleEngine_Match(b *testing.B) {
	engine := NewRuleEngine()
	rules := createTestRules(50)
	engine.LoadRules(rules)

	code := `
func example() {
	password := "secret123"
	fmt.Println(password)
	// TODO: fix this
}
`
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.Match(code, "go")
	}
}

// BenchmarkRuleEngine_LoadRules mide carga de reglas
func BenchmarkRuleEngine_LoadRules(b *testing.B) {
	rules := createTestRules(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine := NewRuleEngine()
		engine.LoadRules(rules)
	}
}

// BenchmarkRuleEngine_RegexCompilation mide compilacion de regex
func BenchmarkRuleEngine_RegexCompilation(b *testing.B) {
	patterns := []string{
		`password\s*=\s*["'][^"']+["']`,
		`api[_-]?key\s*=`,
		`TODO|FIXME|HACK`,
		`fmt\.Print(ln|f)?`,
		`panic\([^)]+\)`,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, pattern := range patterns {
			regexp.MustCompile(pattern)
		}
	}
}
```

### Verificacion

```bash
cd goreview

# Ejecutar benchmarks
go test -bench=. -benchmem ./...

# Ejecutar benchmark especifico
go test -bench=BenchmarkParseDiff -benchmem ./internal/git/

# Comparar benchmarks (requiere benchstat)
go test -bench=. -count=10 ./internal/cache/ > old.txt
# ... hacer cambios ...
go test -bench=. -count=10 ./internal/cache/ > new.txt
benchstat old.txt new.txt
```

---

## Commit 20.2: Agregar profiling

**Mensaje de commit:**
```
perf(goreview): add profiling support

- CPU profiling flag
- Memory profiling flag
- pprof HTTP server option
- Profile analysis helpers
```

### Por que profiling?

El profiling nos permite:
- Identificar funciones que consumen mas CPU
- Encontrar allocaciones excesivas
- Detectar goroutine leaks
- Optimizar hot paths

### `goreview/internal/profiler/profiler.go`

```go
// =============================================================================
// Package profiler - Profiling utilities
// Proporciona herramientas para profiling de CPU y memoria
// =============================================================================

package profiler

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // Registra handlers de pprof
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Profiler maneja la recoleccion de perfiles
type Profiler struct {
	cpuFile    *os.File
	memFile    string
	httpServer *http.Server
	startTime  time.Time
}

// Config configura el profiler
type Config struct {
	CPUProfile  string // Archivo para CPU profile
	MemProfile  string // Archivo para memory profile
	HTTPAddr    string // Direccion para pprof HTTP (ej: ":6060")
	EnableTrace bool   // Habilitar tracing
}

// New crea un nuevo profiler
func New(cfg Config) (*Profiler, error) {
	p := &Profiler{
		memFile:   cfg.MemProfile,
		startTime: time.Now(),
	}

	// Iniciar CPU profiling si se especifico archivo
	if cfg.CPUProfile != "" {
		f, err := os.Create(cfg.CPUProfile)
		if err != nil {
			return nil, fmt.Errorf("failed to create CPU profile: %w", err)
		}
		p.cpuFile = f

		if err := pprof.StartCPUProfile(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to start CPU profile: %w", err)
		}
	}

	// Iniciar servidor HTTP para pprof
	if cfg.HTTPAddr != "" {
		p.httpServer = &http.Server{
			Addr:         cfg.HTTPAddr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		}

		go func() {
			// El pprof ya esta registrado via import
			if err := p.httpServer.ListenAndServe(); err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "pprof server error: %v\n", err)
			}
		}()
	}

	return p, nil
}

// Stop detiene el profiling y guarda los resultados
func (p *Profiler) Stop() error {
	var errs []error

	// Detener CPU profiling
	if p.cpuFile != nil {
		pprof.StopCPUProfile()
		if err := p.cpuFile.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close CPU profile: %w", err))
		}
	}

	// Escribir memory profile
	if p.memFile != "" {
		// Forzar GC para obtener stats precisos
		runtime.GC()

		f, err := os.Create(p.memFile)
		if err != nil {
			errs = append(errs, fmt.Errorf("create memory profile: %w", err))
		} else {
			defer f.Close()
			if err := pprof.WriteHeapProfile(f); err != nil {
				errs = append(errs, fmt.Errorf("write memory profile: %w", err))
			}
		}
	}

	// Detener servidor HTTP
	if p.httpServer != nil {
		if err := p.httpServer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close pprof server: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("profiler stop errors: %v", errs)
	}
	return nil
}

// Stats retorna estadisticas de memoria actuales
func Stats() MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
		HeapAlloc:  m.HeapAlloc,
		HeapSys:    m.HeapSys,
		HeapIdle:   m.HeapIdle,
		HeapInuse:  m.HeapInuse,
	}
}

// MemStats contiene estadisticas de memoria
type MemStats struct {
	Alloc      uint64 // Bytes allocados actualmente
	TotalAlloc uint64 // Bytes allocados en total (acumulado)
	Sys        uint64 // Memoria obtenida del OS
	NumGC      uint32 // Numero de GC ejecutados
	HeapAlloc  uint64 // Bytes en heap allocados
	HeapSys    uint64 // Heap obtenido del OS
	HeapIdle   uint64 // Heap no usado
	HeapInuse  uint64 // Heap en uso
}

// String formatea las estadisticas
func (m MemStats) String() string {
	return fmt.Sprintf(
		"Alloc: %s, HeapAlloc: %s, Sys: %s, NumGC: %d",
		formatBytes(m.Alloc),
		formatBytes(m.HeapAlloc),
		formatBytes(m.Sys),
		m.NumGC,
	)
}

// formatBytes convierte bytes a formato legible
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
```

### `goreview/cmd/goreview/flags_profile.go`

```go
// =============================================================================
// Profile flags para comandos CLI
// Agrega flags de profiling a los comandos
// =============================================================================

package main

import (
	"github.com/spf13/cobra"
)

// ProfileFlags contiene flags de profiling
type ProfileFlags struct {
	CPUProfile  string
	MemProfile  string
	PProfAddr   string
	EnableTrace bool
}

// AddProfileFlags agrega flags de profiling a un comando
func AddProfileFlags(cmd *cobra.Command, flags *ProfileFlags) {
	cmd.Flags().StringVar(
		&flags.CPUProfile,
		"cpuprofile",
		"",
		"Write CPU profile to file",
	)

	cmd.Flags().StringVar(
		&flags.MemProfile,
		"memprofile",
		"",
		"Write memory profile to file",
	)

	cmd.Flags().StringVar(
		&flags.PProfAddr,
		"pprof-addr",
		"",
		"Enable pprof HTTP server (e.g., :6060)",
	)

	cmd.Flags().BoolVar(
		&flags.EnableTrace,
		"trace",
		false,
		"Enable execution tracing",
	)
}
```

### Actualizacion de `goreview/cmd/goreview/review.go`

```go
// Agregar al inicio del comando review:

func runReview(cmd *cobra.Command, args []string) error {
	// Iniciar profiler si se solicito
	var prof *profiler.Profiler
	if profileFlags.CPUProfile != "" || profileFlags.MemProfile != "" || profileFlags.PProfAddr != "" {
		cfg := profiler.Config{
			CPUProfile: profileFlags.CPUProfile,
			MemProfile: profileFlags.MemProfile,
			HTTPAddr:   profileFlags.PProfAddr,
		}

		var err error
		prof, err = profiler.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to start profiler: %w", err)
		}
		defer prof.Stop()

		// Log inicial de memoria
		log.Debug().
			Str("stats", profiler.Stats().String()).
			Msg("Initial memory stats")
	}

	// ... resto del comando ...

	// Al final, antes de retornar
	if prof != nil {
		log.Debug().
			Str("stats", profiler.Stats().String()).
			Msg("Final memory stats")
	}

	return nil
}
```

### Verificacion

```bash
cd goreview

# Ejecutar con CPU profiling
./build/goreview review --cpuprofile=cpu.prof .

# Analizar CPU profile
go tool pprof cpu.prof
# Comandos utiles en pprof:
# top10        - Top 10 funciones por CPU
# top -cum     - Top por tiempo acumulado
# list Match   - Ver codigo de funcion Match
# web          - Abrir visualizacion en navegador

# Ejecutar con memory profiling
./build/goreview review --memprofile=mem.prof .

# Analizar memory profile
go tool pprof mem.prof
# Comandos utiles:
# top10 -inuse_space  - Top por memoria en uso
# top10 -alloc_space  - Top por allocaciones totales

# pprof HTTP (mientras el comando corre)
./build/goreview review --pprof-addr=:6060 .
# En otro terminal:
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## Commit 20.3: Optimizar parsing de diff

**Mensaje de commit:**
```
perf(git): optimize diff parsing

- Reduce string allocations
- Use strings.Builder
- Pre-allocate slices
- Buffer reuse
```

### Por que optimizar el parser?

El parsing de diff es una operacion frecuente que:
- Se ejecuta en cada archivo del review
- Procesa potencialmente miles de lineas
- Crea muchos strings intermedios

### `goreview/internal/git/diff_optimized.go`

```go
// =============================================================================
// Optimized Diff Parser
// Version optimizada con menos allocaciones
// =============================================================================

package git

import (
	"strings"
	"sync"
)

// Pool de strings.Builder para reusar
var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// getBuilder obtiene un builder del pool
func getBuilder() *strings.Builder {
	b := builderPool.Get().(*strings.Builder)
	b.Reset()
	return b
}

// putBuilder devuelve un builder al pool
func putBuilder(b *strings.Builder) {
	// Evitar retener builders muy grandes
	if b.Cap() < 1024*64 { // 64KB max
		builderPool.Put(b)
	}
}

// ParseDiffOptimized parsea un diff con menos allocaciones
func ParseDiffOptimized(diffText string) ([]FileDiff, error) {
	if diffText == "" {
		return nil, nil
	}

	// Pre-estimar numero de archivos (heuristica: 1 por cada "diff --git")
	estimatedFiles := strings.Count(diffText, "diff --git")
	if estimatedFiles == 0 {
		estimatedFiles = 1
	}

	files := make([]FileDiff, 0, estimatedFiles)

	// Parsear linea por linea sin crear slice de strings
	var currentFile *FileDiff
	var currentHunk *Hunk
	var lineNum int

	// Iterar sobre lineas sin split
	start := 0
	for i := 0; i <= len(diffText); i++ {
		// Encontrar fin de linea o fin de string
		if i == len(diffText) || diffText[i] == '\n' {
			line := diffText[start:i]
			start = i + 1
			lineNum++

			if len(line) == 0 {
				continue
			}

			// Procesar linea segun prefijo
			switch {
			case strings.HasPrefix(line, "diff --git"):
				// Nuevo archivo
				if currentFile != nil {
					if currentHunk != nil {
						currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
					}
					files = append(files, *currentFile)
				}
				currentFile = &FileDiff{
					Hunks: make([]Hunk, 0, 4), // Pre-allocar espacio para hunks
				}
				currentHunk = nil

				// Extraer paths sin crear substrings innecesarios
				currentFile.OldPath, currentFile.NewPath = parseGitDiffLine(line)

			case strings.HasPrefix(line, "@@"):
				// Nuevo hunk
				if currentHunk != nil && currentFile != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				}
				currentHunk = parseHunkHeader(line)

			case currentHunk != nil:
				// Linea de contenido
				if len(line) > 0 {
					currentHunk.Lines = append(currentHunk.Lines, parseDiffLine(line))
				}
			}
		}
	}

	// Agregar ultimo archivo/hunk
	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		files = append(files, *currentFile)
	}

	return files, nil
}

// parseGitDiffLine extrae paths de "diff --git a/path b/path"
func parseGitDiffLine(line string) (oldPath, newPath string) {
	// Formato: "diff --git a/path b/path"
	const prefix = "diff --git "
	if !strings.HasPrefix(line, prefix) {
		return "", ""
	}

	rest := line[len(prefix):]

	// Buscar " b/" que separa los paths
	idx := strings.Index(rest, " b/")
	if idx == -1 {
		return "", ""
	}

	oldPath = rest[2:idx] // Skip "a/"
	newPath = rest[idx+3:] // Skip " b/"

	return oldPath, newPath
}

// parseHunkHeader parsea "@@ -1,10 +1,12 @@"
func parseHunkHeader(line string) *Hunk {
	hunk := &Hunk{
		Lines: make([]DiffLine, 0, 32), // Pre-allocar lineas
	}

	// Parseo simple sin regex para velocidad
	var oldStart, oldCount, newStart, newCount int
	fmt.Sscanf(line, "@@ -%d,%d +%d,%d @@",
		&oldStart, &oldCount, &newStart, &newCount)

	hunk.OldStart = oldStart
	hunk.OldCount = oldCount
	hunk.NewStart = newStart
	hunk.NewCount = newCount

	return hunk
}

// parseDiffLine parsea una linea de diff
func parseDiffLine(line string) DiffLine {
	if len(line) == 0 {
		return DiffLine{Type: Context, Content: ""}
	}

	dl := DiffLine{
		Content: line[1:], // Evitar copia si es posible
	}

	switch line[0] {
	case '+':
		dl.Type = Addition
	case '-':
		dl.Type = Deletion
	case ' ':
		dl.Type = Context
	default:
		dl.Type = Context
		dl.Content = line
	}

	return dl
}
```

### `goreview/internal/git/diff_optimized_test.go`

```go
package git

import (
	"testing"
)

// BenchmarkParseDiff_Original vs Optimized
func BenchmarkParseDiff_Original_Medium(b *testing.B) {
	diff := generateDiff(5, 3, 20)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDiff_Optimized_Medium(b *testing.B) {
	diff := generateDiff(5, 3, 20)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiffOptimized(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test de correctitud
func TestParseDiffOptimized_MatchesOriginal(t *testing.T) {
	testCases := []struct {
		name string
		diff string
	}{
		{"small", generateDiff(1, 1, 10)},
		{"medium", generateDiff(5, 3, 20)},
		{"large", generateDiff(20, 10, 50)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original, err1 := ParseDiff(tc.diff)
			optimized, err2 := ParseDiffOptimized(tc.diff)

			if err1 != nil || err2 != nil {
				t.Fatalf("errors: original=%v, optimized=%v", err1, err2)
			}

			if len(original) != len(optimized) {
				t.Fatalf("file count mismatch: original=%d, optimized=%d",
					len(original), len(optimized))
			}

			for i := range original {
				if original[i].OldPath != optimized[i].OldPath {
					t.Errorf("file %d: path mismatch", i)
				}
				if len(original[i].Hunks) != len(optimized[i].Hunks) {
					t.Errorf("file %d: hunk count mismatch", i)
				}
			}
		})
	}
}
```

### Verificacion

```bash
cd goreview

# Comparar benchmarks
go test -bench=BenchmarkParseDiff -benchmem ./internal/git/

# Resultado esperado:
# BenchmarkParseDiff_Original_Medium   10000   120000 ns/op   45000 B/op   800 allocs/op
# BenchmarkParseDiff_Optimized_Medium  15000    80000 ns/op   20000 B/op   300 allocs/op
```

---

## Commit 20.4: Agregar pool de workers

**Mensaje de commit:**
```
perf(review): add worker pool for parallel processing

- Configurable worker count
- Work stealing pattern
- Graceful shutdown
- Error aggregation
```

### Por que worker pool?

Para reviews grandes con muchos archivos:
- Procesar archivos en paralelo
- Limitar goroutines activas
- Evitar saturar la API de LLM
- Balancear carga entre workers

### `goreview/internal/worker/pool.go`

```go
// =============================================================================
// Package worker - Worker pool para procesamiento paralelo
// Implementa un pool de workers con control de concurrencia
// =============================================================================

package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Task representa una tarea a ejecutar
type Task interface {
	Execute(ctx context.Context) error
	ID() string
}

// Result contiene el resultado de una tarea
type Result struct {
	TaskID string
	Error  error
}

// Pool maneja un pool de workers
type Pool struct {
	workers    int
	tasks      chan Task
	results    chan Result
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	started    atomic.Bool
	processed  atomic.Int64
	errors     atomic.Int64
}

// PoolConfig configura el pool
type PoolConfig struct {
	Workers    int // Numero de workers (default: GOMAXPROCS)
	QueueSize  int // Tamano del queue de tareas (default: workers * 2)
}

// NewPool crea un nuevo pool de workers
func NewPool(cfg PoolConfig) *Pool {
	if cfg.Workers <= 0 {
		cfg.Workers = runtime.GOMAXPROCS(0)
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = cfg.Workers * 2
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workers: cfg.Workers,
		tasks:   make(chan Task, cfg.QueueSize),
		results: make(chan Result, cfg.QueueSize),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start inicia los workers
func (p *Pool) Start() {
	if p.started.Swap(true) {
		return // Ya iniciado
	}

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker es la goroutine que procesa tareas
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return

		case task, ok := <-p.tasks:
			if !ok {
				return
			}

			// Ejecutar tarea
			err := task.Execute(p.ctx)

			// Registrar resultado
			p.processed.Add(1)
			if err != nil {
				p.errors.Add(1)
			}

			// Enviar resultado
			select {
			case p.results <- Result{TaskID: task.ID(), Error: err}:
			case <-p.ctx.Done():
				return
			}
		}
	}
}

// Submit envia una tarea al pool
func (p *Pool) Submit(task Task) error {
	if !p.started.Load() {
		return fmt.Errorf("pool not started")
	}

	select {
	case p.tasks <- task:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// Results retorna el canal de resultados
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Stop detiene el pool gracefully
func (p *Pool) Stop() {
	p.cancel()
	close(p.tasks)
	p.wg.Wait()
	close(p.results)
}

// Stats retorna estadisticas del pool
func (p *Pool) Stats() PoolStats {
	return PoolStats{
		Workers:   p.workers,
		Processed: p.processed.Load(),
		Errors:    p.errors.Load(),
		Pending:   len(p.tasks),
	}
}

// PoolStats contiene estadisticas del pool
type PoolStats struct {
	Workers   int
	Processed int64
	Errors    int64
	Pending   int
}
```

### `goreview/internal/worker/file_task.go`

```go
// =============================================================================
// FileTask - Tarea para procesar un archivo
// =============================================================================

package worker

import (
	"context"
	"fmt"
)

// FileTask representa la tarea de revisar un archivo
type FileTask struct {
	id       string
	filePath string
	content  string
	reviewer FileReviewer
}

// FileReviewer interfaz para revisar archivos
type FileReviewer interface {
	ReviewFile(ctx context.Context, path, content string) (*FileReviewResult, error)
}

// FileReviewResult resultado del review de un archivo
type FileReviewResult struct {
	FilePath string
	Issues   []Issue
	Score    float64
	Duration time.Duration
}

// Issue representa un problema encontrado
type Issue struct {
	Line     int
	Column   int
	Severity string
	Message  string
	Rule     string
}

// NewFileTask crea una nueva tarea de archivo
func NewFileTask(path, content string, reviewer FileReviewer) *FileTask {
	return &FileTask{
		id:       fmt.Sprintf("file:%s", path),
		filePath: path,
		content:  content,
		reviewer: reviewer,
	}
}

// ID retorna el identificador de la tarea
func (t *FileTask) ID() string {
	return t.id
}

// Execute ejecuta la tarea
func (t *FileTask) Execute(ctx context.Context) error {
	_, err := t.reviewer.ReviewFile(ctx, t.filePath, t.content)
	return err
}
```

### `goreview/internal/worker/pool_test.go`

```go
package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// mockTask para testing
type mockTask struct {
	id       string
	duration time.Duration
	err      error
}

func (t *mockTask) ID() string { return t.id }
func (t *mockTask) Execute(ctx context.Context) error {
	select {
	case <-time.After(t.duration):
		return t.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestPool_BasicExecution(t *testing.T) {
	pool := NewPool(PoolConfig{Workers: 2, QueueSize: 10})
	pool.Start()
	defer pool.Stop()

	// Enviar tareas
	for i := 0; i < 5; i++ {
		task := &mockTask{
			id:       fmt.Sprintf("task-%d", i),
			duration: 10 * time.Millisecond,
		}
		if err := pool.Submit(task); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	// Recolectar resultados
	results := 0
	timeout := time.After(1 * time.Second)
	for results < 5 {
		select {
		case r := <-pool.Results():
			if r.Error != nil {
				t.Errorf("unexpected error: %v", r.Error)
			}
			results++
		case <-timeout:
			t.Fatal("timeout waiting for results")
		}
	}

	stats := pool.Stats()
	if stats.Processed != 5 {
		t.Errorf("expected 5 processed, got %d", stats.Processed)
	}
}

func TestPool_ErrorHandling(t *testing.T) {
	pool := NewPool(PoolConfig{Workers: 2})
	pool.Start()
	defer pool.Stop()

	expectedErr := errors.New("task failed")
	task := &mockTask{
		id:       "failing-task",
		duration: 10 * time.Millisecond,
		err:      expectedErr,
	}

	pool.Submit(task)

	result := <-pool.Results()
	if result.Error != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, result.Error)
	}

	stats := pool.Stats()
	if stats.Errors != 1 {
		t.Errorf("expected 1 error, got %d", stats.Errors)
	}
}

func TestPool_Cancellation(t *testing.T) {
	pool := NewPool(PoolConfig{Workers: 2})
	pool.Start()

	// Enviar tarea larga
	task := &mockTask{
		id:       "long-task",
		duration: 10 * time.Second, // Muy larga
	}
	pool.Submit(task)

	// Cancelar inmediatamente
	pool.Stop()

	// Pool debe detenerse sin bloquear
}

func TestPool_ConcurrentSubmit(t *testing.T) {
	pool := NewPool(PoolConfig{Workers: 4, QueueSize: 100})
	pool.Start()
	defer pool.Stop()

	var submitted atomic.Int64

	// Enviar desde multiples goroutines
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				task := &mockTask{
					id:       fmt.Sprintf("task-%d-%d", n, j),
					duration: time.Millisecond,
				}
				if err := pool.Submit(task); err == nil {
					submitted.Add(1)
				}
			}
		}(i)
	}

	// Esperar resultados
	time.Sleep(500 * time.Millisecond)

	stats := pool.Stats()
	if stats.Processed < 50 {
		t.Errorf("expected at least 50 processed, got %d", stats.Processed)
	}
}

// Benchmark
func BenchmarkPool_Throughput(b *testing.B) {
	pool := NewPool(PoolConfig{Workers: runtime.GOMAXPROCS(0), QueueSize: 1000})
	pool.Start()
	defer pool.Stop()

	// Consumidor de resultados
	go func() {
		for range pool.Results() {
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		task := &mockTask{
			id:       fmt.Sprintf("task-%d", i),
			duration: 0, // Instantaneo
		}
		pool.Submit(task)
	}
}
```

### Verificacion

```bash
cd goreview

# Tests
go test -v ./internal/worker/

# Benchmarks
go test -bench=. -benchmem ./internal/worker/
```

---

## Commit 20.5: Agregar metricas de rendimiento

**Mensaje de commit:**
```
perf(goreview): add performance metrics collection

- Review duration metrics
- Memory usage tracking
- API latency histograms
- Metrics export (JSON/Prometheus)
```

### Por que metricas?

Las metricas de rendimiento permiten:
- Monitorear degradacion de performance
- Identificar patrones de uso
- Alertar sobre anomalias
- Optimizar basado en datos reales

### `goreview/internal/metrics/metrics.go`

```go
// =============================================================================
// Package metrics - Recoleccion de metricas de rendimiento
// Proporciona metricas de tiempo, contadores y histogramas
// =============================================================================

package metrics

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Collector recolecta metricas
type Collector struct {
	mu         sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	timers     map[string]*Timer
	startTime  time.Time
}

// NewCollector crea un nuevo collector
func NewCollector() *Collector {
	return &Collector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		timers:     make(map[string]*Timer),
		startTime:  time.Now(),
	}
}

// Counter incrementa un contador
type Counter struct {
	value int64
	mu    sync.Mutex
}

func (c *Counter) Inc() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *Counter) Add(n int64) {
	c.mu.Lock()
	c.value += n
	c.mu.Unlock()
}

func (c *Counter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Gauge representa un valor que puede subir o bajar
type Gauge struct {
	value float64
	mu    sync.Mutex
}

func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	g.value = v
	g.mu.Unlock()
}

func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Histogram recolecta distribucion de valores
type Histogram struct {
	values []float64
	mu     sync.Mutex
	max    int // Max valores a mantener
}

func NewHistogram(maxValues int) *Histogram {
	return &Histogram{
		values: make([]float64, 0, maxValues),
		max:    maxValues,
	}
}

func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.values) >= h.max {
		// Rotar: descartar el mas viejo
		h.values = h.values[1:]
	}
	h.values = append(h.values, v)
}

func (h *Histogram) Percentile(p float64) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.values) == 0 {
		return 0
	}

	sorted := make([]float64, len(h.values))
	copy(sorted, h.values)
	sort.Float64s(sorted)

	idx := int(float64(len(sorted)-1) * p / 100)
	return sorted[idx]
}

func (h *Histogram) Stats() HistogramStats {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.values) == 0 {
		return HistogramStats{}
	}

	sorted := make([]float64, len(h.values))
	copy(sorted, h.values)
	sort.Float64s(sorted)

	var sum float64
	for _, v := range sorted {
		sum += v
	}

	return HistogramStats{
		Count: len(sorted),
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
		Avg:   sum / float64(len(sorted)),
		P50:   sorted[len(sorted)*50/100],
		P90:   sorted[len(sorted)*90/100],
		P99:   sorted[len(sorted)*99/100],
	}
}

type HistogramStats struct {
	Count int
	Min   float64
	Max   float64
	Avg   float64
	P50   float64
	P90   float64
	P99   float64
}

// Timer mide duraciones
type Timer struct {
	histogram *Histogram
}

func (t *Timer) Start() *TimerContext {
	return &TimerContext{
		timer: t,
		start: time.Now(),
	}
}

type TimerContext struct {
	timer *Timer
	start time.Time
}

func (tc *TimerContext) Stop() time.Duration {
	d := time.Since(tc.start)
	tc.timer.histogram.Observe(d.Seconds())
	return d
}

// Metodos del Collector

func (c *Collector) Counter(name string) *Counter {
	c.mu.Lock()
	defer c.mu.Unlock()

	if counter, ok := c.counters[name]; ok {
		return counter
	}

	counter := &Counter{}
	c.counters[name] = counter
	return counter
}

func (c *Collector) Gauge(name string) *Gauge {
	c.mu.Lock()
	defer c.mu.Unlock()

	if gauge, ok := c.gauges[name]; ok {
		return gauge
	}

	gauge := &Gauge{}
	c.gauges[name] = gauge
	return gauge
}

func (c *Collector) Histogram(name string) *Histogram {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hist, ok := c.histograms[name]; ok {
		return hist
	}

	hist := NewHistogram(1000)
	c.histograms[name] = hist
	return hist
}

func (c *Collector) Timer(name string) *Timer {
	c.mu.Lock()
	defer c.mu.Unlock()

	if timer, ok := c.timers[name]; ok {
		return timer
	}

	timer := &Timer{histogram: NewHistogram(1000)}
	c.timers[name] = timer
	return timer
}

// Export exporta metricas a JSON
func (c *Collector) Export() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	export := struct {
		Uptime     string                    `json:"uptime"`
		Counters   map[string]int64          `json:"counters"`
		Gauges     map[string]float64        `json:"gauges"`
		Histograms map[string]HistogramStats `json:"histograms"`
		Timers     map[string]HistogramStats `json:"timers"`
	}{
		Uptime:     time.Since(c.startTime).String(),
		Counters:   make(map[string]int64),
		Gauges:     make(map[string]float64),
		Histograms: make(map[string]HistogramStats),
		Timers:     make(map[string]HistogramStats),
	}

	for name, counter := range c.counters {
		export.Counters[name] = counter.Value()
	}

	for name, gauge := range c.gauges {
		export.Gauges[name] = gauge.Value()
	}

	for name, hist := range c.histograms {
		export.Histograms[name] = hist.Stats()
	}

	for name, timer := range c.timers {
		export.Timers[name] = timer.histogram.Stats()
	}

	return json.MarshalIndent(export, "", "  ")
}

// ExportPrometheus exporta en formato Prometheus
func (c *Collector) ExportPrometheus() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sb strings.Builder

	// Counters
	for name, counter := range c.counters {
		sb.WriteString(fmt.Sprintf("# TYPE %s counter\n", name))
		sb.WriteString(fmt.Sprintf("%s %d\n", name, counter.Value()))
	}

	// Gauges
	for name, gauge := range c.gauges {
		sb.WriteString(fmt.Sprintf("# TYPE %s gauge\n", name))
		sb.WriteString(fmt.Sprintf("%s %f\n", name, gauge.Value()))
	}

	// Histograms (summary)
	for name, hist := range c.histograms {
		stats := hist.Stats()
		sb.WriteString(fmt.Sprintf("# TYPE %s summary\n", name))
		sb.WriteString(fmt.Sprintf("%s_count %d\n", name, stats.Count))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.5\"} %f\n", name, stats.P50))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.9\"} %f\n", name, stats.P90))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.99\"} %f\n", name, stats.P99))
	}

	return sb.String()
}
```

### `goreview/internal/metrics/global.go`

```go
// =============================================================================
// Global metrics instance
// Singleton para acceso global a metricas
// =============================================================================

package metrics

import "sync"

var (
	globalCollector *Collector
	once            sync.Once
)

// Global retorna el collector global
func Global() *Collector {
	once.Do(func() {
		globalCollector = NewCollector()
	})
	return globalCollector
}

// Convenience functions para acceso rapido

// IncCounter incrementa un contador global
func IncCounter(name string) {
	Global().Counter(name).Inc()
}

// SetGauge establece un gauge global
func SetGauge(name string, v float64) {
	Global().Gauge(name).Set(v)
}

// ObserveHistogram observa un valor en histograma global
func ObserveHistogram(name string, v float64) {
	Global().Histogram(name).Observe(v)
}

// StartTimer inicia un timer global
func StartTimer(name string) *TimerContext {
	return Global().Timer(name).Start()
}
```

### `goreview/internal/metrics/metrics_test.go`

```go
package metrics

import (
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	c := NewCollector()

	counter := c.Counter("test_counter")
	counter.Inc()
	counter.Inc()
	counter.Add(5)

	if counter.Value() != 7 {
		t.Errorf("expected 7, got %d", counter.Value())
	}
}

func TestGauge(t *testing.T) {
	c := NewCollector()

	gauge := c.Gauge("test_gauge")
	gauge.Set(42.5)

	if gauge.Value() != 42.5 {
		t.Errorf("expected 42.5, got %f", gauge.Value())
	}

	gauge.Set(10.0)
	if gauge.Value() != 10.0 {
		t.Errorf("expected 10.0, got %f", gauge.Value())
	}
}

func TestHistogram(t *testing.T) {
	hist := NewHistogram(100)

	// Agregar valores 1-100
	for i := 1; i <= 100; i++ {
		hist.Observe(float64(i))
	}

	stats := hist.Stats()

	if stats.Count != 100 {
		t.Errorf("expected 100 count, got %d", stats.Count)
	}
	if stats.Min != 1 {
		t.Errorf("expected min 1, got %f", stats.Min)
	}
	if stats.Max != 100 {
		t.Errorf("expected max 100, got %f", stats.Max)
	}
	if stats.Avg != 50.5 {
		t.Errorf("expected avg 50.5, got %f", stats.Avg)
	}
}

func TestTimer(t *testing.T) {
	c := NewCollector()
	timer := c.Timer("test_timer")

	// Simular operacion
	ctx := timer.Start()
	time.Sleep(10 * time.Millisecond)
	duration := ctx.Stop()

	if duration < 10*time.Millisecond {
		t.Errorf("expected at least 10ms, got %v", duration)
	}

	stats := timer.histogram.Stats()
	if stats.Count != 1 {
		t.Errorf("expected 1 measurement, got %d", stats.Count)
	}
}

func TestExportJSON(t *testing.T) {
	c := NewCollector()

	c.Counter("files_processed").Add(10)
	c.Gauge("memory_mb").Set(256.5)
	c.Histogram("response_time").Observe(0.5)

	data, err := c.Export()
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("export returned empty data")
	}
}

func TestExportPrometheus(t *testing.T) {
	c := NewCollector()

	c.Counter("http_requests_total").Add(100)
	c.Gauge("goroutines").Set(50)

	output := c.ExportPrometheus()

	if output == "" {
		t.Error("prometheus export returned empty")
	}

	if !strings.Contains(output, "http_requests_total") {
		t.Error("missing counter in output")
	}
	if !strings.Contains(output, "goroutines") {
		t.Error("missing gauge in output")
	}
}

func BenchmarkCounter_Inc(b *testing.B) {
	counter := &Counter{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}

func BenchmarkHistogram_Observe(b *testing.B) {
	hist := NewHistogram(10000)
	b.RunParallel(func(pb *testing.PB) {
		i := 0.0
		for pb.Next() {
			hist.Observe(i)
			i++
		}
	})
}
```

### Verificacion

```bash
cd goreview

# Tests
go test -v ./internal/metrics/

# Benchmarks
go test -bench=. -benchmem ./internal/metrics/
```

---

## Commit 20.6: Integrar metricas en review engine

**Mensaje de commit:**
```
perf(review): integrate metrics into review engine

- Track review duration
- Count files processed
- Measure LLM latency
- Memory usage gauges
```

### `goreview/internal/review/engine_metrics.go`

```go
// =============================================================================
// Review Engine Metrics Integration
// Instrumentacion del review engine con metricas
// =============================================================================

package review

import (
	"runtime"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/metrics"
)

// MetricNames define nombres de metricas
const (
	MetricReviewsTotal      = "goreview_reviews_total"
	MetricReviewDuration    = "goreview_review_duration_seconds"
	MetricFilesProcessed    = "goreview_files_processed_total"
	MetricIssuesFound       = "goreview_issues_found_total"
	MetricLLMLatency        = "goreview_llm_latency_seconds"
	MetricCacheHits         = "goreview_cache_hits_total"
	MetricCacheMisses       = "goreview_cache_misses_total"
	MetricMemoryUsage       = "goreview_memory_bytes"
	MetricGoroutines        = "goreview_goroutines"
	MetricErrors            = "goreview_errors_total"
)

// InstrumentedEngine envuelve Engine con metricas
type InstrumentedEngine struct {
	engine    *Engine
	collector *metrics.Collector
}

// NewInstrumentedEngine crea un engine con metricas
func NewInstrumentedEngine(engine *Engine) *InstrumentedEngine {
	return &InstrumentedEngine{
		engine:    engine,
		collector: metrics.Global(),
	}
}

// Review ejecuta review con metricas
func (ie *InstrumentedEngine) Review(ctx context.Context, opts ReviewOptions) (*ReviewResult, error) {
	// Incrementar contador de reviews
	ie.collector.Counter(MetricReviewsTotal).Inc()

	// Timer para duracion total
	timer := ie.collector.Timer(MetricReviewDuration).Start()
	defer timer.Stop()

	// Actualizar metricas de memoria antes
	ie.updateMemoryMetrics()

	// Ejecutar review
	result, err := ie.engine.Review(ctx, opts)

	// Registrar errores
	if err != nil {
		ie.collector.Counter(MetricErrors).Inc()
		return result, err
	}

	// Registrar metricas de resultado
	if result != nil {
		ie.collector.Counter(MetricFilesProcessed).Add(int64(len(result.Files)))
		ie.collector.Counter(MetricIssuesFound).Add(int64(result.TotalIssues))
	}

	// Actualizar metricas de memoria despues
	ie.updateMemoryMetrics()

	return result, nil
}

// updateMemoryMetrics actualiza metricas de memoria
func (ie *InstrumentedEngine) updateMemoryMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ie.collector.Gauge(MetricMemoryUsage).Set(float64(m.Alloc))
	ie.collector.Gauge(MetricGoroutines).Set(float64(runtime.NumGoroutine()))
}

// RecordLLMLatency registra latencia de LLM
func (ie *InstrumentedEngine) RecordLLMLatency(duration time.Duration) {
	ie.collector.Histogram(MetricLLMLatency).Observe(duration.Seconds())
}

// RecordCacheHit registra cache hit
func (ie *InstrumentedEngine) RecordCacheHit() {
	ie.collector.Counter(MetricCacheHits).Inc()
}

// RecordCacheMiss registra cache miss
func (ie *InstrumentedEngine) RecordCacheMiss() {
	ie.collector.Counter(MetricCacheMisses).Inc()
}

// Metrics retorna las metricas actuales
func (ie *InstrumentedEngine) Metrics() ([]byte, error) {
	return ie.collector.Export()
}

// MetricsPrometheus retorna metricas en formato Prometheus
func (ie *InstrumentedEngine) MetricsPrometheus() string {
	return ie.collector.ExportPrometheus()
}
```

### Verificacion

```bash
cd goreview

# Ejecutar review y ver metricas
./build/goreview review --metrics . 2>&1 | tail -20
```

---

## Resumen de la Iteracion 20

### Commits:
1. `perf(goreview): add basic benchmarks`
2. `perf(goreview): add profiling support`
3. `perf(git): optimize diff parsing`
4. `perf(review): add worker pool for parallel processing`
5. `perf(goreview): add performance metrics collection`
6. `perf(review): integrate metrics into review engine`

### Archivos:
```
goreview/
├── internal/
│   ├── git/
│   │   ├── diff_benchmark_test.go
│   │   ├── diff_optimized.go
│   │   └── diff_optimized_test.go
│   ├── cache/
│   │   └── cache_benchmark_test.go
│   ├── rules/
│   │   └── rules_benchmark_test.go
│   ├── profiler/
│   │   └── profiler.go
│   ├── worker/
│   │   ├── pool.go
│   │   ├── file_task.go
│   │   └── pool_test.go
│   ├── metrics/
│   │   ├── metrics.go
│   │   ├── global.go
│   │   └── metrics_test.go
│   └── review/
│       └── engine_metrics.go
└── cmd/
    └── goreview/
        └── flags_profile.go
```

### Comandos de verificacion:

```bash
# Ejecutar todos los benchmarks
go test -bench=. -benchmem ./...

# Profiling de CPU
./build/goreview review --cpuprofile=cpu.prof .
go tool pprof cpu.prof

# Profiling de memoria
./build/goreview review --memprofile=mem.prof .
go tool pprof mem.prof

# pprof HTTP
./build/goreview review --pprof-addr=:6060 .
# En navegador: http://localhost:6060/debug/pprof/

# Comparar benchmarks
go test -bench=. -count=5 ./internal/git/ > before.txt
# ... hacer optimizaciones ...
go test -bench=. -count=5 ./internal/git/ > after.txt
benchstat before.txt after.txt
```

---

## Guia Final: Proximos Pasos

Has completado todas las 20 iteraciones del AI Toolkit. El proyecto ahora incluye:

### CLI GoReview (Go)
- Comandos: review, commit, doc, init
- Providers de IA: Ollama, OpenAI
- Cache LRU con persistencia
- Sistema de reglas YAML
- Reportes en multiples formatos
- Profiling y metricas

### GitHub App (TypeScript)
- Webhooks para PRs
- Integracion con Ollama
- Rate limiting y seguridad
- Queue de jobs con retry
- Audit logging

### Infraestructura
- Docker multi-stage
- Docker Compose (dev/prod)
- CI/CD con GitHub Actions
- Security scanning

### Proximas mejoras sugeridas:
1. **Soporte para mas lenguajes** en reglas
2. **Dashboard web** para visualizar reviews
3. **Integracion con otros LLMs** (Anthropic, Gemini)
4. **GitHub Enterprise** support
5. **Metricas en Grafana** con Prometheus
6. **Tests de integracion E2E**
7. **Plugin para VS Code/JetBrains**

---

## Recursos Adicionales

- [Go Performance Tips](https://go.dev/doc/diagnostics)
- [pprof Documentation](https://pkg.go.dev/runtime/pprof)
- [Benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Go Memory Model](https://go.dev/ref/mem)

---

Felicidades por completar el AI Toolkit!
