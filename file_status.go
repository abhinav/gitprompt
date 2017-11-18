package gitprompt

import "errors"

// StatusCode specifies the state of a file in the index or worktree.
type StatusCode byte

// Different values for StatusCode.
//
// From git status --help,
//
//   For paths with merge conflicts, X and Y show the modification states of each
//   side of the merge. For paths that do not have merge conflicts, X shows the
//   status of the index, and Y shows the status of the work tree. For untracked
//   paths, XY are ??. Other status codes can be interpreted as follows:
//
//   o   ' ' = unmodified
//
//   o   M = modified
//
//   o   A = added
//
//   o   D = deleted
//
//   o   R = renamed
//
//   o   C = copied
//
//   o   U = updated but unmerged
//
//   Ignored files are not listed, unless --ignored option is in effect, in which
//   case XY are !!.
const (
	Untracked          StatusCode = '?'
	Unmodified         StatusCode = ' '
	Modified           StatusCode = 'M'
	Added              StatusCode = 'A'
	Deleted            StatusCode = 'D'
	Renamed            StatusCode = 'R'
	Copied             StatusCode = 'C'
	UpdatedButUnmerged StatusCode = 'U'
	Ignored            StatusCode = '!'
)

// FileStatus is the state of a single file in the `git status --porcelain`
// output.
type FileStatus struct {
	Index    StatusCode // X
	WorkTree StatusCode // Y
}

// Untracked returns true if a file is completely untracked.
func (fs *FileStatus) Untracked() bool {
	return fs.Index == Untracked && fs.WorkTree == Untracked
}

// ReadFileStatus reads a FileStatus from the given line of `git status
// --porcelain` output.
func ReadFileStatus(xy string) (FileStatus, error) {
	if len(xy) < 2 {
		return FileStatus{}, errors.New("string too short")
	}

	return FileStatus{
		Index:    StatusCode(xy[0]),
		WorkTree: StatusCode(xy[1]),
	}, nil
}
