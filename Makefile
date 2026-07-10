.PHONY: check fmt fmt-check test test-install vet diff-check

export GOCACHE := $(CURDIR)/.cache/go-build
export GOMODCACHE := $(CURDIR)/.cache/go-mod

check: fmt-check test test-install vet diff-check

fmt:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	gofmt -w ./cmd ./internal

fmt-check:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	@files="$$(gofmt -l ./cmd ./internal)"; if [ -n "$$files" ]; then echo "Files need gofmt:"; echo "$$files"; exit 1; fi

test:
	go test ./...

test-install:
	sh scripts/test-install.sh

vet:
	go vet ./...

diff-check:
	git diff --check
