package chatbot

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/alesr/chatbot/client/openaicli"
	"github.com/alesr/chatbot/storage"
	"github.com/google/uuid"
)

const (
	defaultModel     OpenAIModel = "text-embedding-ada-002"
	defaultChunkSize int         = 500
)

type (
	// OpenAIModel represents the model used by OpenAI.
	OpenAIModel string

	// Client represents the OpenAI client for creating completitions and embeddings.
	Client interface {
		CreateEmbedding(ctx context.Context, in openaicli.EmbbedingRequest) (*openaicli.EmbeddingResponse, error)
		CreateChatCompletition(ctx context.Context, in openaicli.CompletitionRequest) (*openaicli.CompletitionResponse, error)
	}

	// Repository represents the storage repository for storing
	// embeddings and fetching nearest neighbors.
	Repository interface {
		StoreEmbeddings(ctx context.Context, in storage.StoreEmbeddingInput) error
		FetchNearestNeighbor(ctx context.Context, in storage.FetchNearestNeighborInput) (text string, distance float64, errr error)
	}

	// TrainInput represents the input for training.
	TrainInput struct {
		UserID string
		Model  OpenAIModel
		Data   []io.Reader
	}

	// AskInput represents the input for asking questions.
	AskInput struct {
		UserID       string
		CollectionID string
		Question     string
	}

	// Service represents the chatbot service.
	// It provides methods for training by creating embeddings,
	// and asking questions by fetching nearest neighbors and
	// creating completitions.
	Service struct {
		apiKey string
		client Client
		repo   Repository
	}
)

// NewService returns a new chatbot service.
func NewService(apiKey string, client Client, repo Repository) *Service {
	return &Service{
		apiKey: apiKey,
		client: client,
		repo:   repo,
	}
}

// Train trains the chatbot by creating embeddings for the given data.
// It stores both the input data as well as the embeddings in the repository,
// and returns the collection ID.
func (s *Service) Train(ctx context.Context, in TrainInput) (string, error) {
	var collectionID string = "coll-" + uuid.NewString()
	readDataCh := make(chan []string)
	errCh := make(chan error)

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(in.Data))

		for _, d := range in.Data {
			go func(d io.Reader) {
				defer wg.Done()

				chunks, err := readData([]io.Reader{d}, defaultChunkSize)
				if err != nil {
					errCh <- fmt.Errorf("could not read data: %w", err)
					return
				}

				readDataCh <- chunks
			}(d)
		}

		wg.Wait()
		close(readDataCh)
	}()

	go func() {
		var wg sync.WaitGroup

		for chunks := range readDataCh {
			wg.Add(len(chunks))

			for _, chunk := range chunks {
				go func(chunk string) {
					defer wg.Done()

					if err := s.processChunk(
						ctx, in.UserID, collectionID, chunk, string(in.Model),
					); err != nil {
						errCh <- fmt.Errorf("error occurred during training: %w", err)
						return
					}
				}(chunk)
			}
		}

		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return "", err
		}
	}

	return collectionID, nil
}

// processChunk creates embeddings for the given chunk of data,
func (s *Service) processChunk(ctx context.Context, userID, collectionID, chunk, model string) error {
	embedd, err := s.client.CreateEmbedding(ctx, openaicli.EmbbedingRequest{
		Model: model,
		Input: chunk,
	})
	if err != nil {
		return fmt.Errorf("could not create embeddings: %w", err)
	}

	if err := s.repo.StoreEmbeddings(ctx,
		storage.StoreEmbeddingInput{
			ID:           "emb-" + uuid.NewString(),
			UserID:       userID,
			CollectionID: collectionID,
			Model:        string(model),
			Text:         chunk,
			Tokens:       int64(embedd.Usage.TotalTokens),
			Vector:       embedd.Data[0].Embedding,
			CreatedAt:    time.Now().UTC(),
		}); err != nil {
		return fmt.Errorf("could not store vector: %w", err)
	}
	return nil
}

// Ask asks the chatbot a question by fetching the nearest neighbor and creating a chat completition.
func (s *Service) Ask(ctx context.Context, userID, collectionID, question string) (string, error) {
	embedd, err := s.client.CreateEmbedding(ctx, openaicli.EmbbedingRequest{
		Model: string(defaultModel),
		Input: question,
	})
	if err != nil {
		return "", fmt.Errorf("could not create embeddings: %w", err)
	}

	text, _, err := s.repo.FetchNearestNeighbor(ctx, storage.FetchNearestNeighborInput{
		UserID:       userID,
		CollectionID: collectionID,
		Vector:       embedd.Data[0].Embedding,
	})
	if err != nil {
		return "", fmt.Errorf("could not fetch nearest neighbor: %w", err)
	}

	completition, err := s.client.CreateChatCompletition(ctx, openaicli.CompletitionRequest{
		Model: string(defaultModel),
		Messages: []openaicli.Message{
			{
				Role:    "system",
				Content: text,
			},
			{
				Role:    "user",
				Content: question,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("could not create completition: %w", err)
	}

	return completition.Choices[0].Message.Content, nil
}

// readData reads the data from the given readers and returns a slice of strings.
func readData(data []io.Reader, chunkSize int) ([]string, error) {
	chunks := make([]string, 0)
	for _, d := range data {
		scanner := bufio.NewScanner(d)
		scanner.Split(bufio.ScanWords)

		sb := strings.Builder{}

		var counter int

		for scanner.Scan() {
			sb.WriteString(scanner.Text())
			sb.WriteString(" ")
			counter++

			if counter == chunkSize {
				chunks = append(chunks, sb.String())
				sb.Reset()
				counter = 0
			}
		}

		// if there's remaining data in the builder, add it to chunks
		if sb.Len() > 0 {
			chunks = append(chunks, sb.String())
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("could not scan data: %w", err)
		}
	}

	return chunks, nil
}
