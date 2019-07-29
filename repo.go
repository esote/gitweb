package main

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"time"

	"github.com/esote/gitweb/internal/git"
	"github.com/esote/util/tcache"
)

type repository struct {
	Bare        bool
	Description string
	Git         *git.Git
	Name        string

	caches map[string]*tcache.TCache
}

var repos map[string]*repository

func initializeRepos(conf config) error {
	repos = make(map[string]*repository, len(conf.Repos))

	var err error

	const defaultTimeout = 2 * time.Second

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

		r.caches, err = initializeCaches(r.Name, c.CacheDuration)

		if err != nil {
			return err
		}

		repos[r.Name] = &r
	}

	return nil
}

func initializeCaches(name, dur string) (map[string]*tcache.TCache, error) {
	d, err := time.ParseDuration(dur)

	if err != nil {
		return nil, err
	}

	caches := make(map[string]*tcache.TCache, 2)

	caches["log"], err = tcache.NewTCache(d, func() interface{} {
		ret, err := repos[name].Git.Log()

		if err != nil {
			return err
		}

		var page = struct {
			page
			Items []git.LogItem
		}{
			page: page{
				Repo:  repos[name],
				Title: name + " - Log",
			},
			Items: ret,
		}

		var b bytes.Buffer

		w := bufio.NewWriter(&b)

		if err = templates["log"].Execute(w, page); err != nil {
			return err
		}

		if err = w.Flush(); err != nil {
			return err
		}

		return b
	})

	if err != nil {
		return nil, err
	}

	caches["ls"], err = tcache.NewTCache(d, func() interface{} {
		ret, err := repos[name].Git.Ls()

		if err != nil {
			return err
		}

		var page = struct {
			page
			Items []git.LsItem
		}{
			page: page{
				Repo:  repos[name],
				Title: name + " - Files",
			},
			Items: ret,
		}

		var b bytes.Buffer

		w := bufio.NewWriter(&b)

		if err = templates["ls"].Execute(w, page); err != nil {
			return err
		}

		if err = w.Flush(); err != nil {
			return err
		}

		return b
	})

	if err != nil {
		return nil, err
	}

	return caches, nil
}
