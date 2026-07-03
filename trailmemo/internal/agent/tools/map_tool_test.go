package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type fakePOISearcher struct {
	items []POIItem
	err   error
}

func (f fakePOISearcher) SearchPOI(ctx context.Context, city, keyword string, limit int) ([]POIItem, error) {
	return f.items, f.err
}

func TestMapPOIToolSearchesByCityAndKeyword(t *testing.T) {
	tool := NewMapPOITool()

	result, err := tool.Execute(context.Background(), json.RawMessage(`{"city":"杭州","keyword":"美食","limit":3}`))
	if err != nil {
		t.Fatalf("expected search success, got %v", err)
	}
	if !result.Success {
		t.Fatalf("expected successful result: %+v", result)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected data type: %T", result.Data)
	}
	items, ok := data["pois"].([]POIItem)
	if !ok {
		t.Fatalf("unexpected pois type: %T", data["pois"])
	}
	if len(items) == 0 {
		t.Fatal("expected at least one POI")
	}
	if items[0].City != "杭州" {
		t.Fatalf("expected city 杭州, got %s", items[0].City)
	}
}

func TestMapPOIToolReturnsExternalPOIsWhenSearcherSucceeds(t *testing.T) {
	tool := NewMapPOIToolWithSearcher("amap", fakePOISearcher{items: []POIItem{{
		Name: "杭州国家版本馆", City: "杭州", Address: "文润路", Category: "文化场馆", Latitude: 30.293, Longitude: 120.045,
	}}})

	result, err := tool.Execute(context.Background(), json.RawMessage(`{"city":"杭州","keyword":"文化","limit":3}`))
	if err != nil {
		t.Fatalf("expected search success, got %v", err)
	}
	if !result.Success {
		t.Fatalf("expected successful result: %+v", result)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected data type: %T", result.Data)
	}
	if data["source"] != "amap" {
		t.Fatalf("expected amap source, got %v", data["source"])
	}
	if data["fallback"] != false {
		t.Fatalf("expected fallback false, got %v", data["fallback"])
	}
	items, ok := data["pois"].([]POIItem)
	if !ok {
		t.Fatalf("unexpected pois type: %T", data["pois"])
	}
	if len(items) != 1 || items[0].Name != "杭州国家版本馆" {
		t.Fatalf("unexpected pois: %+v", items)
	}
}

func TestMapPOIToolFallsBackToLocalPOIsWhenSearcherFails(t *testing.T) {
	tool := NewMapPOIToolWithSearcher("amap", fakePOISearcher{err: errors.New("amap unavailable")})

	result, err := tool.Execute(context.Background(), json.RawMessage(`{"city":"杭州","keyword":"美食","limit":3}`))
	if err != nil {
		t.Fatalf("expected fallback success, got %v", err)
	}
	if !result.Success {
		t.Fatalf("expected successful fallback result: %+v", result)
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected data type: %T", result.Data)
	}
	if data["source"] != "local_seed" {
		t.Fatalf("expected local_seed source, got %v", data["source"])
	}
	if data["fallback"] != true {
		t.Fatalf("expected fallback true, got %v", data["fallback"])
	}
	if data["warning"] == "" {
		t.Fatalf("expected fallback warning, got %v", data["warning"])
	}
	items, ok := data["pois"].([]POIItem)
	if !ok {
		t.Fatalf("unexpected pois type: %T", data["pois"])
	}
	if len(items) == 0 || items[0].City != "杭州" {
		t.Fatalf("unexpected fallback pois: %+v", items)
	}
}
