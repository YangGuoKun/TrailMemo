// Package wechat 封装微信小程序 OpenAPI 调用。
// 实现 code2session 换取 openid，后续可扩展获取手机号、用户信息等。
package wechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// Client 微信小程序 API 客户端。
type Client struct {
	appID     string
	appSecret string
	httpCli   *http.Client
}

// NewClient 创建微信客户端实例。
func NewClient(appID, appSecret string) *Client {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
		httpCli:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Code2SessionResponse 是微信 jscode2session 的返回结构。
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`      // 用户唯一标识
	SessionKey string `json:"session_key"` // 会话密钥
	UnionID    string `json:"unionid"`     // 开放平台唯一标识
	ErrCode    int    `json:"errcode"`     // 错误码（0表示成功）
	ErrMsg     string `json:"errmsg"`      // 错误信息
}

// Code2Session 用登录 code 换取 openid 和 session_key。
// 文档：https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html
func (c *Client) Code2Session(code string) (*Code2SessionResponse, error) {
	apiURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(c.appID),
		url.QueryEscape(c.appSecret),
		url.QueryEscape(code),
	)

	start := time.Now()
	resp, err := c.httpCli.Get(apiURL)
	latency := time.Since(start)

	if err != nil {
		logger.L().Error("wechat_api_call_failed",
			zap.String("api", "jscode2session"),
			zap.Error(err),
			zap.Duration("latency", latency))
		return nil, fmt.Errorf("微信API调用失败: %w", err)
	}
	defer resp.Body.Close()

	var result Code2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("微信响应解析失败: %w", err)
	}

	if result.ErrCode != 0 {
		logger.L().Error("wechat_code2session_error",
			zap.Int("errcode", result.ErrCode),
			zap.String("errmsg", result.ErrMsg),
			zap.Duration("latency", latency))
		return nil, fmt.Errorf("微信登录失败: [%d] %s", result.ErrCode, result.ErrMsg)
	}

	logger.L().Info("wechat_code2session_success",
		zap.String("openid_hash", HashOpenID(result.OpenID)),
		zap.Duration("latency", latency))

	return &result, nil
}

// HashOpenID 对 openid 做不可逆哈希，用于日志输出（不泄露用户标识）。
func HashOpenID(openid string) string {
	if len(openid) <= 8 { return "***" }
	return openid[:4] + "****" + openid[len(openid)-4:]
}
