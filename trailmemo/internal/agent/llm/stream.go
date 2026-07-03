package llm

import (
	"context"
	"errors"
	"io"

	"github.com/sashabaranov/go-openai"
)

// StreamToChannel starts a streaming chat completion and pushes delta chunks to ch.
// The channel is closed when streaming completes or fails.

func (c *Client) StreamToChannel(ctx context.Context, req ChatRequest, ch chan<- StreamChunk) {
	defer close(ch)

	// Explicitly request streaming from the OpenAI-compatible API
	reqCopy := req
	reqCopy.ResponseFormat = nil // streaming doesn't play well with JSON mode

	messages := toOpenAIMessages(reqCopy.Messages)
	streamReq := openai.ChatCompletionRequest{
		Model:       c.cfg.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   reqCopy.MaxTokens,
		Stream:      true,
	}
	// 调用OpenAI的ChatCompletion API，获取流式响应
	stream, err := c.client.CreateChatCompletionStream(ctx, streamReq)
	if err != nil {
		ch <- StreamChunk{Error: err}
		return
	}
	defer stream.Close() // 确保流在使用后关闭

	for {
		resp, err := stream.Recv() // 接收流式响应
		if errors.Is(err, io.EOF) {
			ch <- StreamChunk{Done: true}
			return
		}
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		if len(resp.Choices) == 0 {
			continue
		}

		delta := resp.Choices[0].Delta
		ch <- StreamChunk{
			Content:   delta.Content,
			ToolCalls: fromOpenAIToolCalls(delta.ToolCalls),
		}
	}
}
