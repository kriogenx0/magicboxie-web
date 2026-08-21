# syntax=docker/dockerfile:1

# ---- Stage 1: frontend ----
FROM node:20-slim AS frontend
WORKDIR /src
COPY frontend ./frontend
COPY internal/web/dist ./internal/web/dist
RUN if [ -f frontend/package.json ]; then \
      cd frontend && npm ci && npm run build; \
    else \
      echo "frontend/package.json not present yet -- keeping placeholder internal/web/dist"; \
    fi

# ---- Stage 2: Go binary ----
FROM golang:1.26 AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -o /out/magicboxie ./cmd/magicboxie

# ---- Stage 3: runtime ----
FROM ubuntu:24.04
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
    && rm -rf /var/lib/apt/lists/*
# Pre-create the config directory so bind-mounting a single config file onto
# /etc/magicboxie/config.yaml works correctly -- Docker mounts a *directory*
# at the target instead of a file if the parent path doesn't already exist
# in the image.
RUN mkdir -p /etc/magicboxie
COPY --from=backend /out/magicboxie /usr/local/bin/magicboxie
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/magicboxie"]
