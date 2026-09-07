# Development

## The workspace

Three modules under one `go.work`. Module paths are unchanged from their
upstream repositories, so imports keep working and published releases stay
valid.

`go build ./...` does **not** work from the repository root: a relative pattern
does not cross module boundaries, and the root is not itself a module. Use
`make build`, or the module path patterns it runs.

## Syncing with upstream

Each component is a `git subtree`, so changes flow both ways.

Pull upstream changes in:

```bash
make sync
```

Push a change made here back to a component's own repository:

```bash
git subtree push --prefix=components/s3nitor \
  https://github.com/luhtaf/s3nitor.git some-branch
```

Two things to know before doing that. Subtree pushes rewrite commits onto the
component's history, so push to a branch and open a pull request rather than
onto the default branch. And a commit that touches more than one component
cannot be split by subtree — keep changes to one component per commit, or the
push carries changes the other repository has no place for.

## Known state of each component

Recorded because it is the kind of thing that is obvious once and forgotten
after:

**s3nitor** — 54 tests, `-race` clean. Requires `CGO_ENABLED=1` for the SQLite
driver; a cgo-less build compiles and then panics at startup, so a green build
is not evidence of a working binary. Do not cross-compile it.

**surisink** — no tests. Its default branch does not build: `go.sum` is missing
entries for viper, zap and modernc.org/sqlite, and `PutObjectTagging` was called
with a `map[string]string` where minio-go v7 requires `*tags.Tags`. Fixed here;
the fix has not been pushed upstream.

**corator** — no tests.

## Before pushing

```bash
make fmt
make vet
make test
```
