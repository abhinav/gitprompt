package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

func main() {
	color.NoColor = false
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var (
	branchColor   = color.New(color.FgMagenta, color.Bold)
	stagedColor   = color.New(color.FgRed)
	conflictColor = color.New(color.FgRed)
	changedColor  = color.New(color.FgBlue)
	cleanColor    = color.New(color.FgGreen)
)

func run() error {
	cmd := exec.Command("git", "status", "--porcelain", "--branch")
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdout pipe to git: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start git %v: %v", cmd.Args, err)
	}

	var s status
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := s.Feed(scanner.Text()); err != nil {
			return fmt.Errorf("failed to parse output of git status: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read output: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil // not a git repo
	}

	fmt.Printf("(%v", branchColor.Sprint(s.Branch))
	if s.Behind > 0 {
		fmt.Printf("↓%d", s.Behind)
	}
	if s.Ahead > 0 {
		fmt.Printf("↑%d", s.Ahead)
	}
	fmt.Printf("|")

	if s.Staged > 0 {
		stagedColor.Printf("●%d", s.Staged)
	}

	if s.Conflicts > 0 {
		conflictColor.Printf("✖%d", s.Conflicts)
	}

	if s.Changed > 0 {
		changedColor.Printf("✚%d", s.Changed)
	}

	if s.Untracked > 0 {
		fmt.Printf("…%d", s.Untracked)
	}

	switch {
	case s.Changed > 0, s.Conflicts > 0, s.Staged > 0, s.Untracked > 0:
	default:
		cleanColor.Printf("✔")
	}

	fmt.Print(")")
	return nil
}

type status struct {
	Branch    string
	Ahead     int64
	Behind    int64
	Staged    int64
	Conflicts int64
	Changed   int64
	Untracked int64
}

func (s *status) Feed(line string) error {
	xy := line[:2]

	switch line[:2] {
	case "##":
		return s.feedBranchInfo(strings.TrimSpace(line[2:]))
	case "??":
		s.Untracked++
		return nil
	}

	// Index status
	switch xy[0] {
	case 'U':
		s.Conflicts++
	case ' ':
		// Ignore files unchanged in index.
	default:
		// Everything else has been staged.
		s.Staged++
	}

	// Files modified in working tree.
	if xy[1] == 'M' {
		s.Changed++
	}

	return nil
}

const (
	_initialCommit = "Initial commit on"
	_noCommitsYet  = "No commits yet on"
	_noBranch      = "no branch"
	_ahead         = "ahead"
	_behind        = "behind"
	_tag           = "tag: "
)

func (s *status) feedBranchInfo(l string) error {
	// We get things like,
	//
	//  ## master
	//  ## master...origin/master
	//  ## master...origin/master [ahead 3]
	//  ## master...origin/master [behind 4]
	//  ## master...origin/master [ahead 3, behind 4]

	switch {
	case strings.Contains(l, _initialCommit) || strings.Contains(l, _noCommitsYet):
		if i := strings.LastIndexByte(l, ' '); i >= 0 {
			s.Branch = l[i+1:]
			return nil
		}
		return fmt.Errorf("could not get branch name from %q", l)
	case strings.Contains(l, _noBranch):
		var err error
		s.Branch, err = getTagNameOrHash()
		if err != nil {
			return err
		}
		return nil
	}

	if i := strings.Index(l, "..."); i >= 0 {
		s.Branch = l[:i]
		l = l[i:]
	} else {
		s.Branch = l
		return nil
	}

	if i := strings.IndexByte(l, ' '); i >= 0 {
		l = l[i+1:]
	} else {
		return nil
	}

	// Drop surround [, ]
	l = l[1 : len(l)-1]
	for _, d := range strings.SplitN(l, ", ", 2) {
		var err error

		if i := strings.Index(d, _ahead); i >= 0 {
			i += len(_ahead) + 1
			s.Ahead, err = strconv.ParseInt(d[i:], 10, 32)
			if err != nil {
				return fmt.Errorf("failed to parse ahead divergence: %v", err)
			}
			continue
		}

		if i := strings.Index(d, _behind); i >= 0 {
			i += len(_behind) + 1
			s.Behind, err = strconv.ParseInt(d[i:], 10, 32)
			if err != nil {
				return fmt.Errorf("failed to parse behind divergence: %v", err)
			}
		}
	}

	return nil
}

func getTagNameOrHash() (string, error) {
	cmd := exec.Command("git", "log", "-1", `--format=%h%d`)
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git log: %v", err)
	}
	out := strings.TrimSpace(string(b))

	var hash string
	if i := strings.IndexByte(out, ' '); i >= 0 {
		hash = out[:i]
	} else {
		return "", fmt.Errorf("could not read hash from %q", out)
	}

	if i := strings.Index(out, _tag); i >= 0 {
		out = out[i+len(_tag):]
	} else {
		return hash, nil
	}

	if i := strings.IndexAny(out, " ,)"); i >= 0 {
		return "tags/" + out[:i], nil
	}

	return "", fmt.Errorf("failed to parse tag name from %q", string(b))
}
