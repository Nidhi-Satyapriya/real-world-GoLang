# 🚀 Quick Start Guide

## Fastest Way to Get Running (5 minutes)

### Option 1: Using Make (Recommended)

```bash
cd week1-device-posture-agent

# Install dependencies
make install-deps

# Terminal 1: Start the API
make run-api

# Terminal 2: Start the Agent (in a new terminal)
make run-agent
```

### Option 2: Manual Setup

**Terminal 1 - Start the Collector API:**
```bash
cd week1-device-posture-agent/collector-api
pip install -r requirements.txt
python main.py
```

**Terminal 2 - Build and Run the Agent:**
```bash
cd week1-device-posture-agent/agent
go build -o agent .
./agent
```

---

## What You Should See

### Terminal 1 (Collector API):
```
╔═══════════════════════════════════════════════════╗
║     📡 COLLECTOR API v1.0 📡                     ║
║     Device Posture Report Receiver               ║
╚═══════════════════════════════════════════════════╝

🚀 Starting FastAPI server...
📍 Listening on: http://localhost:8000
```

### Terminal 2 (Go Agent):
```
╔═══════════════════════════════════════════════════╗
║     🛡️  DEVICE POSTURE AGENT v1.0 🛡️            ║
║     Cisco Secure Client - Training Edition       ║
╚═══════════════════════════════════════════════════╝

[2026-02-16 13:45:00] Collecting device status...
  ✓ Status: HEALTHY
  📍 Hostname: Nisats-MacBook-Pro.local
  🌐 IP Address: 192.168.1.100
  💾 Disk Usage: 78.23%
✓ Report sent successfully
```

---

## Test Commands

```bash
# Test in dry-run mode (no API needed)
make test

# View API docs
open http://localhost:8000/docs

# Check received reports
curl http://localhost:8000/reports

# Check API health
curl http://localhost:8000/health
```

---

## Stopping the Services

Press `Ctrl+C` in each terminal to gracefully stop the services.

---

## Need Help?

See the full `README.md` for detailed explanations, troubleshooting, and architecture details.
