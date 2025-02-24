package db

import "github.com/qdrant/go-client/qdrant"



func NewQdrantDB(host string,port int64)(*qdrant.Client, error) {
	client , err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})

	if err!=nil{
		return nil,err;
	}

	return client,nil;
}