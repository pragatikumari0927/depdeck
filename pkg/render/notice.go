package render

import (
	"slices"

	"github.com/pragatikumari0927/depdeck/pkg/types"
)

type archiveNotice struct {
	Title       string
	Explanation string
	NextStep    string
}

func noticeFor(kind types.FetchKind) archiveNotice {
	switch kind {
	case types.FetchNotFound:
		return archiveNotice{
			"Registry data not found",
			"npm did not return a required record for this Card. The package, a compatible version, or its download count may be missing.",
			"Check the package name and declared version, or try another scan later.",
		}
	case types.FetchNetwork:
		return archiveNotice{
			"The npm request failed",
			"Depdeck could not complete the request to npm.",
			"Check your connection or try again later.",
		}
	case types.FetchTimeout:
		return archiveNotice{
			"The registry took too long",
			"npm did not respond in time.",
			"Try another scan.",
		}
	case types.FetchRateLimit:
		return archiveNotice{
			"The registry needs a moment",
			"npm is limiting requests.",
			"Wait a moment before scanning again.",
		}
	case types.FetchParse:
		return archiveNotice{
			"The record could not be read",
			"Depdeck could not read the registry response or interpret the declared version.",
			"Check the declared version, then run a new scan.",
		}
	default:
		return archiveNotice{
			"Registry details unavailable",
			"Depdeck could not verify the registry details for this Card.",
			"Try another scan later.",
		}
	}
}

func archiveNotices(dep types.Dependency, unknownRank bool) []archiveNotice {
	var notices []archiveNotice
	for _, fetch := range dep.Errors {
		notice := noticeFor(types.FetchKind(fetch.Kind))
		if !slices.Contains(notices, notice) {
			notices = append(notices, notice)
		}
	}
	if len(notices) == 0 && (unknownRank || dep.LastPublish == "" || dep.License == "" || dep.Chaos == nil) {
		return []archiveNotice{{Title: "Some details are unavailable", Explanation: "This scan could not fill every statistic. Fields without a value show Unavailable."}}
	}
	return notices
}
