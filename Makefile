.PHONY: all build run clean

# Name of the binary
BINARY_NAME=go-getpi

# Docker image name
IMAGE_NAME=go-getpi

all: build

build:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) .

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm -v $(PWD)/config.json:/root/config.json $(IMAGE_NAME)

clean:
	rm -f $(BINARY_NAME)
	rm -f application.log
	docker rmi $(IMAGE_NAME)