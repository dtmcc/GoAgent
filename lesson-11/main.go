package main

import (
	"GoAgent/consts"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	einoai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	godotenv.Load()

	cm, err := einoai.NewChatModel(ctx, &einoai.ChatModelConfig{
		Model:   consts.DeepSeekModelFlash,
		APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL: "https://api.deepseek.com/",
	})
	if err != nil {
		log.Fatal(err)
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "go_demo",
		Description:   "回答 Go 学习问题，必要时调用课程和代码诊断工具。",
		Instruction:   "扮演一个面向 Go 学生的学习教练，回答先给结论再解释。",
		Model:         cm,
		MaxIterations: 8,
	})
	if err != nil {
		log.Fatal(err)
	}
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	iter := runner.Query(ctx, "Go 语言中如何实现一个简单的 HTTP 服务器？")

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatal(err)
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		msg, err := event.Output.MessageOutput.GetMessage()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg)
	}
}
