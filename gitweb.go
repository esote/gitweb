package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/esote/cache"
	"github.com/esote/gitweb/internal/git"
	"github.com/esote/gitweb/internal/openbsd"
	"github.com/esote/graceful"
)

type config struct {
	Chroot   string `json:"chroot"`
	HTTPS    bool   `json:"https"`
	HTTPSCrt string `json:"https_crt"`
	HTTPSKey string `json:"https_key"`
	Port     string `json:"port"`
	Repos    []struct {
		Bare          bool     `json:"bare"`
		CacheDuration string   `json:"cache_duration"`
		Description   []string `json:"description"`
		Path          string   `json:"path"`
		Ref           string   `json:"ref"`
		Timeout       string   `json:"timeout"`
	} `json:"repos"`

	OpenBSD        bool        `json:"openbsd"`
	OpenBSDUnveils [][2]string `json:"openbsd_unveils"`
}

type page struct {
	Repo      *repository
	Title     string
	Integrity string
}

type repository struct {
	Bare        bool
	Description []string
	Git         *git.Git
	Name        string

	cache cache.Cache
	d     time.Duration
	mu    sync.Mutex
}

var index []byte
var repos map[string]*repository
var templates map[string]*template.Template

func main() {
	file := "config.json"

	if len(os.Args) > 1 {
		file = os.Args[1]
	}

	var conf *config
	var err error

	if conf, err = parseConfig(file); err != nil {
		log.Fatal(err)
	}

	if err := initializeRepos(conf); err != nil {
		log.Fatal(err)
	}

	if err := initializeTmpls(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", multiplex)
	mux.HandleFunc("/style.css", cssHandler)

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.X25519,
		},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	srv := &http.Server{
		Addr:    conf.Port,
		Handler: mux,

		// will only be used if conf.HTTPS
		TLSConfig:    cfg,
		TLSNextProto: nil,
	}

	listen := func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}

	if conf.HTTPS {
		listen = func() {
			err := srv.ListenAndServeTLS(conf.HTTPSCrt, conf.HTTPSKey)
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}

	graceful.Graceful(srv, listen, os.Interrupt)
}

func headers(w http.ResponseWriter) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "deny")
	w.Header().Set("X-XSS-Protection", "1")
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	headers(w)
	w.Header().Set("Content-Security-Policy", "default-src 'none';")
	w.Header().Set("Content-Type", "text/css")

	if _, err := w.Write([]byte(css)); err != nil {
		log.Println(err)
	}
}

func multiplex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed)
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
		httpError(w, http.StatusNotFound)
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
		httpError(w, http.StatusNotFound)
	}
}

func parseConfig(path string) (*config, error) {
	b, err := ioutil.ReadFile(filepath.Clean(path))

	if err != nil {
		return nil, err
	}

	var conf config

	if err = json.Unmarshal(b, &conf); err != nil {
		return nil, err
	}

	if conf.Port == "" {
		conf.Port = ":8080"
	}

	if conf.HTTPS && (conf.HTTPSCrt == "" || conf.HTTPSKey == "") {
		return nil, errors.New("missing HTTPS crt or key")
	}

	if conf.OpenBSD {
		u := conf.OpenBSDUnveils

		if conf.HTTPS {
			u = append(u, [2]string{conf.HTTPSCrt, "r"},
				[2]string{conf.HTTPSKey, "r"})
		}

		for _, r := range conf.Repos {
			u = append(u, [2]string{r.Path, "r"})
		}

		if err := openbsd.Secure(u); err != nil {
			return nil, err
		}
	}

	if conf.Chroot != "" {
		if err := syscall.Chroot(conf.Chroot); err != nil {
			return nil, err
		}
	}

	return &conf, nil
}

func initializeTmpls() (err error) {
	var tmpls = []struct {
		name, format string
	}{
		{"commit", commitTmpl},
		{"log", logTmpl},
		{"ls", lsTmpl},
		{"show", showTmpl},
	}

	templates = make(map[string]*template.Template, len(tmpls))

	for _, tmpl := range tmpls {
		templates[tmpl.name], err = template.New(tmpl.name).Parse(tmpl.format)

		if err != nil {
			return
		}

		templates[tmpl.name], err = templates[tmpl.name].Parse(layoutTmpl)

		if err != nil {
			return
		}
	}

	tmpl, err := template.New("repos").Parse(reposTmpl)
	if err != nil {
		return
	}

	tmpl, err = tmpl.Parse(layoutTmpl)
	if err != nil {
		return
	}

	var page = struct {
		page
		Repos map[string]*repository
	}{
		page: page{
			Repo:      nil,
			Title:     "Repositories",
			Integrity: integrity,
		},
		Repos: repos,
	}

	var b bytes.Buffer
	if err = tmpl.Execute(&b, page); err != nil {
		return
	}
	index = b.Bytes()
	return
}
func initializeRepos(conf *config) error {
	repos = make(map[string]*repository, len(conf.Repos))

	var err error
	const (
		defaultTimeout       = 2 * time.Second
		defaultCacheDuration = time.Hour
	)

	for _, c := range conf.Repos {
		var timeout = defaultTimeout

		if c.Timeout != "" {
			timeout, err = time.ParseDuration(c.Timeout)
			if err != nil {
				return err
			}
		}

		r := repository{
			Bare:        c.Bare,
			Description: c.Description,
			Git:         git.NewGit(c.Path, c.Ref, timeout),
			Name:        filepath.Base(c.Path),
		}

		if r.Bare {
			r.Name = strings.TrimSuffix(r.Name, ".git")
		}

		if c.CacheDuration == "" {
			r.d = defaultCacheDuration
		} else {
			r.d, err = time.ParseDuration(c.CacheDuration)
			if err != nil {
				return err
			}
		}
		if r.d != 0 {
			r.cache = cache.NewLRU(cacheCount)
		}

		repos[r.Name] = &r
	}

	return nil
}
