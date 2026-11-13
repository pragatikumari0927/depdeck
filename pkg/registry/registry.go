// Package registry fetches npm metadata and downloads for direct dependencies.
package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pragatikumari0927/depdeck/pkg/cache"
	"github.com/pragatikumari0927/depdeck/pkg/types"

	semver "github.com/Masterminds/semver/v3"
)

const (
	npmBase          = "https://registry.npmjs.org"
	npmDownloadsBase = "https://api.npmjs.org/downloads/point/last-week"
	sourceNPM        = "npm"
	sourceNPMDL      = "npm-downloads"
	fetchTimeout     = 10 * time.Second
	maxConcurrent    = 8
	maxMetadataBody  = int64(32 << 20)
	maxDownloadsBody = int64(256 << 10)
)

// Client calls npm and uses the shared response cache.
type Client struct {
	store     *cache.Store
	http      *http.Client
	base      string
	downloads string
	sem       chan struct{}
	timeout   time.Duration
	now       func() time.Time
}

// Option configures the registry client.
type Option func(*Client)

// WithBaseURL sets the npm metadata endpoint. Use it for fixture servers.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.base = trimSlash(u) }
}

// WithDownloadsURL sets the npm downloads endpoint. Use it for fixture servers.
func WithDownloadsURL(u string) Option {
	return func(c *Client) { c.downloads = trimSlash(u) }
}

// New creates a registry client with bounded requests and a cache store.
func New(store *cache.Store, httpClient *http.Client, opts ...Option) *Client {
	if store == nil {
		panic("registry: nil cache store")
	}
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	clientCopy := *httpClient
	previousRedirect := clientCopy.CheckRedirect
	clientCopy.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 same-origin redirects")
		}
		if len(via) > 0 {
			previousURL := via[len(via)-1].URL
			if req.URL.Scheme != previousURL.Scheme || !strings.EqualFold(req.URL.Host, previousURL.Host) {
				return http.ErrUseLastResponse
			}
		}
		if previousRedirect != nil {
			return previousRedirect(req, via)
		}
		return nil
	}
	c := &Client{
		store:     store,
		http:      &clientCopy,
		base:      npmBase,
		downloads: npmDownloadsBase,
		sem:       make(chan struct{}, maxConcurrent),
		timeout:   fetchTimeout,
		now:       time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Fetch enriches a package using a version that matches declared.
func (c *Client) Fetch(ctx context.Context, name, declared string) (*types.Dependency, error) {
	if !ValidPackageName(name) {
		return degradedWithVersion(name, declared, types.FetchParse, "invalid npm package name"), nil
	}
	constraintText, err := npmRange(declared)
	if err != nil {
		dep := degradedWithVersion(name, declared, types.FetchParse, err.Error())
		return dep, nil
	}
	now := c.now()
	dep, hit, err := c.fromCache(ctx, name, declared, constraintText, now)
	if err != nil {
		return nil, fmt.Errorf("cache get %s: %w", name, err)
	}
	if hit {
		return dep, nil
	}
	stale, hasStale, err := c.store.Load(sourceNPM, name, cache.ClassOK)
	if err != nil {
		return nil, fmt.Errorf("cache load %s: %w", name, err)
	}
	if hasStale {
		var cached packument
		if json.Unmarshal(stale.Body, &cached) != nil || cached.Name != name {
			if err := c.store.Delete(sourceNPM, name, cache.ClassOK); err != nil {
				return nil, fmt.Errorf("remove corrupt cache %s: %w", name, err)
			}
			hasStale = false
		}
	}

	return c.fetchHTTP(ctx, name, declared, constraintText, now, stale, hasStale)
}

func (c *Client) fromCache(ctx context.Context, name, declared, constraintText string, now time.Time) (*types.Dependency, bool, error) {
	e, hit, err := c.store.Get(sourceNPM, name, cache.ClassOK, now)
	if err != nil {
		return nil, false, err
	}
	if hit {
		dep, err := parsePackument(name, declared, constraintText, e.Body)
		if err != nil {
			if errors.Is(err, errNoCompatibleVersion) {
				return degradedWithVersion(name, declared, types.FetchNotFound, err.Error()), true, nil
			}
			if delErr := c.store.Delete(sourceNPM, name, cache.ClassOK); delErr != nil {
				return nil, false, delErr
			}
			return nil, false, nil
		}
		fetchErr, opErr := c.attachDownloads(ctx, dep, name, now)
		if opErr != nil {
			return nil, false, opErr
		}
		if fetchErr != nil {
			return degradedWithVersion(name, declared, types.FetchKind(fetchErr.Kind), fetchErr.Detail), true, nil
		}
		return dep, true, nil
	}
	_, hit, err = c.store.Get(sourceNPM, name, cache.ClassNotFound, now)
	if err != nil {
		return nil, false, err
	}
	if hit {
		return degradedWithVersion(name, declared, types.FetchNotFound, "not found (cached)"), true, nil
	}
	return nil, false, nil
}

func (c *Client) fetchHTTP(ctx context.Context, name, declared, constraintText string, now time.Time, stale cache.Entry, hasStale bool) (*types.Dependency, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/"+encodeName(name), nil)
	if err != nil {
		return degradedWithVersion(name, declared, types.FetchNetwork, "could not construct npm request"), nil
	}
	if hasStale && stale.ETag != "" {
		req.Header.Set("If-None-Match", stale.ETag)
	}
	status, headers, body, err := c.request(ctx, req, maxMetadataBody)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		kind, detail := requestFailure(err)

		return degradedWithVersion(name, declared, kind, detail), nil
	}
	switch status {
	case http.StatusNotModified:
		if !hasStale {
			return degradedWithVersion(name, declared, types.FetchNetwork, "304 without cached body"), nil
		}
		if err := c.store.Put(sourceNPM, name, cache.ClassOK, cache.Entry{
			ETag:      stale.ETag,
			Status:    http.StatusOK,
			FetchedAt: now,
			Body:      stale.Body,
		}); err != nil {
			return nil, fmt.Errorf("cache refresh %s: %w", name, err)
		}
		dep, err := parsePackument(name, declared, constraintText, stale.Body)
		if err != nil {
			return degradedWithVersion(name, declared, fetchKindFor(err), err.Error()), nil
		}
		fetchErr, opErr := c.attachDownloads(ctx, dep, name, now)
		if opErr != nil {
			return nil, fmt.Errorf("cache downloads for %s: %w", name, opErr)
		}
		if fetchErr != nil {
			return degradedWithVersion(name, declared, types.FetchKind(fetchErr.Kind), fetchErr.Detail), nil
		}
		return dep, nil
	case http.StatusOK:
		dep, err := parsePackument(name, declared, constraintText, body)
		if err != nil {
			if errors.Is(err, errNoCompatibleVersion) {
				if cacheErr := c.store.Put(sourceNPM, name, cache.ClassOK, cache.Entry{ETag: headers.Get("ETag"), Status: 200, FetchedAt: now, Body: body}); cacheErr != nil {
					return nil, fmt.Errorf("cache metadata %s: %w", name, cacheErr)
				}
			}
			return degradedWithVersion(name, declared, fetchKindFor(err), err.Error()), nil
		}
		if err := c.store.Put(sourceNPM, name, cache.ClassOK, cache.Entry{
			ETag:      headers.Get("ETag"),
			Status:    http.StatusOK,
			FetchedAt: now,
			Body:      body,
		}); err != nil {
			return nil, fmt.Errorf("cache metadata %s: %w", name, err)
		}
		fetchErr, opErr := c.attachDownloads(ctx, dep, name, now)
		if opErr != nil {
			return nil, fmt.Errorf("cache downloads for %s: %w", name, opErr)
		}
		if fetchErr != nil {
			return degradedWithVersion(name, declared, types.FetchKind(fetchErr.Kind), fetchErr.Detail), nil
		}
		return dep, nil
	case http.StatusNotFound:
		if err := c.store.Put(sourceNPM, name, cache.ClassNotFound, cache.Entry{
			Status:    http.StatusNotFound,
			FetchedAt: now,
			Body:      body,
		}); err != nil {
			return nil, fmt.Errorf("cache missing package %s: %w", name, err)
		}
		return degradedWithVersion(name, declared, types.FetchNotFound, "npm returned 404"), nil
	case http.StatusTooManyRequests:
		return degradedWithVersion(name, declared, types.FetchRateLimit, "npm returned 429"), nil
	default:
		return degradedWithVersion(name, declared, types.FetchNetwork, fmt.Sprintf("npm returned %d", status)), nil
	}
}

func (c *Client) attachDownloads(ctx context.Context, dep *types.Dependency, name string, now time.Time) (*types.FetchError, error) {
	if e, hit, err := c.store.Get(sourceNPMDL, name, cache.ClassOK, now); err == nil && hit {
		if n, ok := parseDownloads(e.Body); ok {
			dep.WeeklyDownloads = &n
			return nil, nil
		}
		if err := c.store.Delete(sourceNPMDL, name, cache.ClassOK); err != nil {
			return nil, fmt.Errorf("remove corrupt downloads cache: %w", err)
		}
	} else if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.downloads+"/"+encodeName(name), nil)
	if err != nil {
		return fetchError(types.FetchNetwork, "could not construct npm request"), nil
	}
	status, _, body, err := c.request(ctx, req, maxDownloadsBody)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		kind, detail := requestFailure(err)

		return fetchError(kind, detail), nil
	}
	if status != http.StatusOK {
		kind := types.FetchNetwork
		if status == http.StatusTooManyRequests {
			kind = types.FetchRateLimit
		}
		if status == http.StatusNotFound {
			kind = types.FetchNotFound
		}
		return fetchError(kind, fmt.Sprintf("npm downloads returned %d", status)), nil
	}
	n, ok := parseDownloads(body)
	if !ok {
		return fetchError(types.FetchParse, "invalid downloads response"), nil
	}
	dep.WeeklyDownloads = &n
	if err := c.store.Put(sourceNPMDL, name, cache.ClassOK, cache.Entry{
		Status:    http.StatusOK,
		FetchedAt: now,
		Body:      body,
	}); err != nil {
		return nil, fmt.Errorf("write downloads cache: %w", err)
	}
	return nil, nil
}

type packument struct {
	Name     string `json:"name"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Versions map[string]*versionMeta `json:"versions"`
	Time     map[string]string       `json:"time"`
}

type versionMeta struct {
	License json.RawMessage `json:"license"`
}

var errNoCompatibleVersion = errors.New("no published version matches the declared range")
var errResponseTooLarge = errors.New("registry response exceeds size limit")

func parsePackument(name, declared, rangeText string, body []byte) (*types.Dependency, error) {
	var p packument
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	if p.Name != name {
		return nil, errors.New("registry metadata name does not match requested package")
	}
	constraint, err := semver.NewConstraint(rangeText)
	if err != nil {
		return nil, err
	}
	selected := ""
	latest := p.DistTags.Latest
	if latest != "" && p.Versions[latest] != nil && matches(constraint, latest) {
		selected = latest
	}
	var best *semver.Version
	if selected == "" {
		for candidate, metadata := range p.Versions {
			v, err := semver.NewVersion(candidate)
			if metadata == nil || err != nil || v.Prerelease() != "" || !constraint.Check(v) {
				continue
			}
			if best == nil || v.GreaterThan(best) {
				best = v
				selected = candidate
			}
		}
	}
	meta, ok := p.Versions[selected]
	if selected == "" || !ok {
		return nil, errNoCompatibleVersion
	}
	dep := &types.Dependency{Name: name, Version: declared, License: licenseString(meta.License)}
	if ts, ok := p.Time[selected]; ok {
		if published, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			dep.LastPublish = published.UTC().Format(time.RFC3339)
		} else if published, err := time.Parse(time.RFC3339, ts); err == nil {
			dep.LastPublish = published.UTC().Format(time.RFC3339)
		}
	}
	return dep, nil
}

func matches(c *semver.Constraints, candidate string) bool {
	v, err := semver.NewVersion(candidate)
	return err == nil && c.Check(v)
}

func npmRange(declared string) (string, error) {
	rangeText := strings.TrimSpace(declared)
	if rangeText == "" {
		return "*", nil
	}
	if rangeText == "~0.0.0" {
		return ">=0.0.0 <0.1.0", nil
	}
	lower := strings.ToLower(rangeText)
	for _, prefix := range []string{"npm:", "workspace:", "file:", "link:", "git:", "git+", "http:", "https:", "github:"} {
		if strings.HasPrefix(lower, prefix) {
			return "", errors.New("unsupported npm declaration")
		}
	}
	if strings.Contains(rangeText, "/") {
		return "", errors.New("unsupported npm declaration")
	}
	if _, err := semver.NewConstraint(rangeText); err != nil {
		return "", errors.New("unsupported npm version range")
	}
	return rangeText, nil
}

// ValidPackageName reports whether name is a supported npm package identity.
func ValidPackageName(name string) bool {
	if len(name) == 0 || len(name) > 214 || strings.ToLower(name) != name || strings.TrimSpace(name) != name {
		return false
	}
	parts := strings.Split(name, "/")
	if strings.HasPrefix(name, "@") {
		if len(parts) != 2 || len(parts[0]) < 2 || parts[1] == "" {
			return false
		}
		parts[0] = strings.TrimPrefix(parts[0], "@")
	} else if len(parts) != 1 {
		return false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.HasPrefix(part, ".") || strings.HasPrefix(part, "_") {
			return false
		}
		for _, r := range part {
			isLower := r >= 'a' && r <= 'z'
			isDigit := r >= '0' && r <= '9'
			if !isLower && !isDigit && !strings.ContainsRune("._~-", r) {
				return false
			}
		}
	}
	return true
}

func licenseString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		s = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
				return -1
			}
			return r
		}, s)
		return s
	}
	return ""
}

func parseDownloads(body []byte) (int, bool) {
	var d struct {
		Downloads json.RawMessage `json:"downloads"`
	}
	if err := json.Unmarshal(body, &d); err != nil || len(d.Downloads) == 0 || string(d.Downloads) == "null" {
		return 0, false
	}
	var number json.Number
	if d.Downloads[0] == '"' || json.Unmarshal(d.Downloads, &number) != nil {
		return 0, false
	}
	value, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil || value < 0 || value > int64(int(^uint(0)>>1)) {
		return 0, false
	}
	return int(value), true
}

func (c *Client) request(ctx context.Context, req *http.Request, limit int64) (int, http.Header, []byte, error) {
	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-ctx.Done():
		return 0, nil, nil, ctx.Err()
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.http.Do(req.Clone(requestCtx))
	if err != nil {
		return 0, nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, resp.Header, nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return 0, nil, nil, err
	}
	if int64(len(body)) > limit {
		return resp.StatusCode, resp.Header, nil, errResponseTooLarge
	}
	return resp.StatusCode, resp.Header, body, nil
}

func degradedWithVersion(name, declared string, kind types.FetchKind, detail string) *types.Dependency {
	dep := degraded(name, kind, detail)
	dep.Version = declared
	return dep
}

func fetchError(kind types.FetchKind, detail string) *types.FetchError {
	return &types.FetchError{Source: sourceNPM, Kind: string(kind), Detail: detail}
}

func fetchKindFor(err error) types.FetchKind {
	if errors.Is(err, errNoCompatibleVersion) {
		return types.FetchNotFound
	}
	return types.FetchParse
}

func degraded(name string, kind types.FetchKind, detail string) *types.Dependency {
	return &types.Dependency{
		Name: name,
		Errors: []types.FetchError{{
			Source: sourceNPM,
			Kind:   string(kind),
			Detail: detail,
		}},
	}
}

func requestFailure(err error) (types.FetchKind, string) {
	switch {
	case errors.Is(err, errResponseTooLarge):
		return types.FetchParse, "npm response exceeds size limit"
	case errors.Is(err, context.DeadlineExceeded):
		return types.FetchTimeout, "npm request timed out"
	default:
		return types.FetchNetwork, "npm request failed"
	}
}

func encodeName(name string) string {
	return url.PathEscape(name)
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
