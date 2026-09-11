package render

import (
	"bytes"
	"io"
	"testing"

	"depdeck/pkg/types"
)

func render(t *testing.T, deps []types.Dependency) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := Render(deps, &buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDeterminismSame(t *testing.T) {
	t.Parallel()
	deps := []types.Dependency{
		{Name: "zebra", Version: "1.0.0", Tag: types.TagDependencies},
		{Name: "apple", Version: "2.0.0", Tag: types.TagDevDependencies},
	}
	a := render(t, deps)
	b := render(t, deps)
	if !bytes.Equal(a, b) {
		t.Fatal("same slice produced different bytes")
	}
}

func TestSortTagTiebreaker(t *testing.T) {
	t.Parallel()
	opt := types.Dependency{Name: "dup", Version: "1.0.0", Tag: types.TagOptionalDependencies}
	prod := types.Dependency{Name: "dup", Version: "2.0.0", Tag: types.TagDependencies}
	fwd := render(t, []types.Dependency{opt, prod})
	rev := render(t, []types.Dependency{prod, opt})
	if !bytes.Equal(fwd, rev) {
		t.Fatal("same name different tag produced order-dependent bytes")
	}
	di := bytes.Index(fwd, []byte(`data-tag="dependencies"`))
	oi := bytes.Index(fwd, []byte(`data-tag="optionalDependencies"`))
	if di < 0 || oi < 0 || di > oi {
		t.Fatalf("tag order dependencies=%d optional=%d; want dependencies first", di, oi)
	}
}

func TestDeterminismShuffle(t *testing.T) {
	t.Parallel()
	zebra := types.Dependency{Name: "zebra", Version: "1.0.0", Tag: types.TagDependencies}
	apple := types.Dependency{Name: "apple", Version: "2.0.0", Tag: types.TagDevDependencies}
	fwd := render(t, []types.Dependency{zebra, apple})
	rev := render(t, []types.Dependency{apple, zebra})
	if !bytes.Equal(fwd, rev) {
		t.Fatal("shuffled slice produced different bytes")
	}
	ai := bytes.Index(fwd, []byte("apple"))
	zi := bytes.Index(fwd, []byte("zebra"))
	if ai < 0 || zi < 0 || ai > zi {
		t.Fatalf("apple index %d zebra index %d; want apple before zebra", ai, zi)
	}
	in := []types.Dependency{zebra, apple}
	_ = render(t, in)
	if in[0].Name != "zebra" || in[1].Name != "apple" {
		t.Fatal("Render mutated caller slice")
	}
}

func TestXSSName(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:    "foo<script>&bar",
		Version: "1.0.0",
		Tag:     types.TagDependencies,
	}})
	if !bytes.Contains(out, []byte("foo&lt;script&gt;&amp;bar")) {
		t.Fatal("escaped name missing")
	}
	if bytes.Contains(out, []byte("foo<script>")) {
		t.Fatal("raw name payload present")
	}

	alert := render(t, []types.Dependency{{
		Name:    "<script>alert(1)</script>",
		Version: "1.0.0",
		Tag:     types.TagDependencies,
	}})
	if bytes.Contains(alert, []byte("<script>alert(1)</script>")) {
		t.Fatal("raw alert payload present")
	}
	if !bytes.Contains(alert, []byte("&lt;script&gt;alert(1)&lt;/script&gt;")) {
		t.Fatal("escaped alert name missing")
	}
}

func TestXSSFlavor(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:       "safe",
		Version:    "1.0.0",
		Tag:        types.TagDependencies,
		FlavorText: "foo<script>&bar",
	}})
	if !bytes.Contains(out, []byte("foo&lt;script&gt;&amp;bar")) {
		t.Fatal("escaped flavor missing")
	}
	if bytes.Contains(out, []byte("foo<script>")) {
		t.Fatal("raw flavor payload present")
	}
}

func TestXSSMove(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:        "safe",
		Version:     "1.0.0",
		Tag:         types.TagDependencies,
		SpecialMove: "foo<script>&bar",
	}})
	if !bytes.Contains(out, []byte("foo&lt;script&gt;&amp;bar")) {
		t.Fatal("escaped special move missing")
	}
	if bytes.Contains(out, []byte("foo<script>")) {
		t.Fatal("raw special move payload present")
	}
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	for _, deps := range [][]types.Dependency{nil, {}} {
		var buf bytes.Buffer
		if err := Render(deps, &buf); err != nil {
			t.Fatal(err)
		}
		out := buf.Bytes()
		if !bytes.Contains(out, []byte("<!DOCTYPE html>")) {
			t.Fatal("missing doctype")
		}
		if bytes.Contains(out, []byte(`class="card`)) {
			t.Fatal("empty input emitted a card")
		}
	}
}

func TestDegraded(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{
		Name:    "left-behind",
		Version: "0.0.1",
		Tag:     types.TagOptionalDependencies,
		Errors:  []types.FetchError{{Source: "npm", Kind: string(types.FetchTimeout)}},
	}})
	if !bytes.Contains(out, []byte("left-behind")) {
		t.Fatal("degraded card missing")
	}
	if !bytes.Contains(out, []byte("card--no-data")) {
		t.Fatal("missing card--no-data")
	}
	if !bytes.Contains(out, []byte("timeout")) {
		t.Fatal("missing error kind")
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()
	err := Render([]types.Dependency{{Name: "x", Version: "1", Tag: types.TagDependencies}}, errWriter{err: io.ErrClosedPipe})
	if err != io.ErrClosedPipe {
		t.Fatalf("err=%v want %v", err, io.ErrClosedPipe)
	}
}

func TestNoGeneratedAt(t *testing.T) {
	t.Parallel()
	out := render(t, []types.Dependency{{Name: "x", Version: "1", Tag: types.TagDependencies}})
	if bytes.Contains(out, []byte("generated_at")) {
		t.Fatal("generated_at in HTML")
	}
}

type errWriter struct {
	err error
}

func (w errWriter) Write([]byte) (int, error) {
	return 0, w.err
}
