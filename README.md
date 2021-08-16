[![Build Status](https://travis-ci.org/abhinav/gitprompt.svg?branch=master)](https://travis-ci.org/abhinav/gitprompt)

# Introduction

`gitprompt` provides a prompt component for Zsh and Bash which contains
information about a git repository.

The behavior of `gitprompt` is inspired by
[olivierverdier/zsh-git-prompt] and the [oh-my-zsh plugin] for it.

Example outputs:

- `(master|✔)`: branch `master` with a clean working tree, synced with upstream
- `(feature|…1)`: branch `feature` with one untracked file
- `(feature|✚1…1)`: branch `feature` with one unstaged changed file and one
  untracked file
- `(master|●1)`: branch `master` with one file staged to commit
- `(main↓1|✔)`: branch `main`, one commit behind upstream
- `(main↓2↑1|✔)`: branch `main`, two commits behind upstream and one commit ahead
- `(tags/v0.6.0|✔)`: detached HEAD on a tag
- `(:045f930|✔)`: detached HEAD, not on any branch or tag

# Installation

## Binaries


Pre-built ARM and 64-bit binaries are available for Linux and Mac at
<https://github.com/abhinav/gitprompt/releases>. To install, simply unpack the
archive and put the binaries somewhere on your `$PATH`.

For example, if you have `$HOME/bin` on your `$PATH`,

    OS=$(uname -s)
    ARCH=$(uname -m)
    VERSION=0.7.0
    URL="https://github.com/abhinav/gitprompt/releases/download/v$VERSION/gitprompt_${VERSION}_${OS}_${ARCH}.tar.gz"
    curl -L "$URL" | tar xzv -C ~/bin gitprompt

## Build From Source

If you have Go installed, you can install `gitprompt` from source using the
following command.

    $ go install github.com/abhinav/gitprompt/cmd/gitprompt@latest

# Usage

In Zsh, add `$(gitprompt zsh)` to your prompt. For example, add the following
to your `~/.zshrc`.

```sh
PROMPT='%M:%~ $(gitprompt zsh) $ '
```

In Bash, add `$(gitprompt bash)` to your prompt. For example, add the
following to your `~/.bashrc`.

```sh
PS1='\h:\W $(gitprompt bash) $ '
```

Note that in both cases, the `$(gitprompt ...)` command must be escaped so
that it is not run during prompt declaration. To do this, declare the prompt
with single quotes as done above, or escape the `$` if you're using double
quotes:

```sh
PROMPT="%M:%~ \$(gitprompt zsh) $ "
```

## Large repositories

If you're working inside a large repository where running `git status` takes
too long, use the `-no-git-status` flag with `gitprompt`.

```sh
PROMPT='%M:%~ $(gitprompt -no-git-status zsh) $ '
```

With this flag enabled, only minimal information about the current branch,
tag, or commit hash will be displayed. This is a relatively cheap and fast
operation.

This flag defaults to true if the `GITPROMPT_NO_GIT_STATUS` environment
variable is set. Mixing the environment variable with [direnv], you can
configure this flag on a per-directory basis if necessary.

  [direnv]: https://direnv.net/

# Credits

The functionality provided by `gitprompt` is inspired by
[olivierverdier/zsh-git-prompt] and the [oh-my-zsh plugin] for it.

  [olivierverdier/zsh-git-prompt]: https://github.com/olivierverdier/zsh-git-prompt
  [oh-my-zsh plugin]: https://github.com/robbyrussell/oh-my-zsh/tree/master/plugins/git-prompt
