package memory

import (
	"context"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
)

// SaveURL saves full and short URLs into repository.
func (m *MemoryRepository) SaveURL(ctx context.Context, full, short string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existingShort, exists := m.fullToShort[full]; exists {
		if existingShort == short {
			return nil
		}
		return domain.ErrURLConflict
	}

	if _, exists := m.shortToFull[short]; exists {
		return domain.ErrURLConflict
	}

	m.fullToShort[full] = short
	m.shortToFull[short] = full

	return nil
}

// GetFullURL gets full URL from repository.
func (m *MemoryRepository) GetFullURL(ctx context.Context, short string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	full, exists := m.shortToFull[short]
	if !exists {
		return "", domain.ErrURLNotFound
	}

	return full, nil
}
