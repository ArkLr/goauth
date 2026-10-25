package handlers

import "net/http"

func Register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("register"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("login"))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("logout"))
}

func Protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("protected"))
}
