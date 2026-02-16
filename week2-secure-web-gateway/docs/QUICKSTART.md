# Quick Start Guide

## 5-Minute Setup

### Step 1: Install Dependencies (1 min)

```bash
cd week2-secure-web-gateway
make install
```

### Step 2: Start Policy Engine (1 min)

Open Terminal 1:
```bash
make policy
```

You should see:
```
Starting Cisco SWG Policy Engine...
Loaded 9 blocked domains
INFO:     Uvicorn running on http://0.0.0.0:8000
```

### Step 3: Start Proxy Server (1 min)

Open Terminal 2:
```bash
make proxy
```

You should see:
```
=== Cisco Secure Web Gateway ===
Starting proxy server on port 8080
Policy engine: http://localhost:8000/policy
Blocklist updated: 9 domains blocked
Proxy server listening on http://localhost:8080
```

### Step 4: Test It! (2 min)

Open Terminal 3:

```bash
# Test blocked site (should return 403)
curl -x http://localhost:8080 http://facebook.com

# Test allowed site (should work)
curl -x http://localhost:8080 http://google.com
```

## Using the Makefile

### Run All Tests
```bash
make test
```

### Test Specific Categories
```bash
make test-blocked    # Test blocked domains
make test-allowed    # Test allowed domains
make test-dynamic    # Test adding/removing domains
```

### View Logs
```bash
make logs
```

## Manual Testing

### 1. Check What's Blocked

```bash
curl http://localhost:8000/policy
```

Output:
```json
{
  "blocked": [
    "facebook.com",
    "tiktok.com",
    "twitter.com",
    "instagram.com",
    "reddit.com",
    "youtube.com",
    "gambling.com",
    "bet365.com",
    "pokerstars.com"
  ],
  "total": 9
}
```

### 2. Try Blocked Sites

```bash
# Should get custom 403 page
curl -x http://localhost:8080 http://facebook.com
curl -x http://localhost:8080 http://youtube.com
```

### 3. Try Allowed Sites

```bash
# Should work normally
curl -x http://localhost:8080 http://google.com
curl -x http://localhost:8080 http://github.com
curl -x http://localhost:8080 http://stackoverflow.com
```

### 4. Add Domain to Blocklist

```bash
# Block LinkedIn
curl -X POST "http://localhost:8000/policy/add?domain=linkedin.com"

# Wait 5 minutes or restart proxy for it to update
# Or for immediate testing, restart the proxy server
```

### 5. Remove Domain from Blocklist

```bash
# Unblock YouTube
curl -X DELETE "http://localhost:8000/policy/remove?domain=youtube.com"
```

## Browser Configuration

### Firefox

1. Open Firefox
2. Settings → General → Network Settings → Settings
3. Select "Manual proxy configuration"
4. HTTP Proxy: `localhost`
5. Port: `8080`
6. Check "Use this proxy server for all protocols"
7. Click OK

Now browse to `http://facebook.com` - you should see the blocked page!

### Chrome (via command line)

```bash
# macOS
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
  --proxy-server="localhost:8080" \
  --new-window

# Windows
"C:\Program Files\Google\Chrome\Application\chrome.exe" ^
  --proxy-server="localhost:8080"

# Linux
google-chrome --proxy-server="localhost:8080"
```

## Expected Behavior

### When Accessing Blocked Site

1. **Request**: User tries to access `http://facebook.com`
2. **Proxy Logs**: `Request: GET facebook.com from 127.0.0.1:xxxxx`
3. **Proxy Logs**: `BLOCKED: facebook.com`
4. **Response**: Custom HTML page with "Access Denied by Cisco Security"

### When Accessing Allowed Site

1. **Request**: User tries to access `http://google.com`
2. **Proxy Logs**: `Request: GET google.com from 127.0.0.1:xxxxx`
3. **Proxy Logs**: `ALLOWED: google.com`
4. **Response**: Actual Google homepage

### Blocklist Updates

Every 5 minutes, you'll see in proxy logs:
```
Updating blocklist from policy engine...
Blocklist updated: 9 domains blocked
```

## Troubleshooting

### Policy Engine Won't Start

```bash
# Check if port 8000 is in use
lsof -i :8000

# Kill existing process
kill -9 <PID>

# Try again
make policy
```

### Proxy Won't Start

```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill existing process
kill -9 <PID>

# Try again
make proxy
```

### "Could not load initial blocklist"

This means the proxy can't reach the policy engine. Make sure:

1. Policy engine is running (`make policy`)
2. Policy engine is on port 8000
3. Check: `curl http://localhost:8000/health`

### Browser Not Using Proxy

1. Check proxy settings in browser
2. Try with curl first to verify proxy works
3. Make sure you're using `http://` not `https://` URLs

### No Response from Proxy

```bash
# Check if proxy is running
curl http://localhost:8080

# Should get error (proxy doesn't handle direct requests)
# But if connection refused, proxy isn't running
```

## What to Observe

### In Proxy Logs

```
Request: GET facebook.com from 127.0.0.1:50234
BLOCKED: facebook.com
Request: GET google.com from 127.0.0.1:50235
ALLOWED: google.com
```

### In Policy Engine Logs

```
Policy requested - returning 9 blocked domains
Added domain to blocklist: linkedin.com
Removed domain from blocklist: youtube.com
```

## Next Steps

Once everything works:

1. ✅ Understand the code flow (see main README)
2. ✅ Try adding your own domains
3. ✅ Modify the blocked page HTML
4. ✅ Experiment with different blocklist update intervals
5. ✅ Try the stretch goals (see README)

## Common Use Cases

### Temporary Testing

```bash
# Terminal 1
cd week2-secure-web-gateway/policy-engine
python main.py

# Terminal 2
cd week2-secure-web-gateway/proxy
go run main.go

# Terminal 3
curl -x http://localhost:8080 http://facebook.com
```

### Long Running

```bash
# Use Makefile
make all       # Start both services
make test      # Run tests
make stop      # Stop services
```

## Performance Testing

```bash
# Test multiple requests
for i in {1..10}; do
  curl -x http://localhost:8080 http://google.com -o /dev/null -s -w "Request $i: %{http_code}\n"
done
```

## API Testing

```bash
# View all blocked domains
curl http://localhost:8000/policy/domains | jq

# Health check
curl http://localhost:8000/health

# Add domain
curl -X POST "http://localhost:8000/policy/add?domain=netflix.com"

# Remove domain  
curl -X DELETE "http://localhost:8000/policy/remove?domain=netflix.com"
```

---

**Happy filtering! 🚀**

For detailed documentation, see [README.md](README.md)
