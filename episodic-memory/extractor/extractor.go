package extractor

import (
	"strings"

	"episodic-memory/memory"
)

// CorrectionInput is what a researcher submits when correcting an agent output.
type CorrectionInput struct {
	// Raw is the free-text correction, e.g.
	//   "If domain contains '.ca', check for Canadian data residency requirements"
	Raw string `json:"raw"`

	// Optional structured hints — the extractor uses these if provided,
	// otherwise it infers from the raw text.
	Domain   string            `json:"domain,omitempty"`
	Task     string            `json:"task,omitempty"`
	Source   string            `json:"source,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Extractor turns free-text corrections into structured BehaviorRules.
type Extractor struct{}

func New() *Extractor { return &Extractor{} }

// Extract parses a CorrectionInput into a BehaviorRule ready for storage.
// It splits the raw text into a condition (pattern) and an action using
// common natural-language delimiters ("then", "should", "must", "->", etc.).
func (e *Extractor) Extract(input CorrectionInput) memory.BehaviorRule {
	pattern, action := splitConditionAction(input.Raw)

	rule := memory.BehaviorRule{
		ID:       memory.NewRuleID(),
		Pattern:  strings.TrimSpace(pattern),
		Action:   strings.TrimSpace(action),
		Domain:   input.Domain,
		Task:     input.Task,
		Source:   input.Source,
		Tags:     input.Tags,
		Metadata: input.Metadata,
	}

	if rule.Domain == "" {
		rule.Domain = inferDomain(input.Raw)
	}
	if rule.Task == "" {
		rule.Task = inferTask(input.Raw)
	}

	return rule
}

// splitConditionAction splits "If X, then Y" or "When X, do Y" style text.
var delimiters = []string{
	" then ",
	" -> ",
	" should ",
	" must ",
	", check ",
	", verify ",
	", ensure ",
	", require ",
	", flag ",
	", alert ",
	", block ",
	", allow ",
}

func splitConditionAction(raw string) (string, string) {
	lower := strings.ToLower(raw)
	for _, d := range delimiters {
		idx := strings.Index(lower, d)
		if idx > 0 {
			return raw[:idx], raw[idx+len(d):]
		}
	}
	return raw, raw
}

var domainKeywords = map[string]string{
	".ca":          "canadian-compliance",
	"canada":       "canadian-compliance",
	"gdpr":         "eu-privacy",
	"hipaa":        "healthcare",
	"pci":          "payment-security",
	"aws":          "cloud-aws",
	"azure":        "cloud-azure",
	"gcp":          "cloud-gcp",
	"kubernetes":   "container-orchestration",
	"docker":       "containers",
	"ssl":          "tls-certificates",
	"tls":          "tls-certificates",
	"certificate":  "tls-certificates",
	"dns":          "network-dns",
	"subdomain":    "network-dns",
	"firewall":     "network-security",
	"authentication": "identity",
	"oauth":        "identity",
	"jwt":          "identity",
	"xss":          "web-security",
	"injection":    "web-security",
	"ransomware":   "threat-response",
	"malware":      "threat-response",
	"phishing":     "social-engineering",
}

func inferDomain(text string) string {
	lower := strings.ToLower(text)
	for keyword, domain := range domainKeywords {
		if strings.Contains(lower, keyword) {
			return domain
		}
	}
	return "general"
}

var taskKeywords = map[string]string{
	"subdomain":     "subdomain-discovery",
	"enumerat":      "enumeration",
	"recon":         "reconnaissance",
	"scan":          "scanning",
	"vulnerab":      "vulnerability-assessment",
	"pentest":       "penetration-testing",
	"compliance":    "compliance-check",
	"audit":         "audit",
	"monitor":       "monitoring",
	"incident":      "incident-response",
	"config":        "configuration-review",
	"permission":    "access-review",
	"data residen":  "data-residency",
	"vendor":        "vendor-assessment",
	"risk":          "risk-assessment",
	"deploy":        "deployment",
}

func inferTask(text string) string {
	lower := strings.ToLower(text)
	for keyword, task := range taskKeywords {
		if strings.Contains(lower, keyword) {
			return task
		}
	}
	return "general"
}
