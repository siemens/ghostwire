// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package umem

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// ShmPath is the absolute path name to an instance of a temporary filesystem to
// be used to share memory between processes without making them persistent on a
// permanent mass storage.
//
// Please note the trailing slash.
const ShmPath = "/dev/shm/"

// New creates a new umem of the specified size and returns a file descriptor to
// it. The returned file descriptor has its close-on-exec mode flag set in order
// to not leak it into child processes.
//
// Behind the scenes, the umem is backed by shared memory using an unlinked file
// inside /dev/shm. Callers thus do not need to care about the life cycle of the
// backing file on non-permanent storage, except not to leak them especially in
// long-running programs. Passing them around to other processes is fine,
// though.
//
// Please make sure to check for any error returned, as in this case the
// returned fd number will be zero – which happens to be a valid file
// descriptor.
func New(length int64) (int, error) {
	return newShumem(length, ShmPath)
}

func newShumem(length int64, shmpath string) (int, error) {
	fd, err := createTemp(&unixfs, shmpath)
	if err != nil {
		return 0, err
	}
	if err := unix.Ftruncate(fd, length); err != nil {
		unix.Close(fd)
		return 0, err
	}
	return fd, nil
}

// Map the umem referenced by the specified file descriptor and return the slice
// of memory, if successful.
//
// Map automatically derives the length of the umem to map from the size of the
// backing shared memory file referenced by fd.
func Map(fd int) ([]byte, error) {
	var info unix.Stat_t
	if err := unix.Fstat(fd, &info); err != nil {
		return nil, err
	}
	return unix.Mmap(
		fd,
		0,
		int(info.Size),
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED|unix.MAP_POPULATE)
}

// Unmap the umem memory referenced by b and report the outcome. This is just a
// 1:1 convenience wrapper around [unix.Munmap].
func Unmap(b []byte) error {
	return unix.Munmap(b)
}

// createTemp is the ugly sibling of os.CreateTemp in that it doesn't need to be
// told anything to do its work and it returns a file descriptor to shared
// memory non-permanent storage. Otherwise, it returns an error, such as when no
// temporary file could be newly created. The temporary file immediately gets
// unlinked, so don't loose the file descriptor. The returned file descriptor
// has its close-on-exec mode flag set in order to not leak it into child
// processes.
func createTemp(fs tempFS, prefix string) (int, error) {
	for attempt := 0; attempt < 10; attempt++ {
		path := prefix + strconv.FormatUint(uint64(rand.Int31()), 10)
		fd, err := fs.Open(
			path,
			unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC,
			unix.S_IRUSR|unix.S_IWUSR)
		if os.IsExist(err) {
			continue // play it again, Sam.
		}
		if err := fs.Unlink(path); err != nil {
			// while the file is useable, don't proceed as there's something not
			// correct here.
			_ = fs.Close(fd)
			return 0, fmt.Errorf("cannot unlink temporary file on shared memory file system, reason: %w", err)
		}
		return fd, nil
	}
	return 0, fmt.Errorf("creating a temporary file on shared memory file system failed with too many attempts")
}

// maps to unix package filesystem operations for opening, closing, and
// unlinking temporary files.
var unixfs unixtmpfs

// Dependency injection and mocking galore!
type tempFS interface {
	Open(path string, mode int, perm uint32) (fd int, err error)
	Close(fd int) error
	Unlink(path string) error
}

type unixtmpfs struct{}

func (t *unixtmpfs) Open(path string, mode int, perm uint32) (int, error) {
	return unix.Open(path, mode, perm)
}

func (t *unixtmpfs) Close(fd int) error {
	return unix.Close(fd)
}

func (t *unixtmpfs) Unlink(path string) error {
	return unix.Unlink(path)
}
