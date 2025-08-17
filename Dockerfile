# Multi-stage build for optimized production image
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /src

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build optimized binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags "netgo osusergo static_build" \
    -ldflags "-s -w -extldflags '-static'" \
    -trimpath \
    -o /bin/go_wc \
    ./cmd/go_wc

# Final stage - minimal runtime image
FROM scratch

# Copy CA certificates for HTTPS requests (if needed)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data (if needed)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /bin/go_wc /go_wc

# Set the binary as entrypoint
ENTRYPOINT ["/go_wc"]