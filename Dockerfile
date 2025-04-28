FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o k8s-startup-time

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/k8s-startup-time .
CMD ["./k8s-startup-time"] 