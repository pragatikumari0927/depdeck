package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDependencyOmitempty(t *testing.T) {
	t.Parallel()
	b, err := json.Marshal(Dependency{Name: "react", Version: "^18.2.0", Tag: TagDependencies})
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, key := range []string{
		"resolved_version", "weekly_downloads", "rarity", "chaos",
		"last_publish", "license", "flavor_text", "flavor_source", "errors",
	} {
		if strings.Contains(s, `"`+key+`"`) {
			t.Errorf("omitted field %q present in %s", key, s)
		}
	}
	if strings.Contains(s, `"stars"`) || strings.Contains(s, `"issues"`) {
		t.Errorf("GitHub fields must not exist: %s", s)
	}
}

func TestFetchErrorRoundTrip(t *testing.T) {
	t.Parallel()
	in := FetchError{Source: "npm", Kind: string(FetchNotFound), Detail: "missing"}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out FetchError
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("got %+v want %+v", out, in)
	}
}

func TestEnvelopeMarshal(t *testing.T) {
	t.Parallel()
	b, err := json.Marshal(Envelope{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   "2026-09-10T00:00:00Z",
		GeneratedBy:   "depdeck 0.0.0",
		Tool:          "scan",
		Data:          DeckData{},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, key := range []string{"schema_version", "generated_at", "generated_by", "tool", "data"} {
		if !strings.Contains(s, `"`+key+`"`) {
			t.Errorf("missing %q in %s", key, s)
		}
	}
}
