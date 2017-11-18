package gitprompt

import (
	"errors"
	"strings"
)

// Commit contains information about a commit if we're not no a branch.
type Commit struct {
	ShortID string
	Tags    []string
}

const _tag = "tag:"

// ReadCommit reads the output of `git log -1 --format=%h%d`.
func ReadCommit(l string) (Commit, error) {
	// We get something like,
	//   abcdef1 (HEAD, tag: foo, tag: bar)

	var c Commit
	if i := strings.IndexByte(l, ' '); i >= 0 {
		c.ShortID = l[:i]
		l = l[i+1:]
	} else {
		return c, errors.New("failed to read commit ID")
	}

	// (HEAD, tag: foo, tag: bar)
	l = l[1 : len(l)-1] // drop ( )
	for _, n := range strings.Split(l, ", ") {
		if i := strings.Index(n, _tag); i >= 0 {
			i += len(_tag) + 1
			c.Tags = append(c.Tags, n[i:])
		}
	}

	return c, nil
}
