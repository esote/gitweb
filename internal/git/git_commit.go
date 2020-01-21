package git

import (
	"errors"
	"regexp"
)

// Commit contains details about a commit.
type Commit struct {
	CatFile  []byte
	DiffStat []byte
	Diff     []byte
}

// ErrInvalidHash is used in gitweb to determine if the request error was from a
// bad request or happened running git.
var ErrInvalidHash = errors.New("git: commit: not a hash")

var reNotHash = regexp.MustCompile("[^0-9A-Za-z]")

// Commit retrieves details about a commit.
func (g *Git) Commit(hash string) (*Commit, error) {
	if len(hash) != 40 || reNotHash.MatchString(hash) {
		return nil, ErrInvalidHash
	}

	errs := make(chan error, 3)
	defer close(errs)

	commit := &Commit{}

	go func() {
		var err error
		commit.CatFile, err = g.run("cat-file", "-p", hash)
		errs <- err
	}()

	// empty tree commit hash
	with := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

	if g.hasParents(hash) {
		with = hash + "~"
	}

	go func() {
		var err error
		commit.DiffStat, err = g.run("diff", "--stat", with, hash)
		errs <- err
	}()

	go func() {
		var err error
		commit.Diff, err = g.run("diff", with, hash)
		errs <- err
	}()

	var err error
	for i := 0; i < cap(errs); i++ {
		if err2 := <-errs; err == nil {
			err = err2
		}
	}

	return commit, err
}
