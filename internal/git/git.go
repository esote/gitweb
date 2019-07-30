package git

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// Git is a wrapper for restricted git commands useful to gitweb.
type Git struct {
	path    string
	ref     string
	timeout time.Duration
}

// NewGit creates and initializes a new Git.
func NewGit(path, ref string, timeout time.Duration) *Git {
	return &Git{
		path:    path,
		ref:     ref,
		timeout: timeout,
	}
}

// Ref retrieves the current repository reference.
func (g *Git) Ref() string {
	return g.ref
}

// Utility: check if file is "binary" or printable as plain-text
func (g *Git) binary(file string) bool {
	out, err := g.run("grep", "-I", "--name-only", "-e", ".", "--", file)

	return err != nil || len(out) == 0
}

// Utility: check if file exists according to git
func (g *Git) exists(file string) bool {
	out, err := g.run("cat-file", "-e", g.ref+":"+file)

	return err == nil && len(out) == 0
}

// Utility: check if commit has parents
func (g *Git) hasParents(commit string) bool {
	out, err := g.run("rev-list", "--parents", "-n", "1", commit)

	return err == nil && bytes.Index(out, []byte{' '}) != -1
}

// Utility: run command with timeout
func (g *Git) run(arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	arg = append([]string{"-P", "-C", g.path}, arg...)

	cmd := exec.CommandContext(ctx, "git", arg...)
	cmd.Env = []string{"COLUMNS=80"}

	return cmd.Output()
}
