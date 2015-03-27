default: build

test:
	go test -v ./...

get:
	go get -v github.com/Shopify/sarama
	go get -v github.com/docker/docker/pkg/mflag
	go get -v github.com/stretchr/testify/assert
	go get -v github.com/stretchr/testify/mock

build: test
	go build
