default: build

test:
	go test -v ./...

get:
	go get

build: test
	go build
