# Stage 1: Build the Go binary inside a container
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o task-tracker-api .

# Stage 2: Minimal lightweight execution image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/task-tracker-api .
EXPOSE 8080
CMD ["./task-tracker-api"]