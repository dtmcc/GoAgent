package main

import (
	"GoAgent/consts"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	einoai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	ctx := context.Background()
	godotenv.Load()

	cm, err := einoai.NewChatModel(ctx, &einoai.ChatModelConfig{
		Model:       consts.DeepSeekModelFlash,
		APIKey:      strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL:     "https://api.deepseek.com/",
		Temperature: ptr(float32(0.3)),
	})
	if err != nil {
		log.Fatal(err)
	}
	messages := []*schema.Message{
		schema.SystemMessage("你是一个 Go 后端开发的老师，回答要适合初学者，先给出结论再解释原因"),
		schema.UserMessage("请用三句话解释 Go 的 interface"),
	}

	// 入参其他参数
	//resp, err := cm.Generate(ctx, messages, model.WithTemperature(0.9))
	//resp, err := cm.Generate(ctx, messages, model.WithModel("other-model"))

	resp, err := cm.Generate(ctx, messages)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(resp.Content)
	if resp.ResponseMeta != nil && resp.ResponseMeta.Usage != nil {
		u := resp.ResponseMeta.Usage
		fmt.Printf("Total  Tokens: %d\nPrompt Tokens: %d\nCompletion Tokens: %d", u.TotalTokens, u.PromptTokens, u.CompletionTokens)
	}

	fmt.Println("stream输出 ======================")
	start := time.Now()
	stream, err := cm.Stream(ctx, messages)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer stream.Close()

	firstChunk := true
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if firstChunk {
			firstChunk = false
			fmt.Printf("\n【首token延迟】 %d \n", time.Since(start).Milliseconds())
		}
		fmt.Print(chunk.Content)
	}
	fmt.Printf("\n【总耗时】 %d \n", time.Since(start).Milliseconds())
}
