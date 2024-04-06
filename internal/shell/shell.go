package shell

import (
	"context"

	"github.com/PeronGH/aish/internal/prompt"
	openai "github.com/sashabaranov/go-openai"
)

type AiShell struct {
	openai  *openai.Client
	model   string
	history []openai.ChatCompletionMessage
}

type AiShellConfig struct {
	Openai      *openai.Client
	OpenaiModel string

	PromptName string

	// Data for prompt template
	Username string
	Hostname string
}

func NewAiShell(config AiShellConfig) (*AiShell, string, error) {
	prompt, err := prompt.GetPrompt(config.PromptName, prompt.PromptInit{
		Username: config.Username,
		Hostname: config.Hostname,
	})
	if err != nil {
		return nil, "", err
	}

	return &AiShell{
		openai:  config.Openai,
		model:   config.OpenaiModel,
		history: prompt.Messages,
	}, prompt.InitialPrompt, nil
}

func (s *AiShell) Execute(ctx context.Context, command string) (string, error) {
	s.history = append(
		s.history,
		openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: command},
	)
	response, err := s.openai.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    s.history,
		Temperature: 0,
	})
	if err != nil {
		return "", err
	}
	aiMessage := response.Choices[0].Message
	s.history = append(s.history, aiMessage)
	return aiMessage.Content, nil
}
