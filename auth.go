package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/markbates/goth/gothic"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("name")
	if err == http.ErrNoCookie {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		// some other error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// success - call the next handler
	h.next.ServeHTTP(w, r)
}

// MustAuth wraps a Handler and return an auth handler
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// loginHandler handles the third-party login process.
// format: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	if len(segs) < 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	action := segs[2]
	provider := segs[3]
	providerString := fmt.Sprintf("provider=%s", provider)
	switch action {
	case "login":
		// TODO: Better way to attach provider to gothic
		if r.URL.RawQuery == "" {
			r.URL.RawQuery = providerString
		} else {
			r.URL.RawQuery += "&" + providerString
		}
		gothic.BeginAuthHandler(w, r)
	case "callback":
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			fmt.Fprintln(w, r)
			return
		}
		userName := base64.StdEncoding.EncodeToString([]byte(user.Name))
		http.SetCookie(
			w,
			&http.Cookie{
				Name:  "name",
				Value: userName,
				Path:  "/",
			})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
