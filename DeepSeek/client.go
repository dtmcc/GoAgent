package DeepSeek

import (
	"GoAgent/consts"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func NewClient() *openai.Client {
	godotenv.Load()

	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		fmt.Println("DEEPSEEK_API_KEY is empty")
		return nil
	}
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = consts.DeepSeekBaseURL

	return openai.NewClientWithConfig(cfg)
}
