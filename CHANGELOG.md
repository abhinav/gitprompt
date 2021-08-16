Releases
========

v0.7.0 (unreleased)
-------------------

-   Release pre-built binaries for more systems.


v0.6.0 (2021-06-23)
-------------------

-   Make `gitprompt -no-git-status` faster by dropping reliance on `git log`.


v0.5.0 (2020-04-30)
-------------------

-   **Breaking**: Either `zsh` or `bash` must be specified as a positional
    argument to `gitprompt`.
-   Escape zero-width characters so that tab completion behaves properly.


v0.4.0 (2020-03-04)
-------------------

-   Add `-no-git-status` flag that reports the branch, tag, or hash only. This
    can be useful for large repositories.


v0.3.1 (2019-08-16)
-------------------

-   Don't SIGKILL on `-timeout`. That leaves dirty the index locked.


v0.3.0 (2019-08-16)
-------------------

-   Add a `-timeout` flag to exit early if the repository is too large.


v0.2.0 (2017-11-21)
-------------------

-   Commit SHAs are now prefixed with ":".
-   Don't complain outside git repositories.


v0.1.1 (2017-10-29)
-------------------

-   Include ARM binaries.


v0.1.0 (2017-10-29)
-------------------

-   Initial release.

