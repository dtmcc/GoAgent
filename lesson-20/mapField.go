package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/compose"
)

func mapField(ctx context.Context) {

	// 统计字符串长度
	type Input struct {
		Text    string `json:"text"`
		Keyword string `json:"keyword"`
	}
	charCounter := func(_ context.Context, input string) (output int, err error) {
		return len([]rune(input)), nil
	}

	// 统计关键词出现次数
	type KeywordInput struct {
		FullText string `json:"fullText"`
		Word     string `json:"word"`
	}
	keywordCounter := func(_ context.Context, input KeywordInput) (output int, err error) {
		return strings.Count(input.FullText, input.Word), nil
	}

	// 初始化workflow  输入为Input, 输出为map[string]any
	wf := compose.NewWorkflow[Input, map[string]any]()

	// 第一个工作流 从开始节点start 调用char_count 对应charCounter函数  输入为 Input.Text 字符串  注意此处用 FromField 直接取
	wf.AddLambdaNode("char_count", compose.InvokableLambda(charCounter)).AddInput(compose.START, compose.FromField("Text"))

	// 第二个工作流 从上一个节点 调用keyword_count 对应keywordCounter函数  输入为KeywordInput格式
	// Input.Text->KeywordInput.FullText  Input.Keyword->KeywordInput.Word  组装为KeywordInput格式  注意此处用 MapFields 从前一个结构体字段映射为后一个结构体
	wf.AddLambdaNode("keyword_count", compose.InvokableLambda(keywordCounter)).
		AddInput(compose.START,
			compose.MapFields("Text", "FullText"),
			compose.MapFields("Keyword", "Word"))

	// workflow结束 一共两个工作流 输出结果  对应33行 输出map[string]any
	// char_count 这个流 输出 结果char_count_res 扔到map中
	// keyword_count 这个流 输出 keyword_count_res 扔到map中
	wf.End().
		AddInput("char_count", compose.ToField("char_count_res")).
		AddInput("keyword_count", compose.ToField("keyword_count_res"))

	// 编译
	runner, err := wf.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// 调用 入参Input  出参map[string]any
	result, err := runner.Invoke(ctx, Input{
		Text:    "Go 语言是一门简洁高效的编程语言，Go 的并发模型是 Go 最大的亮点之一。",
		Keyword: "Go",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
