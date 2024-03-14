# Этап сборки Go
FROM golang:latest AS builder

WORKDIR /app

# Копируем зависимости и собираем приложение
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sauk-reader .

# Этап сборки образа
FROM alpine:latest

WORKDIR /root/

# Копируем бинарный файл и файл .env
COPY --from=builder /app/sauk-reader .

# Определяем переменные окружения
ENV LISTEN_PORT 8090
ENV ENDPOINT_URL http://localhost:3000/api/piscanner
ENV RETRY_FILE /tmp/retry

# Запуск приложения при старте контейнера
CMD ["./sauk-reader"]
