# syntax=docker/dockerfile:1

# Dashboard (React) — embedded into the Go binary below.
FROM node:20-alpine AS admin
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY admin/package.json admin/
RUN pnpm install --filter @sooauth/admin --frozen-lockfile
COPY admin/ admin/
RUN pnpm --filter @sooauth/admin build

FROM golang:1.23-alpine AS build
ENV GOTOOLCHAIN=auto
WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
COPY --from=admin /src/admin/dist/ ./internal/adminui/dist/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/sooauth ./cmd/sooauth

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=build /out/sooauth /usr/local/bin/sooauth

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=5 \
  CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" || exit 1

CMD ["sooauth"]
