package store

import (
	"database/sql"
	"github.com/qdrant/go-client/qdrant"
)



type Storage struct{
   Embeddings interface{
		StoreEmbdsInQdrant([]EmbeddingWithPayload)error
		SearchEmbdsQdrant([][]float32) (string, error)
   }
}


func NewStorage(db *sql.DB,quadrantDb *qdrant.Client)Storage{
	return Storage{
		Embeddings: &EmbeddingsStore{quadrantDb},
	}
}