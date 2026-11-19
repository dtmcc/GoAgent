package main

import (
	"context"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	godotenv.Load()

	MapField(ctx)
	WorkFlow(ctx)
}
