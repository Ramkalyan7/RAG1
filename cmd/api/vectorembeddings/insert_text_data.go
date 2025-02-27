package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func CreateVectorEmbedding(textData string) ([][]float32, error) {

	llm, err := ollama.New(ollama.WithModel("nomic-embed-text"))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	ctx := context.Background()
	embs, err := llm.CreateEmbedding(ctx, []string{textData})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return embs, nil
}

func StoreEmbsInQdrant(embs [][]float32, data string) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// client.CreateCollection(context.Background(), &qdrant.CreateCollection{
	// 	CollectionName: "pdf_vector_embeddings_collection",
	// 	VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
	// 		Size:     768,
	// 		Distance: qdrant.Distance_Cosine,
	// 	}),
	// })

	operationInfo, err := client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: "pdf_vector_embeddings_collection",
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewIDNum(1),
				Vectors: qdrant.NewVectors(embs[0]...),
				Payload: qdrant.NewValueMap(map[string]interface{}{
					"payload": data,
				}),
			},
		},
	})

	if err != nil {
		log.Fatalf("Failed to upsert data: %v", err)
	}

	if operationInfo.Status != qdrant.UpdateStatus_Completed {
		log.Fatalf("Upsert operation failed with status: %v", operationInfo.Status)
	}

	log.Println("Upsert operation completed successfully")

}

func CreateQueryVectorEmbedding() ([][]float32, error) {

	llm, err := ollama.New(ollama.WithModel("nomic-embed-text"))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	ctx := context.Background()
	embs, err := llm.CreateEmbedding(ctx, []string{"what are Ram's skills"})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return embs, nil
}

func SearchVectorEmbedding(embs [][]float32) (string, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Qdrant client: %w", err)
	}
	defer client.Close()

	time.Sleep(3 * time.Second)

	searchResults, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "pdf_vector_embeddings_collection",
		Query:          qdrant.NewQuery(embs[0]...),
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(true),
	})
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	if len(searchResults) == 0 {
		return "", fmt.Errorf("no search results found")
	}

	var contextText strings.Builder

	for _, searchResult := range searchResults {
		payload := searchResult.Payload
		textValue, exists := payload["payload"]
		if exists {
			contextText.WriteString(textValue.GetStringValue())
		} else {
			return "", fmt.Errorf("'payload' field not found in search result")
		}
	}

	if contextText.Len() == 0 {
		return "", fmt.Errorf("'payload' field is empty or not a string")
	}

	return contextText.String(), nil
}


func getResponseForUserQuery(contextText string) error {

	llm, err := ollama.New(ollama.WithModel("deepseek-r1"))
	if err != nil {
		return fmt.Errorf("failed to create Ollama client: %w", err)
	}

	contextText = "Use the following retrieved context to answer the query accurately:\n\n" + contextText

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: contextText},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text:"what are Ram's skills"},
			},
		},
	}
	choices, err := llm.GenerateContent(context.Background(), messages)

	if err != nil {
		log.Fatal("err")
		//log.Fatal(choices)
	}

	log.Print(choices.Choices[0].Content)
	return nil
}

func ChunkText(text string, maxChunkSize int) []string {
	words := strings.Split(text, " ")
	var chunks []string

	var currentChunk strings.Builder

	for _, word := range words {
		if currentChunk.Len()+len(word)+1 > maxChunkSize {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}

		currentChunk.WriteString(word)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

type EmbeddingWithPayload struct {
	payload   string
	embedding [][]float32
}

func CreateChunkEmbeddings(chunks []string) ([]EmbeddingWithPayload, error) {
	llm, err := ollama.New(ollama.WithModel("nomic-embed-text"))
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	embeddings, err := llm.CreateEmbedding(context.Background(), chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	if len(embeddings) != len(chunks) {
		return nil, fmt.Errorf("mismatch between number of chunks and embeddings: %d vs %d", len(chunks), len(embeddings))
	}

	var embeddingsWithPayloadSlice []EmbeddingWithPayload

	for i, emb := range embeddings {
		embeddingsWithPayloadSlice = append(embeddingsWithPayloadSlice, EmbeddingWithPayload{
			payload:   chunks[i],
			embedding: [][]float32{emb},
		})
	}
	return embeddingsWithPayloadSlice, nil
}

func StoreMultiplyeEmbsInQdrant(embeddingsWithPayloadSlice []EmbeddingWithPayload) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})

	if err != nil {
		log.Fatalf("failed to create qdrant client: %v", err)
	}
	defer client.Close()

	points := make([]*qdrant.PointStruct,len(embeddingsWithPayloadSlice))

	for i, emb := range embeddingsWithPayloadSlice {
		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i + 1)),
			Vectors: qdrant.NewVectors(emb.embedding[0]...),
			Payload: qdrant.NewValueMap(map[string]interface{}{
				"payload": emb.payload,
			}),
		}
	}

	operationInfo, err := client.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: "pdf_vector_embeddings_collection",
		Points:         points,
	})
	if err != nil {
		log.Fatalf("failed to upsert data: %v", err)
	}

	if operationInfo.Status != qdrant.UpdateStatus_Completed {
		log.Fatalf("upsert operation failed with status: %v", operationInfo.Status)
	}

	log.Println("Batch upsert operation completed successfully")
	
}

func main() {
	embs,err := CreateQueryVectorEmbedding();
	if err!=nil{
		log.Print("embs1")
		return
	}
    
	contextText,err := SearchVectorEmbedding(embs)
	if err!=nil{
		log.Print(err)
		return
	}
	getResponseForUserQuery(contextText);

}
