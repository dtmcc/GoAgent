package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/tool/utils"
)

type SearchRequest struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"用户想学的关键词，自然语言即可，比如：刚学完Gin路由,继续学什么?"`
	Stage string `json:"stage" jsonschema:"required,enum=basic,enum=intermediate,enum=advanced" jsonschema_description:"用户当前学习阶段，比如：初级、中级、高级"`
	Limit int    `json:"limit" jsonschema_description:"返回数量，默认3 最多5"`
}

type Course struct {
	Title       string `json:"title"`
	Stage       string `json:"stage"`
	Description string `json:"description"`
}
type SearchResponse struct {
	Courses []Course `json:"courses"`
	Total   int      `json:"total"`
}

func validStage(stage string) error {
	if stage == "basic" || stage == "intermediate" || stage == "advanced" {
		return nil
	}
	return fmt.Errorf("invalid stage: %s", stage)
}

func courseSearch(_ context.Context, req *SearchRequest) (resp *SearchResponse, err error) {
	err = validStage(req.Stage)
	if err != nil {
		return nil, err
	}

	allCourses := []Course{
		{Title: "课程1", Stage: "basic", Description: "课程1-描述"},
		{Title: "课程2", Stage: "basic", Description: "课程2-描述"},
		{Title: "课程3", Stage: "intermediate", Description: "课程3-描述"},
		{Title: "课程4", Stage: "intermediate", Description: "课程4-描述"},
		{Title: "课程5", Stage: "advanced", Description: "课程5-描述"},
	}

	if req.Limit <= 0 || req.Limit >= 10 {
		req.Limit = 3
	}

	// 正常此处 精准匹配stage 和 模糊匹配title，本次跳过匹配 直接返回前两个课程
	return &SearchResponse{
		Courses: allCourses[:req.Limit],
		Total:   req.Limit,
	}, nil
}
func InferTool(ctx context.Context) {

	callTool, err := utils.InferTool(
		"course_search",
		"在 Go 系列课程库里检索课程。适用：用户想找'我接下来该学什么'、‘哪门课在前哪门课在后‘这类问题。不适用：用户已经明确要学某门课的某一节，直接回答即可。",
		courseSearch,
	)
	if err != nil {
		log.Fatal(err)
	}
	info, err := callTool.Info(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info.Name)
	fmt.Println(info.Desc)

	fmt.Println("===== 正常调用 =====")
	result, err := callTool.InvokableRun(ctx, `{"query": "Go语言", "stage": "basic", "limit": 3}`)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(result)

	// stage 不符合枚举类型
	fmt.Println("===== 异常调用 =====")
	result, err = callTool.InvokableRun(ctx, `{"query": "Go语言", "stage": "basic-xxx", "limit": 3}`)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(result)
}
