// Command depdeck runs the human CLI or the stdio MCP server.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/pragatikumari0927/depdeck/internal/cli"
	mcpface "github.com/pragatikumari0927/depdeck/internal/mcp"
	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/cache"
	"github.com/pragatikumari0927/depdeck/pkg/registry"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

func main() {
	os.Exit(mainCode())
}

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if types.BuildVersion == "dev" {
		types.BuildVersion = installedVersion()
	}
	return dispatch(ctx, os.Args[1:], pipeline.Options{})
}

func installedVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	// A versioned module install has no local VCS stamp. Local builds may
	// carry a generated pseudo-version even without a release version.
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return "dev"
		}
	}
	return info.Main.Version
}

func dispatch(ctx context.Context, args []string, opt pipeline.Options) int {
	if len(args) == 0 || args[0] != "mcp" {
		return cli.Run(ctx, args, opt, os.Stdout, os.Stderr)
	}
	if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
		_, _ = fmt.Fprintln(os.Stdout, "usage: depdeck mcp")
		return 0
	}
	if len(args) != 1 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: depdeck mcp")
		return 2
	}
	if opt.Registry == nil {
		if opt.Cache == nil {
			store, err := cache.New("")
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return 2
			}
			opt.Cache = store
		}
		opt.Registry = registry.New(opt.Cache, nil)
	}
	if err := mcpface.ServeStdio(ctx, opt); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintln(os.Stderr, "MCP session:", err)
		return 2
	}
	return 0
}
