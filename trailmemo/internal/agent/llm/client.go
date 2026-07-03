package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// Client wraps an OpenAI-compatible chat completion provider.
type Client struct {
	cfg    config.LLMConfig
	client *openai.Client
}

// NewClient creates an OpenAI-compatible client using the agent-level LLM config
// (which falls back to the top-level LLM config if not explicitly set).
func NewClient(cfg config.LLMConfig) *Client {
	oc := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		oc.BaseURL = cfg.BaseURL
	}
	return &Client{
		cfg:    cfg,
		client: openai.NewClientWithConfig(oc),
	}
}

// Chat sends a non-streaming chat request and returns a structured result.
// maxTokens caps the total output; toolChoice can force function calling.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	start := time.Now()
	log := logger.FromContext(ctx).With(
		zap.String("model", c.cfg.Model),
		zap.String("provider", c.cfg.Provider),
	)

	messages := toOpenAIMessages(req.Messages) // 转换为OpenAI格式的消息
	temperature := float32(0.7)
	if req.Temperature > 0 {
		temperature = float32(req.Temperature)
	}
	chatReq := openai.ChatCompletionRequest{
		Model:       c.cfg.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   req.MaxTokens,
		Tools:       toOpenAITools(req.Tools),
	}
	// 如果请求指定了工具调用，添加到OpenAI的ChatCompletionRequest中
	if req.ToolChoice != "" {
		chatReq.ToolChoice = req.ToolChoice
	}
	// 如果请求指定了响应格式，添加到OpenAI的ChatCompletionRequest中
	if req.ResponseFormat != nil {
		chatReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}
	// 调用OpenAI的ChatCompletion API
	resp, err := c.client.CreateChatCompletion(ctx, chatReq)
	latency := time.Since(start) // 计算调用延迟

	if err != nil {
		log.Error("llm_call_failed", zap.Error(err), zap.Int64("latency_ms", latency.Milliseconds()))
		return &ChatResponse{Error: err, Latency: latency}, err
	}

	usage := Usage{
		PromptTokens:     resp.Usage.PromptTokens,     // 提示词令牌数
		CompletionTokens: resp.Usage.CompletionTokens, // 完成令牌数
		TotalTokens:      resp.Usage.TotalTokens,      // 总令牌数
	}

	if len(resp.Choices) == 0 {
		return &ChatResponse{Content: "", Usage: usage, Latency: latency}, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	toolCalls := fromOpenAIToolCalls(choice.Message.ToolCalls)

	log.Info("llm_call_success",
		zap.Int("prompt_tokens", usage.PromptTokens),
		zap.Int("completion_tokens", usage.CompletionTokens),
		zap.Int64("latency_ms", latency.Milliseconds()),
		zap.String("finish_reason", string(choice.FinishReason)),
	)

	return &ChatResponse{
		Content:   choice.Message.Content,
		ToolCalls: toolCalls,
		Usage:     usage,
		Latency:   latency,
	}, nil
}

// ParseJSONOutput calls Chat with JSON response format, then unmarshals into target.
// 该函数用于调用Chat方法，将OpenAI的ChatCompletion API的响应解析为JSON格式。
func (c *Client) ParseJSONOutput(ctx context.Context, req ChatRequest, target interface{}) error {
	req.ResponseFormat = &JSONFormat{Type: "json_object"}
	resp, err := c.Chat(ctx, req)
	if err != nil {
		return fmt.Errorf("llm call failed: %w", err)
	}
	if resp.Content == "" {
		return fmt.Errorf("empty response content")
	}
	return json.Unmarshal([]byte(resp.Content), target)
}

// toOpenAIMessages converts the internal Message type to the OpenAI ChatCompletionMessage type.
// 该函数用于将内部的Message类型转换为OpenAI的ChatCompletionMessage类型。
func toOpenAIMessages(msgs []Message) []openai.ChatCompletionMessage {
	out := make([]openai.ChatCompletionMessage, 0, len(msgs)) // 初始化输出切片，容量为输入切片的长度
	// 遍历输入切片，将每个Message转换为OpenAI的ChatCompletionMessage
	for _, m := range msgs {
		om := openai.ChatCompletionMessage{Role: m.Role, Content: m.Content} // 初始化OpenAI的ChatCompletionMessage
		// 如果Message包含工具调用ID，添加到OpenAI的ChatCompletionMessage中
		if m.ToolCallID != "" {
			om.ToolCallID = m.ToolCallID
		}
		// 如果Message包含工具调用，添加到OpenAI的ChatCompletionMessage中
		if len(m.ToolCalls) > 0 {
			om.ToolCalls = make([]openai.ToolCall, 0, len(m.ToolCalls)) // 初始化工具调用切片，容量为输入切片的长度
		}
		for _, tc := range m.ToolCalls {
			om.ToolCalls = append(om.ToolCalls, openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		out = append(out, om) // 将转换后的OpenAI的ChatCompletionMessage添加到输出切片中
	}
	return out
}

// toOpenAITools converts the internal ToolDef type to the OpenAI Tool type.
// 该函数用于将内部的ToolDef类型转换为OpenAI的Tool类型。
func toOpenAITools(tools []ToolDef) []openai.Tool {
	out := make([]openai.Tool, 0, len(tools))
	for _, t := range tools {
		out = append(out, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		})
	}
	return out
}

// fromOpenAIToolCalls converts the OpenAI ToolCall type to the internal ToolCall type.
// 该函数用于将OpenAI的ToolCall类型转换为内部的ToolCall类型。
func fromOpenAIToolCalls(tcs []openai.ToolCall) []ToolCall {
	out := make([]ToolCall, 0, len(tcs))
	for _, tc := range tcs {
		out = append(out, ToolCall{
			ID:   tc.ID,
			Type: string(tc.Type),
			Function: FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}
	return out
}
