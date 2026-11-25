//go:build !windows

package atomicfile

import "os"

// Replace moves oldpath onto newpath, replacing an existing destination.
func Replace(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
