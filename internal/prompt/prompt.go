package prompt

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"text/template"

	"github.com/sashabaranov/go-openai"
)

var errDecodePrompt = errors.New("failed to decode prompt")

//go:embed templates/*
var templates embed.FS

var tmpl = template.Must(
	template.
		New("prompt").
		Funcs(template.FuncMap{
			"escape": func(s string) string {
				quoted := strconv.Quote(s)
				return quoted[1 : len(quoted)-1]
			},
			"promptSymbol": func(username string) string {
				if username == "root" {
					return "#"
				}
				return "$"
			},
		}).
		ParseFS(templates, "templates/*.json"),
)

type PromptInit struct {
	Hostname string
	Username string
}

type Prompt struct {
	Messages      []openai.ChatCompletionMessage `json:"messages"`
	InitialPrompt string                         `json:"initial_prompt"`
}

func GetPrompt(system string, init PromptInit) (*Prompt, error) {
	r, w := io.Pipe()
	go func() {
		tmpl.ExecuteTemplate(w, system+".json", init)
		w.Close()
	}()

	var result Prompt
	err := json.NewDecoder(r).Decode(&result)
	if err != nil {
		return nil, errors.Join(errDecodePrompt, err)
	}

	return &result, nil
}
