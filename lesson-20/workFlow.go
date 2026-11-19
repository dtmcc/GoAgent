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
)

type Input struct {
	UserID   string
	Question string
}

type Profile struct {
	Level string
	Goal  string
}

type KnowLedge struct {
	Chunks []string
}

type Request struct {
	Question string
}

/**
workflow数据流
不关心控制流程，只关心数据流程

STRAT

Input
	｜- load_profile		->	Profile{Level, Goal}
	｜- search_knowledge	->	KnowLedge{Chunks}		->	join_knowledge	-> string
	｜- normal_request		->	Request{Question}

Profile.Level		-> prompt.level
Profile.Goal		-> prompt.goal
join_knowledge		-> prompt.knowledge
Request.Question	-> prompt.question

prompt -> model -> post_process

END

*/

func WorkFlow(ctx context.Context) {

	// 组装chatModel
	cm, err := einoai.NewChatModel(ctx, &einoai.ChatModelConfig{
		APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		Model:   consts.DeepSeekModelFlash,
		BaseURL: consts.DeepSeekBaseURL,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 组装tpl
	tpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是 Go 和 Agent 课程助教。学员水平：｛level｝\n学习目标：｛goal｝\n参考资料：\n｛knowledge｝\n回答要先给结论，再给一个小例子。"),
		schema.UserMessage("{question}"))

	// 定义 load_profile 执行函数		加载人物画像
	loadProfile := func(_ context.Context, input Input) (Profile, error) {
		return Profile{
			Level: "beginner",
			Goal:  "能使用Eino写出自己的学习项目",
		}, nil
	}

	// 定义 search_knowledge 执行函数		检索知识库
	searchKnowledge := func(_ context.Context, question string) (KnowLedge, error) {
		return KnowLedge{
			Chunks: []string{
				"Workflow 适合把多个上游字段映射到一个下游输入。",
				"ChatTempLate 的输入通常是 map［string]any。",
				"ChatModel 的输入是 []*schema.Message，输出是 *schema.Message。",
			},
		}, nil
	}

	// 定义 normal_request 执行函数		脱敏入参
	normalRequest := func(_ context.Context, question string) (Request, error) {
		return Request{
			Question: strings.TrimSpace(question),
		}, nil
	}

	// 定义 join_knowledge 执行函数		将知识库join到一起
	joinKnowledge := func(_ context.Context, knowLedge KnowLedge) (string, error) {
		return "-" + strings.Join(knowLedge.Chunks, "\n-"), nil
	}

	// 定义 post_process 执行函数		后处理model输出 转为string
	postProcess := func(_ context.Context, msg *schema.Message) (string, error) {
		return fmt.Sprintf("【学习指导】\n%s", msg.Content), nil
	}

	// 初始化一个 workflow
	wf := compose.NewWorkflow[Input, string]()

	// load_profile 起始START 取全部Input
	wf.AddLambdaNode("load_profile", compose.InvokableLambda(loadProfile)).AddInput(compose.START)

	// search_knowledge 起始START 取Input.Question
	wf.AddLambdaNode("search_knowledge", compose.InvokableLambda(searchKnowledge)).AddInput(compose.START, compose.FromField("Question"))
	// join_knowledge 起始上一步search_knowledge  取全部输出
	wf.AddLambdaNode("join_knowledge", compose.InvokableLambda(joinKnowledge)).AddInput("search_knowledge")

	// normal_request 起始START 取Input.Question
	wf.AddLambdaNode("normal_request", compose.InvokableLambda(normalRequest)).AddInput(compose.START, compose.FromField("Question"))

	// 添加prompt 替换里面的占位符 数据分别来自  load_profile，join_knowledge，normal_request  并做一些映射
	wf.AddChatTemplateNode("prompt", tpl).
		AddInput("load_profile",
			compose.MapFields("Level", "level"),
			compose.MapFields("Goal", "goal"),
		).
		AddInput("join_knowledge",
			compose.ToField("knowledge"),
		).
		AddInput("normal_request",
			compose.MapFields("Question", "question"),
		)

	// 添加model 输入来自上一步prompt
	wf.AddChatModelNode("model", cm).AddInput("prompt")

	// 添加 post_process 对应postProcess 执行函数  输入来自上一步model
	wf.AddLambdaNode("post_process", compose.InvokableLambda(postProcess)).AddInput("model")

	// 结束 输入来自 post_process  输出字符串
	wf.End().AddInput("post_process")

	runner, err := wf.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}
	result, err := runner.Invoke(ctx, Input{
		UserID:   "10001",
		Question: "Workflow 和 Graph 到底有什么区别？简要回答 100字以内",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
