# Week 2: Secure Web Gateway (SWG)

## 🎯 Project Overview

A **Transparent HTTP Proxy** that filters web traffic based on domain blocklists. This project demonstrates:

- **HTTP Proxy Server** (Go) - Intercepts and filters traffic
- **Policy Engine** (FastAPI) - Manages blocklists dynamically
- **Concurrency** - Handles multiple users simultaneously with goroutines
- **Middleware Pattern** - Request inspection and filtering

## 🏗️ Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   Browser   │────────>│   Go Proxy       │────────>│  Internet   │
│             │  :8080  │   (Filter)       │         │             │
└─────────────┘         └──────────────────┘         └─────────────┘
                               │  ^
                               │  │ GET /policy
                               │  │ (every 5 min)
                               v  │
                        ┌──────────────────┐
                        │  FastAPI Policy  │
                        │     Engine       │
                        │      :8000       │
                        └──────────────────┘
```

### Components

1. **Go HTTP Proxy** (`proxy/main.go`)
   - Listens on port 8080
   - Intercepts HTTP requests
   - Checks domain against blocklist (O(1) map lookup)
   - Blocks or forwards requests
   - Updates blocklist every 5 minutes

2. **FastAPI Policy Engine** (`policy-engine/main.py`)
   - Serves on port 8000
   - Provides `/policy` endpoint with blocklist
   - Allows dynamic add/remove of domains
   - In-memory storage (extendable to database)

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Python 3.9+
- pip

### 1. Start the Policy Engine

```bash
# Install dependencies
cd policy-engine
pip install -r requirements.txt

# Run the server
python main.py
```

The policy engine will start on `http://localhost:8000`

### 2. Start the Go Proxy

```bash
# In a new terminal
cd proxy
go run main.go
```

The proxy will start on `http://localhost:8080`

### 3. Configure Your Browser

**Option A: Firefox**
1. Settings → General → Network Settings → Settings
2. Select "Manual proxy configuration"
3. HTTP Proxy: `localhost`, Port: `8080`
4. Check "Use this proxy server for all protocols"

**Option B: Chrome/Edge (macOS)**
```bash
# Launch with proxy
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --proxy-server="localhost:8080"
```

**Option C: Using curl**
```bash
curl -x http://localhost:8080 http://facebook.com
curl -x http://localhost:8080 http://google.com
```

## 🧪 Testing the Proxy

### Test Blocked Sites

```bash
# Should return 403 Forbidden with custom HTML
curl -x http://localhost:8080 http://facebook.com
curl -x http://localhost:8080 http://tiktok.com
curl -x http://localhost:8080 http://youtube.com
```

### Test Allowed Sites

```bash
# Should return actual content
curl -x http://localhost:8080 http://google.com
curl -x http://localhost:8080 http://github.com
```

### View Current Policy

```bash
# See all blocked domains
curl http://localhost:8000/policy

# Response:
# {
#   "blocked": ["facebook.com", "tiktok.com", ...],
#   "total": 9
# }
```

### Dynamically Update Policy

```bash
# Add a domain to blocklist
curl -X POST "http://localhost:8000/policy/add?domain=linkedin.com"

# Remove a domain from blocklist
curl -X DELETE "http://localhost:8000/policy/remove?domain=youtube.com"

# List all domains
curl http://localhost:8000/policy/domains
```

## 📚 Key Go Concepts Demonstrated

### 1. HTTP Server & Custom Handlers

```go
type ProxyServer struct {
    blocklist map[string]bool
    // ...
}

func (ps *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Custom request handling
}
```

### 2. Goroutines for Concurrency

```go
// Handles multiple users simultaneously
go func() {
    ticker := time.NewTicker(interval)
    for range ticker.C {
        // Update blocklist in background
    }
}()
```

### 3. Thread-Safe Maps with Mutex

```go
type ProxyServer struct {
    blocklist      map[string]bool
    blocklistMutex sync.RWMutex  // Protects concurrent access
}

func (ps *ProxyServer) IsBlocked(host string) bool {
    ps.blocklistMutex.RLock()  // Read lock
    defer ps.blocklistMutex.RUnlock()
    return ps.blocklist[host]
}
```

### 4. HTTP Client for Forwarding

```go
client := &http.Client{
    Timeout: 30 * time.Second,
}
resp, err := client.Do(proxyReq)
```

## 🔍 How It Works

### Request Flow

1. **Browser → Proxy**: User requests `http://facebook.com`
2. **Proxy Inspection**: Extract host from `r.Host` → `"facebook.com"`
3. **Blocklist Lookup**: Check `blocklist["facebook.com"]` → O(1) map lookup
4. **Action**:
   - If **Blocked**: Return HTTP 403 with custom HTML
   - If **Allowed**: Forward request via `http.Client`, return response

### Blocklist Updates

1. **Startup**: Initial GET to `/policy` endpoint
2. **Every 5 minutes**: Background goroutine fetches updated policy
3. **Thread-safe**: Uses `sync.RWMutex` to prevent race conditions

### Domain Matching

The proxy implements **subdomain matching**:
- Blocking `facebook.com` also blocks `www.facebook.com`, `m.facebook.com`, etc.

```go
// Check parent domains
parts := strings.Split(domain, ".")
for i := 1; i < len(parts); i++ {
    parentDomain := strings.Join(parts[i:], ".")
    if ps.blocklist[parentDomain] {
        return true
    }
}
```

## 🎨 Features

- ✅ HTTP Proxy Server
- ✅ Domain Blocklist with O(1) lookup
- ✅ Custom 403 Forbidden page
- ✅ Request forwarding for allowed sites
- ✅ FastAPI policy engine
- ✅ Dynamic blocklist updates (every 5 minutes)
- ✅ Thread-safe concurrent access
- ✅ Subdomain matching
- ✅ RESTful policy management API

## 📊 API Endpoints

### Policy Engine (Port 8000)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Health check |
| GET | `/policy` | Get current blocklist |
| POST | `/policy/add?domain=X` | Add domain to blocklist |
| DELETE | `/policy/remove?domain=X` | Remove domain from blocklist |
| GET | `/policy/domains` | List all blocked domains |

## 🧩 Extending the Project

### Ideas for Enhancement

1. **HTTPS Support**: Add TLS termination (requires certificate generation)
2. **Category-Based Filtering**: Block by category (social, gambling, news)
3. **User Authentication**: Different policies per user
4. **Logging & Analytics**: Track blocked requests
5. **Allowlist**: Whitelist specific sites despite blocklist
6. **Time-Based Rules**: Block social media during work hours
7. **Database Backend**: Replace in-memory blocklist with PostgreSQL
8. **WebUI**: Admin dashboard for policy management

### Performance Optimizations

- **Caching**: Cache allowed/blocked decisions
- **Connection Pooling**: Reuse HTTP connections
- **Load Balancing**: Multiple proxy instances
- **Rate Limiting**: Prevent abuse

## 🐛 Troubleshooting

### Policy Engine Not Reachable

```bash
# Check if policy engine is running
curl http://localhost:8000/health

# Proxy logs will show:
# "Warning: Could not load initial blocklist"
```

### Browser Not Using Proxy

```bash
# Test with curl to isolate browser config
curl -x http://localhost:8080 http://example.com
```

### HTTPS Sites Not Working

This proxy only handles HTTP. For HTTPS, you need:
- TLS termination (MITM with certificates)
- Or implement CONNECT tunneling

## 📖 Learning Outcomes

After completing this project, you understand:

1. **HTTP Proxies**: How traffic interception works
2. **Goroutines**: Concurrent background tasks
3. **Mutex & Thread Safety**: Protecting shared state
4. **HTTP Handlers**: Custom request processing
5. **Client/Server Communication**: Service-to-service communication
6. **Map Data Structure**: O(1) lookups for filtering
7. **Middleware Pattern**: Request inspection pipeline

## 🔗 Related Concepts

- **Cisco Umbrella**: Cloud-based SWG
- **Squid Proxy**: Traditional proxy server
- **DNS Filtering**: Alternative to HTTP proxy
- **VPN**: Network-level tunneling
- **Firewall**: Packet-level filtering

## 📝 Notes

- This is a **learning project** - not production-ready
- No HTTPS/SSL interception (requires CA certificates)
- No authentication or authorization
- Blocklist is in-memory (resets on restart of policy engine)
- Simple forwarding (no caching, compression, etc.)

## 🎓 Next Steps

- Week 3: Add authentication to the proxy
- Week 4: Implement HTTPS support with certificate generation
- Week 5: Build a logging/analytics dashboard
