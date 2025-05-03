# Stage 1: Build the Go app
FROM golang:1.23-alpine AS builder

# Set environment variables for static linking and smaller binaries
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Create app directory
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies and build
RUN go mod download
RUN go build -o healthchecker main.go

# Stage 2: Create a minimal runtime image
FROM scratch

# Copy the compiled binary from the builder
COPY --from=builder /app/healthchecker /healthchecker

# Use an unprivileged user (optional but recommended)
USER 65532:65532

# Set entrypoint
ENTRYPOINT ["/healthchecker"]