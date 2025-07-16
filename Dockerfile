# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o bucket-manager-bff ./cmd/main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/bucket-manager-bff .
COPY .env /app/.env
COPY --from=builder /app/templates /app/templates
EXPOSE 8000
CMD ["./bucket-manager-bff"]