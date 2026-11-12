package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type WeatherInput struct {
	City string `json:"city" jsonschema:"required" jsonschema_description:"城市名称, 例如：北京，天津，上海"`
}

type WeatherOutput struct {
	City    string `json:"city"`
	Weather string `json:"weather"`
	Temp    string `json:"temp"`
}

func queryWeather(ctx context.Context, req *WeatherInput) (resp *WeatherOutput, err error) {
	fake := map[string]WeatherOutput{
		"北京": {"北京", "晴", "26C"},
		"上海": {"上海", "小雨", "24C"},
	}

	if weather, ok := fake[req.City]; ok {
		return &weather, nil
	}

	return &WeatherOutput{req.City, "未知", "未知"}, nil
}
func ToolNode(ctx context.Context) {
	weatherTool, err := utils.InferTool(
		"query_weather",
		"查询指定城市的天气。输入 city，返回天气和温度。",
		queryWeather,
	)
	if err != nil {
		panic(err)
	}
	courseSearchTool, err := utils.InferTool(
		"course_search",
		"搜索 Go Agent 课程。输入 query，返回相关课程标题。",
		courseSearch,
	)
	if err != nil {
		panic(err)
	}
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{weatherTool, courseSearchTool},
	})
	if err != nil {
		panic(err)
	}

	modelOutput := &schema.Message{
		Role: schema.Assistant,
		ToolCalls: []schema.ToolCall{
			{
				ID:   "call_weather_beijing",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "query_weather",
					Arguments: `{"city":"北京"}`,
				},
			},
			{

				ID:   "call_weather_shanghai",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "query_weather",
					Arguments: `{"city":"上海"}`,
				},
			},
			{

				ID:   "call_course_eino",
				Type: "function",
				Function: schema.FunctionCall{
					Name:      "course_search",
					Arguments: `{"query":"Eino","stage":"basic"}`,
				},
			},
		},
	}

	toolResults, err := toolsNode.Invoke(ctx, modelOutput)
	if err != nil {
		panic(err)
	}
	for _, result := range toolResults {
		fmt.Println("===== 结果 =====")
		fmt.Println(result.ToolCallID)
		fmt.Println(result.ToolName)
		fmt.Println(result.Role)
		fmt.Println(result.Content)
	}

}
