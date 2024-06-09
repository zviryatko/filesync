.PHONY: build test

build:
	go build -o bin/$(shell basename $(PWD)) main.go

run:
	go run main.go

test:
	go test -v ./...