package db

import (
	"context"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

func NewQdrantDB(host string, port int) (*qdrant.Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})

	collectionName := "pdf_vector_embeddings_collection"
	ctx := context.Background()

	if err != nil {
		return nil, err
	}

	// Try to get the collection
	Exists, err := client.CollectionExists(ctx, collectionName)

	if err!=nil{
		log.Fatalf("Error while to searching collection: %v", err)
		return nil, err
	}

	if !Exists {
		err=client.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     768,
				Distance: qdrant.Distance_Cosine,
			}),
		})

		if err!=nil{
			log.Fatalf("failed to create collection: %v", err)
		}
	}
	
	return client, nil
}
