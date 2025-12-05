# Dockerfile for gosignaling server

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build Go binary (build entire package, not just main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -o gosignaling .

# Final runtime image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/gosignaling .

# Copy static HTML files if they exist
COPY --from=builder /app/client.html ./client.html
COPY --from=builder /app/rustwasm.html ./rustwasm.html

# Expose port (default 8080)
EXPOSE 8080

# Run the server
CMD ["./gosignaling"]
