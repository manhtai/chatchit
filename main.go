package main

import (
	"encoding/base64"
	"flag"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates",
			t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("name"); err == nil {
		name, _ := base64.StdEncoding.DecodeString(authCookie.Value)
		data["Name"] = string(name)
	}
	t.templ.Execute(w, data)
}

func init() {
	store := sessions.NewFilesystemStore(os.TempDir(), []byte("chatchit"))
	store.MaxLength(math.MaxInt64)
	gothic.Store = store
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the  application.")
	flag.Parse()

	goth.UseProviders(
		// Google+ provider
		gplus.New(
			os.Getenv("GPLUS_KEY"),
			os.Getenv("GPLUS_SECRET"),
			"http://localhost:8080/auth/callback/gplus",
		),
	)

	r := newRoom()
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	// get the room going
	go r.run()
	// start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
