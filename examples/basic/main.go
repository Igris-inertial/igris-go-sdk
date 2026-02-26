package main

import (
	"context"
	"fmt"
	"log"

	igris "github.com/igris-inertial/go-sdk"
)

func main() {
	client := igris.NewClient(
		"https://api.igris-inertial.com",
		"your-api-key",
	)

	ctx := context.Background()

	// Simple inference
	resp, err := client.Infer(ctx, &igris.InferRequest{
		Model: "gpt-4",
		Messages: []igris.Message{
			{Role: "user", Content: "Hello, world!"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
	if resp.Usage != nil {
		fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)
	}

	// Health check
	health, err := client.Health(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Gateway status: %s\n", health.Status)

	// List models
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Models: %v\n", models.Models)

	// List providers
	providers, err := client.Providers.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range providers {
		fmt.Printf("Provider: %s (%s) - %s\n", p.Name, p.Type, p.Status)
	}
}
