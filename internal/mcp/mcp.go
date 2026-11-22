// Package mcpface exposes the shared assessment through the MCP protocol.
package mcpface

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pragatikumari0927/depdeck/internal/pipeline"
	"github.com/pragatikumari0927/depdeck/pkg/flavor"
	"github.com/pragatikumari0927/depdeck/pkg/registry"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

const maxFrameBytes = 16 << 20

type listArgs struct {
	Path   string      `json:"path"`
	Tag    types.Tag   `json:"tag"`
	Flavor flavor.Mode `json:"flavor"`
}

type cardArgs struct {
	Name   string      `json:"name"`
	Flavor flavor.Mode `json:"flavor"`
}

type checkArgs struct {
	Path string `json:"path"`
}

type renderDeckArgs struct {
	Path   string      `json:"path"`
	Out    string      `json:"out"`
	Flavor flavor.Mode `json:"flavor"`
}

// New registers the assessment tools and the canonical HTML Deck renderer.
func New(opt pipeline.Options) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "depdeck", Version: types.BuildVersion}, nil)
	requiredString := map[string]any{"type": "string", "minLength": 1}
	flavorEnum := map[string]any{"type": "string", "enum": []flavor.Mode{flavor.Rule, flavor.None}}
	tagEnum := map[string]any{"type": "string", "enum": []types.Tag{
		types.TagDependencies, types.TagDevDependencies, types.TagOptionalDependencies,
	}}

	register(server, "list_deck", "List assessment data for direct dependencies in a project's root package.json. Use path for the project directory or manifest, tag to select one dependency group, and flavor to choose rule or none. Results include Fetch errors when npm data is unavailable.",
		"path", map[string]any{"path": requiredString, "tag": tagEnum, "flavor": flavorEnum},
		func(ctx context.Context, args listArgs) (any, error) {
			if strings.TrimSpace(args.Path) == "" {
				return nil, errors.New("set path to a project directory or a package.json file")
			}
			requestOpt := opt
			requestOpt.Flavor = args.Flavor
			deps, err := pipeline.Run(ctx, args.Path, requestOpt)
			if err != nil {
				return nil, errors.New("could not assess the project, check that path points to a readable package.json")
			}
			if args.Tag != "" {
				filtered := deps[:0]
				for _, dep := range deps {
					if dep.Tag == args.Tag {
						filtered = append(filtered, dep)
					}
				}
				deps = filtered
			}
			return types.DeckData{Dependencies: deps}, nil
		})

	register(server, "get_card", "Look up facts for one public npm package by name, such as react or @scope/package. Use flavor to choose rule or none. Call list_deck to check project membership. If npm data is unavailable, the Card includes a Fetch error.",
		"name", map[string]any{"name": requiredString, "flavor": flavorEnum},
		func(ctx context.Context, args cardArgs) (any, error) {
			if !registry.ValidPackageName(args.Name) {
				return nil, errors.New("use an npm package name such as react or @scope/package")
			}
			requestOpt := opt
			requestOpt.Flavor = args.Flavor
			dep, err := pipeline.GetCard(ctx, args.Name, requestOpt)
			if err != nil {
				return nil, errors.New("could not complete the assessment for this npm package")
			}
			return dep.StandaloneCard(), nil
		})

	register(server, "check_deck", "Evaluate direct dependencies against the default publish-age and license-declaration policy. Returns the verdict, policy, flagged dependencies, and assessed dependencies with any Fetch errors. Chaos reflects publish age, and Rarity reflects weekly downloads. Use a vulnerability scanner for security findings.",
		"path", map[string]any{"path": requiredString},
		func(ctx context.Context, args checkArgs) (any, error) {
			if strings.TrimSpace(args.Path) == "" {
				return nil, errors.New("set path to a project directory or a package.json file")
			}
			requestOpt := opt
			requestOpt.Flavor = flavor.Rule
			check, err := pipeline.Check(ctx, args.Path, requestOpt)
			if err != nil {
				return nil, errors.New("could not assess the project, check that path points to a readable package.json")
			}
			return check, nil
		})

	register(server, "render_deck", "Create a self-contained Pixel Archive HTML Deck from a project's direct dependencies using depdeck's canonical renderer. Use when the user wants the visual Deck, rather than recreating HTML from assessment data. This may contact npm. Writes a new .html file under the project directory and never overwrites an existing file; out is an optional relative path and defaults to deck.html. Returns the relative output path, Card count, and byte size.",
		"path", map[string]any{
			"path":   requiredString,
			"out":    map[string]any{"type": "string", "description": "Optional new .html path relative to the project directory; defaults to deck.html."},
			"flavor": flavorEnum,
		},
		func(ctx context.Context, args renderDeckArgs) (any, error) {
			if strings.TrimSpace(args.Path) == "" {
				return nil, errors.New("set path to a project directory or a package.json file")
			}
			output, err := cleanDeckOutput(args.Out)
			if err != nil {
				return nil, err
			}
			requestOpt := opt
			requestOpt.Flavor = args.Flavor
			deps, err := pipeline.Run(ctx, args.Path, requestOpt)
			if err != nil {
				return nil, errors.New("could not assess the project, check that path points to a readable package.json")
			}
			manifest, err := pipeline.ManifestPath(args.Path)
			if err != nil {
				return nil, errors.New("could not access the project directory")
			}
			result, err := writeDeckFile(filepath.Dir(manifest), output, deps)
			if err != nil {
				return nil, err
			}
			return result, nil
		})
	return server
}

// ServeStdio runs one MCP session on the process input and output streams.
func ServeStdio(ctx context.Context, opt pipeline.Options) error {
	err := New(opt).Run(ctx, &mcp.StdioTransport{MaxLineLength: maxFrameBytes})
	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) || errors.Is(err, io.ErrUnexpectedEOF) {
		// The SDK closes the session on malformed JSON before dispatch. Stdout
		// stays open, so return a JSON-RPC parse error with the unknown ID.
		const parseError = "{\"jsonrpc\":\"2.0\",\"id\":null,\"error\":{\"code\":-32700,\"message\":\"Parse error\"}}\n"
		if _, writeErr := io.WriteString(os.Stdout, parseError); writeErr != nil {
			return writeErr
		}
		return errors.New("malformed JSON-RPC input")
	}
	return err
}

func register[Args any](server *mcp.Server, name, description, required string, properties map[string]any,
	handle func(context.Context, Args) (any, error),
) {
	tool := &mcp.Tool{
		Name: name, Description: description,
		InputSchema: map[string]any{
			"type": "object", "properties": properties,
			"required": []string{required}, "additionalProperties": false,
		},
	}
	mcp.AddTool(server, tool, func(ctx context.Context, _ *mcp.CallToolRequest, args Args) (_ *mcp.CallToolResult, out any, err error) {
		defer func() {
			if recover() != nil {
				out = nil
				err = errors.New("could not complete the dependency assessment")
			}
		}()
		data, err := handle(ctx, args)
		if err != nil {
			return nil, nil, err
		}
		// The SDK serializes the same envelope as structured content and JSON text.
		return nil, types.Wrap(name, data, time.Now()), nil
	})
}
