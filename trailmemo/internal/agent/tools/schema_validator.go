package tools

import (
	"encoding/json"
	"fmt"
	"math"
)

type toolJSONSchema struct {
	Type       string                        `json:"type"`
	Properties map[string]toolJSONSchemaProp `json:"properties"`
	Required   []string                      `json:"required"`
}

type toolJSONSchemaProp struct {
	Type string `json:"type"`
}

func validateToolArgs(schema json.RawMessage, args json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}
	var parsed toolJSONSchema
	if err := json.Unmarshal(schema, &parsed); err != nil {
		return fmt.Errorf("工具参数schema无效: %w", err)
	}
	if parsed.Type != "" && parsed.Type != "object" {
		return fmt.Errorf("工具参数schema仅支持object类型")
	}

	var values map[string]interface{}
	if err := json.Unmarshal(args, &values); err != nil {
		return fmt.Errorf("工具参数不是有效JSON对象: %w", err)
	}
	for _, name := range parsed.Required {
		if _, ok := values[name]; !ok {
			return fmt.Errorf("缺少必填参数: %s", name)
		}
	}
	for name, value := range values {
		prop, ok := parsed.Properties[name]
		if !ok || prop.Type == "" || value == nil {
			continue
		}
		if !matchesJSONSchemaType(value, prop.Type) {
			return fmt.Errorf("参数%s类型不匹配，期望%s", name, prop.Type)
		}
	}
	return nil
}

func matchesJSONSchemaType(value interface{}, typ string) bool {
	switch typ {
	case "string":
		_, ok := value.(string)
		return ok
	case "integer":
		number, ok := value.(float64)
		return ok && math.Trunc(number) == number
	case "number":
		_, ok := value.(float64)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true
	}
}
