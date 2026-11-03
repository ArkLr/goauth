package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Arklr/goauth/internal/cookies"
	"github.com/Arklr/goauth/internal/models"
	"github.com/Arklr/goauth/internal/repository"
	"github.com/Arklr/goauth/internal/security"
)

type contextKey struct{}

var sessionKey contextKey

// SessionFromContext returns the session that RequireAuth stored.
func SessionFromContext(ctx context.Context) (*models.Session, bool) {
	s, ok := ctx.Value(sessionKey).(*models.Session)
	return s, ok
}

func RequireAuth(sessions *repository.SessionRepository, secure bool) func(http.Handler) http.Handler {
	deny := func(w http.ResponseWriter, r *http.Request) {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(cookies.SessionName)
			if err != nil || c.Value == "" {
				deny(w, r)
				return
			}

			s, err := sessions.GetByToken(r.Context(), security.HashToken(c.Value))
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					cookies.ClearSession(w, secure)
					deny(w, r)
					return
				}
				log.Printf("get session: %v", err)
				jsonError(w, http.StatusInternalServerError, "internal error")
				return
			}

			ctx := context.WithValue(r.Context(), sessionKey, s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// jsonError writes {"error": msg} with the given status.
func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
