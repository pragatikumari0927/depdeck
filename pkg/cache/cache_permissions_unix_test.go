//go:build !windows

package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPutCreatesPrivateCacheDirectoryAndFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "depdeck", "nested")
	store := &Store{Root: root}
	entry := Entry{Status: 200, FetchedAt: time.Now(), Body: []byte("ok")}
	if err := store.Put("npm", "react", ClassOK, entry); err != nil {
		t.Fatal(err)
	}

	dirInfo, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got&0o077 != 0 {
		t.Errorf("new cache directory permissions = %#o, want no group/other access", got)
	}

	fileInfo, err := os.Stat(filepath.Join(root, Key("npm", "react", ClassOK)))
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got&0o077 != 0 {
		t.Errorf("cache file permissions = %#o, want no group/other access", got)
	}
}

func TestPutPreservesExistingCacheDirectoryPermissions(t *testing.T) {
	root := t.TempDir()
	const want os.FileMode = 0o750
	if err := os.Chmod(root, want); err != nil {
		t.Fatal(err)
	}
	store := &Store{Root: root}
	entry := Entry{Status: 200, FetchedAt: time.Now(), Body: []byte("ok")}
	if err := store.Put("npm", "react", ClassOK, entry); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("existing cache directory permissions = %#o, want preserved %#o", got, want)
	}
}
