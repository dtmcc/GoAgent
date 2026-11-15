package main

import (
	"GoAgent/consts"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	einoai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	godotenv.Load()

	tpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("扮演一个{role},回答要先给结论，再解释，适合初学者"),
		schema.UserMessage("{question}"),
	)
	cm, err := einoai.NewChatModel(ctx, &einoai.ChatModelConfig{
		Model:   consts.DeepSeekModelFlash,
		APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL: "https://api.deepseek.com/",
	})
	if err != nil {
		log.Fatal(err)
	}
	postProcess := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (output string, err error) {
		return fmt.Sprintf("【%s】\n %s", "学习教练", input.Content), nil
	})
	countProcess := compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) {
		return fmt.Sprintf("输出：%s，总输出长度 %d", input, len(input)), nil
	})

	chain := compose.NewChain[map[string]any, string]()
	chain.AppendChatTemplate(tpl).AppendChatModel(cm).AppendLambda(postProcess).AppendLambda(countProcess)

	runner, err := chain.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	result, err := runner.Invoke(ctx, map[string]any{"role": "面向Go学生的学习教练", "question": "go语言如何实现多线程,简要回答"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
