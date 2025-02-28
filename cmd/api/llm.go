package main

import (
	"RAG1/internal/store"
	"context"
	"fmt"
	"net/http"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type UserQueryPayload struct{
	Query string `json:"query" validate:"required"`
}

func (app *application) CreateLLMConnection(llmName string)(*ollama.LLM,error){
	llm, err := ollama.New(ollama.WithModel(llmName))

	if err != nil {
		return nil,fmt.Errorf("cannot create the LLM")
	}
	return llm,nil
}


func (app *application) CreateEmbedddings(textData string)([][]float32,error){
	
	llm, err := app.CreateLLMConnection("nomic-embed-text");
	if err != nil {
		return nil,fmt.Errorf("cannot create the LLM")
	}

	ctx := context.Background()
	embs, err := llm.CreateEmbedding(ctx, []string{textData})
	if err != nil {
		return nil,fmt.Errorf("cannot create the embeddings")
	}

	return embs,nil;
}


func (app *application) CreateChunkEmbeddings(chunks []string) ([]store.EmbeddingWithPayload, error) {
	llm, err := app.CreateLLMConnection("nomic-embed-text");
	if err != nil {
		return nil,fmt.Errorf("cannot create the LLM")
	}
	embeddings, err := llm.CreateEmbedding(context.Background(), chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	if len(embeddings) != len(chunks) {
		return nil, fmt.Errorf("mismatch between number of chunks and embeddings: %d vs %d", len(chunks), len(embeddings))
	}

	var embeddingsWithPayloadSlice []store.EmbeddingWithPayload

	for i, emb := range embeddings {
		embeddingsWithPayloadSlice = append(embeddingsWithPayloadSlice, store.EmbeddingWithPayload{
			Payload:   chunks[i],
			Embedding: [][]float32{emb},
		})
	}
	return embeddingsWithPayloadSlice, nil
}


func (app *application) GenerateResponseFromLLm(queryContext,query string)(string,error){
	llm, err := app.CreateLLMConnection("deepseek-r1")
	if err != nil {
		return "",fmt.Errorf("failed to create llm:GenerateResponseFromLLm")
	}

	queryContext = "Use the following retrieved context to answer the query accurately:\n\n" + queryContext

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: queryContext},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text:query},
			},
		},
	}
	choices, err := llm.GenerateContent(context.Background(), messages)

	if err!=nil {
		return "",fmt.Errorf("could not generate response from llm")
	}
	return choices.Choices[0].Content, err;
}



func (app *application) storeUserDataAsEmbeddingsHandler(w http.ResponseWriter,r *http.Request){
	//read pdf
	pdfText ,err := app.ReadPdfFile("cmd/sourcedata/pavancv.pdf")

	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
	//chunk text
	chunks , err := app.CreateOverlappingChunks(pdfText,3,1);
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
	//create embeddings for chunks
	textChunksWithEmbeddings ,err:= app.CreateChunkEmbeddings(chunks);
	
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
	//Pass embeddings to qdrantStore to store them
	err = app.store.Embeddings.StoreEmbdsInQdrant(textChunksWithEmbeddings)
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
	err=app.jsonResponse(w,http.StatusAccepted,"stored data as embeddings sucessfully!");
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
}

func (app *application) userQueryHandler(w http.ResponseWriter,r *http.Request){
	var payload UserQueryPayload;

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	embeddings,err:=app.CreateEmbedddings(payload.Query);
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
	
	contextText,err:=app.store.Embeddings.SearchEmbdsQdrant(embeddings);
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}

	response,err:=app.GenerateResponseFromLLm(contextText,payload.Query)
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
	err=app.jsonResponse(w,http.StatusAccepted,response);
	if err!=nil{
		writeJSONError(w,http.StatusInternalServerError,err.Error())
		return
	}
}