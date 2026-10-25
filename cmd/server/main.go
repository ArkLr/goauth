package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Arklr/goauth/internal/config"
	"github.com/Arklr/goauth/internal/database"
	"github.com/Arklr/goauth/internal/handlers"
	"github.com/Arklr/goauth/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := database.Connect(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	fmt.Println("Conncted")

	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	auth := handlers.NewAuthHandler(userRepo, sessionRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", auth.Register)
	mux.HandleFunc("POST /login", auth.Login)
	mux.HandleFunc("POST /logout", auth.Logout)
	mux.HandleFunc("GET /protected", auth.Protected)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
