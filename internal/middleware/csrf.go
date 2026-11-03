package middleware

import (
	"crypto/subtle"
	"net/http"
)

// RequireCSRF checks the X-CSRF-Token header on state-changing requests
// against the token stored in the session. It must run after RequireAuth.
//
// The React app gets the token from GET /api/me and sends it on every
// POST/PUT/DELETE.
func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods don't change state, so they need no token.
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		s, ok := SessionFromContext(r.Context())
		if !ok {
			jsonError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		sent := r.Header.Get("X-CSRF-Token")
		if sent == "" || subtle.ConstantTimeCompare([]byte(sent), []byte(s.CSRFToken)) != 1 {
			jsonError(w, http.StatusForbidden, "invalid CSRF token")
			return
		}

		next.ServeHTTP(w, r)
	})
}
