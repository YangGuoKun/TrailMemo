package runtime

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/trailmemo/internal/agent/tools"
)

type schemaTool struct{}

func (schemaTool) Name() string        { return "route.search_public" }
func (schemaTool) Description() string { return "search public routes" }
func (schemaTool) Permission() tools.Permission {
	return tools.PermissionRead
}
func (schemaTool) JSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"},"limit":{"type":"integer"}},"required":["city"]}`)
}
func (schemaTool) Execute(context.Context, json.RawMessage) (*tools.ToolResult, error) {
	return &tools.ToolResult{Success: true}, nil
}

func TestConvertToolDefsUsesRegisteredToolSchema(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(schemaTool{})

	defs := convertToolDefs(reg)
	if len(defs) != 1 {
		t.Fatalf("expected 1 tool def, got %d", len(defs))
	}

	got := string(defs[0].Function.Parameters)
	want := `{"type":"object","properties":{"city":{"type":"string"},"limit":{"type":"integer"}},"required":["city"]}`
	if got != want {
		t.Fatalf("expected tool schema %s, got %s", want, got)
	}
}
