# Release process

Publishing is a separate final action. Prepare the artifacts and review the
candidate before authorizing a tag, release, or tap change. See
[published releases](https://github.com/pragatikumari0927/depdeck/releases)
for shipped versions and [the roadmap](PHASES.md) for changes on `main`.

## Candidate preflight

1. Review the candidate commit and the [v1 acceptance and security evidence](https://github.com/pragatikumari0927/depdeck/blob/510113e561b931b1d888b247026322b082d8c30d/.scratch/depdeck-v1/issues/15-security-and-acceptance.md).
2. Run the [test workflow](../.github/workflows/test.yml) on the candidate.
   Manually dispatch [release preflight](../.github/workflows/release-preflight.yml)
   on the candidate ref. Record the run and the tested commit for the six
   archives and the staged macOS Homebrew install.
3. Use Go 1.27.4 for the v1.1.0 candidate. Install the official prebuilt GoReleaser
   OSS 2.20.3 and git-cliff 2.14.2 binaries. Verify their SHA-256 digests
   against the upstream release metadata. Install govulncheck 1.7.0.
4. In a fresh checkout of the exact candidate, confirm `dist/` is absent,
   populate the Go module cache, and run the snapshot commands below.
5. Review `LICENSE`, `THIRD_PARTY_NOTICES.txt`, the changelog, the
   [release notes](../RELEASE_NOTES.md), and the README demo. Check that the
   working tree is clean and verify the expected tag and tap access.

```sh
export GOTOOLCHAIN=local
test "$(go env GOVERSION)" = 'go1.27.4'
go mod download
GOPROXY=off govulncheck -show version ./...
goreleaser check
GOPROXY=off goreleaser release --snapshot --skip=publish
python3 scripts/verify-release.py dist
```

The verifier checks the expected OS and architecture combinations, archive
members, and SHA-256, then runs the host's extracted binary through CLI and MCP
smoke tests. Use a fresh checkout for each snapshot comparison. GoReleaser
writes to `dist/`. Keep it disposable and verify its path before any cleanup.

Preview Conventional Commits since the latest tag with
`git-cliff --config cliff.toml --unreleased`. Use the preview to update the
curated **Unreleased** entries in [CHANGELOG.md](../CHANGELOG.md). For the
approved release, record its version and publication date. Preserve the
existing released sections.

## Source-built Homebrew formula

The [formula template](../release/Formula/depdeck.rb.in) builds
`./cmd/depdeck` from source. Its test checks the version and the Check
envelope. The preflight workflow serves a Git archive of the candidate on
loopback, installs via a disposable tap, and runs `brew test`. It uses
`1.0.0` as the staged formula's fixture version and records the compiler
embedded in the installed executable. The production formula uses the GitHub
tarball and version for the approved tag.

After publication is authorized and the tag exists, download its exact source
tarball. Confirm that the remote tag resolves to the reviewed commit. Set
`tag` to the approved tag, save the source tarball as `depdeck-$tag.tar.gz`,
and generate the formula into a checkout of the tap.

```sh
node scripts/release-formula.mjs "$tag" "depdeck-$tag.tar.gz" /path/to/homebrew-tap/Formula/depdeck.rb
```

Review the URL, version, and SHA-256 in the generated formula. Commit that file
in the tap and open a PR using a separately scoped tap credential. Use the
repository release credential for GitHub Releases. Keep both credentials in
their respective secret stores.

## Publication and rollback

After explicit release approval, tag the exact reviewed commit, build the
GitHub release draft with GoReleaser, attach the reviewed notes, and verify the
archives and checksums before publishing the draft. Then update the tap and
verify public `brew install` and the versioned `go install` command from a
clean environment.

If a published artifact is wrong, stop distribution, mark the release as
withdrawn, and publish corrected artifacts under a new version. Preserve the
old tag and document the change.
