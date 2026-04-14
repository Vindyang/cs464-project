FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/monolith ./cmd/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S app -G app && \
    mkdir -p /app/data && \
    chown -R app:app /app

WORKDIR /app

COPY --from=builder --chown=app:app /out/monolith /app/monolith

USER app

EXPOSE 8080

ENTRYPOINT ["/app/monolith"]