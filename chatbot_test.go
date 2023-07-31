package chatbot

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/alesr/chatbot/client/openaicli"
	"github.com/alesr/chatbot/storage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestTrain(t *testing.T) {
	client := openaicli.New(os.Getenv("OPENAI_API_KEY"), &http.Client{})

	db, err := sqlx.Open(
		"postgres",
		"postgres://user:password@localhost:5432/testdb?sslmode=disable",
	)
	require.NoError(t, err)

	repo := storage.NewPostgres(db)

	svc := NewService(os.Getenv("OPENAI_API_KEY"), client, repo)

	input := TrainInput{
		UserID: "user1",
		Model:  defaultModel,
		Data: []io.Reader{
			strings.NewReader(
				`Unbearable Lightness of Being is an exploration of human life in its intricacies and contradictions. A poignant excerpt reads, "Human life occurs only once, and the reason we cannot determine which of our decisions are good and which are bad is that in a given situation we can make only one decision; we are not granted a second, third, or fourth life for comparison." It beautifully captures the essence of the book's exploration of existence, love, and choices.`,
			),
			strings.NewReader(
				`Milan Kundera is a renowned Czech-born French author, known for his profound and philosophical narratives. His works often blend elements of history, politics, and existentialism. Kundera's most famous novel, "The Unbearable Lightness of Being," is celebrated for its introspective exploration of love and fate, reflecting his unique style of merging philosophy with fiction. He is esteemed for his insightful contributions to 20th-century literature.`,
			),
			strings.NewReader(
				`Neptune, the eighth and farthest known planet from the Sun in our solar system, is known for its striking blue color, attributed to its atmosphere composed of hydrogen, helium, and methane. This gas giant, discovered in 1846, has a strong magnetic field, extreme weather conditions with the fastest wind speeds in the solar system, and a total of 14 known moons. The largest, Triton, is the only large moon in the solar system that orbits in the opposite direction of its planet's rotation, a retrograde orbit.`,
			),
			strings.NewReader(
				`In Roman mythology, Neptune was the god of freshwater and the sea, a counterpart to the Greek god Poseidon. Known for his powerful and volatile nature, he ruled all bodies of water and was often depicted holding a trident. Neptune was also one of the chief gods and was considered the brother of Jupiter and Pluto. The Romans held a festival, Neptunalia, in his honor, which was marked by games and water-based festivities. It's fascinating to note the parallels between the turbulent and mysterious nature of the planet and its namesake god.`,
			),
		},
	}

	got, err := svc.Train(context.TODO(), input)

	t.Log(got)

	require.NoError(t, err)
}

func TestAsk(t *testing.T) {
	client := openaicli.New(os.Getenv("OPENAI_API_KEY"), &http.Client{})

	db, err := sqlx.Open(
		"postgres",
		"postgres://user:password@localhost:5432/testdb?sslmode=disable",
	)
	require.NoError(t, err)

	repo := storage.NewPostgres(db)

	svc := NewService(os.Getenv("OPENAI_API_KEY"), client, repo)

	input := `What neptune the planned has to do with the roman god?`

	got, err := svc.Ask(context.Background(), "user1", "coll-f59fa771-3c98-4c05-8ae4-3b069915f59e", input)
	require.NoError(t, err)

	t.Log(got)
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
