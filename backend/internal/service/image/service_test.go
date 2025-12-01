package image

import (
	"context"
	"os"
	"testing"

	"github.com/guixu633/agent/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

func TestGenerateImage(t *testing.T) {
	// 初始化客户端
	client, err := initTestClient()
	if err != nil {
		t.Skipf("跳过测试: %v", err)
		return
	}

	service := NewService(client)

	// 测试纯文本生成
	t.Run("纯文本提示词生成", func(t *testing.T) {
		req := &model.ImageGenerateRequest{
			Prompt: "一只可爱的羊毛毡小猫咪",
		}

		resp, err := service.GenerateImage(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		
		// 打印结果
		if resp.Description != "" {
			t.Logf("描述: %s", resp.Description)
		}
		if len(resp.Images) > 0 {
			t.Logf("生成了 %d 张图片", len(resp.Images))
		}
	})
}

func initTestClient() (*genai.Client, error) {
	configPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if configPath == "" {
		configPath = "../../configs/gcp/gcp.json"
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", configPath)

	return genai.NewClient(context.Background(), &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  "startup-x-465606",
		Location: "global",
	})
}

