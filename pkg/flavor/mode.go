package flavor

import (
	"github.com/pragatikumari0927/depdeck/pkg/types"
	"errors"
	"fmt"
)

// Mode selects the v1 Flavor behavior.
type Mode string

const (
	// Rule applies deterministic Flavor.
	Rule Mode = "rule"
	// None omits Flavor.
	None Mode = "none"
	// AI is reserved for a later release.
	AI Mode = "ai"
)

// ErrAINotInV1 reports the reserved AI setting.
var ErrAINotInV1 = errors.New("ai Flavor is not available in v1.1")

// Validate checks a requested mode before assessment.
func (m Mode) Validate() error {
	switch m {
	case "", Rule, None:
		return nil
	case AI:
		return ErrAINotInV1
	default:
		return fmt.Errorf("invalid Flavor mode %q", m)
	}
}

// ApplyMode applies or clears Flavor on a Card.
func ApplyMode(dep *types.Dependency, mode Mode) error {
	if err := mode.Validate(); err != nil {
		return err
	}
	if mode == None {
		dep.FlavorText = ""
		dep.SpecialMove = ""
		dep.FlavorSource = ""
		return nil
	}
	Apply(dep)
	return nil
}
