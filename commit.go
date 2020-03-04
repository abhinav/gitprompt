package gitprompt

import (
	"strings"
)

// Commit contains information about a commit if we're not no a branch.
type Commit struct {
	ShortID string
	Branch  string
	Tags    []string
}

const (
	_tag       = "tag: "
	_headPoint = "HEAD -> "
)

// ReadCommit reads the output of `git log -1 --format=%h%d`.
func ReadCommit(l string) (Commit, error) {
	// We get something like,
	//   abcdef1 (HEAD, tag: foo, tag: bar)
	//   abcdef1 (HEAD -> master, origin/master, tag: bar)

	var c Commit
	if i := strings.IndexByte(l, ' '); i >= 0 {
		c.ShortID = l[:i]
		l = l[i+1:]
	} else {
		// Just "abcdef1".
		c.ShortID = l
		return c, nil
	}

	// (HEAD, tag: foo, tag: bar)
	// (HEAD -> master, tag: foo, tag: bar)
	l = l[1 : len(l)-1] // drop ( )
	for _, n := range strings.Split(l, ", ") {
		switch {
		case strings.HasPrefix(n, _tag):
			tag := n[len(_tag):]
			c.Tags = append(c.Tags, tag)
		case strings.HasPrefix(n, _headPoint):
			branch := n[len(_headPoint):]
			c.Branch = branch
		}
	}

	return c, nil
}
