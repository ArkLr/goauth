package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/Arklr/goauth/internal/cookies"
	"github.com/Arklr/goauth/internal/models"
	"github.com/Arklr/goauth/internal/repository"
	"github.com/Arklr/goauth/internal/security"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionTTL      = 24 * time.Hour
	msgInvalidLogin = "invalid email or password"
	msgInternal     = "something went wrong, try again"
)

type AuthHandler struct {
	users         *repository.UserRepository
	sessions      *repository.SessionRepository
	posts         *repository.PostRepository
	secureCookies bool
	dummyHash     []byte
}

func NewAuthHandler(users *repository.UserRepository, sessions *repository.SessionRepository, posts *repository.PostRepository, secureCookies bool) (*AuthHandler, error) {
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
		posts:         posts,
		secureCookies: secureCookies,
		dummyHash:     dummyHash,
	}, nil
}

type registerRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// Register handles POST /api/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if len(req.Password) > 72 {
		writeError(w, http.StatusBadRequest, "password must be at most 72 bytes")
		return
	}
	if req.ConfirmPassword != req.Password {
		writeError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}

	if _, err := h.users.Create(r.Context(), email, string(hash)); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "email already taken")
			return
		}
		log.Printf("create user: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}

	writeJSON(w, http.StatusCreated, nil)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /api/login. On success it sets the session cookie;
// the client then calls GET /api/me to load the user and CSRF token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
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
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}

	cmpErr := bcrypt.CompareHashAndPassword(hash, []byte(req.Password))
	if err != nil || cmpErr != nil {
		writeError(w, http.StatusUnauthorized, msgInvalidLogin)
		return
	}

	token, err := security.GenerateToken()
	if err != nil {
		log.Printf("generate token: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}
	csrfToken, err := security.GenerateToken()
	if err != nil {
		log.Printf("generate csrf token: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
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
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}

	cookies.SetSession(w, token, expires, h.secureCookies)
	writeJSON(w, http.StatusNoContent, nil)
}

// Logout handles POST /api/logout (auth + CSRF protected).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookies.SessionName); err == nil && c.Value != "" {
		if err := h.sessions.Delete(r.Context(), security.HashToken(c.Value)); err != nil {
			log.Printf("delete session: %v", err)
			writeError(w, http.StatusInternalServerError, msgInternal)
			return
		}
	}

	cookies.ClearSession(w, h.secureCookies)
	writeJSON(w, http.StatusNoContent, nil)
}

type meResponse struct {
	User      *models.User `json:"user"`
	CSRFToken string       `json:"csrfToken"`
	ExpiresAt time.Time    `json:"expiresAt"`
}

// Me handles GET /api/me: the current user, session expiry and CSRF token.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}

	user, err := h.users.GetByID(r.Context(), s.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		log.Printf("get user by id: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		User:      user,
		CSRFToken: s.CSRFToken,
		ExpiresAt: s.ExpiresAt,
	})
}

type dashboardResponse struct {
	PostCount   int64         `json:"postCount"`
	RecentPosts []models.Post `json:"recentPosts"`
}

// Dashboard handles GET /api/dashboard.
func (h *AuthHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}

	count, err := h.posts.CountByUser(r.Context(), s.UserId)
	if err != nil {
		log.Printf("count posts: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}

	recent, err := h.posts.List(r.Context(), s.UserId, 3, 0)
	if err != nil {
		log.Printf("list recent posts: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}

	writeJSON(w, http.StatusOK, dashboardResponse{
		PostCount:   count,
		RecentPosts: nonNil(recent),
	})
}
