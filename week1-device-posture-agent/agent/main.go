package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultCollectorURL = "http://localhost:8000/report"
	defaultInterval     = 10 * time.Second
	maxRetries          = 3
)

func main() {
	// Command-line flags
	collectorURL := flag.String("url", defaultCollectorURL, "Collector API URL")
	interval := flag.Duration("interval", defaultInterval, "Report interval (e.g., 10s, 1m)")
	dryRun := flag.Bool("dry-run", false, "Collect data but don't send to API (print to console)")
	flag.Parse()

	// Print banner
	printBanner()

	// Initialize components
	collector := NewSystemCollector()
	reporter := NewReporter(*collectorURL)

	// Create a ticker for periodic execution
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("🚀 Device Posture Agent started\n")
	fmt.Printf("   Collector URL: %s\n", *collectorURL)
	fmt.Printf("   Report Interval: %v\n", *interval)
	fmt.Printf("   Dry Run Mode: %v\n", *dryRun)
	fmt.Printf("   Press Ctrl+C to stop\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Initial collection and report
	collectAndReport(collector, reporter, *dryRun)

	// Main loop
	for {
		select {
		case <-ticker.C:
			collectAndReport(collector, reporter, *dryRun)

		case sig := <-sigChan:
			fmt.Printf("\n📪 Received signal: %v\n", sig)
			fmt.Println("🛑 Shutting down gracefully...")
			return
		}
	}
}

// collectAndReport collects device status and sends it to the collector API
func collectAndReport(collector *SystemCollector, reporter *Reporter, dryRun bool) {
	fmt.Printf("\n[%s] Collecting device status...\n", time.Now().Format("2006-01-02 15:04:05"))

	// Collect device status
	status, err := collector.CollectDeviceStatus()
	if err != nil {
		log.Printf("❌ Error collecting device status: %v\n", err)
		return
	}

	// Print collected data
	printDeviceStatus(status)

	// Send report (or print if dry-run)
	if dryRun {
		fmt.Println("\n🔍 DRY RUN MODE - JSON Payload:")
		jsonData, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		if err := reporter.SendReportWithRetry(status, maxRetries); err != nil {
			log.Printf("❌ Failed to send report: %v\n", err)
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// printDeviceStatus prints the device status in a formatted way
func printDeviceStatus(status *DeviceStatus) {
	statusIcon := "✓"
	if status.Status == StatusUnhealthy {
		statusIcon = "⚠"
	}

	fmt.Printf("  %s Status: %s\n", statusIcon, status.Status)
	fmt.Printf("  📍 Hostname: %s\n", status.Hostname)
	fmt.Printf("  🌐 IP Address: %s\n", status.IP)
	fmt.Printf("  💾 Disk Usage: %.2f%%\n", status.DiskUsage)
	if status.Message != "" {
		fmt.Printf("  💬 Message: %s\n", status.Message)
	}
}

// printBanner prints a nice banner
func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════╗
║     🛡️  DEVICE POSTURE AGENT v1.0 🛡️            ║
║     Cisco Secure Client - Training Edition       ║
╚═══════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
