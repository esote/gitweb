package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"syscall"

	"github.com/esote/gitweb/internal/openbsd"
)

type config struct {
	Chroot   string `json:"chroot"`
	HTTPS    bool   `json:"https"`
	HTTPSCrt string `json:"https_crt"`
	HTTPSKey string `json:"https_key"`
	Port     string `json:"port"`
	Repos    []struct {
		Bare          bool   `json:"bare"`
		CacheDuration string `json:"cache_duration"`
		Description   string `json:"description"`
		Path          string `json:"path"`
		Ref           string `json:"ref"`
		Timeout       string `json:"timeout"`
	} `json:"repos"`

	OpenBSD        bool        `json:"openbsd"`
	OpenBSDUnveils [][2]string `json:"openbsd_unveils"`
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
