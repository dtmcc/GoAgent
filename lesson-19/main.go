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

	cm, err := einoai.NewChatModel(ctx, &einoai.ChatModelConfig{
		Model:   consts.DeepSeekModelFlash,
		APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL: "https://api.deepseek.com/",
	})
	if err != nil {
		log.Fatal(err)
	}

	goTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("扮演 Go 语法老师, 回答要先给结论，再解释，适合初学者, 简要回答 控制在50字内"),
		schema.UserMessage("{question}"),
	)
	agentTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("扮演 Agent 学习教练, 回答时多带上 Eino 对象的对应关系, 简要回答 控制在50字内"),
		schema.UserMessage("{question}"),
	)
	diagTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("扮演 Go 代码诊断工具, 看到报错给出最可能的原因，再给排查路径, 简要回答 控制在50字内"),
		schema.UserMessage("{question}"),
	)

	classifier := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (output map[string]any, err error) {
		q := strings.ToLower(input["question"].(string))
		cat := "go"
		switch {
		case strings.Contains(q, "agent") || strings.Contains(q, "eino") || strings.Contains(q, "tool"):
			cat = "agent"
		case strings.Contains(q, "报错") || strings.Contains(q, "error") || strings.Contains(q, "panic"):
			cat = "diag"
		}
		input["category"] = cat
		return input, nil
	})

	g := compose.NewGraph[map[string]any, *schema.Message]()
	_ = g.AddLambdaNode("classifier", classifier)
	_ = g.AddChatTemplateNode("go_tpl", goTpl)
	_ = g.AddChatTemplateNode("agent_tpl", agentTpl)
	_ = g.AddChatTemplateNode("diag_tpl", diagTpl)
	_ = g.AddChatModelNode("model", cm)

	_ = g.AddEdge(compose.START, "classifier")
	_ = g.AddBranch("classifier", compose.NewGraphBranch(
		func(ctx context.Context, input map[string]any) (endNode string, err error) {
			return input["category"].(string) + "_tpl", nil
		},
		map[string]bool{"go_tpl": true, "agent_tpl": true, "diag_tpl": true},
	))

	_ = g.AddEdge("go_tpl", "model")
	_ = g.AddEdge("agent_tpl", "model")
	_ = g.AddEdge("diag_tpl", "model")

	_ = g.AddEdge("model", compose.END)

	runner, err := g.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	questions := []string{
		"Go 的 interface怎么理解",
		"Eino 的 toolsNode 是什么",
		"panic 报错，nil pointer，怎么排查",
	}

	for _, question := range questions {
		output, err := runner.Invoke(ctx, map[string]any{
			"question": question,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("question：", question, "answer：", output.Content)
	}
}
