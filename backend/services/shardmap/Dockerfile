FROM golang:1.24-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/shardmap ./services/shardmap/cmd/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S app -G app

WORKDIR /app

COPY --from=builder /out/shardmap /app/shardmap

USER app

EXPOSE 8081

ENTRYPOINT ["/app/shardmap"]
