BINARY_AMD64_WAYLAND := trkr
BINARY_AMD64_WAYLAND_RELEASE := trkr-release
BINARY_AMD64_X11 := trkr-x11
BINARY_ARM64 := trkr-arm64
BINARY_WIN64 := trkr.exe

# BUILDFLAGS := -ldflags="-s -w" -gcflags="-m -l"


DOCKERBUILT := .cross-compiler-arm64-built
DOCKERFILE   := Dockerfile.xcompile-arm64
DOCKERIMAGE  := trkr-cross-compiler-arm64

SRCS := $(shell find . -name '*.go' -not -path "./.*")

.PHONY: all clean linux-amd64 linux-wayland linux-arm64 win64

all: linux-wayland

linux-wayland: $(BINARY_AMD64_WAYLAND)

release: $(BINARY_AMD64_WAYLAND_RELEASE)

linux-amd64: $(BINARY_AMD64_X11)

linux-arm64: $(BINARY_ARM64)

win64: $(BINARY_WIN64)

clean: 
	rm -f $(BINARY_AMD64_WAYLAND) $(BINARY_AMD64_X11) $(BINARY_ARM64) $(BINARY_WIN64) $(DOCKERBUILT)


$(BINARY_AMD64_WAYLAND_RELEASE): $(SRCS)
	go build ${BUILDFLAGS} -tags "wayland release" -trimpath -o $(BINARY_AMD64_WAYLAND_RELEASE) ./cmd

$(BINARY_AMD64_WAYLAND): $(SRCS)
	go build ${BUILDFLAGS} -tags "wayland" -o $(BINARY_AMD64_WAYLAND) ./cmd

$(BINARY_AMD64_X11): $(SRCS)
	go build ${BUILDFLAGS} -o $(BINARY_AMD64_X11) ./cmd

$(BINARY_ARM64): $(SRCS) $(DOCKERBUILT)
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
		-e PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig:/usr/share/pkgconfig \
		-e CGO_CFLAGS="-DGRAPHICS_API_OPENGL_ES2 -DPLATFORM_DESKTOP_SDL" \
		-e CGO_LDFLAGS="-lm" \
		$(DOCKERIMAGE) \
		sh -c "go clean -cache && go build -a -tags 'sdl es2' -o $(BINARY_ARM64) ./cmd"

$(BINARY_WIN64): $(SRCS)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o $(BINARY_WIN64) ./cmd

$(DOCKERBUILT): $(DOCKERFILE)
	docker build -t $(DOCKERIMAGE) -f $(DOCKERFILE) .
	@touch $(DOCKERBUILT)
