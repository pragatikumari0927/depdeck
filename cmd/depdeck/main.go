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

	"github.com/Masterminds/semver/v3"
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
	if !ok {
		return "dev"
	}
	return resolveBuildVersion("dev", info)
}

func resolveBuildVersion(linkerVersion string, info *debug.BuildInfo) string {
	if linkerVersion != "dev" {
		return linkerVersion
	}
	return installedVersionFromBuildInfo(info)
}

func installedVersionFromBuildInfo(info *debug.BuildInfo) string {
	const modulePath = "github.com/pragatikumari0927/depdeck"
	if info == nil || info.Main.Path != modulePath || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	// A versioned module install has no local VCS stamp. Local builds may
	// carry VCS metadata and must not be presented as a tagged release.
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return "dev"
		}
	}
	version, err := semver.NewVersion(info.Main.Version)
	if err != nil {
		return "dev"
	}
	return version.String()
}

func dispatch(ctx context.Context, args []string, opt pipeline.Options) int {
	if len(args) == 0 || args[0] != "mcp" {
		return cli.Run(ctx, args, opt, os.Stdout, os.Stderr)
	}
	if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
		_, _ = fmt.Fprintln(os.Stdout, "Usage: depdeck mcp\nStart the MCP server over standard input and output.")
		return 0
	}
	if len(args) != 1 {
		_, _ = fmt.Fprintln(os.Stderr, "depdeck mcp takes no arguments. Run depdeck mcp --help for usage.")
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
