// Package mcpface exposes the shared assessment through the MCP protocol.
package mcpface

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
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

// New registers the three read-only assessment tools on an MCP server.
func New(opt pipeline.Options) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "depdeck", Version: types.BuildVersion}, nil)
	requiredString := map[string]any{"type": "string", "minLength": 1}
	flavorEnum := map[string]any{"type": "string", "enum": []flavor.Mode{flavor.Rule, flavor.None}}
	tagEnum := map[string]any{"type": "string", "enum": []types.Tag{
		types.TagDependencies, types.TagDevDependencies, types.TagOptionalDependencies,
	}}

	register(server, "list_deck", "List every direct dependency of a project as Cards. Use for project-wide questions.",
		"path", map[string]any{"path": requiredString, "tag": tagEnum, "flavor": flavorEnum},
		func(ctx context.Context, args listArgs) (any, error) {
			if strings.TrimSpace(args.Path) == "" {
				return nil, errors.New("path must name a project directory or manifest")
			}
			requestOpt := opt
			requestOpt.Flavor = args.Flavor
			deps, err := pipeline.Run(ctx, args.Path, requestOpt)
			if err != nil {
				return nil, errors.New("could not assess the project manifest")
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

	register(server, "get_card", "Fetch a single package's Card from npm. Use for package-specific questions.",
		"name", map[string]any{"name": requiredString, "flavor": flavorEnum},
		func(ctx context.Context, args cardArgs) (any, error) {
			if !registry.ValidPackageName(args.Name) {
				return nil, errors.New("name must be a supported npm package name")
			}
			requestOpt := opt
			requestOpt.Flavor = args.Flavor
			dep, err := pipeline.GetCard(ctx, args.Name, requestOpt)
			if err != nil {
				return nil, errors.New("could not assess the npm package")
			}
			return dep.StandaloneCard(), nil
		})

	register(server, "check_deck", "Return which dependencies are flagged as risky, and an overall pass/fail. Use before shipping or for CI gating.",
		"path", map[string]any{"path": requiredString},
		func(ctx context.Context, args checkArgs) (any, error) {
			if strings.TrimSpace(args.Path) == "" {
				return nil, errors.New("path must name a project directory or manifest")
			}
			requestOpt := opt
			requestOpt.Flavor = flavor.Rule
			check, err := pipeline.Check(ctx, args.Path, requestOpt)
			if err != nil {
				return nil, errors.New("could not assess the project manifest")
			}
			return check, nil
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
				err = errors.New("dependency assessment failed")
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
