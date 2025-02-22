package main

import (
	"RAG1/internal/db"
	"RAG1/internal/env"
	"RAG1/internal/store"
	"log"
	
	_ "github.com/lib/pq"
)

func main() {

	
	cfg := serverConfig{
	    addr: ":8080",
	    db: dbconfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/rag1?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	log.Println("database connection pool is established")

	app := &application{
		config: cfg,
		store:  store.NewStorage(db),
	}

	mux := app.mount()

    log.Fatal(app.run(mux))
}
