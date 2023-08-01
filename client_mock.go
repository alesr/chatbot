package chatbot

import (
	"context"

	"github.com/alesr/chatbot/client/openaicli"
)

var _ Client = &mockClient{}

type mockClient struct {
	CreateEmbeddingFunc        func(ctx context.Context, in openaicli.EmbbedingRequest) (*openaicli.EmbeddingResponse, error)
	CreateChatCompletitionFunc func(ctx context.Context, in openaicli.CompletitionRequest) (*openaicli.CompletitionResponse, error)
}

func (m *mockClient) CreateEmbedding(ctx context.Context, in openaicli.EmbbedingRequest) (*openaicli.EmbeddingResponse, error) {
	return m.CreateEmbeddingFunc(ctx, in)
}

func (m *mockClient) CreateChatCompletition(ctx context.Context, in openaicli.CompletitionRequest) (*openaicli.CompletitionResponse, error) {
	return m.CreateChatCompletitionFunc(ctx, in)
}
