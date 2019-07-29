package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"syscall"
)

type config struct {
	Chroot string `json:"chroot"`
	Repos  []struct {
		Bare          bool   `json:"bare"`
		CacheDuration string `json:"cache_duration"`
		Description   string `json:"description"`
		Path          string `json:"path"`
		Ref           string `json:"ref"`
		Timeout       string `json:"timeout"`
	} `json:"repos"`
}

func parseConfig(path string) error {
	b, err := ioutil.ReadFile(filepath.Clean(path))

	if err != nil {
		return err
	}

	var conf config

	if err = json.Unmarshal(b, &conf); err != nil {
		return err
	}

	if conf.Chroot != "" {
		if err := syscall.Chroot(conf.Chroot); err != nil {
			return err
		}
	}

	return initializeRepos(conf)
}
