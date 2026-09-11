# One entry point for three modules.
#
# `go build ./...` does not work from a workspace root: a relative pattern does
# not cross module boundaries, and the root is not itself a module. Either the
# module path patterns below or a loop over the directories works; the patterns
# are used so a failure names the component.

MODULES := github.com/luhtaf/corator/... \
           github.com/luhtaf/surisink/... \
           github.com/luhtaf/s3nitor/...

COMPONENTS := corator surisink s3nitor

.PHONY: build test vet fmt tidy sync help

help:
	@echo "build   compile every component"
	@echo "test    run every component's tests (-race)"
	@echo "vet     go vet across the workspace"
	@echo "fmt     gofmt every component"
	@echo "tidy    go mod tidy in each module"
	@echo "sync    pull upstream changes into components/ (see docs/development.md)"

# CGO_ENABLED=1 throughout: s3nitor's SQLite driver requires it, and a build
# without cgo links a stub that only fails at runtime. It costs the other two
# nothing.
build:
	CGO_ENABLED=1 go build $(MODULES)

test:
	CGO_ENABLED=1 go test -race $(MODULES)

vet:
	CGO_ENABLED=1 go vet $(MODULES)

fmt:
	@for c in $(COMPONENTS); do gofmt -w components/$$c; done

tidy:
	@for c in $(COMPONENTS); do (cd components/$$c && go mod tidy); done

# Each component is a git subtree, so upstream changes come in with history
# rather than as a re-copy. Pushing back the other way is the same command with
# `push` — see docs/development.md before using it.
# s3nitor tracks its feature branch, not main: main still holds only the initial
# commit, and everything since lives on the branch. Point this at main once that
# branch is merged.
sync:
	git subtree pull --prefix=components/corator  https://github.com/luhtaf/corator.git  main   --squash
	git subtree pull --prefix=components/surisink https://github.com/luhtaf/surisink.git master --squash
	git subtree pull --prefix=components/s3nitor  https://github.com/luhtaf/s3nitor.git  main --squash
