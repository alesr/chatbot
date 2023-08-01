package chatbot

import (
	"context"

	"github.com/alesr/chatbot/storage"
)

var _ Repository = &mockRepository{}

type mockRepository struct {
	StoreEmbeddingsFunc      func(ctx context.Context, in storage.StoreEmbeddingInput) error
	FetchNearestNeighborFunc func(ctx context.Context, in storage.FetchNearestNeighborInput) (text string, distance float64, errr error)
}

func (m *mockRepository) StoreEmbeddings(ctx context.Context, in storage.StoreEmbeddingInput) error {
	return m.StoreEmbeddingsFunc(ctx, in)
}

func (m *mockRepository) FetchNearestNeighbor(ctx context.Context, in storage.FetchNearestNeighborInput) (text string, distance float64, errr error) {
	return m.FetchNearestNeighborFunc(ctx, in)
}
