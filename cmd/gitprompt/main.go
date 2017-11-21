package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/abhinav/gitprompt"
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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdout pipe to git: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start git %v: %v", cmd.Args, err)
	}

	var s Status
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := s.feed(scanner.Text()); err != nil {
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

// Status holds the information about the worktree necessary to build the
// prompt.
type Status struct {
	Branch    string // empty if not on a branch
	Ahead     int64
	Behind    int64
	Staged    int64
	Conflicts int64
	Changed   int64
	Untracked int64
}

// Feeds a line of `git status --porcelain --branch` to Status.
func (s *Status) feed(line string) error {
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

		s.Branch, err = getTagNameOrHash()
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

func getTagNameOrHash() (string, error) {
	cmd := exec.Command("git", "log", "-1", `--format=%h%d`)
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
