package aish

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/HoneyLLM/aish/internal/prompt"
	"github.com/HoneyLLM/aish/internal/utils"
	"github.com/sashabaranov/go-openai"
)

type Config struct {
	PromptName string

	// Data for prompt template
	Username string
	Hostname string

	// LLM API settings
	Openai      *openai.Client
	OpenaiModel string
}

type Instance struct {
	prompt    *prompt.Prompt
	oaiClient *openai.Client
	model     string
}

func New(config *Config) (*Instance, error) {
	// Read prompt config and get prompt
	promptName := config.PromptName
	if promptName == "" {
		promptName = "ubuntu"
	}

	username := config.Username
	if username == "" {
		username = "root"
	}

	hostname := config.Hostname
	if hostname == "" {
		hostname = "server"
	}

	prompt, err := prompt.GetPrompt(promptName, prompt.PromptInit{
		Username: username,
		Hostname: hostname,
	})
	if err != nil {
		return nil, err
	}

	// Read OpenAI config
	oaiClient := config.Openai
	if oaiClient == nil {
		return nil, errors.New("OpenAI client must not be nil")
	}

	model := config.OpenaiModel
	if model == "" {
		model = openai.GPT4oMini
	}

	return &Instance{
		prompt:    prompt,
		model:     config.OpenaiModel,
		oaiClient: oaiClient,
	}, nil
}

func (i *Instance) streamCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (iter.Seq2[string, error], error) {
	stream, err := i.oaiClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       i.model,
		Messages:    messages,
		Temperature: 0,
	})
	if err != nil {
		return nil, err
	}

	return func(yield func(string, error) bool) {
		defer stream.Close()
		for {
			res, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if !yield(res.Choices[0].Delta.Content, err) {
				return
			}
		}
	}, nil
}

// Respond to a single input, returning the content only without the tailing prompt
func (i *Instance) Handle(ctx context.Context, w io.Writer, input string) error {
	// Initialize messages
	messages := make([]openai.ChatCompletionMessage, len(i.prompt.Messages)+2)
	copy(messages, i.prompt.Messages)

	messages = append(
		i.prompt.Messages,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: i.prompt.InitialPrompt,
		},
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		},
	)

	stream, err := i.streamCompletion(ctx, messages)
	if err != nil {
		return err
	}

	// Stream content line by line, but discard the final one
	var (
		buffer           strings.Builder
		lastLine         string
		lastLineNotEmpty bool
	)

	for content, err := range stream {
		if err != nil {
			return err
		}

		for i, lineContent := range strings.Split(content, "\n") {
			// For the first component
			if i == 0 {
				// It is always in the same line
				// So just write it to buffer
				buffer.WriteString(lineContent)
				continue
			}
			// When new line is encountered

			// Print the last line if it is not empty
			if lastLineNotEmpty {
				fmt.Fprintln(w, lastLine)
			}

			// Mark last line non-empty
			lastLineNotEmpty = true
			// And write buffer to last line
			lastLine = buffer.String()

			// Reset the buffer
			buffer.Reset()
			// And start with this lineContent
			buffer.WriteString(lineContent)
		}
		// Finally, print the last line if it is not empty
		if lastLineNotEmpty {
			fmt.Fprintln(w, lastLine)
		}
		// And discard buffer
		buffer.Reset()
	}

	return nil
}

// Handle a session, read input and stream output line by line
func (i *Instance) HandleSession(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	err := ctx.Err()
	// Quit if ctx is cancelled
	if err != nil {
		return err
	}

	// Initialize messages
	messages := make([]openai.ChatCompletionMessage, len(i.prompt.Messages)+1)
	copy(messages, i.prompt.Messages)

	// Print the initial prompt
	fmt.Fprint(stdout, i.prompt.InitialPrompt, " ")
	// Then add it to message history
	messages = append(
		messages,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: i.prompt.InitialPrompt,
		},
	)

	for command, err := range utils.ReadShellCommand(bufio.NewReader(stdin)) {
		if errors.Is(err, io.EOF) {
			return nil
		}
		messages = append(
			messages,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: command,
			})

		stream, err := i.streamCompletion(ctx, messages)
		if err != nil {
			return err
		}

		var buffer strings.Builder
		for content, err := range stream {
			if err != nil {
				return err
			}
			for i, lineContent := range strings.Split(content, "\n") {
				// For the first component
				if i == 0 {
					// It is always in the same line
					// So just write it to buffer
					buffer.WriteString(lineContent)
					continue
				}
				// When new line is encountered

				// Flush the buffer
				fmt.Fprintln(stdout, buffer.String())
				buffer.Reset()
				// Write this line to the buffer
				buffer.WriteString(lineContent)
			}
			// Finally, flush the buffer if it is not empty
			if buffer.Len() > 0 {
				fmt.Fprint(stdout, buffer.String(), " ")
			}
		}
	}

	return nil
}
