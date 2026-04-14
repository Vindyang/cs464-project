FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/adapter ./services/adapter/cmd/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S app -G app && \
    mkdir -p /app /app/data && \
    chown -R app:app /app

WORKDIR /app

COPY --from=builder --chown=app:app /out/adapter /app/adapter

USER app

EXPOSE 8080

ENTRYPOINT ["/app/adapter"]
