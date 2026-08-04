package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/Evgen-Poloniy/url-shortener/internal/domain"
	"github.com/Evgen-Poloniy/url-shortener/internal/repository/memory"
)

func TestMemoryRepository_SaveURL(t *testing.T) {
	type fields struct {
		fullToShort map[string]string
		shortToFull map[string]string
	}
	type args struct {
		full  string
		short string
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   error
		checkMaps func(t *testing.T, repo *memory.MemoryRepository)
	}{
		{
			name: "success: save new url pair",
			fields: fields{
				fullToShort: make(map[string]string),
				shortToFull: make(map[string]string),
			},
			args: args{
				full:  "https://example.com/long-url",
				short: "aX9kL1mN2o",
			},
			wantErr: nil,
			checkMaps: func(t *testing.T, repo *memory.MemoryRepository) {
				gotFull, err := repo.GetFullURL(context.Background(), "aX9kL1mN2o")
				if err != nil || gotFull != "https://example.com/long-url" {
					t.Errorf("expected full URL 'https://example.com/long-url', got '%s'", gotFull)
				}
			},
		},
		{
			name: "success: idempotent save (same full and same short)",
			fields: fields{
				fullToShort: map[string]string{"https://example.com": "shortKey12"},
				shortToFull: map[string]string{"shortKey12": "https://example.com"},
			},
			args: args{
				full:  "https://example.com",
				short: "shortKey12",
			},
			wantErr: nil,
		},
		{
			name: "error: conflict when full url already exists with different short key",
			fields: fields{
				fullToShort: map[string]string{"https://example.com": "shortKey12"},
				shortToFull: map[string]string{"shortKey12": "https://example.com"},
			},
			args: args{
				full:  "https://example.com",
				short: "newShortKey",
			},
			wantErr: domain.ErrURLConflict,
		},
		{
			name: "error: conflict when short key is already taken by another full url",
			fields: fields{
				fullToShort: map[string]string{"https://other.com": "shortKey12"},
				shortToFull: map[string]string{"shortKey12": "https://other.com"},
			},
			args: args{
				full:  "https://newdomain.com",
				short: "shortKey12",
			},
			wantErr: domain.ErrURLConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewMemoryRepository()
			for f, s := range tt.fields.fullToShort {
				_ = repo.SaveURL(context.Background(), f, s)
			}

			err := repo.SaveURL(context.Background(), tt.args.full, tt.args.short)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("SaveURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkMaps != nil {
				tt.checkMaps(t, repo)
			}
		})
	}
}

func TestMemoryRepository_GetFullURL(t *testing.T) {
	type args struct {
		short string
	}

	tests := []struct {
		name    string
		setup   func(repo *memory.MemoryRepository)
		args    args
		want    string
		wantErr error
	}{
		{
			name: "success: get existing full url",
			setup: func(repo *memory.MemoryRepository) {
				_ = repo.SaveURL(context.Background(), "https://golang.org", "goShortKey")
			},
			args: args{
				short: "goShortKey",
			},
			want:    "https://golang.org",
			wantErr: nil,
		},
		{
			name:  "error: short key not found",
			setup: func(repo *memory.MemoryRepository) {},
			args: args{
				short: "nonExistentKey",
			},
			want:    "",
			wantErr: domain.ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewMemoryRepository()
			tt.setup(repo)

			got, err := repo.GetFullURL(context.Background(), tt.args.short)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetFullURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("GetFullURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryRepository_ConcurrentSafety(t *testing.T) {
	repo := memory.NewMemoryRepository()
	ctx := context.Background()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			_ = repo.SaveURL(ctx, "https://example.com/page", "staticKey")
		}(i)

		go func(id int) {
			defer wg.Done()
			_, _ = repo.GetFullURL(ctx, "staticKey")
		}(i)
	}

	wg.Wait()
}

func TestMemoryRepository_ConcurrentSafety_MultipleKeys(t *testing.T) {
	repo := memory.NewMemoryRepository()
	ctx := context.Background()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			full := fmt.Sprintf("https://example.com/page_%d", id)
			short := fmt.Sprintf("short_%d", id)

			if err := repo.SaveURL(ctx, full, short); err != nil {
				t.Errorf("goroutine %d: failed to save URL: %v", id, err)
				return
			}

			got, err := repo.GetFullURL(ctx, short)
			if err != nil {
				t.Errorf("goroutine %d: failed to get URL: %v", id, err)
				return
			}
			if got != full {
				t.Errorf("goroutine %d: got %s, want %s", id, got, full)
			}
		}(i)
	}

	wg.Wait()
}
