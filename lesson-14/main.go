package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type CalRequest struct {
	Num1 float64 `json:"num1"`
	Num2 float64 `json:"num2"`
	Op   string  `json:"op"`
}

type CalResponse struct {
	Result     float64 `json:"result"`
	Expression string  `json:"expression"`
}

func calculate(_ context.Context, req *CalRequest) (response *CalResponse, err error) {
	switch req.Op {
	case "add":
		return &CalResponse{Result: req.Num1 + req.Num2, Expression: fmt.Sprintf("%.2f %s %.2f = %.2f", req.Num1, req.Op, req.Num2, req.Num1+req.Num2)}, nil
	case "sub":
		return &CalResponse{Result: req.Num1 - req.Num2, Expression: fmt.Sprintf("%.2f %s %.2f = %.2f", req.Num1, req.Op, req.Num2, req.Num1-req.Num2)}, nil
	case "mul":
		return &CalResponse{Result: req.Num1 * req.Num2, Expression: fmt.Sprintf("%.2f %s %.2f = %.2f", req.Num1, req.Op, req.Num2, req.Num1*req.Num2)}, nil
	case "div":
		if req.Num2 == 0 {
			return &CalResponse{Result: 0, Expression: ""}, fmt.Errorf("除零")
		}
		return &CalResponse{Result: req.Num1 / req.Num2, Expression: fmt.Sprintf("%.2f %s %.2f = %.2f", req.Num1, req.Op, req.Num2, req.Num1/req.Num2)}, nil
	}
	return &CalResponse{Result: 0, Expression: ""}, fmt.Errorf("不支持的运算符：%s", req.Op)
}
func main() {
	ctx := context.Background()

	calTool := utils.NewTool(&schema.ToolInfo{
		Name: "calculator",
		Desc: "执行两个数字的加减乘除运算。适用：用户需要做加减乘除等简单算术运算。不适用其他计算",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"num1": {
				Type:     schema.Number,
				Desc:     "第一个数字",
				Required: true,
			},
			"num2": {
				Type:     schema.Number,
				Desc:     "第二个数字",
				Required: true,
			},
			"op": {
				Type:     schema.String,
				Desc:     "运算符",
				Required: true,
				Enum:     []string{"add", "sub", "mul", "div"},
			},
		}),
	}, calculate)

	info, err := calTool.Info(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println(info.Name)
	fmt.Println(info.Desc)

	args := `{"num1": 10, "num2": 5, "op": "add"}`
	result, err := calTool.InvokableRun(ctx, args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println(result, args)
}
