package llm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

// TestNanoBananaImageStyle 测试使用 Gemini Nano Banana 修改图片风格
func TestNanoBananaImageStyle(t *testing.T) {
	ai := getGCP(t)

	ctx := context.Background()

	// 读取图片文件
	imageData, err := os.ReadFile("data/椿.png")
	assert.NoError(t, err, "读取图片文件失败")

	// 创建提示词
	prompt := "帮我把图片修改为羊毛毡的可爱风格，短手短脚的那种可爱玩偶的感觉"

	// 构建请求内容 - 包含文本提示和图片
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
		genai.NewContentFromBytes(imageData, "image/png", genai.RoleUser),
	}

	// 调用 GenerateContent 方法
	resp, err := ai.Models.GenerateContent(
		ctx,
		"gemini-3-pro-image-preview",
		contents,
		nil,
	)
	assert.NoError(t, err, "生成内容失败")
	assert.NotNil(t, resp)

	// 验证响应
	assert.NotEmpty(t, resp.Candidates, "响应候选为空")
	assert.NotEmpty(t, resp.Candidates[0].Content.Parts, "响应内容部分为空")

	// 打印响应信息
	fmt.Printf("=== Gemini 图片风格转换响应 ===\n")
	fmt.Printf("模型: gemini-3-pro-image-preview\n")
	fmt.Printf("提示词: %s\n\n", prompt)

	// 处理响应的各个部分
	for i, part := range resp.Candidates[0].Content.Parts {
		fmt.Printf("--- 响应部分 #%d ---\n", i+1)

		// 如果是文本响应
		if part.Text != "" {
			fmt.Printf("文本内容: %s\n", part.Text)
		}

		// 如果是图片响应（内联数据）
		if part.InlineData != nil {
			fmt.Printf("图片数据:\n")
			fmt.Printf("  MIME类型: %s\n", part.InlineData.MIMEType)
			fmt.Printf("  数据大小: %d 字节\n", len(part.InlineData.Data))

			// 将生成的图片保存到文件
			outputPath := fmt.Sprintf("data/椿_羊毛毡风格_%d.png", i+1)
			err := os.WriteFile(outputPath, part.InlineData.Data, 0644)
			if err != nil {
				t.Logf("保存图片失败: %v", err)
			} else {
				fmt.Printf("  已保存到: %s\n", outputPath)
			}
		}

		fmt.Println()
	}

	// 如果有文本响应，也打印出来
	if resp.Text() != "" {
		fmt.Printf("=== 完整文本响应 ===\n%s\n", resp.Text())
	}
}
