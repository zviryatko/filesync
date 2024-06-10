.PHONY: build test

build:
	go build -o bin/$(shell basename $(PWD)) filesync

run:
	go run filesync

test:
	go test -v ./...