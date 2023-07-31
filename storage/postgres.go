package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type Postgres struct{ *sqlx.DB }

func NewPostgres(dbConn *sqlx.DB) *Postgres {
	return &Postgres{DB: dbConn}
}

type StoreEmbeddingInput struct {
	ID           string
	UserID       string
	CollectionID string
	Model        string
	Text         string
	Tokens       int64
	Vector       []float32
	CreatedAt    time.Time
}

const queryInsertEmbedding string = `INSERT INTO embeddings 
(id, user_id, collection_id, model, text, tokens, vector, created_at) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

func (p *Postgres) StoreEmbeddings(ctx context.Context, in StoreEmbeddingInput) error {
	if _, err := p.ExecContext(
		ctx, queryInsertEmbedding, in.ID, in.UserID,
		in.CollectionID, in.Model, in.Text, in.Tokens,
		pgvector.NewVector(in.Vector), in.CreatedAt,
	); err != nil {
		return fmt.Errorf("could not store vector: %w", err)
	}
	return nil
}

type FetchModelInput struct {
	UserID       string
	CollectionID string
}

const queryFetchModel string = "SELECT model FROM embeddings WHERE user_id = $1 AND collection_id = $2 LIMIT 10"

// SELECT model FROM embeddings WHERE user_id = 'user-1' AND collection_id = 'coll-39f2d70c-bc44-48a3-95ff-e732f62ce893' LIMIT 10

func (p *Postgres) FetchModel(ctx context.Context, in FetchModelInput) (string, error) {
	var model string
	if err := p.GetContext(ctx, &model,
		queryFetchModel,
		in.UserID, in.CollectionID,
	); err != nil {
		return "", fmt.Errorf("could not fetch model: %w", err)
	}
	return model, nil
}

type FetchNearestNeighborInput struct {
	UserID       string
	CollectionID string
	Vector       []float32
}

const queryFetchNearestNeighbor string = `SELECT text, vector <-> $1 AS distance
FROM embeddings
WHERE user_id = $2 AND collection_id = $3
ORDER BY distance ASC
LIMIT 1`

func (p *Postgres) FetchNearestNeighbor(ctx context.Context, in FetchNearestNeighborInput) (string, float64, error) {
	var (
		text     string
		distance float64
	)

	row := p.QueryRowxContext(ctx,
		queryFetchNearestNeighbor,
		pgvector.NewVector(in.Vector), in.UserID, in.CollectionID,
	)

	if err := row.Scan(&text, &distance); err != nil {
		return "", 0, fmt.Errorf("could not fetch nearest neighbor: %w", err)
	}

	return text, distance, nil
}
