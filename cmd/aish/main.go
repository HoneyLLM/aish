package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/PeronGH/aish/internal/shell"
	"github.com/PeronGH/aish/internal/utils"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Read environment variables
	_ = godotenv.Load()

	logFilePath := os.Getenv("LOG_FILE")
	if logFilePath != "" {
		file, err := utils.GetWriter(logFilePath)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(file)
	}

	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	openaiBaseUrl := os.Getenv("OPENAI_BASE_URL")
	if openaiApiKey == "" {
		openaiBaseUrl = "https://api.openai.com/v1"
	}
	openaiModel := os.Getenv("OPENAI_MODEL")
	if openaiModel == "" {
		openaiModel = "gpt-3.5-turbo"
	}
	shellUsername := os.Getenv("SHELL_USERNAME")
	if shellUsername == "" {
		u, _ := user.Current()
		if u == nil {
			shellUsername = "root"
		} else {
			shellUsername = u.Username
		}
	}
	shellHostname := os.Getenv("SHELL_HOSTNAME")
	if shellHostname == "" {
		shellHostname, _ = os.Hostname()
		if shellHostname == "" {
			shellHostname = "server"
		}
	}

	// Create a new OpenAI client
	config := openai.DefaultConfig(openaiApiKey)
	config.BaseURL = openaiBaseUrl
	client := openai.NewClientWithConfig(config)

	aish, initialPrompt, err := shell.NewAiShell(shell.AiShellConfig{
		Openai:      client,
		OpenaiModel: openaiModel,
		PromptName:  "ubuntu",
		Username:    shellUsername,
		Hostname:    shellHostname,
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(initialPrompt)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(" ")
		command, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			break
		}
		output, err := aish.Execute(context.Background(), command)
		if err != nil {
			log.Fatal(err)
			break
		}
		fmt.Print(output)
	}
}
