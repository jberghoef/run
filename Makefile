all: reset install test build

build:
	go build -o run *.go

test:
	go test -v
