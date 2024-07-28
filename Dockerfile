# Use the official Golang image to build the application
FROM --platform=$BUILDPLATFORM golang:1.22.5-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o go-getpi

# Start a new stage from scratch using a multi-architecture compatible base image
FROM --platform=$TARGETPLATFORM alpine:latest

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/go-getpi /usr/local/bin/go-getpi

# Command to run the executable with config file as an argument
CMD ["go-getpi"]
