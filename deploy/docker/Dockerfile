FROM golang:1.25 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY internal ./internal/
COPY cmd ./cmd/
COPY api ./api/
COPY docs ./docs/

RUN go build -v -o test-backend ./cmd/main.go

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /usr/src/app/test-backend .
COPY .env .
CMD ["./test-backend"]