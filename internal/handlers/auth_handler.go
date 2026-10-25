package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"text/template"
	"time"

	"github.com/Arklr/goauth/internal/middleware"
	"github.com/Arklr/goauth/internal/models"
	"github.com/Arklr/goauth/internal/repository"
	"github.com/Arklr/goauth/internal/security"
	"golang.org/x/crypto/bcrypt"

	"github.com/Arklr/goauth/internal/cookies"
)

const (
	sessionTTL      = 24 * time.Hour
	msgInvalidLogin = "invalid email or password"
)

type pageData struct {
	Error     string
	Email     string
	Notice    string
	CSRFToken string
	ExpiresAt time.Time
}

type AuthHandler struct {
	users         *repository.UserRepository
	sessions      *repository.SessionRepository
	tmpl          *template.Template
	secureCookies bool
	dummyHash     []byte
}

func NewAuthHandler(users *repository.UserRepository, sessions *repository.SessionRepository, secureCookies bool) (*AuthHandler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

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
		tmpl:          tmpl,
		secureCookies: secureCookies,
		dummyHash:     dummyHash,
	}, nil
}

func (h *AuthHandler) render(w http.ResponseWriter, status int, name string, data pageData) {
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, "register.html", pageData{})
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := pageData{}
	if r.URL.Query().Get("registered") == "1" {
		data.Notice = "Account created. Please log in."
	}
	h.render(w, http.StatusOK, "login.html", data)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := r.ParseForm(); err != nil {
		h.render(w, http.StatusBadRequest, "register.html", pageData{Error: "invalid form data"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	fail := func(status int, msg string) {
		h.render(w, status, "register.html", pageData{Error: msg, Email: email})
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		fail(http.StatusBadRequest, "invalid email")
		return
	}
	if len(password) < 8 {
		fail(http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if len(password) > 72 {
		fail(http.StatusBadRequest, "password must be at most 72 bytes")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt: %v", err)
		fail(http.StatusInternalServerError, "something went wrong, try again")
		return
	}

	if _, err := h.users.Create(r.Context(), email, string(hash)); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			fail(http.StatusConflict, "email already taken")
			return
		}
		log.Printf("create user: %v", err)
		fail(http.StatusInternalServerError, "something went wrong, try again")
		return
	}

	http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if err := r.ParseForm(); err != nil {
		h.render(w, http.StatusBadRequest, "login.html", pageData{Error: "invalid form data"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")

	fail := func(status int, msg string) {
		h.render(w, status, "login.html", pageData{Error: msg, Email: email})
	}

	if email == "" || password == "" {
		fail(http.StatusBadRequest, "email and password are required")
		return
	}

	hash := h.dummyHash
	user, err := h.users.GetByEmail(r.Context(), email)
	switch {
	case err == nil:
		hash = []byte(user.PasswordHash)
	case errors.Is(err, repository.ErrNotFound):
		// keep the dummy hash so timing stays similar
	default:
		log.Printf("get user: %v", err)
		fail(http.StatusInternalServerError, "something went wrong, try again")
		return
	}

	cmpErr := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil || cmpErr != nil {
		fail(http.StatusUnauthorized, msgInvalidLogin)
		return
	}

	token, err := security.GenerateToken()
	if err != nil {
		log.Printf("generate token: %v", err)
		fail(http.StatusInternalServerError, "something went wrong, try again")
		return
	}
	csrfToken, err := security.GenerateToken()
	if err != nil {
		log.Printf("generate csrf token: %v", err)
		fail(http.StatusInternalServerError, "something went wrong, try again")
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
		fail(http.StatusInternalServerError, "something went wrong, try again")
		return
	}

	cookies.SetSession(w, token, expires, h.secureCookies)
	http.Redirect(w, r, "/protected", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookies.SessionName); err == nil && c.Value != "" {
		if err := h.sessions.Delete(r.Context(), security.HashToken(c.Value)); err != nil {
			log.Printf("delete session: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	cookies.ClearSession(w, h.secureCookies)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request) {
	s, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.users.GetByID(r.Context(), s.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		log.Printf("get user by id: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.render(w, http.StatusOK, "protected.html", pageData{
		Email:     user.Email,
		CSRFToken: s.CSRFToken,
		ExpiresAt: s.ExpiresAt,
	})
}
