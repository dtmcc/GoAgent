package main

import (
	consts "GoAgent"

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

const systemPrompt = `你是一个本机代码仓库分析助手，正在帮用户分析他指定的目录。

每一步回复时，先把你的判断写在 content 字段里（一两句话即可，相当于 ReAct 里的 Thought），
然后在确实需要时调用工具，工具的返回结果会作为下一轮上下文给你。

任务完成（拿到了所有需要的信息）时，不要再调用工具，直接把最终结果写在 content 里。`

func main() {
	question := "请扫描当前目录，告诉我有哪些 Go 文件，并把第一个 Go 文件的前 10 行读出来。"

	client := NewClient()
	if err := runReact(client, question, 5); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewClient() *openai.Client {
	godotenv.Load()

	aipKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if aipKey == "" {
		fmt.Println("DEEPSEEK_API_KEY is empty")
		os.Exit(1)
	}

	cfg := openai.DefaultConfig(aipKey)
	cfg.BaseURL = "https://api.deepseek.com/"

	return openai.NewClientWithConfig(cfg)
}

func runReact(client *openai.Client, question string, maxSteps int) error {
	messages := make([]openai.ChatCompletionMessage, 0, 2)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	tools := buildTools()
	for i := 1; i <= maxSteps; i++ {
		fmt.Printf("\n\n===== 第 %d 轮执行开始 =====\n", i)
		resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model:    consts.DeepSeekModelFlash,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return errors.New("failed to create chat completion: " + err.Error())
		}

		if len(resp.Choices) == 0 {
			return errors.New("model return empty")
		}

		msg := resp.Choices[0].Message
		if msg.Content != "" {
			fmt.Println("Thought: ", msg.Content)
		}

		if len(msg.ToolCalls) == 0 {
			fmt.Println("最终答案： ", msg.Content)
			return nil
		}
		// 如果此字段无值 json序列化 配合omitempty标签会忽略该字段 导致deepseek调用 缺少参数
		if msg.Content == "" {
			msg.Content = ""
		}
		messages = append(messages, msg)
		for _, tool := range msg.ToolCalls {
			fmt.Printf("Action %s : %s", tool.Function.Name, tool.Function.Arguments)
			result := dispatch(tool.Function.Name, tool.Function.Arguments)
			fmt.Printf("Observation %s", result)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: tool.ID,
			})
		}
	}

	return fmt.Errorf("达到最大轮，无结果输出")
}

func buildTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "scan_repo",
				Description: "扫描指定目录下的Go 文件，返回相对路径列表。不递归 vendor、node_modules 等目录",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"root":{"type":"string","description":"目录路径，默认当前目录"}}}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file",
				Description: "读取指定文件的前 N 行",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"需要读取的文件路径"},"limit":{"type":"integer","description":"读取的行数，默认 10"}},"description":"读取指定文件的前 N 行"}`),
			},
		},
	}
}

func dispatch(toolName, toolArgs string) string {
	switch toolName {
	case "scan_repo":
		var p struct {
			Root string `json:"root"`
		}
		_ = json.Unmarshal([]byte(toolArgs), &p)
		if p.Root == "" {
			p.Root = "."
		}

		files, err := scanRepo(p.Root)
		if err != nil {
			return "errors" + err.Error()
		}
		return strings.Join(files, "\n")
	case "read_file":
		var p struct {
			Path  string `json:"path"`
			Limit int    `json:"limit"`
		}
		_ = json.Unmarshal([]byte(toolArgs), &p)
		if p.Path == "" {
			p.Path = "."
		}
		if p.Limit == 0 {
			p.Limit = 10
		}
		lines, err := readFile(p.Path, p.Limit)
		if err != nil {
			return "errors" + err.Error()
		}
		return strings.Join(lines, "\n")
	}
	return "unknow tool: " + toolName
}

func scanRepo(target string) ([]string, error) {
	var files []string
	_ = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, nil
}

func readFile(file string, n int) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	// 如果文件 单行过大 此处默认缓冲区 64kb 会导致 token too long
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) >= n {
			break
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
