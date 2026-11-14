// Package cache stores bounded npm HTTP responses in per-package files.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Class selects the cache lifetime for a response.
type Class string

const (
	// ClassOK identifies successful metadata responses.
	ClassOK Class = "ok"
	// ClassNotFound identifies cached package-not-found responses.
	ClassNotFound Class = "not_found"
)

// Entry is one cached HTTP response.
type Entry struct {
	ETag      string    `json:"etag"`
	Status    int       `json:"status"`
	FetchedAt time.Time `json:"fetched_at"`
	Body      []byte    `json:"body"`
}

// Store reads and writes response entries under Root.
type Store struct {
	Root     string
	Disabled bool
}

const (
	maxMetadataEntrySize  = 48 << 20
	maxDownloadsEntrySize = 384 << 10
	maxMetadataBodySize   = 32 << 20
	maxDownloadsBodySize  = 256 << 10
)

// New creates a cache store. An empty root uses the user's cache directory.
func New(root string) (*Store, error) {
	if root != "" {
		return &Store{Root: root}, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("read user cache dir: %w", err)
	}
	return &Store{Root: filepath.Join(base, "depdeck")}, nil
}

// Key returns the SHA-256 cache filename for a source, package, and class.
func Key(source, name string, class Class) string {
	sum := sha256.Sum256([]byte(source + ":" + name + ":" + string(class)))
	return hex.EncodeToString(sum[:])
}

// Fresh reports whether e is within the TTL for class at now.
func Fresh(e Entry, class Class, now time.Time) bool {
	age := now.Sub(e.FetchedAt)
	if age < 0 {
		return false
	}
	if class == ClassNotFound {
		return age <= time.Hour
	}
	return age <= 24*time.Hour
}

func (s *Store) path(source, name string, class Class) string {
	return filepath.Join(s.Root, Key(source, name, class))
}

// Load returns a structurally valid entry, including an expired entry.
func (s *Store) Load(source, name string, class Class) (Entry, bool, error) {
	var zero Entry
	if s.Disabled {
		return zero, false, nil
	}
	p := s.path(source, name, class)
	limit := int64(maxMetadataEntrySize)
	bodyLimit := maxMetadataBodySize
	if source == "npm-downloads" {
		limit = maxDownloadsEntrySize
		bodyLimit = maxDownloadsBodySize
	}
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return zero, false, nil
	}
	if err != nil {
		return zero, false, fmt.Errorf("stat %s: %w", p, err)
	}
	if !info.Mode().IsRegular() {
		return zero, false, nil
	}
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("read %s: %w", p, err)
	}
	defer func() { _ = f.Close() }()
	info, err = f.Stat()
	if err != nil {
		return zero, false, fmt.Errorf("stat %s: %w", p, err)
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return zero, false, nil
	}
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return zero, false, fmt.Errorf("read %s: %w", p, err)
	}
	if int64(len(raw)) > limit {
		return zero, false, nil
	}
	var e Entry
	if err := json.Unmarshal(raw, &e); err != nil {
		return zero, false, nil
	}
	if len(e.Body) > bodyLimit || e.FetchedAt.IsZero() || (class == ClassOK && (e.Status != 200 || len(e.Body) == 0)) ||
		(class == ClassNotFound && e.Status != 404) {
		return zero, false, nil
	}
	return e, true, nil
}

// Get returns an entry only when it is fresh at now.
func (s *Store) Get(source, name string, class Class, now time.Time) (Entry, bool, error) {
	var zero Entry
	e, ok, err := s.Load(source, name, class)
	if err != nil || !ok {
		return zero, false, err
	}
	if !Fresh(e, class, now) {
		return zero, false, nil
	}
	return e, true, nil
}

func cacheable(class Class, e Entry) bool {
	switch class {
	case ClassOK:
		return e.Status == 200 && len(e.Body) > 0
	case ClassNotFound:
		return e.Status == 404
	default:
		return false
	}
}

// Put atomically replaces an entry when the response is cacheable.
func (s *Store) Put(source, name string, class Class, e Entry) error {
	if s.Disabled || !cacheable(class, e) {
		return nil
	}
	if err := os.MkdirAll(s.Root, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", s.Root, err)
	}
	dest := s.path(source, name, class)
	raw, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("json %s: %w", dest, err)
	}
	tmp, err := os.CreateTemp(s.Root, ".put-*")
	if err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := replaceFile(tmpName, dest); err != nil {
		_ = os.Remove(tmpName) // best effort
		return fmt.Errorf("rename cache %s: %w", dest, err)
	}
	return nil
}

// Delete removes an entry. A missing entry is not an error.
func (s *Store) Delete(source, name string, class Class) error {
	if s.Disabled {
		return nil
	}
	if err := os.Remove(s.path(source, name, class)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove cache entry: %w", err)
	}
	return nil
}
