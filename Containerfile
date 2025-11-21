FROM docker.io/library/golang:1.24-alpine AS builder

ARG VERSION=development
ARG REVISION=development

WORKDIR /build

# Create passwd file for nobody user
RUN echo "nobody:x:65534:65534:Nobody:/:" > /tmp/passwd

# Install build dependencies
RUN apk add --no-cache upx ca-certificates

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.Revision=${REVISION}" \
    -trimpath \
    -o transmission-bot \
    ./cmd/transmission-bot

# Compress with UPX
RUN upx --best --lzma transmission-bot

# Final minimal image
FROM scratch

# Copy passwd file for nobody user
COPY --from=builder /tmp/passwd /etc/passwd

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder --chmod=555 /build/transmission-bot /transmission-bot

# Run as non-root user
USER 65534

ENTRYPOINT ["/transmission-bot"]
