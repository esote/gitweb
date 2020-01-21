package main

import (
	"bytes"
	"time"

	"github.com/esote/gitweb/internal/git"
)

type timePair struct {
	b []byte
	t time.Time
}

const (
	cacheCount = 2

	keyLog int = iota
	keyLs
)

func logCached(repo *repository) ([]byte, error) {
	if repo.cache != nil {
		repo.mu.Lock()
		defer repo.mu.Unlock()

		v, hit := repo.cache.Get(keyLog)
		if hit && time.Now().UTC().Sub(v.(timePair).t) < repo.d {
			return v.(timePair).b, nil
		}
		repo.cache.Delete(keyLog)
	}

	ret, err := repo.Git.Log()
	if err != nil {
		return nil, err
	}

	var page = struct {
		page
		Items []*git.LogItem
	}{
		page: page{
			Repo:      repo,
			Title:     repo.Name + " - Log",
			Integrity: integrity,
		},
		Items: ret,
	}

	var b bytes.Buffer
	if err = templates["log"].Execute(&b, page); err != nil {
		return nil, err
	}

	if repo.cache != nil {
		repo.cache.Add(keyLog, timePair{
			b: b.Bytes(),
			t: time.Now().UTC(),
		})
	}

	return b.Bytes(), nil
}

func lsCached(repo *repository) ([]byte, error) {
	if repo.cache != nil {
		repo.mu.Lock()
		defer repo.mu.Unlock()

		v, hit := repo.cache.Get(keyLs)
		if hit && time.Now().UTC().Sub(v.(timePair).t) < repo.d {
			return v.(timePair).b, nil
		}
		repo.cache.Delete(keyLs)
	}

	ret, err := repo.Git.Ls()

	if err != nil {
		return nil, err
	}

	var page = struct {
		page
		Items []*git.LsItem
	}{
		page: page{
			Repo:      repo,
			Title:     repo.Name + " - Files",
			Integrity: integrity,
		},
		Items: ret,
	}

	var b bytes.Buffer
	if err = templates["ls"].Execute(&b, page); err != nil {
		return nil, err
	}

	if repo.cache != nil {
		repo.cache.Add(keyLs, timePair{
			b: b.Bytes(),
			t: time.Now().UTC(),
		})
	}

	return b.Bytes(), nil
}
