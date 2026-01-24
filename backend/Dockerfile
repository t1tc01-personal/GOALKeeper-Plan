# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goalkeeper-plan-api ./cmd/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /build/goalkeeper-plan-api .

# Copy config file
COPY --from=builder /build/config/config.yaml ./config/

# Expose port
EXPOSE 8000

# Run the application
CMD ["./goalkeeper-plan-api", "api"]

