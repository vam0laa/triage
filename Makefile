# vamola-doctor build helpers
VERSION := $(shell sed -n 's/^const version = "\(.*\)"/\1/p' cmd/vamola-doctor/main.go)
LDFLAGS := -s -w

.PHONY: build test release clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o vamola-doctor ./cmd/vamola-doctor

test:
	go vet ./...
	go test ./...

# Static Linux binaries for GitHub Releases (amd64 + arm64) with checksums.
release: test
	rm -rf dist && mkdir -p dist
	for arch in amd64 arm64; do \
		CGO_ENABLED=0 GOOS=linux GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" -o dist/vamola-doctor ./cmd/vamola-doctor && \
		tar -C dist -czf dist/vamola-doctor-linux-$$arch.tar.gz vamola-doctor && \
		rm dist/vamola-doctor; \
	done
	cd dist && sha256sum *.tar.gz > SHA256SUMS
	@echo "release $(VERSION) built in dist/"

clean:
	rm -rf dist vamola-doctor
