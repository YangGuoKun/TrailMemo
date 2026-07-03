package guardrail

import (
	"encoding/json"
	"fmt"

	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// ValidationResult is the outcome of validating an LLM output against a JSON schema.
type ValidationResult struct {
	Valid   bool     `json:"valid"`              // 是否有效
	Errors  []string `json:"errors,omitempty"`   // 错误信息列表
	RawJSON string   `json:"raw_json,omitempty"` // 原始JSON字符串
}

// Validator checks LLM outputs against expected JSON schemas.
type Validator struct {
	maxRetries int // 最大重试次数
}

// 创建一个新的JSON输出验证器（Validator）实例，用于检查LLM输出是否符合预期的JSON模式。
// 参数：maxRetries（最大重试次数）作为参数，返回一个新的Validator实例
func NewValidator(maxRetries int) *Validator {
	return &Validator{maxRetries: maxRetries}
}

// ValidateJSON ensures raw LLM output is parseable JSON matching the target schema.
// 作用：确保原始LLM输出是可解析的JSON，符合目标模式。
// 参数：原始LLM输出（raw）作为参数，返回验证结果（result）和错误（err）。
func (v *Validator) ValidateJSON(raw string) (*ValidationResult, error) {
	result := &ValidationResult{RawJSON: raw}

	// Strip markdown code fences often returned by models.
	raw = stripMarkdownFences(raw)

	if !json.Valid([]byte(raw)) {
		result.Errors = append(result.Errors, "invalid JSON")
		return result, fmt.Errorf("invalid JSON: %s", raw[:min(len(raw), 200)])
	}

	result.Valid = true
	return result, nil
}

// UnmarshalAndValidate parses raw JSON into target and runs schema checks.
// 作用：将原始JSON字符串解析到目标结构体（target）中，并运行模式检查。
// 参数：原始JSON字符串（raw）作为参数，返回验证结果（result）和错误（err）。
func (v *Validator) UnmarshalAndValidate(raw string, target interface{}) (*ValidationResult, error) {
	result, err := v.ValidateJSON(raw)
	if err != nil {
		return result, err
	}

	raw = stripMarkdownFences(raw)
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "unmarshal failed: "+err.Error())
		logger.L().Warn("guardrail_unmarshal_failed",
			zap.String("raw_preview", raw[:min(len(raw), 200)]),
			zap.Error(err),
		)
		return result, err
	}

	return result, nil
}

// RetryLimit returns how many retries the validator allows if output is invalid.
func (v *Validator) RetryLimit() int {
	return v.maxRetries
}

// stripMarkdownFences strips markdown code fences from the raw LLM output.
// 作用：从原始LLM输出中移除Markdown代码围栏（```json ... ```或``` ... ```）。
// 参数：原始LLM输出（raw）作为参数，返回处理后的字符串（s）。
func stripMarkdownFences(raw string) string {
	s := raw
	// Common patterns:
	// ```json ... ```
	// ``` ... ```
	for len(s) > 0 && (s[0] == '`' || s[0] == '\n') {
		idx := 0
		for idx < len(s) && (s[idx] == '`' || s[idx] == '\n' || s[idx] == 'j' || s[idx] == 's' || s[idx] == 'o' || s[idx] == 'n' || s[idx] == ' ') {
			if s[idx] == '\n' {
				s = s[idx+1:]
				idx = 0
				continue
			}
			idx++
		}
		break
	}
	// Strip trailing ```
	for len(s) > 0 && s[len(s)-1] == '`' {
		s = s[:len(s)-1]
	}
	s = trimRight(s, "\n ")
	return s
}

// trimRight trims characters from the end of a string.
// 作用：从字符串（s）的末尾移除字符（cutset）。
// 参数：字符串（s）作为参数，返回处理后的字符串（s）。
// 注意：此函数会修改原始字符串，而不是返回一个新的字符串。
func trimRight(s string, cutset string) string {
	for len(s) > 0 && contains(cutset, s[len(s)-1]) {
		s = s[:len(s)-1]
	}
	return s
}

// contains checks if a byte is in a string.
// 作用：检查字符（c）是否在字符串（set）中。
// 参数：字符（c）作为参数，返回布尔值（是否包含）。
func contains(set string, c byte) bool {
	for i := 0; i < len(set); i++ {
		if set[i] == c {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
