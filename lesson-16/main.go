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
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/joho/godotenv"
)

type SearchReq struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"用户想学的关键词，自然语言即可，比如：刚学完Gin路由,继续学什么?"`
	Stage string `json:"stage" jsonschema:"required,enum=basic,enum=intermediate,enum=advanced" jsonschema_description:"用户当前学习阶段，比如：初级、中级、高级"`
}

type SearchResp struct {
	Courses []Course `json:"courses"`
}

type Course struct {
	Title string `json:"title"`
	Stage string `json:"stage"`
}

func searchCourse(_ context.Context, req *SearchReq) (resp *SearchResp, err error) {
	allCourses := []Course{
		{Title: "课程1", Stage: "basic"},
		{Title: "课程2", Stage: "basic"},
		{Title: "课程3", Stage: "intermediate"},
		{Title: "课程4", Stage: "intermediate"},
		{Title: "课程5", Stage: "advanced"},
	}

	// 正常此处 精准匹配stage 和 模糊匹配title，本次跳过匹配 直接返回前两个课程
	return &SearchResp{
		Courses: allCourses[:3],
	}, nil
}

func main() {
	ctx := context.Background()
	godotenv.Load()

	cm, err := einoai.NewChatModel(ctx, &einoai.ChatModelConfig{
		APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		Model:   consts.DeepSeekModelFlash,
		BaseURL: "https://api.deepseek.com/",
	})
	if err != nil {
		log.Fatal(err)
	}

	courseTool, err := utils.InferTool(
		"course_search",
		"在 Go 系列课程库里检索课程。适用：用户想找'我接下来该学什么'、‘哪门课在前哪门课在后‘这类问题。不适用：用户已经明确要学某门课的某一节，直接回答即可。",
		searchCourse,
	)
	if err != nil {
		log.Fatal(err)
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "go_coach",
		Description:   "回答 Go 学习问题，必要时调用课程和代码诊断工具。",
		Instruction:   "扮演一个面向 Go 学生的学习教练，回答先给结论再解释。",
		Model:         cm,
		MaxIterations: 3,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{courseTool},
			},
		},
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
			log.Fatal(event.Err)
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
