package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAmapPOISearcherParsesPOIResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v5/place/text" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("key") != "test-key" {
			t.Fatalf("unexpected key: %s", query.Get("key"))
		}
		if query.Get("keywords") != "文化" {
			t.Fatalf("unexpected keywords: %s", query.Get("keywords"))
		}
		if query.Get("region") != "杭州" {
			t.Fatalf("unexpected region: %s", query.Get("region"))
		}
		if query.Get("city_limit") != "true" {
			t.Fatalf("expected city_limit=true, got %s", query.Get("city_limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status":"1",
			"infocode":"10000",
			"pois":[{
				"name":"杭州国家版本馆",
				"cityname":"杭州市",
				"address":"文润路1号",
				"type":"科教文化服务;文化宫",
				"location":"120.045123,30.293456"
			}]
		}`))
	}))
	defer server.Close()

	searcher := NewAmapPOISearcher("test-key")
	searcher.baseURL = server.URL + "/v5/place/text"

	items, err := searcher.SearchPOI(context.Background(), "杭州", "文化", 3)
	if err != nil {
		t.Fatalf("expected search success, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 POI, got %d", len(items))
	}
	item := items[0]
	if item.Name != "杭州国家版本馆" || item.City != "杭州市" || item.Address != "文润路1号" || item.Category != "科教文化服务;文化宫" {
		t.Fatalf("unexpected POI fields: %+v", item)
	}
	if item.Longitude != 120.045123 || item.Latitude != 30.293456 {
		t.Fatalf("unexpected location: %+v", item)
	}
}

func TestAmapPOISearcherReturnsErrorForAmapFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"0","infocode":"10001","info":"INVALID_USER_KEY"}`))
	}))
	defer server.Close()

	searcher := NewAmapPOISearcher("bad-key")
	searcher.baseURL = server.URL

	if _, err := searcher.SearchPOI(context.Background(), "杭州", "文化", 3); err == nil {
		t.Fatal("expected amap failure error")
	}
}
