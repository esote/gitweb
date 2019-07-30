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
func (g *Git) Commit(hash string) (commit *Commit, err error) {
	if len(hash) != 40 || reNotHash.MatchString(hash) {
		return nil, ErrInvalidHash
	}

	commit = &Commit{}

	commit.CatFile, err = g.run("cat-file", "-p", hash)

	if err != nil {
		return nil, err
	}

	// empty tree commit hash
	with := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

	if g.hasParents(hash) {
		with = hash + "~"
	}

	commit.DiffStat, err = g.run("diff", "--stat", with, hash)

	if err != nil {
		return nil, err
	}

	commit.Diff, err = g.run("diff", with, hash)

	if err != nil {
		return nil, err
	}

	return commit, nil
}
