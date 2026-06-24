# --- BUILD STAGE ---
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod ./
# Download Go modules (cached layer)
RUN go mod download

# Copy source code
COPY . .

# Build statically compiled binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server main.go

# --- RUN STAGE ---
FROM alpine:3.20

WORKDIR /app

# Add ca-certificates for secure requests and mkdir for persisted database
RUN apk add --no-cache ca-certificates tzdata \
    && mkdir -p /app/data

# Copy compiled binary from builder
COPY --from=builder /app/server /app/server

EXPOSE 8081

# Command to run
CMD ["/app/server"]
