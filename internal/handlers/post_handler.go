package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Arklr/goauth/internal/middleware"
	"github.com/Arklr/goauth/internal/models"
	"github.com/Arklr/goauth/internal/repository"
)

const (
	maxHeadline = 200
	maxBody     = 10000
	pageSize    = 50
)

type PostHandler struct {
	posts *repository.PostRepository
}

func NewPostHandler(posts *repository.PostRepository) *PostHandler {
	return &PostHandler{posts: posts}
}

// session returns the logged-in session or writes a 401.
func session(w http.ResponseWriter, r *http.Request) (*models.Session, bool) {
	s, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
	}
	return s, ok
}

func postID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

// nonNil makes empty lists encode as [] instead of null.
func nonNil(p []models.Post) []models.Post {
	if p == nil {
		return []models.Post{}
	}
	return p
}

type postRequest struct {
	Headline string `json:"headline"`
	Body     string `json:"body"`
}

// readPost decodes and validates a post body. On failure it writes the
// error response and returns ok=false.
func readPost(w http.ResponseWriter, r *http.Request) (headline, body string, ok bool) {
	var req postRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return "", "", false
	}

	headline = strings.TrimSpace(req.Headline)
	body = strings.TrimSpace(req.Body)

	var msg string
	switch {
	case headline == "":
		msg = "headline is required"
	case utf8.RuneCountInString(headline) > maxHeadline:
		msg = fmt.Sprintf("headline must be at most %d characters", maxHeadline)
	case body == "":
		msg = "body is required"
	case utf8.RuneCountInString(body) > maxBody:
		msg = fmt.Sprintf("body must be at most %d characters", maxBody)
	}
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return "", "", false
	}
	return headline, body, true
}

type listResponse struct {
	Posts []models.Post `json:"posts"`
}

// List handles GET /api/posts. ?filter=mine narrows it to the current user.
func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}

	var authorID int64
	if r.URL.Query().Get("filter") == "mine" {
		authorID = s.UserId
	}

	list, err := h.posts.List(r.Context(), authorID, pageSize, 0)
	if err != nil {
		log.Printf("list posts: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}
	writeJSON(w, http.StatusOK, listResponse{Posts: nonNil(list)})
}

// Get handles GET /api/posts/{id}. Only the owner can read a single post
// (it's used by the edit page); everyone else gets 404.
func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	id, ok := postID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	p, err := h.posts.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("get post: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}
	if p.UserId != s.UserId { // not the owner: look the same as "doesn't exist"
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// Create handles POST /api/posts.
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	headline, body, ok := readPost(w, r)
	if !ok {
		return
	}

	p, err := h.posts.Create(r.Context(), s.UserId, headline, body)
	if err != nil {
		log.Printf("create post: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// Update handles PUT /api/posts/{id}.
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	id, ok := postID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	headline, body, ok := readPost(w, r)
	if !ok {
		return
	}

	p, err := h.posts.Update(r.Context(), id, s.UserId, headline, body)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("update post: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// Delete handles DELETE /api/posts/{id}.
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	id, ok := postID(r)
	if !ok {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	if err := h.posts.Delete(r.Context(), id, s.UserId); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		log.Printf("delete post: %v", err)
		writeError(w, http.StatusInternalServerError, msgInternal)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}
