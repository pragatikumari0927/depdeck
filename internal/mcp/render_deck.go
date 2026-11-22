package mcpface

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pragatikumari0927/depdeck/pkg/render"
	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type renderDeckData struct {
	Output string `json:"output"`
	Cards  int    `json:"cards"`
	Bytes  int    `json:"bytes"`
}

// writeDeckFile writes a new Deck under the project root. It uses Root to keep
// relative paths inside the project even when an output directory is a symlink.
func writeDeckFile(projectRoot, output string, deps []types.Dependency) (renderDeckData, error) {
	output, err := cleanDeckOutput(output)
	if err != nil {
		return renderDeckData{}, err
	}

	var html bytes.Buffer
	if err := render.Render(deps, &html); err != nil {
		return renderDeckData{}, errors.New("could not render the Deck")
	}

	root, err := os.OpenRoot(projectRoot)
	if err != nil {
		return renderDeckData{}, errors.New("could not access the project directory")
	}
	defer func() { _ = root.Close() }()

	file, err := root.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return renderDeckData{}, errors.New("the output file already exists; choose another out path")
	}
	if err != nil {
		return renderDeckData{}, errors.New("could not write the Deck; use an existing directory inside the project")
	}
	created := true
	defer func() {
		if created {
			_ = file.Close()
			_ = root.Remove(output)
		}
	}()

	written, err := file.Write(html.Bytes())
	if err != nil || written != html.Len() {
		return renderDeckData{}, errors.New("could not write the complete Deck")
	}
	if err := file.Sync(); err != nil {
		return renderDeckData{}, errors.New("could not save the Deck")
	}
	if err := file.Close(); err != nil {
		return renderDeckData{}, errors.New("could not close the Deck file")
	}
	created = false

	return renderDeckData{
		Output: filepath.ToSlash(output),
		Cards:  len(deps),
		Bytes:  written,
	}, nil
}

func cleanDeckOutput(output string) (string, error) {
	if output == "" {
		output = "deck.html"
	}
	if strings.TrimSpace(output) != output || !filepath.IsLocal(output) {
		return "", errors.New("set out to a relative .html path inside the project")
	}
	output = filepath.Clean(output)
	if filepath.Ext(output) == "" || !strings.EqualFold(filepath.Ext(output), ".html") {
		return "", errors.New("set out to a relative .html path inside the project")
	}
	return output, nil
}
