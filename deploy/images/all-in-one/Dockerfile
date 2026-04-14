FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS go-builder

WORKDIR /src/backend

RUN apk add --no-cache git

COPY backend/microservice/go.mod backend/microservice/go.sum ./
RUN go mod download

COPY backend/microservice/ ./

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/adapter ./services/adapter/cmd/main.go && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/shardmap ./services/shardmap/cmd/main.go && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/sharding ./services/sharding/cmd/main.go && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/orchestrator ./services/orchestrator/cmd/main.go

FROM oven/bun:1 AS frontend-deps

WORKDIR /src/frontend

COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile

FROM oven/bun:1 AS frontend-builder

WORKDIR /src/frontend

COPY --from=frontend-deps /src/frontend/node_modules ./node_modules
COPY frontend/ ./

ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}

RUN bun run build

FROM oven/bun:1 AS bun-runtime

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates nginx supervisor && \
    rm -rf /var/lib/apt/lists/*

RUN useradd --system --create-home --shell /usr/sbin/nologin app && \
    mkdir -p /app/bin /app/data /app/frontend /var/log/supervisor /run/nginx && \
    chown -R app:app /app /var/log/supervisor

ENV NODE_ENV=production
ENV NEXT_PUBLIC_API_URL=http://localhost:8080
ENV API_INTERNAL_URL=http://127.0.0.1:8080
ENV GATEWAY_URL=http://127.0.0.1:8084
ENV NEXT_PUBLIC_BETTER_AUTH_URL=http://localhost:3000
ENV FRONTEND_URL=http://localhost:3000
ENV ADAPTER_URL=http://127.0.0.1:8080
ENV SHARDMAP_URL=http://127.0.0.1:8081
ENV SHARDING_URL=http://127.0.0.1:8083
ENV ORCHESTRATOR_URL=http://127.0.0.1:8082
ENV Omnishard_DB_PATH=/app/data/Omnishard.db
ENV Omnishard_SHARDMAP_DB_PATH=/app/data/Omnishard-shardmap.db
ENV PATH=/app/frontend/node_modules/.bin:$PATH

COPY --from=go-builder /out/adapter /app/bin/adapter
COPY --from=go-builder /out/shardmap /app/bin/shardmap
COPY --from=go-builder /out/sharding /app/bin/sharding
COPY --from=go-builder /out/orchestrator /app/bin/orchestrator

COPY --from=bun-runtime /usr/local/bin/bun /usr/local/bin/bun
COPY --from=frontend-builder /src/frontend/package.json /app/frontend/package.json
COPY --from=frontend-builder /src/frontend/node_modules /app/frontend/node_modules
COPY --from=frontend-builder /src/frontend/.next /app/frontend/.next
COPY --from=frontend-builder /src/frontend/public /app/frontend/public
COPY --from=frontend-builder /src/frontend/next.config.ts /app/frontend/next.config.ts

COPY backend/microservice/services/gateway/docs /usr/share/nginx/html/docs
COPY deploy/images/all-in-one/nginx.conf /etc/nginx/conf.d/default.conf
COPY deploy/images/all-in-one/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY deploy/images/all-in-one/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh && \
    chown -R app:app /app/frontend && \
    rm -f /etc/nginx/sites-enabled/default

EXPOSE 3000 8080

ENTRYPOINT ["/entrypoint.sh"]