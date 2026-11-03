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
	"github.com/Arklr/goauth/internal/spa"
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
	postRepo := repository.NewPostRepository(pool)

	auth, err := handlers.NewAuthHandler(userRepo, sessionRepo, postRepo, cfg.CookieSecure)
	if err != nil {
		log.Fatal(err)
	}
	posts := handlers.NewPostHandler(postRepo)

	requireAuth := middleware.RequireAuth(sessionRepo, cfg.CookieSecure)

	// authed: logged-in only. protect: logged-in + valid X-CSRF-Token.
	authed := func(h http.HandlerFunc) http.Handler { return requireAuth(h) }
	protect := func(h http.HandlerFunc) http.Handler {
		return requireAuth(middleware.RequireCSRF(h))
	}

	mux := http.NewServeMux()

	// ---- JSON API ----
	mux.HandleFunc("POST /api/register", auth.Register)
	mux.HandleFunc("POST /api/login", auth.Login)
	mux.Handle("POST /api/logout", protect(auth.Logout))

	mux.Handle("GET /api/me", authed(auth.Me))
	mux.Handle("GET /api/dashboard", authed(auth.Dashboard))

	mux.Handle("GET /api/posts", authed(posts.List))
	mux.Handle("POST /api/posts", protect(posts.Create))
	mux.Handle("GET /api/posts/{id}", authed(posts.Get))
	mux.Handle("PUT /api/posts/{id}", protect(posts.Update))
	mux.Handle("DELETE /api/posts/{id}", protect(posts.Delete))

	// Unknown API routes get a JSON 404 instead of the React app.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}` + "\n"))
	})

	// ---- React app (built by `npm run build` in web/) ----
	mux.Handle("GET /", spa.Handler("web/dist"))

	log.Println("listening on :8080", addr)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
