FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary and static files from builder
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static

# Create required directories
RUN mkdir -p /app/static/uploads /app/db/sqlite

EXPOSE 8080

CMD ["./main"] 