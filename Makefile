BINARY := magicbox
IMAGE := magicbox
# Local dev (run-local/dev) listens on :8090, separate from Docker's :8080
# (docker-compose.yml), so both can run side by side without a port clash.
URL := $(if $(MAGICBOX_URL),$(MAGICBOX_URL),http://localhost:8090)

.PHONY: build build-local build-web build-go run run-local dev open test tidy

# Default target: builds the full multi-stage Docker image (frontend, Go
# binary, ffmpeg runtime) -- no local node/npm/go toolchain required, and
# this is what actually gets deployed (see Dockerfile/docker-compose.yml).
build:
	docker build -t $(IMAGE) .

# Host-based build for fast local dev iteration without Docker overhead;
# requires node/npm and go installed locally.
build-local: build-web build-go

build-web:
	@if [ -f frontend/package.json ]; then \
		cd frontend && npm ci && npm run build ; \
	else \
		echo "frontend/package.json not found yet -- using placeholder internal/web/dist" ; \
	fi

build-go:
	go build -o bin/$(BINARY) ./cmd/magicbox

# Runs the app the same way it's actually deployed: via Docker (see
# docker-compose.yml). Bootstraps configs/magicbox.yaml from the example on
# first run, since docker-compose bind-mounts it directly.
run:
	@test -f configs/magicbox.yaml || { \
		cp configs/magicbox.example.yaml configs/magicbox.yaml; \
		echo "Created configs/magicbox.yaml from the example -- set auth.password_hash" \
		     "(generate one with: docker run --rm $(IMAGE) hash-password '<password>')" \
		     "before logging in."; \
	}
	docker compose up --build

# Runs the locally-built binary directly (no Docker), against
# configs/magicbox.local.yaml -- for fast local dev iteration. Static
# assets (movies/music) and app data live under ./content and ./data.
run-local: build-local
	@test -f configs/magicbox.local.yaml || { \
		echo "configs/magicbox.local.yaml not found -- copy configs/magicbox.example.yaml," \
		     "point movies_dir/music_dir/data_dir at local paths (e.g. content/movies," \
		     "content/music, data), and set auth.password_hash." ; \
		exit 1 ; \
	}
	@mkdir -p content/movies content/music data
	MAGICBOX_CONFIG=configs/magicbox.local.yaml ./bin/$(BINARY)

# Same as run-local, but also opens the server in the browser once it's
# ready to accept connections. Ctrl+C stops the server.
dev: build-local
	@test -f configs/magicbox.local.yaml || { \
		echo "configs/magicbox.local.yaml not found -- copy configs/magicbox.example.yaml," \
		     "point movies_dir/music_dir/data_dir at local paths (e.g. content/movies," \
		     "content/music, data), and set auth.password_hash." ; \
		exit 1 ; \
	}
	@mkdir -p content/movies content/music data
	@( \
		i=0 ; \
		until curl -sf "$(URL)" >/dev/null 2>&1 || [ $$i -ge 100 ]; do sleep 0.2; i=$$((i+1)); done ; \
		if curl -sf "$(URL)" >/dev/null 2>&1; then \
			$(MAKE) open ; \
		else \
			echo "Server did not respond at $(URL) within 20s -- not opening browser" >&2 ; \
		fi \
	) &
	MAGICBOX_CONFIG=configs/magicbox.local.yaml ./bin/$(BINARY)

# Opens the running server in the default browser. Override the target URL
# with MAGICBOX_URL=... if your local config listens on a different address.
open:
	@open "$(URL)" 2>/dev/null || xdg-open "$(URL)" 2>/dev/null || echo "Open $(URL) in your browser"

test:
	go test ./...

tidy:
	go mod tidy
