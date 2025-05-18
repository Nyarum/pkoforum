package api

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// translateText translates text between English and Russian using OpenAI
func (app *App) translateText(ctx context.Context, text, targetLang string) (string, error) {
	var prompt string
	if targetLang == "ru" {
		prompt = fmt.Sprintf("Translate the following English text to Russian:\n\n%s\n\nAnswer with only translated variant without anything else, if you can't translate, return the original text", text)
	} else {
		prompt = fmt.Sprintf("Translate the following Russian text to English:\n\n%s\n\nAnswer with only translated variant without anything else, if you can't translate, return the original text", text)
	}

	resp, err := app.openai.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "deepseek-chat",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("translation error: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no translation received")
	}

	return resp.Choices[0].Message.Content, nil
}
