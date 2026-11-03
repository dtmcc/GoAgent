package main

import (
	consts "GoAgent"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	question := "go中的channel 如何理解"
	stream, err := NewChatStream()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = stream.StreamAsk(context.Background(), question)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type ChatStream struct {
	client *openai.Client
	model  string
}

func NewChatStream() (*ChatStream, error) {
	godotenv.Load()

	aipKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if aipKey == "" {
		return nil, errors.New("DEEPSEEK_API_KEY is empty")
	}

	cfg := openai.DefaultConfig(aipKey)
	cfg.BaseURL = "https://api.deepseek.com/"

	return &ChatStream{
		client: openai.NewClientWithConfig(cfg),
		model:  consts.DeepSeekModelFlash,
	}, nil
}

func (s *ChatStream) StreamAsk(ctx context.Context, question string) error {
	stream, err := s.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: s.model,
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
		Stream: true,
	})
	if err != nil {
		return nil
	}
	defer stream.Close()

	fmt.Println("------------ 回复：")

	var (
		full         strings.Builder
		firstTokenAt time.Time
		startedAt    = time.Now()
	)

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		content := chunk.Choices[0].Delta.Content
		if content == "" {
			continue
		}
		if firstTokenAt.IsZero() {
			firstTokenAt = time.Now()
		}
		fmt.Println(content)
		full.WriteString(content)
	}
	fmt.Println("----------------------")
	if !firstTokenAt.IsZero() {
		fmt.Printf("首个token输出耗时%s", firstTokenAt.Sub(startedAt).Round(time.Millisecond))
	}

	fmt.Printf("总耗时%s", time.Since(startedAt).Round(time.Millisecond))
	fmt.Println(full.String())

	return nil
}
