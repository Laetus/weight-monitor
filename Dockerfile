FROM golang:alpine AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Get HTTPS certificates 
RUN apk add --no-cache ca-certificates

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY src/ src 
COPY templates/ templates 

# Build the application
RUN go build -o main src/main.go

# Build a small image
FROM scratch
COPY --from=builder /build/main /
COPY templates/ /templates/ 

# Get HTTPS certificates 
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt 

# Expose port
EXPOSE 8080

# Command to run
ENTRYPOINT ["/main"]
