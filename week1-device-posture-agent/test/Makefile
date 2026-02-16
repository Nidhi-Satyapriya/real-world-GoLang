.PHONY: help build run-agent run-api clean test install-deps

help:
	@echo "📋 Week 1: Device Posture Agent - Available Commands"
	@echo ""
	@echo "  make install-deps  - Install all dependencies (Go + Python)"
	@echo "  make build         - Build the Go agent binary"
	@echo "  make run-agent     - Run the Go agent"
	@echo "  make run-api       - Run the Python collector API"
	@echo "  make test          - Run agent in dry-run mode"
	@echo "  make clean         - Remove build artifacts"
	@echo ""

install-deps:
	@echo "📦 Installing Python dependencies..."
	cd collector-api && pip install -r requirements.txt
	@echo "✓ Python dependencies installed"
	@echo ""
	@echo "📦 Checking Go installation..."
	go version
	@echo "✓ Go is installed"

build:
	@echo "🔨 Building Go agent..."
	cd agent && go build -o agent .
	@echo "✓ Agent built successfully: agent/agent"

run-agent: build
	@echo "🚀 Starting Device Posture Agent..."
	cd agent && ./agent

run-api:
	@echo "📡 Starting Collector API..."
	cd collector-api && python main.py

test: build
	@echo "🧪 Running agent in dry-run mode..."
	cd agent && ./agent -dry-run -interval=5s

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f agent/agent
	rm -rf collector-api/__pycache__
	@echo "✓ Cleaned"
