# messaging-gateway (SVC-01) — multi-stage. go-shared es módulo privado:
# el token se pasa como BuildKit secret (no queda en capas ni history).
#   GITHUB_TOKEN=$(gh auth token) docker compose build messaging-gateway

# ── deps ─────────────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS deps
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata
ENV GOPRIVATE=github.com/hornosg/*
COPY go.mod go.sum ./
RUN --mount=type=secret,id=github_token \
    sh -c 'if [ -s /run/secrets/github_token ]; then \
      git config --global url."https://$(cat /run/secrets/github_token)@github.com/".insteadOf "https://github.com/"; \
    fi' && \
    go mod download

# ── builder (binario prod) ───────────────────────────────────────────────────
FROM deps AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -trimpath -o /out/gateway ./src

# ── development (Air hot reload; el código se monta por volumen) ──────────────
FROM deps AS development
RUN apk add --no-cache curl && go install github.com/air-verse/air@latest
WORKDIR /app
EXPOSE 8101
CMD ["air", "-c", ".air.toml"]

# ── runtime (prod) ───────────────────────────────────────────────────────────
FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata && adduser -S -D app
COPY --from=builder /out/gateway /usr/local/bin/gateway
USER app
EXPOSE 8101
ENTRYPOINT ["gateway"]
