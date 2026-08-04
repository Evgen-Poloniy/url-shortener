package memory

import "sync"

// MemoryRepository represents RAM memory repository for save URLs.
type MemoryRepository struct {
	mu          sync.RWMutex
	shortToFull map[string]string // Key: short URL -> Value: full URL
	fullToShort map[string]string // Key: full URL -> Value: short URL
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		shortToFull: make(map[string]string),
		fullToShort: make(map[string]string),
	}
}
