package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultAmapPOIURL = "https://restapi.amap.com/v5/place/text"

type AmapPOISearcher struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewAmapPOISearcher 创建一个新的高德地图 POI 搜索器。
// api_key 是高德地图 API 密钥，用于身份验证。
// 返回一个指向 AmapPOISearcher 结构体的指针。
// 如果 api_key 为空，返回 nil。
func NewAmapPOISearcher(apiKey string) *AmapPOISearcher {
	return &AmapPOISearcher{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: defaultAmapPOIURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// SearchPOI 搜索高德地图上的 POI 数据。
// 支持指定城市、关键词和返回数量的限制。
// 返回通用 POI 项的列表。
func (s *AmapPOISearcher) SearchPOI(ctx context.Context, city string, keyword string, limit int) ([]POIItem, error) {
	if s == nil || s.apiKey == "" {
		return nil, fmt.Errorf("高德地图API key未配置")
	}
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	endpoint, err := url.Parse(s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("高德POI接口地址无效: %w", err)
	}
	query := endpoint.Query()
	query.Set("key", s.apiKey)
	query.Set("keywords", strings.TrimSpace(keyword))
	query.Set("region", strings.TrimSpace(city))
	query.Set("city_limit", "true")
	query.Set("page_size", strconv.Itoa(limit))
	query.Set("show_fields", "business")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建高德POI请求失败: %w", err)
	}
	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("高德POI请求失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("高德POI请求HTTP状态异常: %d", resp.StatusCode)
	}

	var payload amapPOIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("解析高德POI响应失败: %w", err)
	}
	if payload.Status != "1" {
		info := strings.TrimSpace(payload.Info)
		if info == "" {
			info = payload.Infocode
		}
		return nil, fmt.Errorf("高德POI搜索失败: %s", info)
	}

	items := make([]POIItem, 0, len(payload.POIs))
	for _, poi := range payload.POIs {
		item, ok := poi.toPOIItem()
		if !ok {
			continue
		}
		items = append(items, item)
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}

type amapPOIResponse struct {
	Status   string        `json:"status"`
	Info     string        `json:"info"`
	Infocode string        `json:"infocode"`
	POIs     []amapPOIItem `json:"pois"`
}

type amapPOIItem struct {
	Name     string      `json:"name"`
	CityName string      `json:"cityname"`
	Address  stringValue `json:"address"`
	Type     string      `json:"type"`
	Location string      `json:"location"`
}

// toPOIItem 将高德 POI 项转换为通用 POI 项。
func (p amapPOIItem) toPOIItem() (POIItem, bool) {
	lng, lat, ok := parseAmapLocation(p.Location)
	if !ok {
		return POIItem{}, false
	}
	return POIItem{
		Name:      p.Name,
		City:      p.CityName,
		Address:   string(p.Address),
		Category:  p.Type,
		Latitude:  lat,
		Longitude: lng,
	}, true
}

// parseAmapLocation 解析高德 POI 项中的位置字符串。
func parseAmapLocation(location string) (float64, float64, bool) {
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return 0, 0, false
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, false
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, false
	}
	return lng, lat, true
}

type stringValue string

// UnmarshalJSON 实现 stringValue 类型的 JSON 反序列化。
// 支持字符串和字符串数组的 JSON 输入，将它们转换为 stringValue 类型的值。
func (s *stringValue) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*s = stringValue(text)
		return nil
	}
	var texts []string
	if err := json.Unmarshal(data, &texts); err == nil {
		*s = stringValue(strings.Join(texts, " "))
		return nil
	}
	*s = ""
	return nil
}
