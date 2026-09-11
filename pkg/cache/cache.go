package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Class string

const (
	ClassOK       Class = "ok"
	ClassNotFound Class = "not_found"
)

type Entry struct {
	ETag      string    `json:"etag"`
	Status    int       `json:"status"`
	FetchedAt time.Time `json:"fetched_at"`
	Body      []byte    `json:"body"`
}

type Store struct {
	Root     string
	Disabled bool
}

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

func Key(source, name string, class Class) string {
	sum := sha256.Sum256([]byte(source + ":" + name + ":" + string(class)))
	return hex.EncodeToString(sum[:])
}

func Fresh(e Entry, class Class, now time.Time) bool {
	age := now.Sub(e.FetchedAt)
	if class == ClassNotFound {
		return age <= time.Hour
	}
	return age <= 24*time.Hour
}

func (s *Store) path(source, name string, class Class) string {
	return filepath.Join(s.Root, Key(source, name, class))
}

func (s *Store) Get(source, name string, class Class, now time.Time) (Entry, bool, error) {
	var zero Entry
	if s.Disabled {
		return zero, false, nil
	}
	p := s.path(source, name, class)
	raw, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, false, nil
		}
		return zero, false, fmt.Errorf("read %s: %w", p, err)
	}
	var e Entry
	if err := json.Unmarshal(raw, &e); err != nil {
		return zero, false, fmt.Errorf("json %s: %w", p, err)
	}
	if !Fresh(e, class, now) {
		return zero, false, nil
	}
	return e, true, nil
}

func cacheable(status int) bool {
	if status == 429 || status >= 500 || status < 200 {
		return false
	}
	return true
}

func (s *Store) Put(source, name string, class Class, e Entry) error {
	if s.Disabled || !cacheable(e.Status) {
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
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName) // best effort
		return fmt.Errorf("rename cache %s: %w", dest, err)
	}
	return nil
}
