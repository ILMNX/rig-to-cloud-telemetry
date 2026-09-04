.PHONY: init infra-up infra-down run-sim run-edge run-cloud

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
	$(PY) simulator/wits_generator.py

run-edge:
	cd edge-daemon && go run ./cmd/daemon

run-cloud:
	cd cloud-backend && go run ./cmd/ingest
