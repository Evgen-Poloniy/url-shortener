package memory_test

import (
	"context"
	"testing"

	"github.com/Evgen-Poloniy/url-shortener/internal/repository/memory"
)

func TestNewMemoryRepository(t *testing.T) {
	repo := memory.NewMemoryRepository()
	if repo == nil {
		t.Fatal("expected non-nil MemoryRepository")
	}

	err := repo.SaveURL(context.Background(), "https://example.com", "short1")
	if err != nil {
		t.Fatalf("unexpected error on fresh repo: %v", err)
	}
}
