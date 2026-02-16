# Week 2 Project Summary: Secure Web Gateway

## 📋 Project Overview

**Project Name:** Secure Web Gateway (SWG)  
**Theme:** Building a Traffic Filter Proxy  
**Primary Goal:** Understand Concurrency and Middleware in Go  
**Difficulty:** Intermediate  
**Technologies:** Go, Python (FastAPI), HTTP, REST APIs

## 🎯 What We Built

A complete HTTP proxy system that:
1. Intercepts web traffic
2. Filters requests based on a dynamic blocklist
3. Blocks forbidden domains with custom pages
4. Forwards allowed requests to the internet
5. Updates policies from a centralized engine

## 🏗️ System Architecture

### High-Level Design

```
┌──────────────┐
│   Browser    │  User makes request
│   (Client)   │
└──────┬───────┘
       │ HTTP Request
       │
       v
┌──────────────────────────────────────┐
│     Go HTTP Proxy Server             │
│     (Port 8080)                      │
│                                      │
│  ┌────────────────────────────────┐ │
│  │  1. Receive Request            │ │
│  │  2. Extract Host               │ │
│  │  3. Check Blocklist (O(1))     │ │
│  │  4. Block or Forward           │ │
│  └────────────────────────────────┘ │
└──────────────┬───────────────────────┘
               │
     ┌─────────┴─────────┐
     │                   │
     v                   v
┌─────────────┐    ┌─────────────┐
│  Blocked?   │    │  Allowed?   │
│  Return 403 │    │  Forward to │
│  Custom Page│    │  Internet   │
└─────────────┘    └─────────────┘
     ^
     │ Policy Updates
     │ (every 5 min)
     │
┌────────────────────────┐
│  FastAPI Policy Engine │
│  (Port 8000)           │
│                        │
│  GET /policy           │
│  POST /policy/add      │
│  DELETE /policy/remove │
└────────────────────────┘
```

### Component Breakdown

#### 1. Go Proxy Server (`proxy/main.go`)

**Responsibilities:**
- Listen for incoming HTTP requests
- Parse request host/domain
- Check against blocklist
- Block or forward requests
- Periodic blocklist updates

**Key Structures:**

```go
type ProxyServer struct {
    blocklist      map[string]bool    // O(1) lookup
    blocklistMutex sync.RWMutex       // Thread safety
    policyURL      string             // Policy engine endpoint
}
```

**Core Methods:**
- `ServeHTTP()` - Main request handler
- `IsBlocked()` - Check if domain is blocked
- `UpdateBlocklist()` - Fetch policy from FastAPI
- `StartPeriodicUpdate()` - Background goroutine for updates
- `forwardRequest()` - Proxy allowed requests
- `serveBlockedPage()` - Return 403 HTML

#### 2. FastAPI Policy Engine (`policy-engine/main.py`)

**Responsibilities:**
- Store blocklist in memory
- Serve blocklist via REST API
- Allow dynamic policy updates
- Provide admin endpoints

**Key Endpoints:**

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/policy` | Return current blocklist |
| POST | `/policy/add?domain=X` | Add domain to blocklist |
| DELETE | `/policy/remove?domain=X` | Remove domain |
| GET | `/policy/domains` | List all blocked domains |
| GET | `/health` | Health check |

## 🔑 Key Go Concepts Demonstrated

### 1. HTTP Server & Custom Handlers

```go
// ProxyServer implements http.Handler interface
func (ps *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Custom request handling logic
}

// Start server
server := &http.Server{
    Addr:    ":8080",
    Handler: proxy,
}
server.ListenAndServe()
```

**Learning Points:**
- `http.Handler` interface
- `http.Server` configuration
- Request/Response writing
- Status codes and headers

### 2. Goroutines for Concurrency

```go
// Each request handled in its own goroutine (automatic)
// Manual goroutine for background tasks
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        ps.UpdateBlocklist()
    }
}()
```

**Learning Points:**
- Automatic goroutines per request
- Manual goroutines with `go` keyword
- Time tickers for periodic tasks
- Background processing

### 3. Maps for O(1) Lookup

```go
type ProxyServer struct {
    blocklist map[string]bool  // Key: domain, Value: true
}

// O(1) lookup
if ps.blocklist[domain] {
    // Domain is blocked
}
```

**Learning Points:**
- Map data structure
- O(1) lookup performance
- Boolean flags as values
- Case-insensitive matching

### 4. Mutex for Thread Safety

```go
type ProxyServer struct {
    blocklist      map[string]bool
    blocklistMutex sync.RWMutex  // Protects blocklist
}

// Reading (many readers allowed)
func (ps *ProxyServer) IsBlocked(host string) bool {
    ps.blocklistMutex.RLock()      // Read lock
    defer ps.blocklistMutex.RUnlock()
    return ps.blocklist[host]
}

// Writing (exclusive access)
func (ps *ProxyServer) UpdateBlocklist() {
    ps.blocklistMutex.Lock()        // Write lock
    defer ps.blocklistMutex.Unlock()
    ps.blocklist = newBlocklist
}
```

**Learning Points:**
- Race conditions in concurrent code
- `sync.RWMutex` for reader/writer locks
- `RLock()` for read-only access
- `Lock()` for write access
- `defer` for automatic unlocking

### 5. HTTP Client for Forwarding

```go
// Create client with timeout
client := &http.Client{
    Timeout: 30 * time.Second,
}

// Forward request
resp, err := client.Do(proxyReq)
if err != nil {
    http.Error(w, "Error forwarding", http.StatusBadGateway)
    return
}

// Copy response back
io.Copy(w, resp.Body)
```

**Learning Points:**
- HTTP client creation
- Timeout configuration
- Request forwarding
- Response streaming with `io.Copy`

### 6. JSON Encoding/Decoding

```go
type PolicyResponse struct {
    Blocked []string `json:"blocked"`
}

// Decode JSON response
var policy PolicyResponse
err := json.NewDecoder(resp.Body).Decode(&policy)
```

**Learning Points:**
- Struct tags for JSON mapping
- `json.Decoder` for streaming
- Error handling for parsing

## 🔄 Request Flow

### Scenario 1: Blocked Domain

```
1. User → Proxy: GET http://facebook.com
2. Proxy: Extract host → "facebook.com"
3. Proxy: Check blocklist → blocklist["facebook.com"] = true
4. Proxy: Generate 403 HTML page
5. Proxy → User: HTTP 403 Forbidden + Custom HTML
```

**Log Output:**
```
Request: GET facebook.com from 127.0.0.1:50234
BLOCKED: facebook.com
```

### Scenario 2: Allowed Domain

```
1. User → Proxy: GET http://google.com
2. Proxy: Extract host → "google.com"
3. Proxy: Check blocklist → blocklist["google.com"] = false
4. Proxy → Internet: Forward GET http://google.com
5. Internet → Proxy: HTTP 200 + HTML content
6. Proxy → User: Forward response
```

**Log Output:**
```
Request: GET google.com from 127.0.0.1:50235
ALLOWED: google.com
```

### Scenario 3: Blocklist Update

```
1. Timer: 5 minutes elapsed
2. Goroutine: Wake up
3. Proxy → Policy Engine: GET http://localhost:8000/policy
4. Policy Engine → Proxy: {"blocked": ["facebook.com", ...]}
5. Proxy: Parse JSON
6. Proxy: Acquire write lock
7. Proxy: Update blocklist map
8. Proxy: Release write lock
9. Proxy: Back to sleep for 5 minutes
```

**Log Output:**
```
Updating blocklist from policy engine...
Blocked domain: facebook.com
Blocked domain: tiktok.com
...
Blocklist updated: 9 domains blocked
```

## 🎨 Features Implemented

### Core Features

- ✅ HTTP Proxy Server on port 8080
- ✅ Domain blocklist with O(1) lookup
- ✅ Custom 403 Forbidden HTML page
- ✅ Request forwarding for allowed sites
- ✅ FastAPI policy engine on port 8000
- ✅ Periodic policy updates (5 minutes)
- ✅ Thread-safe concurrent access
- ✅ Subdomain matching
- ✅ RESTful policy management

### Advanced Features

- ✅ Graceful error handling
- ✅ Structured logging
- ✅ Configurable timeouts
- ✅ Case-insensitive domain matching
- ✅ Port stripping from hosts
- ✅ Dynamic policy updates via API
- ✅ Health check endpoints

## 🧪 Testing Strategy

### 1. Service Health Checks

```bash
curl http://localhost:8000/health  # Policy engine
curl http://localhost:8080         # Proxy (expect error)
```

### 2. Blocked Domain Tests

```bash
curl -x http://localhost:8080 http://facebook.com  # Expect 403
curl -x http://localhost:8080 http://tiktok.com    # Expect 403
```

### 3. Allowed Domain Tests

```bash
curl -x http://localhost:8080 http://google.com   # Expect 200/301
curl -x http://localhost:8080 http://github.com   # Expect 200/301
```

### 4. Dynamic Policy Tests

```bash
# Add domain
curl -X POST "http://localhost:8000/policy/add?domain=test.com"

# Verify
curl http://localhost:8000/policy | grep test.com

# Remove domain
curl -X DELETE "http://localhost:8000/policy/remove?domain=test.com"
```

### 5. Subdomain Tests

```bash
curl -x http://localhost:8080 http://www.facebook.com  # Also blocked
curl -x http://localhost:8080 http://m.facebook.com    # Also blocked
```

## 📊 Performance Characteristics

### Blocklist Lookup

- **Time Complexity:** O(1) - Hash map lookup
- **Space Complexity:** O(n) - n = number of blocked domains
- **Thread Safety:** Multiple concurrent readers, single writer

### Request Handling

- **Concurrency:** Unlimited concurrent connections (goroutine per request)
- **Throughput:** Limited by network and CPU
- **Latency:** 
  - Blocked: ~1ms (map lookup + HTML generation)
  - Allowed: Network latency + upstream response time

### Memory Usage

- **Blocklist:** ~100 bytes per domain
- **Goroutines:** ~2KB stack per request
- **Total:** Minimal for typical workloads

## 🔒 Security Considerations

### Current Implementation

- ❌ No authentication/authorization
- ❌ No HTTPS support (plaintext only)
- ❌ No request validation
- ❌ No rate limiting
- ❌ In-memory storage (volatile)

### Production Requirements Would Need

1. **Authentication:** API keys, JWT tokens
2. **HTTPS:** TLS interception with certificates
3. **Authorization:** Role-based access control
4. **Logging:** Audit trail of all requests
5. **Rate Limiting:** Prevent abuse
6. **Persistence:** Database for blocklist
7. **Monitoring:** Metrics and alerting
8. **Input Validation:** Sanitize domain names

## 🚀 Deployment Considerations

### Development

```bash
# Terminal 1
cd policy-engine && python main.py

# Terminal 2
cd proxy && go run main.go
```

### Production (Conceptual)

```yaml
# Docker Compose example
services:
  policy-engine:
    image: swg-policy:latest
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql://...
  
  proxy:
    image: swg-proxy:latest
    ports:
      - "8080:8080"
    environment:
      - POLICY_URL=http://policy-engine:8000/policy
    depends_on:
      - policy-engine
    replicas: 3  # Load balancing
```

## 📈 Potential Enhancements

### Short-term

1. **Caching:** Cache policy locally to reduce API calls
2. **Metrics:** Prometheus metrics for monitoring
3. **Logging:** Structured logging (JSON) for log aggregation
4. **Configuration:** Environment variables for ports, URLs
5. **Health Checks:** Kubernetes-ready liveness/readiness probes

### Medium-term

1. **HTTPS Support:** TLS termination and certificate management
2. **Database:** PostgreSQL for persistent blocklist
3. **Categories:** Group domains by category (social, gambling, etc.)
4. **User Auth:** Different policies per user/group
5. **Admin UI:** Web dashboard for policy management

### Long-term

1. **ML Integration:** Automatic threat detection
2. **Content Inspection:** Deep packet inspection
3. **Load Balancing:** Multiple proxy instances
4. **Cloud Deployment:** AWS/GCP/Azure
5. **API Gateway:** Full-featured API management

## 🎓 Learning Outcomes

After completing this project, you understand:

### Go Programming

- HTTP server implementation
- Custom handler interfaces
- Goroutines and concurrency
- Mutex and thread safety
- Maps and data structures
- HTTP client usage
- JSON encoding/decoding
- Error handling patterns

### Network Programming

- HTTP proxy mechanics
- Request/response forwarding
- Status codes and headers
- Port and protocol handling
- Client/server communication

### Software Architecture

- Microservices design
- REST API design
- Policy engine pattern
- Middleware concept
- Service communication
- Background processing

### DevOps & Testing

- Service orchestration
- Testing strategies
- Logging and debugging
- Makefile automation
- Deployment considerations

## 🔗 Real-World Applications

This proxy architecture is used in:

- **Cisco Umbrella:** Cloud-delivered security
- **Zscaler:** Zero trust security platform
- **Squid Proxy:** Web caching and filtering
- **Corporate Firewalls:** Content filtering
- **Parental Controls:** Home network filtering
- **CDN Edge Nodes:** Content delivery and filtering

## 📚 Code Statistics

- **Total Lines:** ~550 (Go + Python)
  - Go: ~280 lines
  - Python: ~150 lines
  - Documentation: ~1200+ lines

- **Files:** 11
  - Source: 4 (2 Go, 2 Python config)
  - Documentation: 4
  - Configuration: 3

- **Key Functions:** 10
  - Go: 7 methods
  - Python: 6 endpoints

## 🎯 Success Metrics

### What "Done" Looks Like

- ✅ Both services start without errors
- ✅ Proxy blocks all default domains
- ✅ Proxy allows non-blocked domains
- ✅ Policy updates automatically every 5 minutes
- ✅ Can add/remove domains dynamically
- ✅ All tests pass
- ✅ Comprehensive documentation

### Skills Acquired

- ✅ Building HTTP servers in Go
- ✅ Concurrent programming with goroutines
- ✅ Thread-safe data structures
- ✅ REST API design and consumption
- ✅ HTTP client/server patterns
- ✅ Microservice architecture
- ✅ Testing network services

## 📝 Next Steps

1. **Week 3:** Add authentication and authorization
2. **Week 4:** Implement HTTPS with TLS
3. **Week 5:** Build analytics dashboard
4. **Week 6:** Database integration
5. **Week 7:** Deploy to cloud (AWS/GCP)

---

## 📖 Related Documentation

- [Main README](../README.md) - Project overview
- [Documentation](README.md) - Detailed guide
- [Quick Start](QUICKSTART.md) - 5-minute setup

---

**Project Status:** ✅ Complete and fully functional  
**Difficulty:** ⭐⭐⭐ Intermediate  
**Time to Complete:** 3-4 hours (with documentation)  
**Prerequisites:** Basic Go knowledge, HTTP understanding
