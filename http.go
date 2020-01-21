package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"github.com/esote/gitweb/internal/git"
)

func httpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func httpLog(w http.ResponseWriter, r *http.Request, repo *repository) {
	b, err := logCached(repo)
	if err != nil {
		if err == context.DeadlineExceeded {
			httpError(w, http.StatusRequestTimeout)
		} else {
			httpError(w, http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

func httpLs(w http.ResponseWriter, r *http.Request, repo *repository) {
	b, err := lsCached(repo)
	if err != nil {
		if err == context.DeadlineExceeded {
			httpError(w, http.StatusRequestTimeout)
		} else {
			httpError(w, http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

func httpCommit(w http.ResponseWriter, r *http.Request, repo *repository, hash string) {
	out, err := repo.Git.Commit(hash)

	if err != nil {
		switch err {
		case git.ErrInvalidHash:
			httpError(w, http.StatusBadRequest)
		case context.DeadlineExceeded:
			httpError(w, http.StatusRequestTimeout)
		default:
			httpError(w, http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	var page = struct {
		page
		Commit *git.Commit
	}{
		page: page{
			Repo:      repo,
			Title:     repo.Name + " - Commit " + hash,
			Integrity: integrity,
		},
		Commit: out,
	}

	var b bytes.Buffer
	if err = templates["commit"].Execute(&b, page); err != nil {
		log.Println(err)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if _, err = io.Copy(w, &b); err != nil {
		log.Println(err)
	}
}

func httpFile(w http.ResponseWriter, r *http.Request, repo *repository, file string) {
	if repo.Bare {
		httpError(w, http.StatusNotFound)
		return
	}

	out, err := repo.Git.Show(file)

	if err != nil {
		switch err {
		case git.ErrNotExist:
			httpError(w, http.StatusBadRequest)
		case context.DeadlineExceeded:
			httpError(w, http.StatusRequestTimeout)
		default:
			httpError(w, http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	var page = struct {
		page
		git.Show
	}{
		page: page{
			Repo:      repo,
			Title:     repo.Name + " - File " + file,
			Integrity: integrity,
		},
		Show: out,
	}

	var b bytes.Buffer
	if err = templates["show"].Execute(&b, page); err != nil {
		log.Println(err)
		httpError(w, http.StatusInternalServerError)
		return
	}
	if _, err = io.Copy(w, &b); err != nil {
		log.Println(err)
	}
}

func httpIndex(w http.ResponseWriter) {
	if _, err := w.Write(index); err != nil {
		log.Println(err)
	}
}
