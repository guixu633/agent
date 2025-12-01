package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	"golang.org/x/oauth2/google"
)

const (
	DefaultRegion  = "global"
	DefaultProject = "startup-x-465606"
)

func NewClaude(ctx context.Context, configPath string) (*anthropic.Client, error) {
	jsonData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %v, please set GOOGLE_APPLICATION_CREDENTIALS to the path of the service account key file", err)
	}
	creds, err := google.CredentialsFromJSON(ctx, jsonData,
		"https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %v", err)
	}
	client := anthropic.NewClient(
		vertex.WithCredentials(ctx, DefaultRegion, DefaultProject, creds),
		option.WithHeaderAdd("anthropic-beta", "interleaved-thinking-2025-05-14"),
	)
	return &client, nil
}
