package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/abhinav/gitprompt"
)

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

type color []byte

var (
	_branchColor   = color("\x1b[35;1m") // magenta; bold
	_stagedColor   = color("\x1b[31m")   // red
	_conflictColor = _stagedColor        // red
	_changedColor  = color("\x1b[34m")   // blue
	_cleanColor    = color("\x1b[32m")   // green

	_reset = color("\x1b[0m")
)

type shell struct {
	// Byte sequences to write before and after each zero-width sequence
	// of characters in the prompt.
	ZWOpen, ZWClose []byte
}

var (
	_zsh  = shell{ZWOpen: []byte("%{"), ZWClose: []byte("%}")}
	_bash = shell{ZWOpen: []byte{0x01}, ZWClose: []byte{0x02}}
)

type promptBuffer struct {
	bytes.Buffer

	Shell shell

	// Used for AppendInt.
	intbuff [4]byte
}

func (w *promptBuffer) Itoa(i int64) {
	w.Write(strconv.AppendInt(w.intbuff[:0], i, 10))
}

func (w *promptBuffer) Color(c color) {
	w.Write(w.Shell.ZWOpen)
	w.Write(c)
	w.Write(w.Shell.ZWClose)
}

func (w *promptBuffer) Reset() {
	w.Color(_reset)
}

const (
	_check    = '✔'
	_cross    = '✖'
	_dot      = '●'
	_dots     = '…'
	_down     = '↓'
	_parClose = ')'
	_parOpen  = '('
	_pipe     = '|'
	_plus     = '✚'
	_up       = '↑'
)

func run(args []string, stdout io.Writer) error {
	var (
		timeout  time.Duration
		noStatus bool
	)

	flag := flag.NewFlagSet("gitprompt", flag.ContinueOnError)
	flag.Usage = func() {
		fmt.Fprintf(flag.Output(), "usage: %v shell\n", flag.Name())
		flag.PrintDefaults()
	}

	flag.DurationVar(&timeout, "timeout", 0,
		"amount of time the 'git status' command is allowed to take; unlimited if 0")

	flag.BoolVar(&noStatus, "no-git-status", os.Getenv("GITPROMPT_NO_GIT_STATUS") == "1",
		"display only branch, tags or hash without calling git status (default: $GITPROMPT_NO_GIT_STATUS)")

	if err := flag.Parse(args); err != nil {
		return err
	}

	if flag.NArg() == 0 {
		flag.Usage()
		return nil
	}

	var shell shell
	switch flag.Arg(0) {
	case "bash":
		shell = _bash
	case "zsh":
		shell = _zsh
	default:
		return fmt.Errorf("unsupported shell %q", flag.Arg(0))
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
		s, err = gitBranchName(ctx)
	}
	if err != nil {
		return err
	}
	if s == nil {
		return nil // not a git repo
	}

	w := promptBuffer{Shell: shell}
	buildPrompt(&w, s)
	w.WriteTo(stdout)
	return nil
}

func buildPrompt(w *promptBuffer, s *Status) {
	w.WriteRune(_parOpen)
	defer w.WriteRune(_parClose)

	w.Color(_branchColor)
	w.WriteString(s.Branch)
	w.Reset()

	if s.BranchOnly {
		return
	}

	if s.Behind > 0 {
		w.WriteRune(_down)
		w.Itoa(s.Behind)
	}

	if s.Ahead > 0 {
		w.WriteRune(_up)
		w.Itoa(s.Ahead)
	}

	w.WriteRune(_pipe)

	if s.Staged > 0 {
		w.Color(_stagedColor)
		w.WriteRune(_dot)
		w.Itoa(s.Staged)
		w.Reset()
	}

	if s.Conflicts > 0 {
		w.Color(_conflictColor)
		w.WriteRune(_cross)
		w.Itoa(s.Conflicts)
		w.Reset()
	}

	if s.Changed > 0 {
		w.Color(_changedColor)
		w.WriteRune(_plus)
		w.Itoa(s.Changed)
		w.Reset()
	}

	if s.Untracked > 0 {
		w.WriteRune(_dots)
		w.Itoa(s.Untracked)
	}

	switch {
	case s.Changed > 0, s.Conflicts > 0, s.Staged > 0, s.Untracked > 0:
	default:
		w.Color(_cleanColor)
		w.WriteRune(_check)
		w.Reset()
	}
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

func gitBranchName(ctx context.Context) (*Status, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run git rev-parse: %v", err)
	}

	if out := strings.TrimSpace(string(b)); out != "HEAD" {
		return &Status{
			Branch:     out,
			BranchOnly: true,
		}, nil
	}

	cmd = exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
	cmd.Stderr = os.Stderr
	b, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run git rev-parse: %v", err)
	}

	return &Status{
		Branch:     ":" + strings.TrimSpace(string(b)),
		BranchOnly: true,
	}, nil
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
