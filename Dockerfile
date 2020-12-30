FROM golang:alpine AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY src/ .

# Build the application
RUN go build -o main main.go

# Build a small image
FROM scratch
COPY --from=builder /build/main /

# Command to run
ENTRYPOINT ["/main"]
