FROM golang AS builder

# Install git and build tools
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o go_runner ./cmd/go_runner

FROM alpine:latest

# Install git for repo operations
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/go_runner .

# Create data directory
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./go_runner"]