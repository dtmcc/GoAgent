package main

import (
	"GoAgent/DeepSeek"
	"GoAgent/consts"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

func main() {
	client := DeepSeek.NewClient()

	questions := []string{
		"上海今天天气怎么样？",
		"纽约现在几点？",
		"计算一下 3 * 4 等于多少？",
		"Go 语言怎么样？",
		"北京今天天气怎么样？",
	}

	tools := buildTools()
	for i, question := range questions {
		fmt.Println("==================================")
		fmt.Printf("Question %d: %s\n", i+1, question)
		runOnce(client, tools, question)
		fmt.Println()
	}
}

func runOnce(client *openai.Client, tools []openai.Tool, question string) {
	messages := make([]openai.ChatCompletionMessage, 0, 2)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "扮演一个助手，根据用户的提问选择合适的工具，如果没有工具 直接回答",
	}, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       consts.DeepSeekModelFlash,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.3,
	})
	if err != nil {
		fmt.Println("Error:", err)
	}

	msg := resp.Choices[0].Message
	messages = append(messages, msg)
	if len(msg.ToolCalls) == 0 {
		fmt.Println("Assistant:", msg.Content)
		return
	}
	if len(msg.ToolCalls) > 0 {
		for _, call := range msg.ToolCalls {
			fmt.Println("tool Call:", call.Function.Name, call.Function.Arguments)
			toolResp, err := dispatch(call.Function.Name, call.Function.Arguments)
			if err != nil {
				fmt.Println("tool Error:", err)
			}
			fmt.Println("tool Response:", toolResp)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    toolResp,
				Name:       call.Function.Name,
				ToolCallID: call.ID,
			})
		}
	}

	// 再次调用
	resp2, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       consts.DeepSeekModelFlash,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.3,
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Assistant:", resp2.Choices[0].Message.Content)
}

func dispatch(name, args string) (string, error) {
	switch name {
	case "get_weather":
		var p struct {
			City string `json:"city"`
		}
		_ = json.Unmarshal([]byte(args), &p)
		return getWeather(p.City)
	case "get_current_time":
		var p struct {
			Tz string `json:"tz"`
		}
		_ = json.Unmarshal([]byte(args), &p)
		return getCurrentTime(p.Tz)
	case "cal_tool":
		var p struct {
			Expr string `json:"expr"`
		}
		_ = json.Unmarshal([]byte(args), &p)
		return calTool(p.Expr)
	}
	return "", errors.New("unknown function: " + name)
}

func buildTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_weather",
				Description: "获取城市天气",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string","description":"城市名称"}},"required":["city"]}`),
			},
		}, {
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_current_time",
				Description: "获取当前时间",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"tz":{"type":"string","description":"时区"}},"required":["tz"]}`),
			},
		}, {
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "cal_tool",
				Description: "计算工具",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"expr":{"type":"string","description":"计算表达式"}},"required":["expr"]}`),
			},
		},
	}
}

func getWeather(city string) (string, error) {
	weatherMap := map[string]string{
		"上海": "38度",
		"北京": "-3度",
	}
	if _, ok := weatherMap[city]; !ok {
		return "", errors.New("城市未找到")
	}
	return weatherMap[city], nil
}

func getCurrentTime(tz string) (string, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", err
	}
	return time.Now().In(loc).Format(time.DateTime), nil
}

func calTool(expr string) (string, error) {
	var a, b float64
	var op string
	if _, err := fmt.Sscanf(strings.TrimSpace(expr), "%f %s %f", &a, &op, &b); err != nil {
		return "", errors.New("无法解析" + expr)
	}

	switch op {
	case "+":
		return fmt.Sprintf("%.2f", a+b), nil
	case "-":
		return fmt.Sprintf("%.2f", a-b), nil
	case "*":
		return fmt.Sprintf("%.2f", a*b), nil
	case "/":
		if b == 0 {
			return "", errors.New("除零")
		}
		return fmt.Sprintf("%.2f", a/b), nil
	}
	return "", errors.New("不支持的运算符：" + op)
}
