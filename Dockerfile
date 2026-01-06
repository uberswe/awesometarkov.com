# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies for CGO (needed for image processing)
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o awesometarkov .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/awesometarkov .

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/resources ./resources

# Copy fonts for OG image generation
COPY --from=builder /app/fonts ./fonts

# Expose port
EXPOSE 8080

# Run the application
CMD ["./awesometarkov"]
