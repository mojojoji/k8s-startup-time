FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o health-server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/health-server .
EXPOSE 8080
CMD ["./health-server"] 