# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy only the files needed for the build
COPY main.go .
COPY go.mod .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o health-server

# Final stage
FROM scratch

# Copy only the binary
COPY --from=builder /app/health-server /health-server

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/health-server"] 