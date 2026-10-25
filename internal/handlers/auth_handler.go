package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/Arklr/goauth/internal/models"
	"github.com/Arklr/goauth/internal/repository"
	"github.com/Arklr/goauth/internal/security"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "session"
	sessionTTL        = 24 * time.Hour
	msgInvalidLogin   = "invalid email or password"
)

type AuthHandler struct {
	users         *repository.UserRepository
	sessions      *repository.SessionRepository
	secureCookies bool
	dummyHash     []byte
}

func NewAuthHandler(users *repository.UserRepository, sessions *repository.SessionRepository, secureCookies bool) (*AuthHandler, error) {
	dummy, err := security.GenerateToken()
	if err != nil {
		return nil, err
	}
	dummyHash, err := bcrypt.GenerateFromPassword([]byte(dummy), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("dummy hash: %w", err)
	}

	return &AuthHandler{
		users:         users,
		sessions:      sessions,
		secureCookies: secureCookies,
		dummyHash:     dummyHash,
	}, nil

}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB cap

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if len(password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	if len(password) > 72 {
		http.Error(w, "password must be at most 72 bytes", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err = h.users.Create(r.Context(), email, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			http.Error(w, "email already taken", http.StatusConflict)
			return
		}
		log.Printf("create user: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("registered\n"))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB cap

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")
	if email == "" || password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	hash := h.dummyHash
	user, err := h.users.GetByEmail(r.Context(), email)
	switch {
	case err == nil:
		hash = []byte(user.PasswordHash)
	case errors.Is(err, repository.ErrNotFound):
		// fall through with dummyHash
	default:
		log.Printf("get user: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	cmpErr := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil || cmpErr != nil {
		http.Error(w, msgInvalidLogin, http.StatusUnauthorized)
		return
	}

	token, err := security.GenerateToken()
	if err != nil {
		log.Printf("generate token: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	csrfToken, err := security.GenerateToken()
	if err != nil {
		log.Printf("generate csrf token: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	expires := time.Now().Add(sessionTTL)
	err = h.sessions.Create(r.Context(), &models.Session{
		Token:     security.HashToken(token),
		UserId:    user.Id,
		CSRFToken: csrfToken,
		ExpiresAt: expires,
	})

	if err != nil {
		log.Printf("create session: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	fmt.Fprintf(w, "logged in\ncsrf_token: %s\n", csrfToken)

}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("logout\n"))
}

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("protected\n"))
}
