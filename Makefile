.PHONY: init infra-up infra-down run-sim run-edge run-cloud run-dashboard test

PYTHON ?= python3
VENV ?= simulator/.venv
PIP := $(VENV)/bin/pip
PY := $(VENV)/bin/python

init:
	@echo "==> Go workspace modules"
	cd shared && go mod tidy
	cd edge-daemon && go mod tidy
	cd cloud-backend && go mod tidy
	@echo "==> Python simulator venv"
	$(PYTHON) -m venv $(VENV)
	$(PIP) install --upgrade pip
	$(PIP) install -r simulator/requirements.txt
	@echo "==> Dashboard (Node)"
	cd dashboard && npm install
	@echo "==> Init complete"

infra-up:
	docker compose up -d

infra-down:
	docker compose down

run-sim:
	cd simulator && $(CURDIR)/$(VENV)/bin/python wits_generator.py

run-edge:
	@if [ -z "$$SERIAL_PORT" ]; then echo "SERIAL_PORT is required (from make run-sim output)"; exit 1; fi
	cd edge-daemon && go run ./cmd/daemon

run-cloud:
	cd cloud-backend && go run ./cmd/ingest

run-dashboard:
	cd dashboard && npm run dev

test:
	cd shared && go test ./...
	cd edge-daemon && go test ./...
	cd cloud-backend && go test ./...
