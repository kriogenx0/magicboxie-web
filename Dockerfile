# syntax=docker/dockerfile:1

# ---- Stage 1: frontend ----
FROM node:20-slim AS frontend
WORKDIR /src
COPY web ./web
COPY internal/web/dist ./internal/web/dist
RUN if [ -f web/package.json ]; then \
      cd web && npm ci && npm run build; \
    else \
      echo "web/package.json not present yet -- keeping placeholder internal/web/dist"; \
    fi

# ---- Stage 2: Go binary ----
FROM golang:1.26 AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -o /out/magicbox ./cmd/magicbox

# ---- Stage 3: runtime ----
FROM ubuntu:24.04
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
    && rm -rf /var/lib/apt/lists/*
# Pre-create the config directory so bind-mounting a single config file onto
# /etc/magicbox/config.yaml works correctly -- Docker mounts a *directory*
# at the target instead of a file if the parent path doesn't already exist
# in the image.
RUN mkdir -p /etc/magicbox
COPY --from=backend /out/magicbox /usr/local/bin/magicbox
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/magicbox"]
