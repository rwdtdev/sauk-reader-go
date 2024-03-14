# Dockerfile for running the Go binary
FROM golang:latest

WORKDIR /app

COPY sauk-reader .

CMD ["./sauk-reader"]
