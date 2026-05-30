BINARY_AMD64 := trkr
BINARY_ARM64 := trkr-arm64
BINARY_WIN64 := trkr-x64.exe

DOCKER_IMAGE := cross-compiler-arm64
DOCKERFILE   := Dockerfile.xcompile-arm64

SRCS := $(shell find . -name '*.go' -not -path "./.*")

all: $(BINARY_AMD64)

$(BINARY_AMD64): $(SRCS)
	go build -o $(BINARY_AMD64) -tags wayland ./cmd

$(DOCKER_IMAGE): $(DOCKERFILE)
	docker build -t $(DOCKER_IMAGE) -f $(DOCKERFILE) .
	@touch $(DOCKER_IMAGE)

$(BINARY_ARM64): $(SRCS) $(DOCKER_IMAGE)
	docker run --rm \
		-v "$(shell pwd)":/workspace/trkr \
		-w /workspace/trkr \
		-u $(shell id -u):$(shell id -g) \
		-e GOCACHE=/workspace/trkr/.go-cache \
		-e GOPATH=/workspace/trkr/.go-path \
		-e GOOS=linux \
		-e GOARCH=arm64 \
		-e CGO_ENABLED=1 \
		-e CC=aarch64-linux-gnu-gcc \
		$(DOCKER_IMAGE) \
		go build -o $(BINARY_ARM64) ./cmd

$(BINARY_WIN64): $(SRCS)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o $(BINARY_WIN64) ./cmd


.PHONY: all clean

clean:
	rm -f $(BINARY_AMD64) $(BINARY_ARM64) $(BINARY_WIN64) $(DOCKER_IMAGE)
