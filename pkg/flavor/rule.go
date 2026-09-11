package flavor

import (
	"hash/fnv"
	"math/rand"
	"strings"

	"depdeck/pkg/types"
)

func Apply(dep *types.Dependency) {
	h := fnv.New64a()
	_, _ = h.Write([]byte(dep.Name))
	r := rand.New(rand.NewSource(int64(h.Sum64())))

	if len(dep.Errors) > 0 {
		moves := []string{"Name Echo", "Fog Recall", "Bare Label"}
		texts := []string{
			"{name} is known only by its name.",
			"The roster still lists {name}.",
			"{name} waits without stats.",
		}
		dep.FlavorSource = types.FlavorRuleNameOnly
		dep.SpecialMove = moves[r.Intn(len(moves))]
		dep.FlavorText = strings.ReplaceAll(texts[r.Intn(len(texts))], "{name}", dep.Name)
		return
	}

	dep.FlavorSource = types.FlavorRuleFull
	if dep.Rarity == types.RarityLegendary || dep.WeeklyDownloads >= 10000 {
		moves := []string{"Crowd Surge", "Legendary Draw", "Wide Pull"}
		texts := []string{
			"Legendary {name} still draws a crowd.",
			"{name} holds Legendary weekly reach.",
		}
		dep.SpecialMove = moves[r.Intn(len(moves))]
		dep.FlavorText = strings.ReplaceAll(texts[r.Intn(len(texts))], "{name}", dep.Name)
		return
	}

	moves := []string{"Quiet Fetch", "Steady Import", "Small Orbit"}
	texts := []string{
		"{name} ships with modest weekly downloads.",
		"{name} is judged by publish recency, not fame.",
	}
	if dep.Chaos != nil && *dep.Chaos >= 0.6 {
		texts = []string{
			"{name} shows stale publish recency.",
			"{name} carries high Chaos from last publish.",
		}
	}
	dep.SpecialMove = moves[r.Intn(len(moves))]
	dep.FlavorText = strings.ReplaceAll(texts[r.Intn(len(texts))], "{name}", dep.Name)
}
