.PHONY: build clean test all goimed goime-dict goimec

GO := go

build: goimed goime-dict goimec

goimed:
	$(GO) build -o $@ ./cmd/goimed

goime-dict:
	$(GO) build -o $@ ./cmd/goime-dict

goimec:
	$(GO) build -o $@ ./cmd/goimec

clean:
	rm -f goimed goime-dict goimec

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

release:
	goreleaser release --clean

.PHONY: release

all: clean build test
