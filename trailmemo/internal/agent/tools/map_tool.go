package tools

import (
	"context"
	"encoding/json"
	"strings"
)

type POISearcher interface {
	SearchPOI(ctx context.Context, city string, keyword string, limit int) ([]POIItem, error)
}

type MapPOITool struct {
	source   string
	searcher POISearcher
}

type POIItem struct {
	Name      string  `json:"name"`
	City      string  `json:"city"`
	Address   string  `json:"address"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type poiSearchArgs struct {
	City    string `json:"city"`
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
}

func NewMapPOITool() *MapPOITool { return &MapPOITool{} }

func NewMapPOIToolWithSearcher(source string, searcher POISearcher) *MapPOITool {
	return &MapPOITool{source: strings.TrimSpace(source), searcher: searcher}
}

func (t *MapPOITool) Name() string { return "map.poi_search" }
func (t *MapPOITool) Description() string {
	return "搜索城市POI打卡点。可按城市和关键词筛选，返回名称、地址、分类和经纬度。"
}
func (t *MapPOITool) Permission() Permission { return PermissionRead }
func (t *MapPOITool) JSONSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"city":{"type":"string","description":"城市名"},"keyword":{"type":"string","description":"兴趣关键词，如美食/文化/夜景"},"limit":{"type":"integer","description":"最多返回条数"}}}`)
}

func (t *MapPOITool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params poiSearchArgs
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Success: false, Error: "参数解析失败"}, err
	}
	if params.Limit <= 0 || params.Limit > 10 {
		params.Limit = 5
	}

	if t.searcher != nil {
		items, err := t.searcher.SearchPOI(ctx, params.City, params.Keyword, params.Limit)
		if err == nil && len(items) > 0 {
			source := t.source
			if source == "" {
				source = "external"
			}
			return &ToolResult{Success: true, Data: map[string]interface{}{
				"pois":     items,
				"fallback": false,
				"source":   source,
			}}, nil
		}
		warning := "外部地图POI搜索未返回结果，已使用本地种子数据"
		if err != nil {
			warning = err.Error()
		}
		items = searchLocalPOIs(params.City, params.Keyword, params.Limit)
		return &ToolResult{Success: true, Data: map[string]interface{}{
			"pois":     items,
			"fallback": true,
			"source":   "local_seed",
			"warning":  warning,
		}}, nil
	}

	items := searchLocalPOIs(params.City, params.Keyword, params.Limit)
	return &ToolResult{Success: true, Data: map[string]interface{}{
		"pois":     items,
		"fallback": true,
		"source":   "local_seed",
	}}, nil
}

// searchLocalPOIs 搜索本地 POI 数据。
func searchLocalPOIs(city string, keyword string, limit int) []POIItem {
	city = strings.TrimSpace(city)
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	result := make([]POIItem, 0, limit)
	for _, item := range localPOISeeds {
		if city != "" && !strings.Contains(item.City, city) {
			continue
		}
		text := strings.ToLower(item.Name + item.Address + item.Category)
		if keyword != "" && !strings.Contains(text, keyword) {
			continue
		}
		result = append(result, item)
		if len(result) >= limit {
			return result
		}
	}
	if len(result) > 0 {
		return result
	}
	for _, item := range localPOISeeds {
		if city != "" && !strings.Contains(item.City, city) {
			continue
		}
		result = append(result, item)
		if len(result) >= limit {
			return result
		}
	}
	return result
}

var localPOISeeds = []POIItem{
	{Name: "西湖断桥", City: "杭州", Address: "杭州市西湖区北山街", Category: "景点 文化", Latitude: 30.259244, Longitude: 120.148934},
	{Name: "河坊街", City: "杭州", Address: "杭州市上城区河坊街", Category: "美食 文化", Latitude: 30.245123, Longitude: 120.171676},
	{Name: "龙井村", City: "杭州", Address: "杭州市西湖区龙井村", Category: "茶文化 美食", Latitude: 30.224192, Longitude: 120.10362},
	{Name: "宽窄巷子", City: "成都", Address: "成都市青羊区长顺街", Category: "美食 文化", Latitude: 30.669, Longitude: 104.059},
	{Name: "太古里", City: "成都", Address: "成都市锦江区中纱帽街", Category: "购物 夜景 美食", Latitude: 30.6535, Longitude: 104.0809},
	{Name: "沙面岛", City: "广州", Address: "广州市荔湾区沙面", Category: "文化 摄影", Latitude: 23.1095, Longitude: 113.2386},
	{Name: "永庆坊", City: "广州", Address: "广州市荔湾区恩宁路", Category: "美食 文化", Latitude: 23.1176, Longitude: 113.2452},
}
