package git

import (
	"bytes"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/esote/util/pool"
)

// LogStat is the LogItem file diff stats.
type LogStat struct {
	Changed    uint64
	Insertions uint64
	Deletions  uint64
}

// LogItem is the parsed columnar output of git log.
type LogItem struct {
	Time    time.Time
	Hash    string
	Name    string
	Subject string
	Stat    LogStat
}

// Log retrieves the simple commit history.
func (g *Git) Log() ([]*LogItem, error) {
	const l = 6

	out, err := g.run("log", "--format=%aI%n%H%n%an%n%s",
		"--shortstat", g.ref)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(out[:len(out)-1], []byte{'\n'})

	if len(lines)%l != 0 {
		return nil, errors.New("git: log: output line count mismatch")
	}

	ret := make([]*LogItem, len(lines)/l)
	p := pool.New(5, 100)
	errs := make(chan error, 1)
	defer close(errs)

	f := func(args ...interface{}) {
		var err error
		ret[args[0].(int)/l], err = parseLogItem(args[1].([][]byte))
		if err != nil {
			select {
			case errs <- err:
			default:
			}
		}
	}

	for i := 0; i < len(lines); i += l {
		p.Enlist(true, f, i, lines[i:i+l])
	}

	p.Close(false)

	return ret, nil
}

var reNum = regexp.MustCompile("[0-9]+")

func parseLogItem(raw [][]byte) (item *LogItem, err error) {
	// raw[0] = time
	// raw[1] = hash
	// raw[2] = author
	// raw[3] = subject
	// raw[4] = empty line
	// raw[5] = shortstat (files changed, insertions, deletions)

	item = &LogItem{}

	const iso8601 = "2006-01-02T15:04:05-07:00"

	item.Time, err = time.Parse(iso8601, string(raw[0]))

	if err != nil {
		return nil, err
	}

	item.Hash = string(raw[1])
	item.Name = string(raw[2])
	item.Subject = string(raw[3])

	// nums[0] = file(s) changed
	// nums[1] = insertions
	// nums[2] = deletions
	nums := reNum.FindAll(raw[5], -1)

	if nums == nil {
		return nil, errors.New("git: log: no shortstat numbers found")
	}

	if len(nums) < 2 {
		return nil, errors.New("git: log: weird shortstat numbers")
	}

	i := 0

	item.Stat.Changed, err = strconv.ParseUint(string(nums[i]), 10, 64)

	if err != nil {
		return nil, err
	}

	i++

	if bytes.Contains(raw[5], []byte("insert")) {
		item.Stat.Insertions, err = strconv.ParseUint(string(nums[i]), 10, 64)

		if err != nil {
			return nil, err
		}

		i++
	}

	if bytes.Contains(raw[5], []byte("delet")) {
		item.Stat.Deletions, err = strconv.ParseUint(string(nums[i]), 10, 64)

		if err != nil {
			return nil, err
		}
	}

	return
}
