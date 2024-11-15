# Этап сборки Go
FROM golang:latest AS builder

WORKDIR /app

# Копируем зависимости и собираем приложение
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GODEBUG=x509ignoreCN=0 go build -a -installsuffix cgo -o sauk-reader .

# Этап сборки образа
FROM alpine:latest
RUN apk --no-cache add ca-certificates bash

WORKDIR /root/

# Копируем бинарный файл
COPY --from=builder /app/sauk-reader .

# Определяем переменные окружения
ENV LISTEN_PORT 8090
ENV ENDPOINT_URL http://doss.rwdt.ru/api/piscanner
ENV RETRY_FILE /tmp/retry

# Запуск приложения при старте контейнера
CMD ["./sauk-reader"]
