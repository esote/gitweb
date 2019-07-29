package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/esote/gitweb/internal/git"
	"github.com/esote/util/tcache"
)

func writeError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func writeCached(w http.ResponseWriter, cache *tcache.TCache) {
	ret := cache.Next()

	switch ret.(type) {
	case bytes.Buffer:
		b := ret.(bytes.Buffer)

		if _, err := b.WriteTo(w); err != nil {
			log.Println(err)
		}

		return
	case error:
		log.Println(ret)
	}

	writeError(w, http.StatusInternalServerError)
}

func httpLog(w http.ResponseWriter, r *http.Request, repo *repository) {
	writeCached(w, repo.caches["log"])
}

func httpLs(w http.ResponseWriter, r *http.Request, repo *repository) {
	writeCached(w, repo.caches["ls"])
}

func httpCommit(w http.ResponseWriter, r *http.Request, repo *repository, hash string) {
	out, err := repo.Git.Commit(hash)

	if err != nil {
		if err == git.ErrInvalidHash {
			writeError(w, http.StatusBadRequest)
		} else {
			writeError(w, http.StatusInternalServerError)
		}
		return
	}

	var page = struct {
		page
		Commit *git.Commit
	}{
		page: page{
			Repo:  repo,
			Title: repo.Name + " - Commit " + hash,
		},
		Commit: out,
	}

	if err = templates["commit"].Execute(w, page); err != nil {
		log.Println(err)
	}
}

func httpFile(w http.ResponseWriter, r *http.Request, repo *repository, file string) {
	if repo.Bare {
		writeError(w, http.StatusBadRequest)
		return
	}

	out, err := repo.Git.Show(file)

	if err != nil {
		if err == git.ErrNotExist {
			writeError(w, http.StatusBadRequest)
		} else {
			writeError(w, http.StatusInternalServerError)
		}

		return
	}

	var page = struct {
		page
		git.Show
	}{
		page: page{
			Repo:  repo,
			Title: repo.Name + " - File " + file,
		},
		Show: out,
	}

	if err = templates["show"].Execute(w, page); err != nil {
		log.Println(err)
	}
}

func httpIndex(w http.ResponseWriter) {
	if _, err := w.Write(index); err != nil {
		log.Println(err)
	}
}
