## DEVELOPMENT STAGE
FROM golang:1.26-trixie AS development

# Set direktori kerja
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies.
#
# Bukan `go work sync`: perintah itu tidak mengunduh apa pun, melainkan menulis
# ulang direktif `go` di go.work dan tiap go.mod mengikuti versi toolchain image.
# Akibatnya versi yang tercatat di repo ikut berubah tanpa disengaja, dan build
# gagal begitu versi image berbeda dari versi yang tertulis.
RUN go -C app mod download && go -C modules/stunting mod download

# Install Watch tool untuk live reload saat development
RUN go install github.com/air-verse/air@latest

# Instal Delve untuk debugging
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Build Producer dengan CGO untuk Kafka support
RUN go build -o /app/main ./app/main.go

EXPOSE 2025

# CMD ["air", "--build.cmd", "go build -o /app/main /app/webcore/main.go", "--build.bin", "/app/webcore", "--debug.host", "0.0.0.0", "--debug.port", "2345"]
CMD ["air", "-c", "/app/air.toml"]

## BUILD STAGE
FROM golang:1.26-alpine3.23 AS builder

# Install packages required for CGO and librdkafka
# build-base: Provides GCC, make, and other build tools (equivalent to build-essential in Debian/Ubuntu)
# librdkafka-dev: Development headers for librdkafka
# librdkafka: Runtime library for librdkafka
RUN apk add --no-cache \
    gcc \
    musl-dev \
    librdkafka-dev \
    pkgconfig \
    build-base

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Copy go mod files
# COPY go.work go.work.sum ./

# Download dependencies.
#
# Bukan `go work sync`: perintah itu tidak mengunduh apa pun, melainkan menulis
# ulang direktif `go` di go.work dan tiap go.mod mengikuti versi toolchain image.
# Akibatnya versi yang tercatat di repo ikut berubah tanpa disengaja, dan build
# gagal begitu versi image berbeda dari versi yang tertulis.
RUN go -C app mod download && go -C modules/stunting mod download

# Build the application
# Dengan Kafka ditanam ke dalam binary
RUN CGO_ENABLED=1 GOOS=linux go build -tags musl -o /app/main /app/app/main.go

# # Jika Tanpa Kafka
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main /app/webcore/main.go

# Build the migrate tool
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate app/init/migrate.go

## PRODUCTION STAGE
FROM alpine:3.23 AS production

# Install runtime dependencies for librdkafka
# The binary is built with CGO_ENABLED=1 and dynamically links to librdkafka
RUN apk add --no-cache \
    ca-certificates \
    librdkafka \
    libssl3 \
    libcrypto3 \
    zlib \
    libstdc++ \
    libgcc

# Create user and group with limited privileges (non-root)
RUN addgroup -g 2000 webcore && \
    adduser -D -u 2000 -G webcore webcore

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main /usr/local/bin/webcore
COPY --from=builder /app/migrate /usr/local/bin/migrate

# Copy configuration files
COPY --from=builder /app/config.yaml.example ./config.yaml
COPY --from=builder /app/access.yaml.example ./access.yaml
COPY --from=builder /app/webcore/init .

# Set ownership of working directory to webcore user
RUN chown -R webcore:webcore /app

# Switch to webcore user
USER webcore

# Expose port
EXPOSE 2025

# Run the application
CMD ["webcore", "stunting"]
