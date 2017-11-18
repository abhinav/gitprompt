package gitprompt

import (
	"fmt"
	"strconv"
	"strings"
)

// Branch contains information about the branch that we are on.
type Branch struct {
	Name   string // empty if not on a branch
	Remote string // empty if no remote
	Ahead  int64
	Behind int64
}

const (
	_initialCommit = "Initial commit on"
	_noCommitsYet  = "No commits yet on"
	_noBranch      = "no branch"
	_ahead         = "ahead"
	_behind        = "behind"
)

// ReadBranch reads branch information from `git status --porcelain --branch`.
func ReadBranch(l string) (Branch, error) {
	// We get things like,
	//
	//  ## master
	//  ## master...origin/master
	//  ## master...origin/master [ahead 3]
	//  ## master...origin/master [behind 4]
	//  ## master...origin/master [ahead 3, behind 4]

	if l[:3] == "## " {
		l = l[3:]
	}

	var br Branch
	switch {
	case strings.Contains(l, _initialCommit) || strings.Contains(l, _noCommitsYet):
		if i := strings.LastIndexByte(l, ' '); i >= 0 {
			br.Name = l[i+1:]
			return br, nil
		}
		return br, fmt.Errorf("could not get branch name from %q", l)
	case strings.Contains(l, _noBranch):
		// Not on a branch.
		return br, nil
	}

	// master...origin/master
	if i := strings.Index(l, "..."); i >= 0 {
		br.Name = l[:i]
		l = l[i+3:]
	} else {
		br.Name = l
		return br, nil
	}

	// origin/master [behind 4]
	if i := strings.IndexByte(l, ' '); i >= 0 {
		br.Remote = l[:i]
		l = l[i+1:]
	} else {
		br.Remote = l
		return br, nil
	}

	// [ahead 3, behind 4]
	l = l[1 : len(l)-1] // drop [ ]
	for _, d := range strings.SplitN(l, ", ", 2) {
		if i := strings.Index(d, _ahead); i >= 0 {
			i += len(_ahead) + 1
			var err error
			br.Ahead, err = strconv.ParseInt(d[i:], 10, 32)
			if err != nil {
				return br, fmt.Errorf("failed to parse ahead divergence from %q: %v", d, err)
			}
		} else if i := strings.Index(d, _behind); i >= 0 {
			i += len(_behind) + 1
			var err error
			br.Behind, err = strconv.ParseInt(d[i:], 10, 32)
			if err != nil {
				return br, fmt.Errorf("failed to parse behind divergence from %q: %v", d, err)
			}
		}
	}

	return br, nil
}
