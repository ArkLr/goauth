package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
)

const maxBodyBytes = 1 << 20

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

var errBadJSON = errors.New("invalid request body")

// readJSON decodes a JSON request body into dst.
//
// It requires Content-Type: application/json. Browsers can't send that
// cross-site without a CORS preflight (which we never allow), so this also
// stops plain HTML forms on other sites from posting to the API.
func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct != "application/json" {
		return errBadJSON
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errBadJSON
	}
	return nil
}
