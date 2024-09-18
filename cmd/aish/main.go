package main

import (
	"context"
	"os"

	"github.com/HoneyLLM/aish/internal/utils"
	"github.com/HoneyLLM/aish/pkg/aish"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Read environment variables
	_ = godotenv.Load()

	// Set up the logger
	logFilePath := os.Getenv("LOG_FILE")
	if logFilePath != "" {
		file, err := utils.GetWriter(logFilePath)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(file)
	}

	// Create a new OpenAI client
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	openaiBaseUrl := os.Getenv("OPENAI_BASE_URL")
	if openaiBaseUrl == "" {
		openaiBaseUrl = "https://api.openai.com/v1"
	}
	config := openai.DefaultConfig(openaiApiKey)
	config.BaseURL = openaiBaseUrl
	client := openai.NewClientWithConfig(config)

	openaiModel := os.Getenv("OPENAI_MODEL")
	promptOs := os.Getenv("PROMPT_OS")
	shellUsername := os.Getenv("AISH_USERNAME")
	shellHostname := os.Getenv("AISH_HOSTNAME")
	shellCommand := os.Getenv("AISH_COMMAND")

	aiShell, err := aish.New(&aish.Config{
		PromptName:  promptOs,
		Username:    shellUsername,
		Hostname:    shellHostname,
		Openai:      client,
		OpenaiModel: openaiModel,
	})
	if err != nil {
		log.Fatal(err)
	}

	if shellCommand != "" {
		err = aiShell.Handle(context.Background(), os.Stdout, shellCommand)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = aiShell.HandleSession(context.Background(), os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
