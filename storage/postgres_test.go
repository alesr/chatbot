package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestStoreVector(t *testing.T) {
	db := setupDB(t)
	defer teardownDB(t, db)

	repo := NewPostgres(db)

	err := repo.StoreEmbeddings(context.TODO(), StoreEmbeddingInput{
		ID:           uuid.New().String(),
		CollectionID: uuid.New().String(),
		Model:        "test-model",
		Text:         "test-text",
		Tokens:       1,
		Vector:       vectorInputHelper(t),
		CreatedAt:    time.Time{}.Add(1),
	})
	require.NoError(t, err)
}

func setupDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open(
		"postgres",
		"postgres://user:password@localhost:5432/testdb?sslmode=disable",
	)
	require.NoError(t, err)
	return db
}

func teardownDB(t *testing.T, db *sqlx.DB) {
	err := db.Close()
	require.NoError(t, err)
}

// pq: expected 1536 dimensions
func vectorInputHelper(t *testing.T) []float32 {
	vector := make([]float32, 1536)
	for i := 0; i < 1536; i++ {
		vector[i] = float32(i)
	}
	return vector
}
