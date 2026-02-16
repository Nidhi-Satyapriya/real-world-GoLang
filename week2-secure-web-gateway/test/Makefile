.PHONY: help policy proxy all clean test logs install

help:
	@echo "Cisco Secure Web Gateway - Make Commands"
	@echo ""
	@echo "Available commands:"
	@echo "  make install       - Install all dependencies"
	@echo "  make policy        - Run FastAPI policy engine"
	@echo "  make proxy         - Run Go proxy server"
	@echo "  make all           - Run both services (requires tmux)"
	@echo "  make test          - Run tests against the services"
	@echo "  make test-blocked  - Test blocked domains"
	@echo "  make test-allowed  - Test allowed domains"
	@echo "  make logs          - Show logs from both services"
	@echo "  make clean         - Clean build artifacts"
	@echo ""

install:
	@echo "Installing dependencies..."
	@echo "Installing Python dependencies..."
	cd policy-engine && pip install -r requirements.txt
	@echo "Installing Go dependencies..."
	cd proxy && go mod tidy
	@echo "✓ All dependencies installed"

policy:
	@echo "Starting FastAPI Policy Engine on port 8000..."
	cd policy-engine && python main.py

proxy:
	@echo "Starting Go Proxy Server on port 8080..."
	cd proxy && go run main.go

all:
	@echo "Starting all services..."
	@echo "Starting Policy Engine..."
	@cd policy-engine && python main.py > ../logs/policy.log 2>&1 & echo $$! > ../logs/policy.pid
	@sleep 2
	@echo "Starting Proxy Server..."
	@cd proxy && go run main.go > ../logs/proxy.log 2>&1 & echo $$! > ../logs/proxy.pid
	@echo ""
	@echo "✓ All services started"
	@echo ""
	@echo "Policy Engine: http://localhost:8000"
	@echo "Proxy Server:  http://localhost:8080"
	@echo ""
	@echo "To test:"
	@echo "  make test"
	@echo ""
	@echo "To stop:"
	@echo "  make stop"

stop:
	@echo "Stopping services..."
	@if [ -f logs/policy.pid ]; then \
		kill `cat logs/policy.pid` 2>/dev/null || true; \
		rm logs/policy.pid; \
		echo "✓ Policy engine stopped"; \
	fi
	@if [ -f logs/proxy.pid ]; then \
		kill `cat logs/proxy.pid` 2>/dev/null || true; \
		rm logs/proxy.pid; \
		echo "✓ Proxy server stopped"; \
	fi

test: test-health test-blocked test-allowed test-policy

test-health:
	@echo "Testing service health..."
	@echo -n "Policy Engine: "
	@curl -s http://localhost:8000/health | grep -q "healthy" && echo "✓ OK" || echo "✗ FAILED"
	@sleep 1

test-policy:
	@echo ""
	@echo "Testing Policy Engine API..."
	@echo -n "GET /policy: "
	@curl -s http://localhost:8000/policy | grep -q "blocked" && echo "✓ OK" || echo "✗ FAILED"
	@echo -n "GET /policy/domains: "
	@curl -s http://localhost:8000/policy/domains | grep -q "total" && echo "✓ OK" || echo "✗ FAILED"

test-blocked:
	@echo ""
	@echo "Testing blocked domains (should return 403)..."
	@echo -n "facebook.com: "
	@curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" http://facebook.com | grep -q "403" && echo "✓ BLOCKED" || echo "✗ FAILED"
	@echo -n "tiktok.com: "
	@curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" http://tiktok.com | grep -q "403" && echo "✓ BLOCKED" || echo "✗ FAILED"
	@echo -n "youtube.com: "
	@curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" http://youtube.com | grep -q "403" && echo "✓ BLOCKED" || echo "✗ FAILED"

test-allowed:
	@echo ""
	@echo "Testing allowed domains (should work)..."
	@echo -n "google.com: "
	@curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" http://google.com | grep -q "200\|301\|302" && echo "✓ ALLOWED" || echo "✗ FAILED"
	@echo -n "github.com: "
	@curl -x http://localhost:8080 -s -o /dev/null -w "%{http_code}" http://github.com | grep -q "200\|301\|302" && echo "✓ ALLOWED" || echo "✗ FAILED"

test-dynamic:
	@echo ""
	@echo "Testing dynamic policy updates..."
	@echo "Adding 'example.com' to blocklist..."
	@curl -s -X POST "http://localhost:8000/policy/add?domain=example.com" | grep -q "added" && echo "✓ Added" || echo "✗ Failed"
	@sleep 1
	@echo "Verifying blocklist..."
	@curl -s http://localhost:8000/policy | grep -q "example.com" && echo "✓ Verified in policy" || echo "✗ Not found"
	@echo ""
	@echo "Removing 'example.com' from blocklist..."
	@curl -s -X DELETE "http://localhost:8000/policy/remove?domain=example.com" | grep -q "removed" && echo "✓ Removed" || echo "✗ Failed"

logs:
	@mkdir -p logs
	@echo "Showing service logs..."
	@echo ""
	@echo "=== Policy Engine Logs ==="
	@if [ -f logs/policy.log ]; then tail -20 logs/policy.log; else echo "No logs found"; fi
	@echo ""
	@echo "=== Proxy Server Logs ==="
	@if [ -f logs/proxy.log ]; then tail -20 logs/proxy.log; else echo "No logs found"; fi

clean:
	@echo "Cleaning build artifacts..."
	@rm -f proxy/proxy
	@rm -rf proxy/__pycache__ policy-engine/__pycache__
	@rm -rf logs/*
	@echo "✓ Cleaned"

build:
	@echo "Building Go proxy binary..."
	cd proxy && go build -o proxy main.go
	@echo "✓ Binary created: proxy/proxy"

run-binary: build
	@echo "Running proxy from binary..."
	cd proxy && ./proxy

# Development helpers
dev-policy:
	cd policy-engine && uvicorn main:app --reload --port 8000

dev-proxy:
	cd proxy && go run main.go

# Docker commands (future)
docker-build:
	@echo "Building Docker images..."
	docker build -t swg-policy policy-engine/
	docker build -t swg-proxy proxy/

docker-run:
	@echo "Running Docker containers..."
	docker-compose up -d

docker-stop:
	docker-compose down
