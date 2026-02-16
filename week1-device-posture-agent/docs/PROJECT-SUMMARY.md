#  Project Summary: Week 1 Device Posture Agent

## What I Built

A complete **Device Posture Agent** system with two components:

1. **Go Agent** - Monitors system health and reports to a central server
2. **Python Collector API** - Receives and processes health reports

---

## 📁 Complete Project Structure

```
week1-device-posture-agent/
│
├── 📖 Documentation
│   ├── README.md              # Comprehensive guide (9000+ words)
│   ├── QUICKSTART.md          # 5-minute getting started guide
│   └── PROJECT-SUMMARY.md     # This file
│
├── 🔧 Build & Test Tools
│   ├── Makefile               # Build automation commands
│   └── .gitignore             # Git ignore rules
│
├── 🐹 Go Agent (agent/)
│   ├── main.go                # Orchestration & periodic execution
│   ├── collector.go           # System data collection (OS interaction)
│   ├── reporter.go            # HTTP client & API communication
│   ├── models.go              # Data structures & constants
│   └── go.mod                 # Go module definition
│
└── Python Collector API (collector-api/)
    ├── main.py                # FastAPI application with endpoints
    └── requirements.txt       # Python dependencies
```
---

## 🏗️ Architecture Breakdown

### Go Agent Components

#### 1. **models.go** (Data Layer)
- **Purpose**: Defines data schema
- **Key Structs**: `DeviceStatus`

#### 2. **collector.go** (System Interaction Layer)
- **Purpose**: Gathers system information
- **Key Functions**:
  - `GetHostname()` - Uses `os` package
  - `GetLocalIP()` - Uses `net` package
  - `GetDiskUsage()` - Uses `os/exec` package (cross-platform)
  - `CollectDeviceStatus()` - Orchestrates collection

#### 3. **reporter.go** (Network Layer)
- **Purpose**: Sends data to API
- **Key Functions**:
  - `SendReport()` - HTTP POST with JSON
  - `SendReportWithRetry()` - Exponential backoff retry logic
- **Features**: Timeout handling, error wrapping

#### 4. **main.go** (Orchestration Layer)
- **Purpose**: Coordinates everything
- **Key Features**:
  - Periodic execution with `time.Ticker`
  - Graceful shutdown (Ctrl+C handling)
  - Formatted console output
### Python Collector API

#### **main.py** (API Layer)
- **Framework**: FastAPI
- **Endpoints**:
  - `POST /report` - Receive device status
  - `GET /reports` - List all reports
  - `GET /reports/unhealthy` - Filter unhealthy devices
  - `GET /reports/{hostname}` - Get device-specific reports
  - `GET /health` - API health check
  - `DELETE /reports` - Clear stored data
- **Features**:
  - Pydantic validation
  - In-memory storage
  - Alert detection
  - Auto-generated OpenAPI docs

---

## 🔬 How It Works - Detailed Flow

### Every 10 Seconds:

```
┌─────────────────────────────────────────────────────────┐
│ 1. TIMER FIRES                                          │
│    time.Ticker sends message on channel                 │
│    → main.go receives and calls collectAndReport()     │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 2. COLLECT SYSTEM DATA                                  │
│    collector.GetHostname()                              │
│    → os.Hostname() → "Nisats-MacBook-Pro.local"        │
│                                                          │
│    collector.GetLocalIP()                               │
│    → net.InterfaceAddrs() → "192.168.1.100"            │
│                                                          │
│    collector.GetDiskUsage()                             │
│    → exec.Command("df -h /") → parse output → 78.23%   │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 3. EVALUATE HEALTH                                      │
│    if diskUsage > 90.0:                                 │
│        status = "UNHEALTHY"                             │
│    else:                                                │
│        status = "HEALTHY"                               │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 4. CREATE DEVICE STATUS STRUCT                          │
│    DeviceStatus{                                        │
│        Hostname: "Nisats-MacBook-Pro.local",            │
│        IP: "192.168.1.100",                             │
│        DiskUsage: 78.23,                                │
│        Status: "HEALTHY",                               │
│        Timestamp: time.Now(),                           │
│        Message: "All systems operational"               │
│    }                                                    │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 5. SERIALIZE TO JSON                                    │
│    json.Marshal(status)                                 │
│    → {"hostname":"...","ip":"...","disk_usage":78.23,...}│
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 6. HTTP POST REQUEST                                    │
│    POST http://localhost:8000/report                    │
│    Content-Type: application/json                       │
│    Body: {JSON payload}                                 │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 7. API RECEIVES & VALIDATES                             │
│    FastAPI + Pydantic validate JSON structure           │
│    → Ensure all required fields present                 │
│    → Ensure disk_usage is 0-100                         │
│    → Parse timestamp                                    │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 8. STORE & ALERT                                        │
│    reports_db.append(data)                              │
│                                                          │
│    if status == "UNHEALTHY":                            │
│        print("🚨 ALERT: Device is CRITICAL!")          │
│        return {"alert": true, ...}                      │
│    else:                                                │
│        return {"msg": "Report received"}                │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 9. AGENT LOGS SUCCESS                                   │
│    "✓ Report sent successfully"                         │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│ 10. WAIT FOR NEXT TICK                                  │
│     select { case <-ticker.C: ... }                     │
│     → Sleeps until next 10-second interval              │
└─────────────────────────────────────────────────────────┘
                         ↓
                    [REPEAT]
```

---

**Status**: Complete and Production-Ready  
**Last Updated**: February 8, 2026  
**Version**: 1.0.0
