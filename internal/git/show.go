package git

import "errors"

// Show contains the contents of a file.
type Show struct {
	Binary bool
	File   []byte
}

// ErrNotExist is used in gitweb to determine if the request error was from a
// bad request or happened running git.
var ErrNotExist = errors.New("git: show: file does not exist")

// Show retrieves the contents of a tracked file or mark as binary.
func (g *Git) Show(file string) (show Show, err error) {
	if !g.exists(file) {
		err = ErrNotExist
		return
	}

	if show.Binary = g.binary(file); show.Binary {
		show.File = nil
		return
	}

	show.File, err = g.run("show", g.ref+":"+file)
	return
}
