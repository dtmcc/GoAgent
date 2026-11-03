package main

import (
	consts "GoAgent"

	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

const weakPrompt = `你是一个代码审查专家，请回答用户问题`
const structuredPrompt = `你是一个面向本机代码仓库的分析助手。

回答要求：
1. 先给结论，再展开细节
2. 涉及代码的地方贴关键片段，并说明这段代码在做什么
3. 看不准的地方明确写“看不准”，附上需要再确定的文件路径
4. 报告结构固定：概述 / 技术栈 / 目录分层 / 调用链路 / 业务清单 / 改进建议`

func main() {
	client, err := NewClient()
	if err != nil {
		panic(err)
	}

	question := `我有一个项目放在 /Users/lianghao/PhpStorm/js.dtm.cc 目录，希望你帮我梳理
1. 这是什么类型的项目，用了哪些技术栈
2. 一次Http请求 从入口到数据库 大概经过哪几层
3. 我现在要接收维护，大概从哪个目录，哪个文件开始阅读？`

	fmt.Println("====== 弱 prompt ======")
	answer1, err := Chat(client, weakPrompt, question)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(answer1)
	fmt.Println(" ")
	answer2, err := Chat(client, structuredPrompt, question)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(answer2)
}

func NewClient() (*openai.Client, error) {
	godotenv.Load()

	aipKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if aipKey == "" {
		return nil, errors.New("DEEPSEEK_API_KEY is empty")
	}

	cfg := openai.DefaultConfig(aipKey)
	cfg.BaseURL = "https://api.deepseek.com/"

	return openai.NewClientWithConfig(cfg), nil
}

func Chat(client *openai.Client, prompt, question string) (string, error) {
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: consts.DeepSeekModelFlash,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", err
	}
	return resp.Choices[0].Message.Content, err
}
