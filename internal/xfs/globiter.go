// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package xfs

import (
	"errors"
	"io/fs"
	"iter"
	"os"
	"path"

	"github.com/bmatcuk/doublestar/v4"
)

// Glob iterates over the files matching the specified pattern, supporting “**”.
// Glob does not produce any directories.
func Glob(pattern string) iter.Seq2[string, fs.DirEntry] {
	return func(yield func(string, fs.DirEntry) bool) {
		base, pattern := doublestar.SplitPattern(path.Clean(pattern))
		fsys := os.DirFS(base)
		_ = doublestar.GlobWalk(fsys, pattern, func(relpath string, d fs.DirEntry) error {
			if !yield(path.Join(base, relpath), d) {
				return errors.New("aborted")
			}
			return nil
		}, doublestar.WithFilesOnly())
	}
}
