.PHONY: build clean test all vet release goimed goime-dict goimec

GO := go

# 检测平台，Windows 下追加 .exe 后缀
ifeq ($(OS),Windows_NT)
  EXT := .exe
else
  EXT :=
endif

build: goimed goime-dict goimec

goimed:
	$(GO) build -o goimed$(EXT) ./cmd/goimed

goime-dict:
	$(GO) build -o goime-dict$(EXT) ./cmd/goime-dict

goimec:
	$(GO) build -o goimec$(EXT) ./cmd/goimec

clean:
	rm -f goimed$(EXT) goime-dict$(EXT) goimec$(EXT)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

release:
	goreleaser release --clean


all: clean build test
