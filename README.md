[![Build Status](https://travis-ci.org/abhinav/gitprompt.svg?branch=master)](https://travis-ci.org/abhinav/gitprompt)

Introduction
============

`gitprompt` provides a prompt component for Zsh and Bash which contains
information about a git repository. The behavior of `gitprompt` is inspired by
[olivierverdier/zsh-git-prompt] and the [oh-my-zsh plugin] for it.

Installation
============

Binaries
--------

Pre-built 64-bit binaries are available for Linux and Mac at
<https://github.com/abhinav/gitprompt/releases>. To install, simply unpack the
archive and put the binaries somewhere on your `$PATH`.

For example, if you have `$HOME/bin` on your `$PATH`,

    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    VERSION=v0.1.0
    URL="https://github.com/abhinav/gitprompt/releases/download/$VERSION/gitprompt.$VERSION.$OS.amd64.tar.gz"
    curl -L "$URL" | tar xv -C ~/bin

Build From Source
-----------------

If you have Go installed, you can install `gitprompt` from source using the
following command.

    $ go get -u github.com/abhinav/gitprompt/cmd/gitprompt

Usage
=====

In either Bash or Zsh, add `$(gitprompt)` to your prompt. Be sure to escape it
so that the command isn't executed when you're declaring the prompt.

For example, in Bash, you can add the following to your .bashrc.

```sh
PS1='\h:\W $(gitprompt) $ '
```

Similarly, in Zsh, you can add the following to your .zshrc.

```sh
PROMPT='%M:%~ $(gitprompt) $ '
```

Credits
=======

The functionality provided by `gitprompt` is inspired by
[olivierverdier/zsh-git-prompt] and the [oh-my-zsh plugin] for it.

  [olivierverdier/zsh-git-prompt]: https://github.com/olivierverdier/zsh-git-prompt
  [oh-my-zsh plugin]: https://github.com/robbyrussell/oh-my-zsh/tree/master/plugins/git-prompt
