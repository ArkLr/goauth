package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
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

type postsPageData struct {
	Error         string
	CSRFToken     string
	CurrentUserId int64
	Posts         []models.Post
	Mine          bool         // list shows only the current user's posts
	Post          *models.Post // used by the edit page
	Headline      string
	Body          string
}

type PostHandler struct {
	posts *repository.PostRepository
	tmpl  *template.Template
}

func NewPostHandler(posts *repository.PostRepository) (*PostHandler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &PostHandler{posts: posts, tmpl: tmpl}, nil
}

func (h *PostHandler) render(w http.ResponseWriter, status int, name string, data postsPageData) {
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

// session returns the logged-in session or redirects to /login.
func session(w http.ResponseWriter, r *http.Request) (*models.Session, bool) {
	s, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	return s, ok
}

func postID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

func validate(r *http.Request) (headline, body, msg string) {
	headline = strings.TrimSpace(r.PostFormValue("headline"))
	body = strings.TrimSpace(r.PostFormValue("body"))

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
	return
}

func (h *PostHandler) showList(w http.ResponseWriter, r *http.Request, s *models.Session, status int, errMsg, headline, body string) {
	// ?filter=mine narrows the list to the current user's posts.
	mine := r.URL.Query().Get("filter") == "mine"
	var authorID int64
	if mine {
		authorID = s.UserId
	}

	list, err := h.posts.List(r.Context(), authorID, pageSize, 0)
	if err != nil {
		log.Printf("list posts: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.render(w, status, "posts.html", postsPageData{
		Error:         errMsg,
		CSRFToken:     s.CSRFToken,
		CurrentUserId: s.UserId,
		Posts:         list,
		Mine:          mine,
		Headline:      headline,
		Body:          body,
	})
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	h.showList(w, r, s, http.StatusOK, "", "", "")
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}

	headline, body, msg := validate(r)
	if msg != "" {
		h.showList(w, r, s, http.StatusBadRequest, msg, headline, body)
		return
	}

	if _, err := h.posts.Create(r.Context(), s.UserId, headline, body); err != nil {
		log.Printf("create post: %v", err)
		h.showList(w, r, s, http.StatusInternalServerError, "something went wrong, try again", headline, body)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	id, ok := postID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	headline, body, msg := validate(r)
	if msg != "" {
		h.render(w, http.StatusBadRequest, "post_edit.html", postsPageData{
			Error:     msg,
			CSRFToken: s.CSRFToken,
			Post:      &models.Post{Id: id},
			Headline:  headline,
			Body:      body,
		})
		return
	}

	_, err := h.posts.Update(r.Context(), id, s.UserId, headline, body)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("update post: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	id, ok := postID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if err := h.posts.Delete(r.Context(), id, s.UserId); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("delete post: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *PostHandler) EditPage(w http.ResponseWriter, r *http.Request) {
	s, ok := session(w, r)
	if !ok {
		return
	}
	id, ok := postID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	p, err := h.posts.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("get post: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if p.UserId != s.UserId { // not the owner: look the same as "doesn't exist"
		http.NotFound(w, r)
		return
	}

	h.render(w, http.StatusOK, "post_edit.html", postsPageData{
		CSRFToken: s.CSRFToken,
		Post:      p,
		Headline:  p.Headline,
		Body:      p.Body,
	})
}
