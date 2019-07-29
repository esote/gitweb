package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/esote/graceful"
)

func headers(w http.ResponseWriter) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "deny")
	w.Header().Set("X-XSS-Protection", "1")
}

func css(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed)
		return
	}

	headers(w)
	w.Header().Set("Content-Security-Policy", "default-src 'none';")
	w.Header().Set("Content-Type", "text/css")

	const file = `body {
	background-color: #fff;
	color: #000;
	font-family: monospace;
	font-size: 14px;
}

td, th {
	padding: 0 0.5em;
}

th {
	text-align: left;
}

tr:hover {
	background-color: #eee;
}

.num {
	text-align: right;
}

.desc {
	color: #444;
}
`

	if _, err := w.Write([]byte(file)); err != nil {
		log.Println(err)
	}
}

func multiplex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed)
		return
	}

	headers(w)
	w.Header().Set("Content-Security-Policy", "default-src 'none';"+
		"style-src 'self';")

	paths := strings.Split(r.URL.Path[1:], "/")

	if len(paths) < 1 || paths[0] == "" {
		httpIndex(w)
		return
	}

	repo, ok := repos[paths[0]]

	if !ok {
		writeError(w, http.StatusNotFound)
		return
	}

	l := len(paths)

	switch {
	case l == 1:
		httpLog(w, r, repo)
	case l == 2 && paths[1] == "files":
		httpLs(w, r, repo)
	case l >= 3 && paths[1] == "file":
		httpFile(w, r, repo, strings.Join(paths[2:], "/"))
	case l >= 3 && paths[1] == "commit":
		httpCommit(w, r, repo, paths[2])
	default:
		writeError(w, http.StatusNotFound)
	}
}

func main() {
	config := "config.json"

	if len(os.Args) > 1 {
		config = os.Args[1]
	}

	if err := parseConfig(config); err != nil {
		log.Fatal(err)
	}

	if err := initializeTemplates(); err != nil {
		log.Fatal(err)
	}

	// pre-parse index page
	if err := initializeIndex(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", multiplex)
	mux.HandleFunc("/style.css", css)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	graceful.Graceful(srv, func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}, os.Interrupt)
}
