# Go Learning Journey

A comprehensive collection of Go projects focusing on network security and systems programming.

## 📚 Projects

### Week 1: Device Posture Agent

**Theme:** Building a Host Security Monitor  
**Goal:** Understand structs, JSON, and system programming

An agent that collects system metrics and reports them to a central server.

- ✅ System metrics collection (CPU, memory, disk, network)
- ✅ HTTP reporting to backend server
- ✅ FastAPI analytics dashboard
- ✅ Real-time monitoring

📁 [View Project →](week1-device-posture-agent/)

### Week 2: Secure Web Gateway (SWG)

**Theme:** Building a Traffic Filter Proxy  
**Goal:** Understand concurrency and middleware

An HTTP proxy that filters web traffic based on dynamic blocklists.

- ✅ HTTP proxy server with domain blocking
- ✅ FastAPI policy engine for blocklist management
- ✅ Concurrent request handling with goroutines
- ✅ Thread-safe map operations with mutex
- ✅ Dynamic policy updates

📁 [View Project →](week2-secure-web-gateway/)

### Week 3: Zero Trust Access (ZTNA)

**Theme:** "Never Trust, Always Verify"  
**Goal:** Authentication, JWTs, and tying the Control Plane to the Data Plane

A Zero Trust Enforcer gateway that validates JWT tokens and enforces role-based access to protected resources.

- ✅ JWT validation middleware (HMAC-SHA256 signature verification)
- ✅ Middleware chaining (AccessLogger → RequestTimeout → JWTAuth)
- ✅ Context-based request timeouts and claim propagation
- ✅ Role-based access control (admin allowed, guest denied)
- ✅ FastAPI Identity Provider for token issuance
- ✅ Structured access logging (user, timestamp, success/fail)

📁 [View Project →](week3-ztna/)

## 🚀 Quick Start

### Prerequisites

```bash
# Install Go
brew install go
go version

# Install Python 3.9+
brew install python3
python3 --version
```

### Running Projects

Each project has its own README with detailed instructions:

```bash
# Week 1: Device Posture Agent
cd week1-device-posture-agent
make help

# Week 2: Secure Web Gateway
cd week2-secure-web-gateway
make help

# Week 3: Zero Trust Access
cd week3-ztna
cat README.md
```

## 🎓 Learning Path

| Week | Project | Key Concepts | Difficulty |
|------|---------|--------------|------------|
| 1 | Device Posture Agent | Structs, JSON, HTTP Client | ⭐⭐ Beginner |
| 2 | Secure Web Gateway | Concurrency, Middleware, Maps | ⭐⭐⭐ Intermediate |
| 3 | Zero Trust Access | Middleware Chaining, Context, JWTs | ⭐⭐⭐ Intermediate |

## 📖 Resources

- [Go Documentation](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)

## 🛠️ Development Tools

```bash
go run .           # Run the program
go test ./...      # Run tests
go build           # Build binary
gofmt -w .         # Format code
go mod tidy        # Clean dependencies
```

## 📝 Assignment 1

Basic Go exercises including factorial and fibonacci calculations.

📁 [View Assignment →](assignment-1/)
