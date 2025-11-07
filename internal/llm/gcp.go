package llm

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type GCP struct {
	genai *genai.Client
}

type GCPOptions struct {
	Project  string
	Location string
}

const (
	DefaultProject  = "startup-x-465606"
	DefaultLocation = "global"
)

func NewGCP(opts *GCPOptions) (*GCP, error) {
	project := opts.Project
	location := opts.Location
	if project == "" {
		project = DefaultProject
	}
	if location == "" {
		location = DefaultLocation
	}

	// 检查环境变量
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable not set")
	}

	ai, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  project,
		Location: location,
	})
	if err != nil {
		return nil, err
	}

	return &GCP{
		genai: ai,
	}, nil
}
