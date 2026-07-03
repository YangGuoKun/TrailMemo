package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trailmemo/internal/agent/llm"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// Case is a single golden test case.
type Case struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"` // recommend / route / safety / fallback
	Input       llm.ChatRequest  `json:"input"`
	Checks      []Check          `json:"checks"`
}

// Check is one assertion to run against the LLM output.
type Check struct {
	Type    string `json:"type"`    // "contains" / "not_contains" / "json_valid" / "field_present" / "field_range"
	Field   string `json:"field"`   // json path
	Value   string `json:"value"`   // expected value or substring
	Min     int    `json:"min,omitempty"`
	Max     int    `json:"max,omitempty"`
}

// Report is a single evaluation report.
type Report struct {
	Total     int           `json:"total"`
	Passed    int           `json:"passed"`
	Failed    int           `json:"failed"`
	Duration  time.Duration `json:"duration"`
	CaseResults []CaseResult `json:"cases"`
}

type CaseResult struct {
	Name    string   `json:"name"`
	Passed  bool     `json:"passed"`
	Errors  []string `json:"errors,omitempty"`
}

// Runner executes eval cases against a live LLM client.
type Runner struct {
	client       *llm.Client
	cases        []Case
}

func NewRunner(client *llm.Client, cases []Case) *Runner {
	return &Runner{client: client, cases: cases}
}

// Run executes all cases and returns a report.
func (r *Runner) Run(ctx context.Context) *Report {
	report := &Report{Total: len(r.cases)}
	start := time.Now()

	for _, c := range r.cases {
		result := r.runCase(ctx, c)
		report.CaseResults = append(report.CaseResults, result)
		if result.Passed {
			report.Passed++
		} else {
			report.Failed++
		}
	}

	report.Duration = time.Since(start)
	logger.FromContext(ctx).Info("eval_completed",
		zap.Int("total", report.Total),
		zap.Int("passed", report.Passed),
		zap.Int("failed", report.Failed),
		zap.Duration("duration", report.Duration),
	)
	return report
}

func (r *Runner) runCase(ctx context.Context, c Case) CaseResult {
	result := CaseResult{Name: c.Name}

	resp, err := r.client.Chat(ctx, c.Input)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("llm error: %v", err))
		return result
	}

	for _, check := range c.Checks {
		if err := runCheck(resp.Content, check); err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	result.Passed = len(result.Errors) == 0
	return result
}

func runCheck(content string, c Check) error {
	switch c.Type {
	case "json_valid":
		if !json.Valid([]byte(content)) {
			return fmt.Errorf("expected valid JSON, got: %.100s", content)
		}
	case "contains":
		if !contains(content, c.Value) {
			return fmt.Errorf("expected output to contain %q", c.Value)
		}
	case "not_contains":
		if contains(content, c.Value) {
			return fmt.Errorf("expected output to NOT contain %q", c.Value)
		}
	default:
		return nil
	}
	return nil
}

func contains(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && (s == sub || len(s) >= len(sub) && findSub(s, sub))
}

func findSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
