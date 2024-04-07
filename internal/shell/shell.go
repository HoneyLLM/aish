package shell

import (
	"context"

	"github.com/PeronGH/aish/internal/prompt"
	"github.com/PeronGH/aish/internal/utils"
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

func (s *AiShell) Execute(ctx context.Context, command string) (<-chan string, error) {
	s.history = append(
		s.history,
		openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: command},
	)
	response, err := s.openai.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    s.history,
		Temperature: 0,
	})
	if err != nil {
		return nil, err
	}

	tokenCh := make(chan string)
	go func() {
		defer response.Close()
		defer close(tokenCh)

		for {
			msg, err := response.Recv()
			if err != nil {
				return
			}

			tokenCh <- msg.Choices[0].Delta.Content
		}
	}()

	return utils.StringToLineChannel(tokenCh), nil
}

func (s *AiShell) GetHistory() []openai.ChatCompletionMessage {
	return s.history
}

func (s *AiShell) AddAiMessage(message string) {
	s.history = append(
		s.history,
		openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: message},
	)
}
