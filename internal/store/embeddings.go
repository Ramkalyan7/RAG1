package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/qdrant/go-client/qdrant"
)


type EmbeddingWithPayload struct {
	Payload   string
	Embedding [][]float32
}


type EmbeddingsStore struct{
	qdrant *qdrant.Client
}

func (e *EmbeddingsStore) StoreEmbdsInQdrant(embeddingsWithPayloads []EmbeddingWithPayload)error{
	points := make([]*qdrant.PointStruct,len(embeddingsWithPayloads))

	for i, emb := range embeddingsWithPayloads {
		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i + 1)),
			Vectors: qdrant.NewVectors(emb.Embedding[0]...),
			Payload: qdrant.NewValueMap(map[string]interface{}{
				"payload": emb.Payload,
			}),
		}
	}

	operationInfo, err := e.qdrant.Upsert(context.Background(), &qdrant.UpsertPoints{
		CollectionName: "pdf_vector_embeddings_collection",
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("failed to create qdrant client: %v", err)
	}

	if operationInfo.Status != qdrant.UpdateStatus_Acknowledged && operationInfo.Status != qdrant.UpdateStatus_Completed {
		return fmt.Errorf("upsert operation failed") 
	}

	return nil
}


func (e *EmbeddingsStore)SearchEmbdsQdrant(embds [][]float32) (string, error){
	searchResults, err := e.qdrant.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "pdf_vector_embeddings_collection",
		Query:          qdrant.NewQuery(embds[0]...),
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
		if !exists {
			return "", fmt.Errorf("'payload' field not found in search result")
		}
		contextText.WriteString(textValue.GetStringValue())
	}

	if contextText.Len() == 0 {
		return "", fmt.Errorf("'payload' field is empty or not a string")
	}

	return contextText.String(), nil
}