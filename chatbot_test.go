package chatbot

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/alesr/chatbot/client/openaicli"
	"github.com/alesr/chatbot/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrain(t *testing.T) {
	client := mockClient{
		CreateEmbeddingFunc: func(ctx context.Context, in openaicli.EmbbedingRequest) (*openaicli.EmbeddingResponse, error) {
			return &openaicli.EmbeddingResponse{
				Usage: openaicli.Usage{
					TotalTokens: 1,
				},
				Data: []openaicli.Embedding{
					{
						Embedding: []float32{1.0, 2.0, 3.0},
					},
				},
			}, nil
		},
	}

	repo := mockRepository{
		StoreEmbeddingsFunc: func(ctx context.Context, in storage.StoreEmbeddingInput) error {
			return nil
		},
	}

	svc := NewService("test-api-key", &client, &repo)

	input := TrainInput{
		UserID: "test-user",
		Model:  defaultModel,
		Data: []io.Reader{
			strings.NewReader("word1 word2 word3"),
		},
	}

	collectionID, err := svc.Train(context.Background(), input)
	require.NoError(t, err)

	// Returned ID is unpredicatable, but it should not be empty
	assert.NotEmpty(t, collectionID)

	// Returned ID starts with "coll-"
	assert.Equal(t, "coll-", collectionID[:5])

	// The rest of the ID is a valid UUID
	_, err = uuid.Parse(collectionID[5:])
	assert.NoError(t, err)
}

func TestAsk(t *testing.T) {
	client := mockClient{
		CreateEmbeddingFunc: func(ctx context.Context, in openaicli.EmbbedingRequest) (*openaicli.EmbeddingResponse, error) {
			return &openaicli.EmbeddingResponse{
				Usage: openaicli.Usage{
					TotalTokens: 1,
				},
				Data: []openaicli.Embedding{
					{
						Embedding: []float32{1.0, 2.0, 3.0},
					},
				},
			}, nil
		},
		CreateChatCompletitionFunc: func(ctx context.Context, in openaicli.CompletitionRequest) (*openaicli.CompletitionResponse, error) {
			return &openaicli.CompletitionResponse{
				Choices: []openaicli.Choice{
					{
						Message: openaicli.Message{
							Content: "42",
						},
					},
				},
			}, nil
		},
	}

	repo := mockRepository{
		StoreEmbeddingsFunc: func(ctx context.Context, in storage.StoreEmbeddingInput) error {
			return nil
		},
		FetchNearestNeighborFunc: func(ctx context.Context, in storage.FetchNearestNeighborInput) (text string, distance float64, err error) {
			return "", 0.0, nil
		},
	}

	svc := NewService("test-api-key", &client, &repo)

	userID := "test-user"
	collectionID := "coll-" + uuid.NewString()
	question := "What is the meaning of life?"

	answer, err := svc.Ask(context.Background(), userID, collectionID, question)
	require.NoError(t, err)

	assert.Equal(t, "42", answer)
}

func TestReadData(t *testing.T) {
	tests := []struct {
		name      string
		data      []string
		chunkSize int
		expected  []string
	}{
		{
			name:      "Single reader, single word",
			data:      []string{"word"},
			chunkSize: 1,
			expected:  []string{"word "},
		},
		{
			name:      "Single reader, multiple words, chunk size 1",
			data:      []string{"word1 word2 word3"},
			chunkSize: 1,
			expected:  []string{"word1 ", "word2 ", "word3 "},
		},
		{
			name:      "Single reader, multiple words, chunk size 2",
			data:      []string{"word1 word2 word3 word4"},
			chunkSize: 2,
			expected:  []string{"word1 word2 ", "word3 word4 "},
		},
		{
			name:      "Single reader, multiple words, chunk size greater than number of words",
			data:      []string{"word1 word2 word3"},
			chunkSize: 5,
			expected:  []string{"word1 word2 word3 "},
		},
		{
			name:      "Multiple readers, single word each",
			data:      []string{"word1", "word2"},
			chunkSize: 1,
			expected:  []string{"word1 ", "word2 "},
		},
		{
			name:      "Multiple readers, multiple words each, chunk size 1",
			data:      []string{"word1 word2", "word3 word4"},
			chunkSize: 1,
			expected:  []string{"word1 ", "word2 ", "word3 ", "word4 "},
		},
		{
			name:      "Multiple readers, multiple words each, chunk size 2",
			data:      []string{"word1 word2 word3", "word4 word5 word6"},
			chunkSize: 2,
			expected:  []string{"word1 word2 ", "word3 ", "word4 word5 ", "word6 "},
		},
		{
			name:      "Multiple readers, multiple words each, chunk size greater than number of words",
			data:      []string{"word1 word2", "word3 word4"},
			chunkSize: 3,
			expected:  []string{"word1 word2 ", "word3 word4 "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readers := make([]io.Reader, len(tt.data))
			for i, data := range tt.data {
				readers[i] = strings.NewReader(data)
			}

			result, err := readData(readers, tt.chunkSize)
			require.NoError(t, err)

			require.Equal(t, tt.expected, result)
		})
	}
}
