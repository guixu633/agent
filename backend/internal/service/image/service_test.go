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
	// 注意：此测试需要 OSS 客户端，暂时跳过
	// TODO: 添加 OSS mock 或集成测试
	t.Skip("需要 OSS 客户端，暂时跳过")
	
	// 初始化客户端
	// client, err := initTestClient()
	// if err != nil {
	// 	t.Skipf("跳过测试: %v", err)
	// 	return
	// }
	// service := NewService(client, nil)

	// 测试纯文本生成
	t.Run("纯文本提示词生成", func(t *testing.T) {
		req := &model.ImageGenerateRequest{
			Prompt: "一只可爱的羊毛毡小猫咪",
		}

		resp, err := service.GenerateImage(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// 打印结果
		for i, part := range resp.Parts {
			t.Logf("Part %d (%s):", i, part.Type)
			if part.Type == "text" {
				t.Logf("  Text: %s", part.Text)
			} else if part.Type == "image" {
				t.Logf("  Image: %s, Size: %d", part.Image.MimeType, len(part.Image.Data))
			}
		}
	})
}

func initTestClient() (*genai.Client, error) {
	configPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if configPath == "" {
		configPath = "../../../configs/gcp/gcp.json"
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", configPath)

	return genai.NewClient(context.Background(), &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  "startup-x-465606",
		Location: "global",
	})
}
