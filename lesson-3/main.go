package main

import (
	consts "GoAgent"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()

	question := "请解释一下go中的interface 使用简洁语言，直接输出常规文本即可"
	chat, err := NewCoach()
	if err != nil {
		fmt.Println(err)
		return
	}
	answer, usage, err := chat.Ask(ctx, question)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("问题：", question, "答案：", answer)
	fmt.Println("Usage:", usage)
}

type Coach struct {
	client *openai.Client
	model  string
}

func NewCoach() (*Coach, error) {
	aipKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if aipKey == "" {
		return nil, errors.New("DEEPSEEK_API_KEY is empty")
	}

	cfg := openai.DefaultConfig(aipKey)
	cfg.BaseURL = "https://api.deepseek.com/"

	return &Coach{
		client: openai.NewClientWithConfig(cfg),
		model:  consts.DeepSeekModelFlash,
	}, nil
}

func (c *Coach) Ask(ctx context.Context, question string) (string, openai.Usage, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一个 Go 后端开发的老师，回答要适合初学者，先给出结论再解释原因",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
		MaxCompletionTokens: 1024,
		Temperature:         0.3,
	})
	if err != nil {
		return "", openai.Usage{}, err
	}
	if len(resp.Choices) == 0 {
		return "", openai.Usage{}, err
	}
	return resp.Choices[0].Message.Content, resp.Usage, err
}
