package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/Arklr/goauth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users    *repository.UserRepository
	sessions *repository.SessionRepository
}

func NewAuthHandler(users *repository.UserRepository, sessions *repository.SessionRepository) *AuthHandler {
	return &AuthHandler{users: users, sessions: sessions}
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
	w.Write([]byte("login\n"))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("logout\n"))
}

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("protected\n"))
}
