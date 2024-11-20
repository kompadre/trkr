EXE := trkr
SRC := $(sh find . --name "*.go" | xargs echo)

all:	$(EXE) $(EXE).arm64 # main.arm64

$(EXE):	$(SRC)
	go build -o $(EXE) cmd/main.go

$(EXE).arm64: $(SRC)
	GOARCH=arm64 GOOS=linux CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o $(EXE).arm64 -tags linux,sdl,es2 cmd/main.go

run:	$(EXE)
	./$(EXE)

clean:
	rm $(EXE) $(EXE).arm64
