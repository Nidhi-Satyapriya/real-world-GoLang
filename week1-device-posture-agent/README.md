# Week 1: Device Posture Agent - Complete Guide

## Overview

This project simulates a simplified version of **Cisco Secure Client**. The system consists of two components:

1. **Go Agent** (`agent/`) - Monitors device health and sends reports
2. **Python Collector API** (`collector-api/`) - Receives and processes reports

---

## 🏗️ Architecture & Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                      Device Posture Agent (Go)                   │
│                                                                   │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   Collector  │───▶│   Models     │───▶│   Reporter   │      │
│  │  (System     │    │  (Data       │    │  (HTTP       │      │
│  │   Data)      │    │   Structs)   │    │   Client)    │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│         │                    │                    │              │
│         └────────────────────┼────────────────────┘              │
│                              │                                   │
│                      ┌───────▼────────┐                          │
│                      │   main.go      │                          │
│                      │  (Orchestrator)│                          │
│                      └───────┬────────┘                          │
│                              │ Every 10 seconds                  │
└──────────────────────────────┼───────────────────────────────────┘
                               │
                               │ HTTP POST (JSON)
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│              Collector API (Python FastAPI)                      │
│                                                                   │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   /report    │───▶│  Validation  │───▶│  Storage     │      │
│  │  (Endpoint)  │    │  (Pydantic)  │    │  (In-Memory) │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│                              │                                   │
│                              ▼                                   │
│                      ┌───────────────┐                           │
│                      │ Alert Logic   │                           │
│                      │ (if UNHEALTHY)│                           │
│                      └───────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📁 Project Structure

```
week1-device-posture-agent/
├── agent/                    # Go Agent (Device Monitor)
│   ├── main.go              # Main orchestrator & ticker logic
│   ├── collector.go         # System data collection
│   ├── reporter.go          # HTTP client & reporting logic
│   ├── models.go            # Data structures (DeviceStatus)
│   └── go.mod               # Go module definition
│
├── collector-api/           # Python API (Report Receiver)
│   ├── main.py              # FastAPI application
│   └── requirements.txt     # Python dependencies
│
└── README.md                # This file
```

---

## 🔍 Detailed Component Breakdown

### 1️⃣ **models.go** - Data Structures

**Purpose**: Defines the data contract between the Go agent and Python API.

**Key Components**:

```go
type DeviceStatus struct {
    Hostname     string    `json:"hostname"`      // Device identifier
    IP           string    `json:"ip"`            // Local IP address
    DiskUsage    float64   `json:"disk_usage"`    // Disk usage %
    Status       string    `json:"status"`        // HEALTHY/UNHEALTHY
    Timestamp    time.Time `json:"timestamp"`     // When collected
    Message      string    `json:"message"`       // Human-readable status
}
```

**JSON Tags**: The `` `json:"hostname"` `` tags tell Go how to serialize structs to JSON. This ensures the Python API receives properly forma

---

### 2️⃣ **collector.go** - System Data Collection

**Purpose**: Interacts with the operating system to gather device information.
---

### 3️⃣ **reporter.go** - HTTP Communication

**Purpose**: Sends collected data to the Collector API via HTTP POST.

---

### 4️⃣ **main.go** - Orchestration & Timing

**Purpose**: Coordinates the entire agent lifecycle.

##### **Ticker (Every 10 Seconds)**
```go
ticker := time.NewTicker(10 * time.Second)
```
- Creates a channel that sends a message every 10 seconds
- Non-blocking: allows the agent to do other things

---

### 5️⃣ **main.py** - Collector API (Python FastAPI)

**Purpose**: Receives, validates, and processes device reports.

**Alert Logic**:
```python
if data.status == "UNHEALTHY":
    print(f"🚨 ALERT: Device {data.hostname} is CRITICAL!")
    return {"alert": True, "action_required": "Immediate attention needed"}
```

**Response Example**:
```json
{
  "msg": "Report received - UNHEALTHY device detected",
  "alert": true,
  "device": "Nisats-MacBook-Pro",
  "action_required": "Immediate attention needed"
}
```

##### `GET /reports`
- Returns all stored reports (in-memory)
- Supports pagination with `?limit=50`

##### `GET /reports/unhealthy`
- Filters only unhealthy devices
- Useful for dashboard/monitoring

##### `GET /health`
- Health check endpoint
- Returns API status and report count

---

## 🚀 Setup & Running Instructions

### Prerequisites

**Go Agent**:
- Go 1.21 or higher
- No external dependencies (uses standard library only)

**Python API**:
- Python 3.8+
- pip (Python package manager)

---

**Expected Output**:
```
╔═══════════════════════════════════════════════════╗
║     🛡️  DEVICE POSTURE AGENT v1.0 🛡️            ║
║     Cisco Secure Client - Training Edition       ║
╚═══════════════════════════════════════════════════╝

🚀 Device Posture Agent started
   Collector URL: http://localhost:8000/report
   Report Interval: 10s
   Dry Run Mode: false
   Press Ctrl+C to stop
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[2026-02-16 13:45:00] Collecting device status...
  ✓ Status: HEALTHY
  📍 Hostname: Nisats-MacBook-Pro.local
  🌐 IP Address: 192.168.1.100
  💾 Disk Usage: 78.23%
  💬 Message: All systems operational
✓ Report sent successfully: {"msg":"Report received successfully",...}
```

---
