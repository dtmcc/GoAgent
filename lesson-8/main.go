package main

import (
	"GoAgent/DeepSeek"
	"GoAgent/consts"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Plan struct {
	Goal  string `json:"goal"`
	Steps []Step `json:"steps"`
}

type Step struct {
	ID     int    `json:"id"`     // 步骤ID
	Action string `json:"action"` // 步骤行为
	Why    string `json:"why"`    // 为什么做这一步
	Output string `json:"output"` // 做这一步的产出是什么
}

const planterPrompt = `你是一个本机 Go 仓库分析任务的 Planner。把用户给的分析目标拆成 4 到 6 个可执行步骤。

每个步骤必须包含：
1. action：具体做什么。允许用的动作是：扫描目录、读取 README/go.mod、读取入口文件、读取 router/handLer、读取 service、或者其他
2. why：为什么这一步现在做（要确认什么事实）
3. output：这一步做完会有什么产出（目录结构 /文件清单/一条结论/一份草稿报告）

第一步必须是扫描目录，最后一步必须是汇总。中间步骤围绕信息缺口安排。
严格按 JSON 输出，不要用 markdown 代码块包裹：
{"goal":"...","steps":[{"id":1,"action":"...","why":"","output":"..."},{"id":2,"action":"...","why":"","output":"..."}]}
`

func main() {
	goal := "请彻底分析我本机的一个 PHP  项目（路径 /Users/lianghao/PhpStorm/js.dtm.cc ），覆盖技术栈、目录分层、启动链路、API入口、业务链路等 "
	client := DeepSeek.NewClient()
	plans, err := makePlan(client, goal)
	if err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}
	fmt.Println("目标：", plans.Goal)
	for _, step := range plans.Steps {
		fmt.Println("\n步骤：", step.ID, "\n行为：", step.Action, "\n原因：", step.Why, "\n产出：", step.Output)
		fmt.Println()
	}
}

func makePlan(client *openai.Client, goal string) (Plan, error) {
	var plan Plan
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: consts.DeepSeekModelFlash,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: planterPrompt},
			{Role: openai.ChatMessageRoleUser, Content: goal},
		},
		Stream: false,
	})
	if err != nil {
		return plan, err
	}
	if len(resp.Choices) == 0 {
		fmt.Println("no choices error: ", err)
		return plan, err
	}

	content := resp.Choices[0].Message.Content
	// 此处可能需要 去除输出的 空格 反引号等特殊字符
	content = strings.TrimSpace(content)
	err = json.Unmarshal([]byte(content), &plan)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return plan, err
	}
	return plan, nil
}
