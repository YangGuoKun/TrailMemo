package tools

import (
	"context"
	"encoding/json"
	"testing"
)

type schemaTestTool struct {
	called bool
}

func (t *schemaTestTool) Name() string        { return "schema.test" }
func (t *schemaTestTool) Description() string { return "schema validation test tool" }
func (t *schemaTestTool) Permission() Permission {
	return PermissionRead
}
func (t *schemaTestTool) JSONSchema() json.RawMessage {
	return json.RawMessage(`{
		"type":"object",
		"properties":{
			"route_id":{"type":"integer"},
			"keyword":{"type":"string"},
			"public":{"type":"boolean"}
		},
		"required":["route_id"]
	}`)
}
func (t *schemaTestTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	t.called = true
	return &ToolResult{Success: true, Data: "ok"}, nil
}

func TestRegistryRejectsMissingRequiredToolArg(t *testing.T) {
	reg := NewRegistry()
	tool := &schemaTestTool{}
	reg.Register(tool)

	result, err := reg.Execute(context.Background(), "schema.test", json.RawMessage(`{"keyword":"杭州"}`))
	if err != nil {
		t.Fatalf("expected validation result without hard error, got %v", err)
	}
	if result.Success {
		t.Fatalf("expected validation failure, got %+v", result)
	}
	if tool.called {
		t.Fatal("tool should not execute when required args are missing")
	}
}

func TestRegistryRejectsWrongToolArgType(t *testing.T) {
	reg := NewRegistry()
	tool := &schemaTestTool{}
	reg.Register(tool)

	result, err := reg.Execute(context.Background(), "schema.test", json.RawMessage(`{"route_id":"123"}`))
	if err != nil {
		t.Fatalf("expected validation result without hard error, got %v", err)
	}
	if result.Success {
		t.Fatalf("expected validation failure, got %+v", result)
	}
	if tool.called {
		t.Fatal("tool should not execute when arg type is invalid")
	}
}

func TestRegistryExecutesToolWhenArgsMatchSchema(t *testing.T) {
	reg := NewRegistry()
	tool := &schemaTestTool{}
	reg.Register(tool)

	result, err := reg.Execute(context.Background(), "schema.test", json.RawMessage(`{"route_id":123,"keyword":"杭州","public":true}`))
	if err != nil {
		t.Fatalf("expected execution success, got %v", err)
	}
	if !result.Success {
		t.Fatalf("expected successful tool result, got %+v", result)
	}
	if !tool.called {
		t.Fatal("tool should execute when args match schema")
	}
}
