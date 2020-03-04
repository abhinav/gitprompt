package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/abhinav/gitprompt"
	"github.com/fatih/color"
)

func main() {
	color.NoColor = false
	log.SetFlags(0)
	if err := run(os.Args[1:]); err != nil {
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

func run(args []string) error {
	var (
		timeout  time.Duration
		noStatus bool
	)

	flag := flag.NewFlagSet("gitprompt", flag.ContinueOnError)
	flag.DurationVar(&timeout, "timeout", 0,
		"amount of time the 'git status' command is allowed to take; unlimited if 0")

	flag.BoolVar(&noStatus, "no-git-status", os.Getenv("GITPROMPT_NO_GIT_STATUS") == "1",
		"display only branch, tags or hash without calling git status (default: $GITPROMPT_NO_GIT_STATUS)")

	if err := flag.Parse(args); err != nil {
		return err
	}

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var (
		s   *Status
		err error
	)
	if !noStatus {
		s, err = gitStatus(ctx)
	} else {
		s, err = gitLog(ctx)
	}
	if err != nil {
		return err
	}
	if s == nil {
		return nil // not a git repo
	}

	fmt.Printf("(%v", branchColor.Sprint(s.Branch))
	defer fmt.Print(")")

	if s.BranchOnly {
		return nil
	}

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

	return nil
}

func gitStatus(ctx context.Context) (*Status, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain", "--branch")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("open stdout pipe to git: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start git %v: %v", cmd.Args, err)
	}

	var s Status
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := s.feed(ctx, scanner.Text()); err != nil {
			return nil, fmt.Errorf("parse output of git status: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read output: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if err != context.DeadlineExceeded {
			err = nil // not a git repo
		}
		return nil, err
	}

	return &s, nil
}

func gitLog(ctx context.Context) (*Status, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "-1", `--format=%h%d`)
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run git log: %v", err)
	}

	out := strings.TrimSpace(string(b))
	c, err := gitprompt.ReadCommit(out)
	if err != nil {
		return nil, fmt.Errorf("read commit information from %q: %v", out, err)
	}

	s := Status{BranchOnly: true}
	switch {
	case len(c.Branch) > 0:
		s.Branch = c.Branch
	case len(c.Tags) > 0:
		s.Branch = "tags/" + c.Tags[0]
	default:
		s.Branch = c.ShortID
	}

	return &s, nil
}

// Status holds the information about the worktree necessary to build the
// prompt.
type Status struct {
	Branch     string // empty if not on a branch
	BranchOnly bool

	Ahead     int64
	Behind    int64
	Staged    int64
	Conflicts int64
	Changed   int64
	Untracked int64
}

// Feeds a line of `git status --porcelain --branch` to Status.
func (s *Status) feed(ctx context.Context, line string) error {
	if line[:2] == "##" {
		branch, err := gitprompt.ReadBranch(line)
		if err != nil {
			return fmt.Errorf("cannot read branch information from %q: %v", line, err)
		}

		if branch.Name != "" {
			s.Branch = branch.Name
			s.Ahead = branch.Ahead
			s.Behind = branch.Behind
			return nil
		}

		s.Branch, err = getTagNameOrHash(ctx)
		if err != nil {
			return err
		}

		return nil
	}

	f, err := gitprompt.ReadFileStatus(line)
	if err != nil {
		return fmt.Errorf("cannot read file status from %q: %v", line, err)
	}

	if f.Untracked() {
		s.Untracked++
		return nil
	}

	// Index status
	switch f.Index {
	case gitprompt.UpdatedButUnmerged:
		s.Conflicts++
	case gitprompt.Unmodified:
		// Ignore files unchanged in index.
	default:
		// Everything else has been staged.
		s.Staged++
	}

	// Files modified in working tree.
	if f.WorkTree == gitprompt.Modified {
		s.Changed++
	}

	return nil
}

func getTagNameOrHash(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "-1", `--format=%h%d`)
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run git log: %v", err)
	}
	out := strings.TrimSpace(string(b))
	c, err := gitprompt.ReadCommit(out)
	if err != nil {
		return "", fmt.Errorf("failed to read commit information from %q: %v", out, err)
	}

	if len(c.Tags) > 0 {
		return "tags/" + c.Tags[0], nil
	}
	return ":" + c.ShortID, nil
}
