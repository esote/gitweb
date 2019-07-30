package git

import (
	"bytes"
	"errors"
	"os"
	"strconv"
)

// Tree object types
const (
	LsBlob = iota
	LsTree
)

// LsItem is the parsed columnar output of git ls-tree.
type LsItem struct {
	Mode os.FileMode
	Type int
	Hash string
	Size int64
	Name string
}

// Ls retrieves the list of tracked files.
func (g *Git) Ls() ([]LsItem, error) {
	out, err := g.run("ls-tree", "-lr", g.ref)

	if err != nil {
		return nil, err
	}

	b := bytes.Split(out[:len(out)-1], []byte{'\n'})

	ret := make([]LsItem, len(b))

	for i := range b {
		item, err := parseLsItem(b[i])

		if err != nil {
			return nil, err
		}

		ret[i] = *item
	}

	return ret, nil
}

func parseLsItem(raw []byte) (item *LsItem, err error) {
	// ls-tree with -l has 3 space sep and 1 tab sep

	// first[0] = mode
	// first[1] = type
	// first[2] = hash
	// first[3] = size and name
	first := bytes.SplitN(raw, []byte{' '}, 4)

	if len(first) != 4 {
		return nil, errors.New("git: ls: first split failed")
	}

	// second[0] = size
	// second[1] = name
	second := bytes.SplitN(first[3], []byte{'\t'}, 2)

	if len(second) != 2 {
		return nil, errors.New("git: ls: second split failed")
	}

	item = &LsItem{}

	mode, err := strconv.ParseUint(string(first[0]), 10, 32)

	if err != nil {
		return nil, err
	}

	item.Mode = os.FileMode(mode)

	switch string(first[1]) {
	case "blob":
		item.Type = LsBlob
	case "tree":
		item.Type = LsTree
	default:
		return nil, errors.New("git: ls: unknown object type")
	}

	item.Hash = string(first[2])

	second[0] = bytes.TrimLeft(second[0], " ")

	if second[0][0] == '-' {
		item.Size = 0
	} else {
		item.Size, err = strconv.ParseInt(string(second[0]), 10, 64)

		if err != nil {
			return nil, err
		}
	}

	item.Name = string(second[1])

	return
}
