# Iteracion 21: Sistema de Memoria Cognitiva (GoMemory)

## Objetivos

- Interface Memory para abstraccion
- Working Memory (buffer circular en RAM)
- Session Memory (LRU + TTL)
- Long-term Memory (BadgerDB persistente)
- Hebbian Learning (conexiones que se refuerzan)
- Busqueda semantica con embeddings
- Feature flag para activar/desactivar

## Tiempo Estimado: 8 horas

## Prerequisitos

- Iteraciones 00-06 completadas (especialmente Cache)
- Providers de IA funcionando
- BadgerDB como dependencia

---

## Commit 21.1: Agregar configuracion de memoria

**Mensaje de commit:**
```
feat(config): add memory configuration with feature flag

- Add MemoryConfig and HebbianConfig structs
- Feature flag disabled by default
- Add --memory CLI flag
```

### `goreview/internal/config/config.go` (agregar al final)

```go
// MemoryConfig configura el sistema de memoria cognitiva
type MemoryConfig struct {
	// Enabled activa/desactiva el sistema de memoria
	// Cuando esta deshabilitado, usa NoopMemory (zero overhead)
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`

	// Dir es el directorio donde se persiste la memoria
	Dir string `yaml:"dir" mapstructure:"dir"`

	// WorkingSize es el numero maximo de items en working memory
	WorkingSize int `yaml:"working_size" mapstructure:"working_size"`

	// SessionTTL es el tiempo de vida de items en session memory
	SessionTTL time.Duration `yaml:"session_ttl" mapstructure:"session_ttl"`

	// Hebbian configura el aprendizaje hebbiano
	Hebbian HebbianConfig `yaml:"hebbian" mapstructure:"hebbian"`

	// SemanticSearch habilita busqueda por embeddings
	SemanticSearch bool `yaml:"semantic_search" mapstructure:"semantic_search"`
}

// HebbianConfig configura el aprendizaje hebbiano
type HebbianConfig struct {
	// Enabled activa el reforzamiento de conexiones
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`

	// DecayRate es la tasa de decay exponencial (0.0 - 1.0)
	DecayRate float64 `yaml:"decay_rate" mapstructure:"decay_rate"`

	// MinActivation es el umbral minimo de activacion
	MinActivation float64 `yaml:"min_activation" mapstructure:"min_activation"`
}
```

### `goreview/internal/config/defaults.go` (agregar funcion)

```go
// DefaultMemoryConfig returns default memory configuration
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		Enabled:        false, // Deshabilitado por default
		Dir:            ".cache/goreview/memory",
		WorkingSize:    100,
		SessionTTL:     time.Hour,
		SemanticSearch: true,
		Hebbian: HebbianConfig{
			Enabled:       true,
			DecayRate:     0.1,
			MinActivation: 0.01,
		},
	}
}
```

### Agregar al struct Config principal

```go
type Config struct {
	// ... campos existentes ...

	// Memory settings (Feature Flag)
	Memory MemoryConfig `yaml:"memory" mapstructure:"memory"`
}
```

---

## Commit 21.2: Crear interface Memory y NoopMemory

**Mensaje de commit:**
```
feat(memory): add Memory interface and NoopMemory

- Define Memory interface with Store, Search, Get
- Add Item and SearchResult structs
- Implement NoopMemory for zero-overhead when disabled
- Add factory function
```

### `goreview/internal/memory/memory.go`

```go
package memory

import (
	"context"
	"time"
)

// Item representa una unidad de memoria
type Item struct {
	// ID unico de la memoria
	ID string `json:"id"`

	// Content es el contenido principal
	Content string `json:"content"`

	// Type indica el tipo de memoria
	Type ItemType `json:"type"`

	// Metadata contiene informacion adicional
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Tags para categorizar
	Tags []string `json:"tags,omitempty"`

	// Importance score (0.0 - 1.0)
	Importance float64 `json:"importance"`

	// Activation nivel de activacion actual
	Activation float64 `json:"activation"`

	// CreatedAt timestamp de creacion
	CreatedAt time.Time `json:"created_at"`

	// AccessedAt ultimo acceso
	AccessedAt time.Time `json:"accessed_at"`

	// AccessCount numero de accesos
	AccessCount int `json:"access_count"`
}

// ItemType define los tipos de memoria
type ItemType string

const (
	TypeIssue   ItemType = "issue"
	TypePattern ItemType = "pattern"
	TypeFix     ItemType = "fix"
	TypeContext ItemType = "context"
)

// SearchResult representa un resultado de busqueda
type SearchResult struct {
	Item      Item    `json:"item"`
	Score     float64 `json:"score"`
	MatchType string  `json:"match_type"`
}

// Stats contiene estadisticas del sistema
type Stats struct {
	WorkingCount   int     `json:"working_count"`
	SessionCount   int     `json:"session_count"`
	LongTermCount  int     `json:"longterm_count"`
	TotalCount     int     `json:"total_count"`
	DiskUsageBytes int64   `json:"disk_usage_bytes"`
	AvgActivation  float64 `json:"avg_activation"`
}

// Memory define la interface del sistema de memoria
type Memory interface {
	// Store guarda un item en memoria
	Store(ctx context.Context, item Item) error

	// Search busca items relevantes
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)

	// Get obtiene un item por ID
	Get(ctx context.Context, id string) (*Item, error)

	// GetRelated obtiene items relacionados via Hebbian
	GetRelated(ctx context.Context, id string, limit int) ([]SearchResult, error)

	// Update actualiza un item existente
	Update(ctx context.Context, item Item) error

	// Delete elimina un item
	Delete(ctx context.Context, id string) error

	// RecordAccess registra accesos para Hebbian learning
	RecordAccess(ctx context.Context, ids ...string) error

	// Stats retorna estadisticas
	Stats(ctx context.Context) (*Stats, error)

	// Close cierra conexiones
	Close() error
}
```

### `goreview/internal/memory/noop.go`

```go
package memory

import "context"

// NoopMemory es una implementacion que no hace nada
// Se usa cuando memory.enabled = false para zero overhead
type NoopMemory struct{}

// NewNoop crea una instancia de NoopMemory
func NewNoop() *NoopMemory {
	return &NoopMemory{}
}

func (n *NoopMemory) Store(ctx context.Context, item Item) error {
	return nil
}

func (n *NoopMemory) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	return nil, nil
}

func (n *NoopMemory) Get(ctx context.Context, id string) (*Item, error) {
	return nil, nil
}

func (n *NoopMemory) GetRelated(ctx context.Context, id string, limit int) ([]SearchResult, error) {
	return nil, nil
}

func (n *NoopMemory) Update(ctx context.Context, item Item) error {
	return nil
}

func (n *NoopMemory) Delete(ctx context.Context, id string) error {
	return nil
}

func (n *NoopMemory) RecordAccess(ctx context.Context, ids ...string) error {
	return nil
}

func (n *NoopMemory) Stats(ctx context.Context) (*Stats, error) {
	return &Stats{}, nil
}

func (n *NoopMemory) Close() error {
	return nil
}

// Verify interface compliance at compile time
var _ Memory = (*NoopMemory)(nil)
```

### `goreview/internal/memory/factory.go`

```go
package memory

import (
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

// New crea una instancia de Memory basada en la configuracion
func New(cfg config.MemoryConfig, provider providers.Provider) (Memory, error) {
	if !cfg.Enabled {
		return NewNoop(), nil
	}

	return NewStore(cfg, provider)
}
```

---

## Commit 21.3: Implementar Working Memory

**Mensaje de commit:**
```
feat(memory): add WorkingMemory with circular buffer

- Thread-safe circular buffer using container/ring
- O(1) Add, Get, Remove via index map
- Automatic overflow handling
- Predicate-based search
```

### `goreview/internal/memory/working.go`

```go
package memory

import (
	"container/ring"
	"sync"
	"time"
)

// WorkingMemory implementa un buffer circular thread-safe
type WorkingMemory struct {
	mu      sync.RWMutex
	buffer  *ring.Ring
	index   map[string]*ring.Ring
	size    int
	maxSize int
}

// NewWorkingMemory crea una nueva working memory
func NewWorkingMemory(maxSize int) *WorkingMemory {
	return &WorkingMemory{
		buffer:  ring.New(maxSize),
		index:   make(map[string]*ring.Ring),
		maxSize: maxSize,
	}
}

// Add agrega un item a working memory
// Retorna el item desplazado si el buffer esta lleno
func (w *WorkingMemory) Add(item Item) *Item {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Si ya existe, actualizar
	if elem, exists := w.index[item.ID]; exists {
		elem.Value = item
		return nil
	}

	// Obtener item que sera desplazado
	var displaced *Item
	if w.size >= w.maxSize {
		if w.buffer.Value != nil {
			old := w.buffer.Value.(Item)
			displaced = &old
			delete(w.index, old.ID)
		}
	}

	// Agregar nuevo item
	item.AccessedAt = time.Now()
	w.buffer.Value = item
	w.index[item.ID] = w.buffer
	w.buffer = w.buffer.Next()

	if w.size < w.maxSize {
		w.size++
	}

	return displaced
}

// Get obtiene un item por ID
func (w *WorkingMemory) Get(id string) (*Item, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if elem, exists := w.index[id]; exists {
		item := elem.Value.(Item)
		return &item, true
	}
	return nil, false
}

// Search busca items que coincidan con el predicado
func (w *WorkingMemory) Search(predicate func(Item) bool, limit int) []Item {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var results []Item
	w.buffer.Do(func(v interface{}) {
		if v == nil || len(results) >= limit {
			return
		}
		item := v.(Item)
		if predicate(item) {
			results = append(results, item)
		}
	})
	return results
}

// All retorna todos los items
func (w *WorkingMemory) All() []Item {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var items []Item
	w.buffer.Do(func(v interface{}) {
		if v != nil {
			items = append(items, v.(Item))
		}
	})
	return items
}

// Remove elimina un item por ID
func (w *WorkingMemory) Remove(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if elem, exists := w.index[id]; exists {
		elem.Value = nil
		delete(w.index, id)
		w.size--
		return true
	}
	return false
}

// Size retorna el numero de items
func (w *WorkingMemory) Size() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.size
}

// Clear limpia toda la memoria
func (w *WorkingMemory) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer = ring.New(w.maxSize)
	w.index = make(map[string]*ring.Ring)
	w.size = 0
}
```

---

## Commit 21.4: Implementar Session Memory

**Mensaje de commit:**
```
feat(memory): add SessionMemory with LRU and TTL

- LRU cache using container/list
- TTL-based expiration
- Background cleanup goroutine
- Access tracking
```

### `goreview/internal/memory/session.go`

```go
package memory

import (
	"container/list"
	"sync"
	"time"
)

// SessionMemory implementa un cache LRU con TTL
type SessionMemory struct {
	mu      sync.RWMutex
	items   map[string]*sessionEntry
	order   *list.List
	maxSize int
	ttl     time.Duration
	stopCh  chan struct{}
}

type sessionEntry struct {
	item      Item
	element   *list.Element
	expiresAt time.Time
}

// NewSessionMemory crea una nueva session memory
func NewSessionMemory(maxSize int, ttl time.Duration) *SessionMemory {
	sm := &SessionMemory{
		items:   make(map[string]*sessionEntry),
		order:   list.New(),
		maxSize: maxSize,
		ttl:     ttl,
		stopCh:  make(chan struct{}),
	}

	go sm.cleanup()

	return sm
}

// Add agrega o actualiza un item
func (s *SessionMemory) Add(item Item) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	item.AccessedAt = now

	if entry, exists := s.items[item.ID]; exists {
		entry.item = item
		entry.expiresAt = now.Add(s.ttl)
		s.order.MoveToFront(entry.element)
		return
	}

	for len(s.items) >= s.maxSize {
		s.evictOldest()
	}

	elem := s.order.PushFront(item.ID)
	s.items[item.ID] = &sessionEntry{
		item:      item,
		element:   elem,
		expiresAt: now.Add(s.ttl),
	}
}

// Get obtiene un item y actualiza su posicion LRU
func (s *SessionMemory) Get(id string) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.items[id]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		s.removeEntry(id, entry)
		return nil, false
	}

	entry.item.AccessedAt = time.Now()
	entry.item.AccessCount++
	s.order.MoveToFront(entry.element)

	return &entry.item, true
}

// All retorna todos los items validos
func (s *SessionMemory) All() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	var items []Item
	for _, entry := range s.items {
		if now.Before(entry.expiresAt) {
			items = append(items, entry.item)
		}
	}
	return items
}

// Remove elimina un item
func (s *SessionMemory) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.items[id]
	if !exists {
		return false
	}

	s.removeEntry(id, entry)
	return true
}

// Size retorna el numero de items
func (s *SessionMemory) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// Close detiene el cleanup goroutine
func (s *SessionMemory) Close() {
	close(s.stopCh)
}

func (s *SessionMemory) evictOldest() {
	elem := s.order.Back()
	if elem == nil {
		return
	}

	id := elem.Value.(string)
	if entry, exists := s.items[id]; exists {
		s.removeEntry(id, entry)
	}
}

func (s *SessionMemory) removeEntry(id string, entry *sessionEntry) {
	s.order.Remove(entry.element)
	delete(s.items, id)
}

func (s *SessionMemory) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanupExpired()
		}
	}
}

func (s *SessionMemory) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, entry := range s.items {
		if now.After(entry.expiresAt) {
			s.removeEntry(id, entry)
		}
	}
}
```

---

## Commit 21.5: Implementar Long-term Memory con BadgerDB

**Mensaje de commit:**
```
feat(memory): add LongTermMemory with BadgerDB

- Persistent storage using BadgerDB
- Secondary indices for type and tag queries
- Automatic garbage collection
- Disk usage tracking
```

### Agregar dependencia

```bash
cd goreview
go get github.com/dgraph-io/badger/v4
```

### `goreview/internal/memory/longterm.go`

```go
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// LongTermMemory implementa almacenamiento persistente
type LongTermMemory struct {
	db   *badger.DB
	path string
}

// NewLongTermMemory crea una nueva long-term memory
func NewLongTermMemory(dir string) (*LongTermMemory, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	opts := badger.DefaultOptions(dir).
		WithLogger(nil).
		WithValueLogFileSize(64 << 20).
		WithIndexCacheSize(32 << 20)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	ltm := &LongTermMemory{
		db:   db,
		path: dir,
	}

	go ltm.runGC()

	return ltm, nil
}

// Store guarda un item
func (l *LongTermMemory) Store(item Item) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Set(l.itemKey(item.ID), data)
	})
}

// StoreWithIndex guarda un item con indices secundarios
func (l *LongTermMemory) StoreWithIndex(item Item) error {
	return l.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if err := txn.Set(l.itemKey(item.ID), data); err != nil {
			return err
		}

		// Index by type
		typeKey := []byte("idx:type:" + string(item.Type) + ":" + item.ID)
		if err := txn.Set(typeKey, nil); err != nil {
			return err
		}

		// Index by tags
		for _, tag := range item.Tags {
			tagKey := []byte("idx:tag:" + tag + ":" + item.ID)
			if err := txn.Set(tagKey, nil); err != nil {
				return err
			}
		}

		return nil
	})
}

// Get obtiene un item por ID
func (l *LongTermMemory) Get(id string) (*Item, error) {
	var item Item

	err := l.db.View(func(txn *badger.Txn) error {
		entry, err := txn.Get(l.itemKey(id))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}

		return entry.Value(func(val []byte) error {
			return json.Unmarshal(val, &item)
		})
	})

	if err != nil {
		return nil, err
	}
	if item.ID == "" {
		return nil, nil
	}
	return &item, nil
}

// Delete elimina un item
func (l *LongTermMemory) Delete(id string) error {
	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(l.itemKey(id))
	})
}

// DeleteWithIndex elimina un item y sus indices
func (l *LongTermMemory) DeleteWithIndex(id string) error {
	item, err := l.Get(id)
	if err != nil || item == nil {
		return err
	}

	return l.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(l.itemKey(id)); err != nil {
			return err
		}

		typeKey := []byte("idx:type:" + string(item.Type) + ":" + id)
		txn.Delete(typeKey)

		for _, tag := range item.Tags {
			tagKey := []byte("idx:tag:" + tag + ":" + id)
			txn.Delete(tagKey)
		}

		return nil
	})
}

// All itera sobre todos los items
func (l *LongTermMemory) All(fn func(Item) bool) error {
	return l.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("item:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var item Item
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &item)
			})
			if err != nil {
				continue
			}

			if !fn(item) {
				break
			}
		}
		return nil
	})
}

// FindByType busca items por tipo
func (l *LongTermMemory) FindByType(itemType ItemType, limit int) ([]Item, error) {
	var items []Item

	err := l.db.View(func(txn *badger.Txn) error {
		prefix := []byte("idx:type:" + string(itemType) + ":")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid() && len(items) < limit; it.Next() {
			key := string(it.Item().Key())
			id := strings.TrimPrefix(key, string(prefix))

			item, err := l.Get(id)
			if err == nil && item != nil {
				items = append(items, *item)
			}
		}
		return nil
	})

	return items, err
}

// FindByTag busca items por tag
func (l *LongTermMemory) FindByTag(tag string, limit int) ([]Item, error) {
	var items []Item

	err := l.db.View(func(txn *badger.Txn) error {
		prefix := []byte("idx:tag:" + tag + ":")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid() && len(items) < limit; it.Next() {
			key := string(it.Item().Key())
			id := strings.TrimPrefix(key, string(prefix))

			item, err := l.Get(id)
			if err == nil && item != nil {
				items = append(items, *item)
			}
		}
		return nil
	})

	return items, err
}

// Count retorna el numero de items
func (l *LongTermMemory) Count() (int, error) {
	var count int
	err := l.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("item:")
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})
	return count, err
}

// DiskUsage retorna el uso de disco en bytes
func (l *LongTermMemory) DiskUsage() (int64, error) {
	var size int64

	err := filepath.Walk(l.path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// Close cierra la base de datos
func (l *LongTermMemory) Close() error {
	return l.db.Close()
}

func (l *LongTermMemory) itemKey(id string) []byte {
	return []byte("item:" + id)
}

func (l *LongTermMemory) runGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		for {
			err := l.db.RunValueLogGC(0.5)
			if err != nil {
				break
			}
		}
	}
}
```

---

## Commit 21.6: Implementar Hebbian Learning

**Mensaje de commit:**
```
feat(memory): add Hebbian learning for connections

- 'Neurons that fire together wire together'
- Exponential decay for unused connections
- Bounded weight growth (0.0 - 1.0)
- Persistent connections in BadgerDB
```

### `goreview/internal/memory/hebbian.go`

```go
package memory

import (
	"encoding/json"
	"math"
	"sort"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
)

// HebbianLearning implementa el aprendizaje hebbiano
type HebbianLearning struct {
	mu          sync.RWMutex
	connections map[string]map[string]*Connection
	db          *badger.DB
	config      config.HebbianConfig
}

// Connection representa una conexion entre dos items
type Connection struct {
	SourceID   string    `json:"source_id"`
	TargetID   string    `json:"target_id"`
	Weight     float64   `json:"weight"`
	CoOccur    int       `json:"co_occur"`
	LastAccess time.Time `json:"last_access"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewHebbianLearning crea una nueva instancia
func NewHebbianLearning(db *badger.DB, cfg config.HebbianConfig) *HebbianLearning {
	h := &HebbianLearning{
		connections: make(map[string]map[string]*Connection),
		db:          db,
		config:      cfg,
	}

	h.loadConnections()

	if cfg.Enabled {
		go h.runDecay()
	}

	return h
}

// RecordCoAccess registra que items fueron accedidos juntos
func (h *HebbianLearning) RecordCoAccess(ids []string) {
	if !h.config.Enabled || len(ids) < 2 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()

	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			h.strengthen(ids[i], ids[j], now)
			h.strengthen(ids[j], ids[i], now)
		}
	}
}

// GetConnected retorna los items mas conectados
func (h *HebbianLearning) GetConnected(id string, limit int) []Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, exists := h.connections[id]
	if !exists {
		return nil
	}

	var result []Connection
	for _, conn := range conns {
		if conn.Weight >= h.config.MinActivation {
			result = append(result, *conn)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Weight > result[j].Weight
	})

	if len(result) > limit {
		result = result[:limit]
	}

	return result
}

// GetConnectionWeight retorna el peso entre dos items
func (h *HebbianLearning) GetConnectionWeight(id1, id2 string) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, exists := h.connections[id1]; exists {
		if conn, exists := conns[id2]; exists {
			return conn.Weight
		}
	}
	return 0
}

func (h *HebbianLearning) strengthen(sourceID, targetID string, now time.Time) {
	if h.connections[sourceID] == nil {
		h.connections[sourceID] = make(map[string]*Connection)
	}

	conn, exists := h.connections[sourceID][targetID]
	if !exists {
		conn = &Connection{
			SourceID:  sourceID,
			TargetID:  targetID,
			Weight:    0.1,
			CreatedAt: now,
		}
		h.connections[sourceID][targetID] = conn
	}

	conn.CoOccur++
	conn.LastAccess = now
	conn.Weight = math.Min(1.0, conn.Weight+0.1*(1.0-conn.Weight))

	h.persistConnection(conn)
}

func (h *HebbianLearning) runDecay() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		h.applyDecay()
	}
}

func (h *HebbianLearning) applyDecay() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()

	for sourceID, conns := range h.connections {
		for targetID, conn := range conns {
			hoursSinceAccess := now.Sub(conn.LastAccess).Hours()
			decayFactor := math.Exp(-h.config.DecayRate * hoursSinceAccess)
			conn.Weight *= decayFactor

			if conn.Weight < h.config.MinActivation {
				delete(conns, targetID)
				h.deleteConnection(sourceID, targetID)
			} else {
				h.persistConnection(conn)
			}
		}

		if len(conns) == 0 {
			delete(h.connections, sourceID)
		}
	}
}

func (h *HebbianLearning) persistConnection(conn *Connection) {
	if h.db == nil {
		return
	}

	data, _ := json.Marshal(conn)
	key := []byte("hebb:" + conn.SourceID + ":" + conn.TargetID)

	h.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

func (h *HebbianLearning) deleteConnection(sourceID, targetID string) {
	if h.db == nil {
		return
	}

	key := []byte("hebb:" + sourceID + ":" + targetID)
	h.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (h *HebbianLearning) loadConnections() {
	if h.db == nil {
		return
	}

	h.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("hebb:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var conn Connection
			it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &conn)
			})

			if h.connections[conn.SourceID] == nil {
				h.connections[conn.SourceID] = make(map[string]*Connection)
			}
			h.connections[conn.SourceID][conn.TargetID] = &conn
		}
		return nil
	})
}

// Stats retorna estadisticas
func (h *HebbianLearning) Stats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalConns := 0
	var totalWeight float64

	for _, conns := range h.connections {
		totalConns += len(conns)
		for _, conn := range conns {
			totalWeight += conn.Weight
		}
	}

	avgWeight := 0.0
	if totalConns > 0 {
		avgWeight = totalWeight / float64(totalConns)
	}

	return map[string]interface{}{
		"total_connections": totalConns,
		"unique_sources":    len(h.connections),
		"average_weight":    avgWeight,
	}
}
```

---

## Commit 21.7: Implementar busqueda semantica

**Mensaje de commit:**
```
feat(memory): add semantic search with embeddings

- Cosine similarity for embedding comparison
- BM25 keyword search as fallback
- Integration with provider embeddings
```

### `goreview/internal/memory/search.go`

```go
package memory

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

// SemanticSearch implementa busqueda usando embeddings
type SemanticSearch struct {
	provider providers.Provider
	enabled  bool
}

// NewSemanticSearch crea un nuevo buscador semantico
func NewSemanticSearch(provider providers.Provider, enabled bool) *SemanticSearch {
	return &SemanticSearch{
		provider: provider,
		enabled:  enabled,
	}
}

// Search busca items semanticamente similares
func (s *SemanticSearch) Search(ctx context.Context, query string, items []Item, limit int) ([]SearchResult, error) {
	if !s.enabled || s.provider == nil {
		return s.keywordSearch(query, items, limit), nil
	}

	queryEmbed, err := s.provider.GetEmbedding(ctx, query)
	if err != nil {
		return s.keywordSearch(query, items, limit), nil
	}

	var results []SearchResult

	for _, item := range items {
		itemEmbed, err := s.getItemEmbedding(ctx, item)
		if err != nil {
			continue
		}

		similarity := cosineSimilarity(queryEmbed, itemEmbed)

		if similarity > 0.3 {
			results = append(results, SearchResult{
				Item:      item,
				Score:     similarity,
				MatchType: "semantic",
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (s *SemanticSearch) getItemEmbedding(ctx context.Context, item Item) ([]float64, error) {
	if embed, ok := item.Metadata["embedding"].([]float64); ok {
		return embed, nil
	}

	return s.provider.GetEmbedding(ctx, item.Content)
}

func (s *SemanticSearch) keywordSearch(query string, items []Item, limit int) []SearchResult {
	queryTerms := tokenize(query)
	var results []SearchResult

	for _, item := range items {
		score := s.bm25Score(queryTerms, item)
		if score > 0 {
			results = append(results, SearchResult{
				Item:      item,
				Score:     score,
				MatchType: "keyword",
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (s *SemanticSearch) bm25Score(queryTerms []string, item Item) float64 {
	content := strings.ToLower(item.Content)
	contentTerms := tokenize(content)

	k1 := 1.2
	b := 0.75
	avgDocLen := 100.0

	docLen := float64(len(contentTerms))
	var score float64

	termFreq := make(map[string]int)
	for _, term := range contentTerms {
		termFreq[term]++
	}

	for _, queryTerm := range queryTerms {
		tf := float64(termFreq[queryTerm])
		if tf == 0 {
			continue
		}

		idf := 1.0

		numerator := tf * (k1 + 1)
		denominator := tf + k1*(1-b+b*docLen/avgDocLen)
		score += idf * numerator / denominator
	}

	return score
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	})
	return words
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

---

## Commit 21.8: Implementar Store completo

**Mensaje de commit:**
```
feat(memory): add Store with 3-tier integration

- Integrate Working, Session, Long-term memories
- Automatic promotion based on importance
- Hebbian learning on search
- Complete Memory interface implementation
```

### `goreview/internal/memory/store.go`

```go
package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

// Store implementa Memory interface con los 3 niveles
type Store struct {
	working  *WorkingMemory
	session  *SessionMemory
	longterm *LongTermMemory
	hebbian  *HebbianLearning
	search   *SemanticSearch
	config   config.MemoryConfig
}

// NewStore crea un nuevo store de memoria
func NewStore(cfg config.MemoryConfig, provider providers.Provider) (*Store, error) {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory dir: %w", err)
	}

	ltm, err := NewLongTermMemory(cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to init long-term memory: %w", err)
	}

	hebbian := NewHebbianLearning(ltm.db, cfg.Hebbian)

	return &Store{
		working:  NewWorkingMemory(cfg.WorkingSize),
		session:  NewSessionMemory(500, cfg.SessionTTL),
		longterm: ltm,
		hebbian:  hebbian,
		search:   NewSemanticSearch(provider, cfg.SemanticSearch),
		config:   cfg,
	}, nil
}

func (s *Store) Store(ctx context.Context, item Item) error {
	if item.ID == "" {
		item.ID = generateID(item.Content)
	}

	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.AccessedAt = now

	displaced := s.working.Add(item)

	if displaced != nil {
		s.session.Add(*displaced)
		s.promoteToLongTerm(*displaced)
	}

	return nil
}

func (s *Store) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	var candidates []Item

	candidates = append(candidates, s.working.All()...)
	candidates = append(candidates, s.session.All()...)

	s.longterm.All(func(item Item) bool {
		candidates = append(candidates, item)
		return len(candidates) < 1000
	})

	results, err := s.search.Search(ctx, query, candidates, limit)
	if err != nil {
		return nil, err
	}

	var accessedIDs []string
	for _, r := range results {
		accessedIDs = append(accessedIDs, r.Item.ID)
	}
	s.hebbian.RecordCoAccess(accessedIDs)

	return results, nil
}

func (s *Store) Get(ctx context.Context, id string) (*Item, error) {
	if item, found := s.working.Get(id); found {
		return item, nil
	}

	if item, found := s.session.Get(id); found {
		return item, nil
	}

	return s.longterm.Get(id)
}

func (s *Store) GetRelated(ctx context.Context, id string, limit int) ([]SearchResult, error) {
	connections := s.hebbian.GetConnected(id, limit)

	var results []SearchResult
	for _, conn := range connections {
		item, err := s.Get(ctx, conn.TargetID)
		if err != nil || item == nil {
			continue
		}
		results = append(results, SearchResult{
			Item:      *item,
			Score:     conn.Weight,
			MatchType: "hebbian",
		})
	}

	return results, nil
}

func (s *Store) Update(ctx context.Context, item Item) error {
	item.AccessedAt = time.Now()

	if existing, found := s.working.Get(item.ID); found {
		item.AccessCount = existing.AccessCount + 1
		s.working.Add(item)
		return nil
	}

	if existing, found := s.session.Get(item.ID); found {
		item.AccessCount = existing.AccessCount + 1
		s.session.Add(item)
		return nil
	}

	return s.longterm.StoreWithIndex(item)
}

func (s *Store) Delete(ctx context.Context, id string) error {
	s.working.Remove(id)
	s.session.Remove(id)
	return s.longterm.DeleteWithIndex(id)
}

func (s *Store) RecordAccess(ctx context.Context, ids ...string) error {
	s.hebbian.RecordCoAccess(ids)
	return nil
}

func (s *Store) Stats(ctx context.Context) (*Stats, error) {
	ltCount, _ := s.longterm.Count()
	diskUsage, _ := s.longterm.DiskUsage()
	hebbianStats := s.hebbian.Stats()

	return &Stats{
		WorkingCount:   s.working.Size(),
		SessionCount:   s.session.Size(),
		LongTermCount:  ltCount,
		TotalCount:     s.working.Size() + s.session.Size() + ltCount,
		DiskUsageBytes: diskUsage,
		AvgActivation:  hebbianStats["average_weight"].(float64),
	}, nil
}

func (s *Store) Close() error {
	s.session.Close()
	return s.longterm.Close()
}

func (s *Store) promoteToLongTerm(item Item) {
	if item.Importance >= 0.5 || item.AccessCount >= 3 {
		s.longterm.StoreWithIndex(item)
	}
}

func generateID(content string) string {
	hash := sha256.Sum256([]byte(content + time.Now().String()))
	return hex.EncodeToString(hash[:16])
}

var _ Memory = (*Store)(nil)
```

---

## Commit 21.9: Agregar tests de memoria

**Mensaje de commit:**
```
test(memory): add memory system tests

- Test NoopMemory
- Test WorkingMemory overflow
- Test SessionMemory TTL
- Test Store integration
```

### `goreview/internal/memory/memory_test.go`

```go
package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNoopMemory(t *testing.T) {
	mem := NewNoop()
	ctx := context.Background()

	t.Run("Store returns nil", func(t *testing.T) {
		err := mem.Store(ctx, Item{ID: "test"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("Search returns empty", func(t *testing.T) {
		results, err := mem.Search(ctx, "query", 10)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if results != nil {
			t.Errorf("expected nil results, got %v", results)
		}
	})

	t.Run("Close returns nil", func(t *testing.T) {
		err := mem.Close()
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestWorkingMemory(t *testing.T) {
	t.Run("Add and Get", func(t *testing.T) {
		wm := NewWorkingMemory(10)

		item := Item{
			ID:      "test-1",
			Content: "test content",
			Type:    TypeIssue,
		}

		displaced := wm.Add(item)
		if displaced != nil {
			t.Error("expected no displacement")
		}

		got, found := wm.Get("test-1")
		if !found {
			t.Fatal("expected to find item")
		}
		if got.Content != "test content" {
			t.Errorf("expected 'test content', got %s", got.Content)
		}
	})

	t.Run("Overflow displaces oldest", func(t *testing.T) {
		wm := NewWorkingMemory(3)

		for i := 0; i < 5; i++ {
			wm.Add(Item{
				ID:      fmt.Sprintf("item-%d", i),
				Content: fmt.Sprintf("content %d", i),
			})
		}

		_, found := wm.Get("item-0")
		if found {
			t.Error("item-0 should have been displaced")
		}

		_, found = wm.Get("item-1")
		if found {
			t.Error("item-1 should have been displaced")
		}

		for i := 2; i < 5; i++ {
			_, found := wm.Get(fmt.Sprintf("item-%d", i))
			if !found {
				t.Errorf("item-%d should exist", i)
			}
		}
	})

	t.Run("Concurrent access", func(t *testing.T) {
		wm := NewWorkingMemory(100)
		var wg sync.WaitGroup

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				wm.Add(Item{
					ID:      fmt.Sprintf("concurrent-%d", n),
					Content: "test",
				})
			}(i)
		}

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				wm.Get(fmt.Sprintf("concurrent-%d", n))
			}(i)
		}

		wg.Wait()

		if wm.Size() != 50 {
			t.Errorf("expected 50 items, got %d", wm.Size())
		}
	})
}

func TestSessionMemory(t *testing.T) {
	t.Run("Add and Get", func(t *testing.T) {
		sm := NewSessionMemory(10, time.Hour)
		defer sm.Close()

		item := Item{ID: "test-1", Content: "content"}
		sm.Add(item)

		got, found := sm.Get("test-1")
		if !found {
			t.Fatal("expected to find item")
		}
		if got.Content != "content" {
			t.Errorf("expected 'content', got %s", got.Content)
		}
	})

	t.Run("LRU eviction", func(t *testing.T) {
		sm := NewSessionMemory(3, time.Hour)
		defer sm.Close()

		sm.Add(Item{ID: "1"})
		sm.Add(Item{ID: "2"})
		sm.Add(Item{ID: "3"})

		sm.Get("1")

		sm.Add(Item{ID: "4"})

		_, found := sm.Get("2")
		if found {
			t.Error("item 2 should have been evicted")
		}

		_, found = sm.Get("1")
		if !found {
			t.Error("item 1 should still exist")
		}
	})

	t.Run("TTL expiration", func(t *testing.T) {
		sm := NewSessionMemory(10, 50*time.Millisecond)
		defer sm.Close()

		sm.Add(Item{ID: "expiring"})

		_, found := sm.Get("expiring")
		if !found {
			t.Fatal("item should exist")
		}

		time.Sleep(100 * time.Millisecond)

		_, found = sm.Get("expiring")
		if found {
			t.Error("item should have expired")
		}
	})
}
```

---

## Resumen de la Iteracion 21

### Commits:
1. `feat(config): add memory configuration with feature flag`
2. `feat(memory): add Memory interface and NoopMemory`
3. `feat(memory): add WorkingMemory with circular buffer`
4. `feat(memory): add SessionMemory with LRU and TTL`
5. `feat(memory): add LongTermMemory with BadgerDB`
6. `feat(memory): add Hebbian learning for connections`
7. `feat(memory): add semantic search with embeddings`
8. `feat(memory): add Store with 3-tier integration`
9. `test(memory): add memory system tests`

### Archivos:
```
goreview/internal/memory/
├── memory.go        # Interface y tipos
├── noop.go          # NoopMemory (zero overhead)
├── factory.go       # Factory function
├── working.go       # Working memory (RAM)
├── session.go       # Session memory (LRU+TTL)
├── longterm.go      # Long-term (BadgerDB)
├── hebbian.go       # Hebbian learning
├── search.go        # Semantic search
├── store.go         # Store principal
└── memory_test.go   # Tests

goreview/internal/config/
├── config.go        # + MemoryConfig
└── defaults.go      # + DefaultMemoryConfig
```

### Verificacion:
```bash
cd goreview
go mod tidy
go test ./internal/memory/... -v
go build -o build/goreview ./cmd/goreview
./build/goreview review --staged --memory
```

---

## Siguiente Iteracion

Esta es la ultima iteracion del proyecto base. Posibles extensiones futuras:

- **22-MEMORY-CLI.md** - Comandos `goreview memory stats/clear`
- **23-KNOWLEDGE-GRAPH.md** - Visualizacion del grafo de conocimiento
- **24-MULTI-REPO.md** - Memoria compartida entre repositorios
