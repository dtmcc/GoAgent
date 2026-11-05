package main

import (
	"GoAgent/consts"

	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

// 实现 React

const systemPrompt = "你是一位 Go 语言开发导师，尽量用通俗易懂的语言回答，先给出结论，后给出释义。你可以调用 scan_dir、read_file 两个工具来观察用户本地代码，需要看文件时请直接调用工具，不要说没有权限。"

func main() {
	question := "分析当前目录，给出当前目录下全部的 .go 文件，并读取第一份文件的前 10 行，额外返回读取的第一份文件的行数"

	client := newClient()

	err := runReact(client, question, 5)
	if err != nil {
		fmt.Println("出错：", err)
		os.Exit(1)
	}
}

func runReact(client *openai.Client, question string, maxStep int) error {
	messages := make([]openai.ChatCompletionMessage, 0, 2)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	// 一定要把自己开发的工具带进去
	tools := buildTools()
	for i := 1; i <= maxStep; i++ {
		fmt.Printf("\r===== 开始第 %d 轮调用 =====", i)
		resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model:    consts.DeepSeekModelFlash,
			Messages: messages,
			Stream:   false,
			Tools:    tools, // 入参进去
		})
		if err != nil {
			return err
		}
		if len(resp.Choices) == 0 {
			return nil
		}
		msg := resp.Choices[0].Message
		if msg.Content == "" {
			msg.Content = ""
		}
		// 保留原样输出 送回
		messages = append(messages, msg)

		if len(msg.ToolCalls) == 0 {
			fmt.Println("最终输出-调用结束", msg.Content)
			return nil
		}

		for _, tool := range msg.ToolCalls {
			fmt.Printf("\n ACTION %s", tool.Function.Name)
			dispatchRes, err := dispatch(tool.Function.Name, tool.Function.Arguments)
			fmt.Printf("\n Observation %s", dispatchRes)
			if err != nil {
				return errors.New("")
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    dispatchRes,
				ToolCallID: tool.ID,
			})
		}
	}
	fmt.Println("超过最大循环上限，仍未输出")
	return nil
}

func buildTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "scan_dir",
				Description: "递归列出指定目录下所有 .go 文件，返回每行一个绝对/相对路径",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"要扫描的目录，缺省为当前目录 ."}},"required":["path"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file",
				Description: "读取指定文件的前 N 行",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"文件路径"},"limit":{"type":"integer","description":"最多读取行数，缺省 10"}},"required":["path"]}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "count_lines",
				Description: "读取某个文件的行数 返回行数",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"文件路径"}},"required":["path"]}`),
			},
		},
	}
}

func dispatch(function string, arguments string) (string, error) {
	switch function {
	case "scan_dir":
		var p struct {
			Path string `json:"path"`
		}
		if p.Path == "" {
			p.Path = "."
		}
		_ = json.Unmarshal([]byte(arguments), &p)
		return scanDir(p.Path)
	case "read_file":
		var p struct {
			Path  string `json:"path"`
			Limit int    `json:"limit"`
		}
		_ = json.Unmarshal([]byte(arguments), &p)
		files, err := scanDir(p.Path)
		if err != nil {
			return "", errors.New("scan_dir 出错")
		}
		firstPath := strings.Split(files, "\n")[0]
		content, err := readFile(firstPath, p.Limit)
		if err != nil {
			return "", errors.New("read_file 出错")
		}
		return content, nil
	case "count_lines":
		var p struct {
			Path string `json:"path"`
		}
		_ = json.Unmarshal([]byte(arguments), &p)
		countLines, err := countLines(p.Path)
		if err != nil {
			return "", errors.New("count_lines 出错")
		}
		return fmt.Sprintf("当前文件路径：%s 行数：%d", p.Path, countLines), nil
	default:
		return "", errors.New("未知函数")
	}
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var lines int
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for sc.Scan() {
		lines++
	}
	if err := sc.Err(); err != nil {
		return 0, err
	}
	return lines, nil
}

func readFile(firstPath string, limit int) (string, error) {
	f, err := os.Open(firstPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) >= limit {
			break
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

func scanDir(path string) (string, error) {
	var files []string
	_ = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})
	return strings.Join(files, "\n"), nil
}

func newClient() *openai.Client {
	godotenv.Load()

	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		fmt.Println("DEEPSEEK_API_KEY is empty")
		return nil
	}
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = "https://api.deepseek.com/"

	return openai.NewClientWithConfig(cfg)
}
