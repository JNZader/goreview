package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// HebbianLearnerImpl implements HebbianLearner with persistent storage.
// It implements the principle "neurons that fire together wire together".
type HebbianLearnerImpl struct {
	mu           sync.RWMutex
	db           *badger.DB
	learningRate float64
	decayRate    float64
	minStrength  float64
}

// HebbianOptions configures Hebbian learning.
type HebbianOptions struct {
	Dir          string
	LearningRate float64 // How fast associations strengthen (0-1)
	DecayRate    float64 // How fast associations decay (0-1)
	MinStrength  float64 // Minimum strength before pruning (0-1)
}

const (
	associationPrefix = "assoc:"
)

// NewHebbianLearner creates a new Hebbian learning instance.
func NewHebbianLearner(opts HebbianOptions) (*HebbianLearnerImpl, error) {
	badgerOpts := badger.DefaultOptions(opts.Dir)
	badgerOpts.Logger = nil

	db, err := badger.Open(badgerOpts)
	if err != nil {
		return nil, fmt.Errorf("opening badger db: %w", err)
	}

	// Set defaults
	learningRate := opts.LearningRate
	if learningRate <= 0 || learningRate > 1 {
		learningRate = 0.1
	}

	decayRate := opts.DecayRate
	if decayRate <= 0 || decayRate > 1 {
		decayRate = 0.01
	}

	minStrength := opts.MinStrength
	if minStrength <= 0 || minStrength > 1 {
		minStrength = 0.1
	}

	return &HebbianLearnerImpl{
		db:           db,
		learningRate: learningRate,
		decayRate:    decayRate,
		minStrength:  minStrength,
	}, nil
}

// Compile-time interface check.
var _ HebbianLearner = (*HebbianLearnerImpl)(nil)

// Strengthen increases the association between two entries.
// Uses Hebbian rule: w_new = w_old + learningRate * (1 - w_old)
func (h *HebbianLearnerImpl) Strengthen(ctx context.Context, sourceID, targetID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.makeKey(sourceID, targetID)

	return h.db.Update(func(txn *badger.Txn) error {
		assoc, err := h.getAssociation(txn, key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		if assoc == nil {
			assoc = &Association{
				SourceID:  sourceID,
				TargetID:  targetID,
				Strength:  0,
				CreatedAt: time.Now(),
			}
		}

		// Hebbian learning: asymptotic approach to 1.0
		assoc.Strength = assoc.Strength + h.learningRate*(1-assoc.Strength)
		assoc.CoActivations++
		assoc.UpdatedAt = time.Now()

		return h.setAssociation(txn, key, assoc)
	})
}

// Weaken decreases the association between two entries.
// Uses anti-Hebbian rule: w_new = w_old - decayRate * w_old
func (h *HebbianLearnerImpl) Weaken(ctx context.Context, sourceID, targetID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := h.makeKey(sourceID, targetID)

	return h.db.Update(func(txn *badger.Txn) error {
		assoc, err := h.getAssociation(txn, key)
		if err == badger.ErrKeyNotFound {
			return nil // Nothing to weaken
		}
		if err != nil {
			return err
		}

		// Anti-Hebbian: decay toward 0
		assoc.Strength = assoc.Strength - h.decayRate*assoc.Strength
		assoc.UpdatedAt = time.Now()

		// Remove if below threshold
		if assoc.Strength < h.minStrength {
			return txn.Delete([]byte(key))
		}

		return h.setAssociation(txn, key, assoc)
	})
}

// GetAssociations returns associations for an entry.
func (h *HebbianLearnerImpl) GetAssociations(ctx context.Context, id string) ([]*Association, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	associations := make([]*Association, 0)
	prefix := []byte(associationPrefix + id + ":")

	err := h.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var assoc Association
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &assoc)
			})
			if err != nil {
				continue
			}

			associations = append(associations, &assoc)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("getting associations: %w", err)
	}

	// Also get reverse associations (where id is target)
	err = h.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			// Skip if not an association or already included
			if len(key) < len(associationPrefix) {
				continue
			}

			var assoc Association
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &assoc)
			})
			if err != nil {
				continue
			}

			// Include if target matches
			if assoc.TargetID == id {
				associations = append(associations, &assoc)
			}
		}
		return nil
	})

	return associations, err
}

// Decay applies time-based decay to all associations.
// Uses exponential decay: w_new = w_old * exp(-decayRate * timeDelta)
func (h *HebbianLearnerImpl) Decay(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	keysToUpdate := make(map[string]*Association)
	keysToDelete := make([]string, 0)
	now := time.Now()

	// Find all associations and calculate decay
	err := h.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(associationPrefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(associationPrefix)); it.ValidForPrefix([]byte(associationPrefix)); it.Next() {
			item := it.Item()
			key := string(item.Key())

			var assoc Association
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &assoc)
			})
			if err != nil {
				continue
			}

			// Calculate time-based decay
			timeDelta := now.Sub(assoc.UpdatedAt).Hours() / 24.0 // Days since update
			decayFactor := math.Exp(-h.decayRate * timeDelta)
			assoc.Strength = assoc.Strength * decayFactor

			if assoc.Strength < h.minStrength {
				keysToDelete = append(keysToDelete, key)
			} else {
				assocCopy := assoc
				keysToUpdate[key] = &assocCopy
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("scanning associations: %w", err)
	}

	// Update and delete in batch
	return h.db.Update(func(txn *badger.Txn) error {
		// Delete weak associations
		for _, key := range keysToDelete {
			if err := txn.Delete([]byte(key)); err != nil {
				return err
			}
		}

		// Update remaining associations
		for key, assoc := range keysToUpdate {
			data, err := json.Marshal(assoc)
			if err != nil {
				continue
			}
			if err := txn.Set([]byte(key), data); err != nil {
				return err
			}
		}
		return nil
	})
}

// Prune removes associations below minimum strength.
func (h *HebbianLearnerImpl) Prune(ctx context.Context, minStrength float64) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	keysToDelete := make([]string, 0)

	err := h.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(associationPrefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(associationPrefix)); it.ValidForPrefix([]byte(associationPrefix)); it.Next() {
			item := it.Item()

			var assoc Association
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &assoc)
			})
			if err != nil {
				continue
			}

			if assoc.Strength < minStrength {
				keysToDelete = append(keysToDelete, string(item.Key()))
			}
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("scanning for prune: %w", err)
	}

	if len(keysToDelete) == 0 {
		return 0, nil
	}

	err = h.db.Update(func(txn *badger.Txn) error {
		for _, key := range keysToDelete {
			if err := txn.Delete([]byte(key)); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("deleting weak associations: %w", err)
	}

	return len(keysToDelete), nil
}

// Close releases resources.
func (h *HebbianLearnerImpl) Close() error {
	return h.db.Close()
}

// Helper methods

func (h *HebbianLearnerImpl) makeKey(sourceID, targetID string) string {
	return associationPrefix + sourceID + ":" + targetID
}

func (h *HebbianLearnerImpl) getAssociation(txn *badger.Txn, key string) (*Association, error) {
	item, err := txn.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	var assoc Association
	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &assoc)
	})
	if err != nil {
		return nil, err
	}

	return &assoc, nil
}

func (h *HebbianLearnerImpl) setAssociation(txn *badger.Txn, key string, assoc *Association) error {
	data, err := json.Marshal(assoc)
	if err != nil {
		return fmt.Errorf("marshaling association: %w", err)
	}
	return txn.Set([]byte(key), data)
}

// GetAllAssociations returns all associations (for debugging/export).
func (h *HebbianLearnerImpl) GetAllAssociations(ctx context.Context) ([]*Association, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	associations := make([]*Association, 0)

	err := h.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(associationPrefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(associationPrefix)); it.ValidForPrefix([]byte(associationPrefix)); it.Next() {
			item := it.Item()

			var assoc Association
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &assoc)
			})
			if err != nil {
				continue
			}

			associations = append(associations, &assoc)
		}
		return nil
	})

	return associations, err
}

// Stats returns Hebbian learner statistics.
func (h *HebbianLearnerImpl) Stats(ctx context.Context) (total int64, avgStrength float64, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var sumStrength float64

	err = h.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(associationPrefix)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(associationPrefix)); it.ValidForPrefix([]byte(associationPrefix)); it.Next() {
			item := it.Item()

			var assoc Association
			ierr := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &assoc)
			})
			if ierr != nil {
				continue
			}

			total++
			sumStrength += assoc.Strength
		}
		return nil
	})

	if total > 0 {
		avgStrength = sumStrength / float64(total)
	}

	return total, avgStrength, err
}
