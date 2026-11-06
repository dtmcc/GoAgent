package main

import (
	"GoAgent/DeepSeek"
	"GoAgent/consts"
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

type UserProfile struct {
	UserID       int
	Level        string // 新手 / 入门/ 进阶
	Goal         string // 用户的学习目标
	CurrentStage string // 现在卡在哪
	Preference   string // 偏好（比如喜欢短示例、不喜欢套话）
}

type Profile struct {
	UserID           string
	AuthSubject      string    // 关联到登录身份
	UserType         string    // 学生/ 在职 / 自学
	SkillLevel       string    // Go 水平
	GoalType         string    // 学基础 / 做项目/ 学 Agent
	PurchasedCourses []string  // 买过哪些课
	CurrentTopic     string    // 当前在学什么
	CurrentStage     string    // 阶段
	UpdatedAt        time.Time // 上次更新时间，用于判断画像新不新
}

const profile = `用户画像：

- Go 水平：新手
- 目标：完成 Gin + GORM 项目实战
- 当前阶段：刚学完 Gin 路由

最近摘要：
- 上次卡点：GORM 报 no such table- 已给建议：检查 AutoMigrate 是否注册了 Model- 下一步：运行迁移并验证表结构`

func main() {
	client := DeepSeek.NewClient()
	question := "我下面要继续学什么?"
	fmt.Println("===== 无用户画像 =====")

	_ = ask(client, question)

	fmt.Println("===== 有用户画像 =====")
	_ = ask(client, profile+"\n"+question)
}

func ask(client *openai.Client, question string) error {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: consts.DeepSeekModelFlash,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "扮演一个面向Go学生的讲师，教练，回答问题具体，可执行"},
				{Role: openai.ChatMessageRoleUser, Content: question},
			},
		},
	)
	if err != nil {
		fmt.Println("model return error,", err)
		return err
	}
	if len(resp.Choices) == 0 {
		fmt.Println("model not return")
		return nil
	}

	fmt.Println(resp.Choices[0].Message.Content)
	return nil
}
