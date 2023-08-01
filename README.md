# chatbot
Implements a chatbot with custom embeddings using OpenAI APIs and pgVector.

## Example:

The following example shows how to train a model and ask a question.
When calling the Train method, the service will read the data from the
provided io.Reader, split it into chunks and create a openAI embedding for
each chunk. The embeddings are then stored in a pgVector database, together
with the original text, user and collection ID.
Note that each user can have multiple collections, and each collection can
have multiple embeddings.

When calling the Ask method, the service will create an embedding for the
question and compare it to the embeddings in the collection using cosine
similarity. The answer with the highest similarity is returned.

Then the method request the OpenAI API to generate a chat completion using the
text from the most similar embedding as the prompt, and the question.

```go
func ExampleService_Train() {
	client := openaicli.New(os.Getenv("OPENAI_API_KEY"), &http.Client{})

	db, _ := sqlx.Open(
		"postgres",
		"postgres://user:password@localhost:5432/testdb?sslmode=disable",
	)

	repo := storage.NewPostgres(db)

	svc := chatbot.NewService(os.Getenv("OPENAI_API_KEY"), client, repo)

	input := chatbot.TrainInput{
		UserID: "user1",
		Model:  "text-embedding-ada-002",
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

	collectionID, _ := svc.Train(context.TODO(), input)

	fmt.Print(collectionID)

	// Output example: coll-00000000-0000-0000-0000-000000000000
}


func ExampleService_Ask() {
	client := openaicli.New(os.Getenv("OPENAI_API_KEY"), &http.Client{})

	db, _ := sqlx.Open(
		"postgres",
		"postgres://user:password@localhost:5432/testdb?sslmode=disable",
	)

	repo := storage.NewPostgres(db)

	svc := chatbot.NewService(os.Getenv("OPENAI_API_KEY"), client, repo)

	input := `What neptune the planned has to do with the roman god?`

	answer, _ := svc.Ask(context.Background(), "user1", "coll-00000000-0000-0000-0000-000000000000", input)

	fmt.Print(answer)

	// Output example: Given the deep influence of ancient mythology on naming celestial bodies, the planet was named after the Roman god of the sea, Neptune, who held a similar role to the Greek god Poseidon. The Roman god was associated with the sea, freshwater, and other bodies of water, symbolizing both their tranquil and tempestuous aspects. The name Neptune was chosen to capture the mysterious and powerful nature of the planet, which lies so distant in the outer reaches of our solar system.
}
```
