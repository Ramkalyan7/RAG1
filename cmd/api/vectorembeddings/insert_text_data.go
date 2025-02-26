package main

import (

	"context"
	"fmt"
	"log"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func CreateVectorEmbedding( textData string) ([][]float32, error) {
	
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

func StoreEmbsInQdrant(embs [][]float32,data string) {
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
					"payload":     data,
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
	embs, err := llm.CreateEmbedding(ctx, []string{"what are the projects that Ram have done"})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return embs, nil
}

func SearchVectorEmbedding(embs [][]float32)(string, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	time.Sleep(3 * time.Second)

	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "pdf_vector_embeddings_collection",
		Query:          qdrant.NewQuery(embs[0]...),
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(true),
	})

	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	if len(searchResult) == 0 {
		return "", fmt.Errorf("no search results found")
	}
	payload := searchResult[0].Payload
	textValue, exists := payload["payload"]
	if !exists {
		return "", fmt.Errorf("'text' field not found in payload")
	}

	contextText := textValue.GetStringValue()
	if contextText == "" {
		return "", fmt.Errorf("'text' field is empty or not a string")
	}

	return contextText, nil
}

func getResponseForUserQuery(contextText string) error {


	llm, err := ollama.New(ollama.WithModel("deepseek-r1"))
	if err != nil {
		return fmt.Errorf("failed to create Ollama client: %w", err)
	}

	contextText = "Use the following retrieved context to answer the query accurately:\n\n" + contextText
	//log.Print(contextText)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: contextText},
			},
		},
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "what are the projects that Ram have done"},
			},
		},
	}
	choices, err := llm.GenerateContent(context.Background(), messages)

	if err!=nil{
		log.Fatal("err")
		//log.Fatal(choices)
	}

	log.Print(choices.Choices[0].Content)
	return nil
}

func main() {
	embs,err:=CreateQueryVectorEmbedding()
	if err!=nil{
		log.Fatal(err)
	}

	ctx,err:=SearchVectorEmbedding(embs)
	if err!=nil{
		log.Fatal(err)
	}

	getResponseForUserQuery(ctx);
}
