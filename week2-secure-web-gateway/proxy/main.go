package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// PolicyResponse represents the response from the policy engine
type PolicyResponse struct {
	Blocked []string `json:"blocked"`
}

// ProxyServer handles HTTP proxy requests with domain blocking
type ProxyServer struct {
	blocklist      map[string]bool
	blocklistMutex sync.RWMutex
	policyURL      string
}

// NewProxyServer creates a new proxy server instance
func NewProxyServer(policyURL string) *ProxyServer {
	return &ProxyServer{
		blocklist: make(map[string]bool),
		policyURL: policyURL,
	}
}

// UpdateBlocklist fetches the blocklist from the policy engine
func (ps *ProxyServer) UpdateBlocklist() error {
	resp, err := http.Get(ps.policyURL)
	if err != nil {
		return fmt.Errorf("failed to fetch policy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("policy engine returned status: %d", resp.StatusCode)
	}

	var policy PolicyResponse
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return fmt.Errorf("failed to decode policy: %w", err)
	}

	// Update the blocklist with write lock
	ps.blocklistMutex.Lock()
	defer ps.blocklistMutex.Unlock()

	// Clear and rebuild the blocklist
	ps.blocklist = make(map[string]bool)
	for _, domain := range policy.Blocked {
		ps.blocklist[strings.ToLower(domain)] = true
		log.Printf("Blocked domain: %s", domain)
	}

	log.Printf("Blocklist updated: %d domains blocked", len(ps.blocklist))
	return nil
}

// StartPeriodicUpdate starts a goroutine that updates the blocklist periodically
func (ps *ProxyServer) StartPeriodicUpdate(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("Updating blocklist from policy engine...")
			if err := ps.UpdateBlocklist(); err != nil {
				log.Printf("Error updating blocklist: %v", err)
			}
		}
	}()
}

// IsBlocked checks if a domain is in the blocklist
func (ps *ProxyServer) IsBlocked(host string) bool {
	ps.blocklistMutex.RLock()
	defer ps.blocklistMutex.RUnlock()

	// Remove port if present
	domain := strings.Split(host, ":")[0]
	domain = strings.ToLower(domain)

	// Check exact match
	if ps.blocklist[domain] {
		return true
	}

	// Check if any parent domain is blocked (e.g., www.facebook.com matches facebook.com)
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		parentDomain := strings.Join(parts[i:], ".")
		if ps.blocklist[parentDomain] {
			return true
		}
	}

	return false
}

// ServeHTTP handles incoming proxy requests
func (ps *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}

	log.Printf("Request: %s %s from %s", r.Method, host, r.RemoteAddr)

	// Check if the domain is blocked
	if ps.IsBlocked(host) {
		log.Printf("BLOCKED: %s", host)
		ps.serveBlockedPage(w, host)
		return
	}

	// Allow the request - forward it to the actual destination
	log.Printf("ALLOWED: %s", host)
	ps.forwardRequest(w, r)
}

// serveBlockedPage returns a 403 Forbidden page
func (ps *ProxyServer) serveBlockedPage(w http.ResponseWriter, host string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Access Denied</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.3);
            text-align: center;
            max-width: 500px;
        }
        h1 {
            color: #e74c3c;
            margin-top: 0;
        }
        .blocked-icon {
            font-size: 72px;
            color: #e74c3c;
        }
        .domain {
            background: #f8f9fa;
            padding: 10px;
            border-radius: 5px;
            margin: 20px 0;
            font-family: monospace;
            word-break: break-all;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="blocked-icon">🚫</div>
        <h1>Access Denied by Cisco Security</h1>
        <p>The website you are trying to access has been blocked by your organization's security policy.</p>
        <div class="domain">%s</div>
        <p><small>If you believe this is an error, please contact your IT administrator.</small></p>
    </div>
</body>
</html>`, host)

	fmt.Fprint(w, html)
}

// forwardRequest forwards the request to the actual destination
func (ps *ProxyServer) forwardRequest(w http.ResponseWriter, r *http.Request) {
	// Build the target URL
	targetURL := r.URL.String()
	if !strings.HasPrefix(targetURL, "http") {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		targetURL = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path)
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}
	}

	// Create a new request
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		log.Printf("Error creating request: %v", err)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Execute the request
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Error forwarding request", http.StatusBadGateway)
		log.Printf("Error forwarding request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	// Configuration
	proxyPort := "8080"
	policyURL := "http://localhost:8000/policy"
	updateInterval := 5 * time.Minute

	log.Println("=== Cisco Secure Web Gateway ===")
	log.Printf("Starting proxy server on port %s", proxyPort)
	log.Printf("Policy engine: %s", policyURL)
	log.Printf("Blocklist update interval: %v", updateInterval)

	// Create proxy server
	proxy := NewProxyServer(policyURL)

	// Initial blocklist load
	log.Println("Loading initial blocklist...")
	if err := proxy.UpdateBlocklist(); err != nil {
		log.Printf("Warning: Could not load initial blocklist: %v", err)
		log.Println("Proxy will retry in background. Using empty blocklist for now.")
	}

	// Start periodic updates
	proxy.StartPeriodicUpdate(updateInterval)

	// Start the HTTP server
	server := &http.Server{
		Addr:         ":" + proxyPort,
		Handler:      proxy,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Proxy server listening on http://localhost:%s", proxyPort)
	log.Println("Configure your browser to use this proxy")
	log.Fatal(server.ListenAndServe())
}
