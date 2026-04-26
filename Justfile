build:
    mkdir -p bin
    go build -o bin/sloppy -ldflags "-s -w" cmd/main.go

fmt:
    gofmt -s -w .