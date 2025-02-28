package main

import (
	"RAG1/internal/store"
	"log"
	"net/http"
	

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config serverConfig
	store store.Storage
}

type serverConfig struct{
	addr string
	db dbconfig
	qdb qdrantDbConfig
	env string
}

type dbconfig struct{
	addr string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime string
}

type qdrantDbConfig struct{
	host string
	port int
}

func (app *application) mount()http.Handler{
	r:=chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	//r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1",func(r chi.Router) {
		r.Get("/health",app.healthCheckHandler)
		r.Get("/storedata",app.storeUserDataAsEmbeddingsHandler)
		r.Get("/query",app.userQueryHandler)
	});


	return r;
}

func (app *application) run(mux http.Handler)error {
	srv := &http.Server{
		Addr: app.config.addr,
		Handler: mux,
	}

	log.Printf("server running on port %s",app.config.addr)

	return srv.ListenAndServe()
}
