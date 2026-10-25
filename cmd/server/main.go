package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Arklr/goauth/internal/config"
	"github.com/Arklr/goauth/internal/database"
	"github.com/Arklr/goauth/internal/handlers"
	"github.com/Arklr/goauth/internal/middleware"
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

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)

	fmt.Println("Connected")

	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	auth, err := handlers.NewAuthHandler(userRepo, sessionRepo, cfg.CookieSecure)
	if err != nil {
		log.Fatal(err)
	}

	requireAuth := middleware.RequireAuth(sessionRepo, cfg.CookieSecure)

	mux := http.NewServeMux()

	// static files (CSS)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// pages
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /register", auth.RegisterPage)
	mux.HandleFunc("GET /login", auth.LoginPage)

	// actions
	mux.HandleFunc("POST /register", auth.Register)
	mux.HandleFunc("POST /login", auth.Login)
	mux.HandleFunc("POST /logout", auth.Logout)

	// protected
	mux.Handle("GET /protected", requireAuth(http.HandlerFunc(auth.Protected)))
	log.Println("listening on :8080", addr)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
