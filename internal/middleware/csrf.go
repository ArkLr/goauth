package middleware

import (
	"crypto/subtle"
	"net/http"
)

func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods don't change state, so they need no token.
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		// The session comes from RequireAuth, which must run first.
		s, ok := SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		// Header for fetch/AJAX clients, form field for HTML forms.
		sent := r.Header.Get("X-CSRF-Token")
		if sent == "" {
			sent = r.PostFormValue("csrf_token")
		}

		if subtle.ConstantTimeCompare([]byte(sent), []byte(s.CSRFToken)) != 1 {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
