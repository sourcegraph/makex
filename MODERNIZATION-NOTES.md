# Modernization notes

Judgment calls made while moving this repo from GOPATH to Go modules (Go 1.26).

* **`golang.org/x/tools` pinned to `v0.0.0-20180401165715-ac136b6c2db7`, and
  `golang.org/x/tools/godoc v0.1.0-deprecated` excluded.**
  `rwvfs.FileSystem` embeds `vfs.FileSystem` from `golang.org/x/tools/godoc/vfs`.
  That interface gained a `RootType` method in April 2018, and none of rwvfs's
  implementations provide it, so rwvfs does not compile against any later x/tools.
  In September 2025 `godoc` was deleted from x/tools and republished as the
  standalone module `golang.org/x/tools/godoc v0.1.0-deprecated`; the go command
  prefers it by longest-path match, so it has to be excluded for the package to
  resolve out of the pinned x/tools instead. The alternative was inlining rwvfs
  under `internal/`, which would have changed makex's exported API
  (`makex.NewFileSystem` takes an `rwvfs.FileSystem`), so pinning won.

* **`github.com/sourcegraph/rwvfs` and `github.com/neelance/parallel` are
  unmaintained.** Both have a single 2016 commit and no tags, so they are pinned
  at pseudo-versions. Neither has a successor at a different import path.

* **`sourcegraph.com/sourcegraph/rwvfs` remains as an indirect requirement.**
  rwvfs's own subpackages (`testutil`, `cloudstoragevfs`, `httpvfs`) still import
  the old vanity path, and because rwvfs has no `go.mod` the go command pulls
  those imports into the module graph. Nothing in makex builds against it; it
  resolves only from the module proxy's cache, since the `sourcegraph.com` vanity
  host no longer serves `go-import` metadata. Removing it would require a
  `replace` directive, which downstream importers of makex would ignore.

* **`TestExternal` now skips at runtime when `make` is not on `PATH`.** It shells
  out to the system `make` via `External`, which is not present in every container.
  The test still runs on `ubuntu-latest`.

* **The `Makefile`'s Windows (`OS=Windows_NT`) branches were dropped** along with
  the GOPATH-derived `GOBIN` computation and the `find`-based staleness check.
  `go install ./cmd/makex` picks its own install directory and the build cache
  handles staleness. The Go code itself is OS-independent — no build tags, no
  `_windows.go` files — and `GOOS=windows`/`GOOS=darwin` `go build` and `go vet`
  were verified locally, but CI only tests on Linux.

* **`make_test.go` and `makefile.go` were left un-`gofmt`'d.** They carry
  pre-existing drift from older gofmt rules (struct-literal key alignment and the
  Go 1.19 doc-comment reflow). Reformatting them is unrelated churn; `go build`,
  `go vet` and `go test` are unaffected.
