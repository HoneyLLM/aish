package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/user"
	"strings"

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
	promptOs := os.Getenv("PROMPT_OS")
	if promptOs == "" {
		promptOs = "ubuntu"
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
	shellCommand := os.Getenv("SHELL_COMMAND")

	// Create a new OpenAI client
	config := openai.DefaultConfig(openaiApiKey)
	config.BaseURL = openaiBaseUrl
	client := openai.NewClientWithConfig(config)

	aish, initialPrompt, err := shell.NewAiShell(shell.AiShellConfig{
		Openai:      client,
		OpenaiModel: openaiModel,
		PromptName:  promptOs,
		Username:    shellUsername,
		Hostname:    shellHostname,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Info("New session", "user", shellUsername, "host", shellHostname)

	if shellCommand != "" {
		log.Info("User", "command", shellCommand)
		output, err := aish.Execute(context.Background(), shellCommand)
		if err != nil {
			log.Fatal(err)
			return
		}
		utils.HandleChannel(
			output,
			func(line string, isLast bool) {
				if !isLast {
					fmt.Print(line)
					fmt.Print("\n")
				}
			},
			func(t []string) {
				fullMsg := strings.Join(t, "\n")
				log.Info("AI", "output", fullMsg)
				aish.AddAiMessage(fullMsg)
			},
		)
		return
	}

	fmt.Print(initialPrompt)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(" ")
		command, err := reader.ReadString('\n')
		command = strings.TrimRight(command, "\n")
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Info("User", "command", command)

		output, err := aish.Execute(context.Background(), command)
		if err != nil {
			log.Fatal(err)
			return
		}
		utils.HandleChannel(
			output,
			func(line string, isLast bool) {
				fmt.Print(line)
				if !isLast {
					fmt.Print("\n")
				}
			},
			func(t []string) {
				fullMsg := strings.Join(t, "\n")
				log.Info("AI", "output", fullMsg)
				aish.AddAiMessage(fullMsg)
			},
		)
	}
}
