package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Reporter handles sending device status to the collector API
type Reporter struct {
	collectorURL string
	httpClient   *http.Client
}

// NewReporter creates a new Reporter instance
func NewReporter(collectorURL string) *Reporter {
	return &Reporter{
		collectorURL: collectorURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendReport sends device status to the collector API
func (r *Reporter) SendReport(status *DeviceStatus) error {
	// Marshal the status to JSON
	jsonData, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal device status: %w", err)
	}

	// Create HTTP POST request
	req, err := http.NewRequest("POST", r.collectorURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "DevicePostureAgent/1.0")

	// Send the request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("collector API returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✓ Report sent successfully: %s\n", string(body))
	return nil
}

// SendReportWithRetry attempts to send the report with retry logic
func (r *Reporter) SendReportWithRetry(status *DeviceStatus, maxRetries int) error {
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.SendReport(status)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 2 * time.Second
			fmt.Printf("⚠ Failed to send report (attempt %d/%d): %v\n", attempt, maxRetries, err)
			fmt.Printf("  Retrying in %v...\n", waitTime)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
