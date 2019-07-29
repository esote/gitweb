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

	var commit = &Commit{}

	out, err := g.run("git", "-P", "-C", g.path, "cat-file", "-p", hash)

	if err != nil {
		return nil, err
	}

	commit.CatFile = out

	// empty tree commit hash
	with := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

	if g.hasParents(hash) {
		with = hash + "~"
	}

	out, err = g.run("git", "-P", "-C", g.path, "diff", "--stat", with,
		hash)

	if err != nil {
		return nil, err
	}

	commit.DiffStat = out

	out, err = g.run("git", "-P", "-C", g.path, "diff", with, hash)

	if err != nil {
		return nil, err
	}

	commit.Diff = out

	return commit, nil
}
