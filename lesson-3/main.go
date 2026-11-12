package main

import (
	"GoAgent/consts"
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()

	// 单轮对话
	OneChat(ctx)

	// 多轮对话
	MultiChat(ctx)
}

type Coach struct {
	client *openai.Client
	model  string
}

func OneChat(ctx context.Context) {
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

func MultiChat(ctx context.Context) {
	chat, err := NewCoach()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = chat.ChatAsk(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func NewCoach() (*Coach, error) {
	godotenv.Load()

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

func (c *Coach) ChatAsk(ctx context.Context) error {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "你是一个 Go 后端开发的老师，回答要适合初学者，先给出结论再解释原因",
		},
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("开始对话（输入quit退出）。。。")
	for {
		fmt.Println("你：")
		if !scanner.Scan() {
			continue
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			return nil
		}
		if input == "" {
			continue
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})
		resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: messages,
		})
		if err != nil {
			fmt.Println(err)
			messages = messages[0 : len(messages)-1]
			continue
		}
		reply := resp.Choices[0].Message.Content
		fmt.Println("AI助手：", reply)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: reply,
		})
	}
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
		return "", openai.Usage{}, nil
	}
	return resp.Choices[0].Message.Content, resp.Usage, err
}
