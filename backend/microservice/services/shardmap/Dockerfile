FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/shardmap ./services/shardmap/cmd/main.go

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S app -G app && \
    mkdir -p /app /app/data && \
    chown -R app:app /app

WORKDIR /app

COPY --from=builder --chown=app:app /out/shardmap /app/shardmap

USER app

EXPOSE 8081

ENTRYPOINT ["/app/shardmap"]
