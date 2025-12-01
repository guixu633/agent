package llm

import (
	"context"
	"os"

	"google.golang.org/genai"
)

const (
	DefaultRegion  = "global"
	DefaultProject = "startup-x-465606"
)

func NewGenAI(configPath string) (*genai.Client, error) {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", configPath)

	ai, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  DefaultProject,
		Location: DefaultRegion,
	})
	if err != nil {
		return nil, err
	}

	return ai, nil
}
