# Week 3 — Zero Trust Access (ZTNA)

> **"Never Trust, Always Verify."**

A Zero Trust Enforcer gateway (Go) + Identity Provider (FastAPI/Python) demonstrating JWT-based authentication, middleware chaining, context timeouts, and role-based access control.

## Architecture

```
                ┌──────────────┐         ┌──────────────────────┐
  curl /login → │  FastAPI IdP │ → JWT   │                      │
                │  (port 9000) │         │   Go Gateway         │
                └──────────────┘         │   (port 8080)        │
                                         │                      │
  curl /hr  ────────────────────────────→│ AccessLogger         │
  + Bearer token                         │   → RequestTimeout   │
                                         │     → JWTAuth        │
                                         │       → HR Handler   │
                                         └──────────────────────┘
```

## Quick Start

### 1. Set the shared secret

```bash
export JWT_SECRET="super-secret-key-change-me-in-prod"
```

### 2. Start the Identity Provider (Python)

```bash
cd idp
pip install -r requirements.txt
uvicorn main:app --port 9000
```

### 3. Start the Go Gateway

```bash
cd gateway
go run main.go
```

### 4. Login and get a token

```bash
# Admin user (will be allowed through)
curl -s -X POST http://localhost:9000/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq

# Guest user (will be denied by gateway)
curl -s -X POST http://localhost:9000/login \
  -H "Content-Type: application/json" \
  -d '{"username":"guest","password":"guest123"}' | jq
```

### 5. Access the protected resource

```bash
# Replace <TOKEN> with the access_token from step 4
curl -s http://localhost:8080/hr -H "Authorization: Bearer <TOKEN>" | jq
```

## Expected Results

| User  | Role  | Gateway Response         |
|-------|-------|--------------------------|
| admin | admin | 200 — "Hello from HR"    |
| guest | guest | 403 — access denied      |
| none  | —     | 401 — missing/invalid    |

## Go Concepts Covered

- **Middleware chaining** — `Chain()` composes AccessLogger → RequestTimeout → JWTAuth
- **`context` package** — `context.WithTimeout` enforces 5s request deadlines; user claims propagated via context values
- **JWT validation** — HMAC-SHA256 signature verification, claims extraction, role enforcement

## Files

```
week3-ztna/
├── gateway/
│   ├── main.go              # Entry point, route setup, middleware chain
│   ├── middleware/
│   │   └── auth.go          # JWTAuth, AccessLogger, RequestTimeout, Chain
│   ├── handlers/
│   │   └── hr.go            # Protected HR resource handler
│   ├── go.mod
│   └── go.sum
├── idp/
│   ├── main.py              # FastAPI /login endpoint, JWT issuance
│   └── requirements.txt
└── README.md
```
