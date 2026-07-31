BINARY := magicbox
IMAGE := magicbox
URL := $(if $(MAGICBOX_URL),$(MAGICBOX_URL),http://localhost:8080)

.PHONY: build build-local build-web build-go run open test tidy

# Default target: builds the full multi-stage Docker image (frontend, Go
# binary, ffmpeg runtime) -- no local node/npm/go toolchain required, and
# this is what actually gets deployed (see Dockerfile/docker-compose.yml).
build:
	docker build -t $(IMAGE) .

# Host-based build for fast local dev iteration without Docker overhead
# (used by `make run`); requires node/npm and go installed locally.
build-local: build-web build-go

build-web:
	@if [ -f web/package.json ]; then \
		cd web && npm ci && npm run build ; \
	else \
		echo "web/package.json not found yet -- using placeholder internal/web/dist" ; \
	fi

build-go:
	go build -o bin/$(BINARY) ./cmd/magicbox

run: build-go
	MAGICBOX_CONFIG=$${MAGICBOX_CONFIG:-configs/magicbox.local.yaml} ./bin/$(BINARY)

# Opens the running server in the default browser. Override the target URL
# with MAGICBOX_URL=... if your local config listens on a different address.
open:
	@open "$(URL)" 2>/dev/null || xdg-open "$(URL)" 2>/dev/null || echo "Open $(URL) in your browser"

test:
	go test ./...

tidy:
	go mod tidy
