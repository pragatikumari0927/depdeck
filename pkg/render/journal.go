package render

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

func fieldJournal(dep types.Dependency, rarity, group, published string) [2]string {
	shelf := "The dependency group is unavailable."
	if group != "" {
		shelf = "Filed with the " + strings.ToLower(group) + " dependencies."
	}
	if len(dep.Errors) != 0 {
		return [2]string{
			"The seal is Unknown. This scan could not verify weekly downloads.",
			shelf + " The publish date and license could not be verified.",
		}
	}

	activity := "The seal is Unknown because the weekly download count is unavailable."
	if dep.WeeklyDownloads != nil {
		downloads := *dep.WeeklyDownloads
		count := strconv.Itoa(downloads) + " downloads"
		switch downloads {
		case 0:
			count = "zero downloads"
		case 1:
			count = "one download"
		}
		activity = "npm recorded " + count + " for the assessed week."
		if rarity == "Unknown" {
			activity += " The seal is Unknown because the rank could not be verified."
		} else {
			activity += " This entry carries the " + rarity + " seal."
		}
	}

	publication := "The assessed version has no verified publish date."
	if published != "" {
		publication = "The assessed version has a recorded publish date of " + published
		// Use the assessed publish-age band. The renderer has no clock.
		if dep.Chaos != nil {
			switch *dep.Chaos {
			case 0:
				publication += ". This date falls in the most recent publish-age band"
			case 0.6:
				publication += ", two to under five years before assessment"
			case 1:
				publication += ", at least five years before assessment"
			}
		}
		publication += "."
	}

	license := "The npm record lists no license."
	switch {
	case utf8.RuneCountInString(dep.License) > 80:
		license = "See the License field for the full npm declaration."
	case dep.License != "":
		license = "The npm record lists " + dep.License + " as the license."
	}
	return [2]string{activity, shelf + " " + publication + " " + license}
}
