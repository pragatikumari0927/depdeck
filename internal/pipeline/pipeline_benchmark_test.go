package pipeline_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/cache"
	"github.com/pragatikumari0927/depdeck/pkg/registry"
)

func BenchmarkPipelineColdWarmScan47(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/downloads/point/last-week/") {
			_, _ = fmt.Fprint(w, `{"downloads":1250}`)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		_, _ = fmt.Fprintf(w, `{"name":%q,"dist-tags":{"latest":"1.0.0"},"versions":{"1.0.0":{"license":"MIT"}},"time":{"1.0.0":"2026-01-01T00:00:00Z"}}`, name)
	}))
	defer server.Close()

	root := b.TempDir()
	project := filepath.Join(root, "project")
	if err := os.Mkdir(project, 0700); err != nil {
		b.Fatal(err)
	}
	var body strings.Builder
	body.WriteString(`{"dependencies":{`)
	for i := range 47 {
		if i > 0 {
			body.WriteByte(',')
		}
		fmt.Fprintf(&body, "%q:%q", fmt.Sprintf("fixture-%03d", i), "^1.0.0")
	}
	body.WriteString("}}")
	if err := os.WriteFile(filepath.Join(project, "package.json"), []byte(body.String()), 0600); err != nil {
		b.Fatal(err)
	}

	makeOptions := func(cacheRoot string) pipeline.Options {
		store, err := cache.New(cacheRoot)
		if err != nil {
			b.Fatal(err)
		}
		client := registry.New(store, server.Client(),
			registry.WithBaseURL(server.URL),
			registry.WithDownloadsURL(server.URL+"/downloads/point/last-week"),
		)
		fixedNow := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
		return pipeline.Options{Cache: store, Registry: client, Now: func() time.Time { return fixedNow }}
	}

	b.Run("cold", func(b *testing.B) {
		options := make([]pipeline.Options, b.N)
		for i := range options {
			options[i] = makeOptions(filepath.Join(root, fmt.Sprintf("cold-cache-%d", i)))
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			deps, err := pipeline.Run(context.Background(), project, options[i])
			if err != nil || len(deps) != 47 {
				b.Fatalf("cold scan returned %d Cards, error %v", len(deps), err)
			}
		}
	})

	b.Run("warm", func(b *testing.B) {
		options := makeOptions(filepath.Join(root, "warm-cache"))
		if deps, err := pipeline.Run(context.Background(), project, options); err != nil || len(deps) != 47 {
			b.Fatalf("warmup scan returned %d Cards, error %v", len(deps), err)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			deps, err := pipeline.Run(context.Background(), project, options)
			if err != nil || len(deps) != 47 {
				b.Fatalf("warm scan returned %d Cards, error %v", len(deps), err)
			}
		}
	})
}
