FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS monolith-builder

WORKDIR /src/backend

COPY backend/monolith/go.mod backend/monolith/go.sum ./
RUN go mod download

COPY backend/monolith/ ./

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/monolith ./cmd/main.go

FROM node:22-alpine AS frontend-deps

WORKDIR /src/frontend

COPY frontend/package.json ./
RUN npm install --no-audit --no-fund

FROM node:22-alpine AS frontend-builder

WORKDIR /src/frontend

COPY --from=frontend-deps /src/frontend/node_modules ./node_modules
COPY frontend/ ./

ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}

RUN npm run build

FROM node:22-alpine

RUN apk add --no-cache ca-certificates supervisor tzdata

RUN addgroup -S app && \
    adduser -S app -G app && \
    mkdir -p /app/bin /app/data /app/frontend /var/log/supervisor && \
    chown -R app:app /app /var/log/supervisor

ENV NODE_ENV=production
ENV NEXT_PUBLIC_API_URL=http://localhost:8080
ENV API_INTERNAL_URL=http://127.0.0.1:8080
ENV GATEWAY_URL=http://127.0.0.1:8080
ENV NEXT_PUBLIC_BETTER_AUTH_URL=http://localhost:3000
ENV FRONTEND_URL=http://localhost:3000
ENV Omnishard_DB_PATH=/app/data/Omnishard.db
ENV Omnishard_SHARDMAP_DB_PATH=/app/data/Omnishard-shardmap.db

COPY --from=monolith-builder --chown=app:app /out/monolith /app/bin/monolith

COPY --from=frontend-builder --chown=app:app /src/frontend/.next/standalone /app/frontend
COPY --from=frontend-builder --chown=app:app /src/frontend/.next/static /app/frontend/.next/static
COPY --from=frontend-builder --chown=app:app /src/frontend/public /app/frontend/public

COPY deploy/images/all-in-one-monolith/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

EXPOSE 3000 8080

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]