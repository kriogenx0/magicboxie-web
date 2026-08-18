BINARY := magicbox
IMAGE := magicbox
# Local dev (run-local/dev) listens on :8090, separate from Docker's :8080
# (docker-compose.yml), so both can run side by side without a port clash.
URL := $(if $(MAGICBOX_URL),$(MAGICBOX_URL),http://localhost:8090)

# Raspberry Pi (Raspbian/Raspberry Pi OS) bare-metal target: builds and
# runs directly on the Pi via systemd (deploy/systemd/magicbox.service)
# instead of Docker -- lighter on Pi-class hardware and avoids multi-arch
# image builds. Run pi-* targets ON the Pi itself.
PI_GO_VERSION := 1.26.5
PI_ARCH := $(shell uname -m | sed -e 's/aarch64/arm64/' -e 's/armv7l/armv6l/')
# apt's go/nodejs are almost always too old for this repo's go.mod/Vite --
# pi-setup installs Go to /usr/local/go instead of relying on PATH, since a
# non-login `make` invocation right after pi-setup won't have re-sourced
# /etc/profile.d yet.
PI_GO := $(shell command -v go 2>/dev/null || echo /usr/local/go/bin/go)
# Where movies_dir/music_dir live on the Pi's own filesystem (e.g. an
# external drive) -- override with `make pi-setup PI_CONTENT_DIR=/mnt/media`.
PI_CONTENT_DIR ?= /content
# mDNS name this Pi answers to (http://$(PI_HOSTNAME).local:8080) -- lets a
# magicboxie-device Pi find this server without a fixed/static IP.
PI_HOSTNAME ?= magicboxie

.PHONY: default build build-local build-web build-go run run-local dev restart deploy open test tidy pi-setup pi-install pi-run pi-logs

# Keep the no-argument workflow aligned with `make dev`.
default: dev

# Builds the full multi-stage Docker image (frontend, Go binary, ffmpeg
# runtime) -- no local node/npm/go toolchain required, and this is what gets
# deployed (see Dockerfile/docker-compose.yml).
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

restart: build-local
	@test -f configs/magicbox.local.yaml || { \
		echo "configs/magicbox.local.yaml not found -- copy configs/magicbox.example.yaml," \
		     "point movies_dir/music_dir/data_dir at local paths (e.g. content/movies," \
		     "content/music, data), and set auth.password_hash." ; \
		exit 1 ; \
	}
	@mkdir -p content/movies content/music data
	@pkill -f './bin/magicbox' || true
	@MAGICBOX_CONFIG=configs/magicbox.local.yaml ./bin/$(BINARY) > /tmp/magicbox.log 2>&1 & \
	 echo "Restarted MagicBox (PID $$!)"

# Deploy the published container image and refresh this app's production
# compose/nginx/TLS configuration. Server bootstrap remains a one-time task
# handled by deploy/server_setup.sh.
deploy:
	./deploy/deploy.sh

# One-time, run on the Pi itself: installs Go (official tarball -- apt's
# package is far behind go.mod's requirement) + Node (via NodeSource --
# apt's version varies too much by Raspbian release to trust for Vite) +
# ffmpeg, creates the `magicbox` system user the systemd unit runs as, and
# the directories it needs. Safe to re-run.
pi-setup:
	@echo "Detected arch: $(PI_ARCH)"
	sudo apt-get update
	sudo apt-get install -y curl ffmpeg avahi-daemon
	@if [ "$$(hostname)" != "$(PI_HOSTNAME)" ]; then \
		echo "Setting hostname to $(PI_HOSTNAME) (was $$(hostname))" ; \
		sudo hostnamectl set-hostname "$(PI_HOSTNAME)" ; \
		sudo sed -i "s/127\\.0\\.1\\.1.*/127.0.1.1\t$(PI_HOSTNAME)/" /etc/hosts ; \
	fi
	sudo systemctl enable --now avahi-daemon
	@echo "Reachable at http://$(PI_HOSTNAME).local:8080 once the service is running (see pi-run)."
	@if ! $(PI_GO) version 2>/dev/null | grep -q "go$(PI_GO_VERSION)"; then \
		echo "Installing Go $(PI_GO_VERSION) ($(PI_ARCH))" ; \
		curl -fsSL "https://go.dev/dl/go$(PI_GO_VERSION).linux-$(PI_ARCH).tar.gz" -o /tmp/go.tar.gz ; \
		sudo rm -rf /usr/local/go ; \
		sudo tar -C /usr/local -xzf /tmp/go.tar.gz ; \
		rm /tmp/go.tar.gz ; \
		echo 'export PATH=$$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null ; \
	fi
	@command -v node >/dev/null 2>&1 || { \
		curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - ; \
		sudo apt-get install -y nodejs ; \
	}
	sudo id magicbox >/dev/null 2>&1 || sudo useradd --system --create-home --home-dir /opt/magicbox --shell /usr/sbin/nologin magicbox
	sudo mkdir -p /opt/magicbox/bin /etc/magicbox /var/lib/magicbox $(PI_CONTENT_DIR)/movies $(PI_CONTENT_DIR)/music
	sudo chown -R magicbox:magicbox /opt/magicbox /var/lib/magicbox $(PI_CONTENT_DIR)
	@echo "Done. Open a new shell (so /usr/local/go/bin is on PATH), then: make pi-install"

# Builds on the Pi and installs the binary + systemd unit as the `magicbox`
# system user. Never overwrites an existing /etc/magicbox/config.yaml --
# that holds the auth password hash and TMDB token, set by hand once (see
# configs/magicbox.example.yaml's own comments) and preserved across
# re-installs/updates.
pi-install:
	cd frontend && npm ci && npm run build
	$(PI_GO) build -o bin/$(BINARY) ./cmd/magicbox
	sudo install -m 0755 -o magicbox -g magicbox bin/$(BINARY) /opt/magicbox/bin/$(BINARY)
	sudo test -f /etc/magicbox/config.yaml || { \
		sudo install -m 0640 -o magicbox -g magicbox configs/magicbox.example.yaml /etc/magicbox/config.yaml ; \
		echo "Created /etc/magicbox/config.yaml from the example -- set auth.password_hash" \
		     "(generate with: /opt/magicbox/bin/$(BINARY) hash-password '<password>')" \
		     "and tmdb.api_read_token before starting the service." ; \
	}
	sudo install -m 0644 deploy/systemd/magicbox.service /etc/systemd/system/magicbox.service
	sudo systemctl daemon-reload
	sudo systemctl enable magicbox
	@echo "Installed. Run 'make pi-run' to (re)start it."

# Rebuilds + reinstalls (pi-install), then (re)starts the systemd service
# and tails its logs -- Ctrl+C stops following logs without stopping the
# service. Safe to re-run after a git pull to deploy an update.
pi-run: pi-install
	sudo systemctl restart magicbox
	$(MAKE) pi-logs

pi-logs:
	sudo journalctl -u magicbox -f

# Opens the running server in the default browser. Override the target URL
# with MAGICBOX_URL=... if your local config listens on a different address.
open:
	@open "$(URL)" 2>/dev/null || xdg-open "$(URL)" 2>/dev/null || echo "Open $(URL) in your browser"

test:
	go test ./...

tidy:
	go mod tidy
