# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o commitgen main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/commitgen .

# Create non-root user
RUN addgroup -g 1001 -S commitgen && \
    adduser -u 1001 -S commitgen -G commitgen

# Change ownership of the binary
RUN chown commitgen:commitgen /root/commitgen

# Switch to non-root user
USER commitgen

# Set the entrypoint
ENTRYPOINT ["./commitgen"]