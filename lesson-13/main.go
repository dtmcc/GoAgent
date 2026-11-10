package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
	message()
	template()
	messagesPlaceholder()
}

func messagesPlaceholder() {
	ctx := context.Background()

	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("扮演一个面向 Go 学习教练"),
		schema.MessagesPlaceholder("history", false),
		schema.UserMessage("{question}"),
	)
	history := []*schema.Message{
		schema.UserMessage("我刚学完 Gin 路由"),
		schema.AssistantMessage("接下来可以试着 GROM 做接口", nil),
	}
	msg, err := template.Format(ctx, map[string]any{
		"history":  history,
		"question": "三句话介 绍 channel",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range msg {
		fmt.Printf("[%s]%s\n", m.Role, m.Content)
	}
}

func message() {
	// assistant 发起的一次工具调用 没有正文。只有一个 ToolCall
	toolCall := schema.ToolCall{
		ID: "call_1",
		Function: schema.FunctionCall{
			Name:      "list_files",
			Arguments: `{"dir":"."}`,
		},
	}

	messages := []*schema.Message{
		schema.SystemMessage("扮演一个 Go 学习教练，可以在需要时调用工具"),
		schema.UserMessage("当前目录有哪些文件?"),

		// 模型决定调用工具
		schema.AssistantMessage("", []schema.ToolCall{toolCall}),

		// 工具结果回填
		schema.ToolMessage("file1.txt,file2.txt,file3.txt", "call_1"),

		// 模型拿到工具结果后 给出最终回复
		schema.AssistantMessage("当前目录有三个文件: file1.txt, file2.txt, file3.txt", nil),
	}

	for i, m := range messages {
		fmt.Printf("%d: %s\n", i+1, m.Role)
		if m.Content != "" {
			fmt.Printf("	content: %s\n", m.Content)
		}
		for _, call := range m.ToolCalls {
			fmt.Printf("	Tool Call: %s(%s)\n", call.Function.Name, call.Function.Arguments)
		}
		if m.ToolCallID != "" {
			fmt.Printf("	Tool Call ID: %s\n", m.ToolCallID)
		}
	}
}
func template() {
	ctx := context.Background()

	// {} 中英文模式下 不同 此处仅可以是英文模式下的半角花括号
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("扮演一个面向 Go 学生的{role}，水平：{level}。回答先给结论。"),
		schema.UserMessage("{question}"),
	)

	msg, err := template.Format(ctx, map[string]any{
		"role":     "学习教练",
		"level":    "中级",
		"question": "三句话介绍 channel",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range msg {
		fmt.Printf("[%s]%s\n", m.Role, m.Content)
	}
}
