all: reset install test build

build:
	go build -o run *.go

test:
	go test -v

reset:
	rm -rf cache.db

install:
	dep ensure

update:
	dep ensure -update
