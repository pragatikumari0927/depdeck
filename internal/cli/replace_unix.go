//go:build !windows

package cli

import "os"

func replaceOutput(from, to string) error { return os.Rename(from, to) }
