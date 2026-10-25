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
	_, _ = userRepo, sessionRepo

	mux := http.NewServeMux()

	mux.HandleFunc("/register", handlers.Register)
	mux.HandleFunc("/login", handlers.Login)
	mux.HandleFunc("/logout", handlers.Logout)
	mux.HandleFunc("/protected", handlers.Protected)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
