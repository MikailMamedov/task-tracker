# Этап 1: Сборка бинарника внутри Linux
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o task-tracker-api .

# Этап 2: Финальный легковесный образ
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/task-tracker-api .
EXPOSE 8080
CMD ["./task-tracker-api"]