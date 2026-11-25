# triage build helpers
VERSION := $(shell sed -n 's/^const version = "\(.*\)"/\1/p' cmd/triage/main.go)
LDFLAGS := -s -w

.PHONY: build test release clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o triage ./cmd/triage

test:
	go vet ./...
	go test ./...

# Static Linux binaries for GitHub Releases (amd64 + arm64) with checksums.
release: test
	rm -rf dist && mkdir -p dist
	for arch in amd64 arm64; do \
		CGO_ENABLED=0 GOOS=linux GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" -o dist/triage ./cmd/triage && \
		tar -C dist -czf dist/triage-linux-$$arch.tar.gz triage && \
		rm dist/triage; \
	done
	cd dist && sha256sum *.tar.gz > SHA256SUMS
	@echo "release $(VERSION) built in dist/"

clean:
	rm -rf dist triage
