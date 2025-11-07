package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

func TestGCP(t *testing.T) {
	// 创建 GCP 客户端
	ai, err := NewGenAI("../../configs/gcp/gcp.json")
	assert.NoError(t, err)

	// 创建生成请求
	ctx := context.Background()
	prompt := "你好，请用一句话介绍你自己。"

	// 调用 GenerateContent 方法
	resp, err := ai.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", genai.Text(prompt), nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// 验证响应
	assert.NotEmpty(t, resp.Candidates)
	assert.NotEmpty(t, resp.Candidates[0].Content.Parts)

	// 打印响应内容
	fmt.Printf("模型响应: %v\n", resp.Candidates[0].Content.Parts[0])

	// 也可以使用 Text() 方法获取文本
	fmt.Printf("文本内容: %s\n", resp.Text())
}

func TestListModels(t *testing.T) {
	// 创建 GCP 客户端
	ai, err := NewGenAI("../../configs/gcp/gcp.json")
	assert.NoError(t, err)

	ctx := context.Background()

	// 方法1: 使用 All() 列出所有模型（推荐）
	t.Log("=== 使用 All() 列出所有可用模型 ===")
	count := 0
	for model, err := range ai.Models.All(ctx) {
		if err != nil {
			t.Fatalf("遍历模型时出错: %v", err)
		}
		count++
		fmt.Printf("模型 #%d: %s\n", count, model.Name)
		if model.DisplayName != "" {
			fmt.Printf("  显示名称: %s\n", model.DisplayName)
		}
		if model.Description != "" {
			fmt.Printf("  描述: %s\n", model.Description)
		}
		fmt.Println()
	}
	t.Logf("总共找到 %d 个模型", count)
	assert.Greater(t, count, 0, "应该至少有一个可用模型")

	// 方法2: 使用 List() 分页列出模型
	t.Log("\n=== 使用 List() 分页列出模型（第一页，最多5个）===")
	page, err := ai.Models.List(ctx, &genai.ListModelsConfig{
		PageSize: 5,
	})
	assert.NoError(t, err)

	pageModels := page.Items
	assert.NotEmpty(t, pageModels, "第一页应该有模型")

	for i, model := range pageModels {
		fmt.Printf("页面模型 #%d: %s\n", i+1, model.Name)
	}

	t.Logf("第一页共有 %d 个模型", len(pageModels))
	if page.NextPageToken != "" {
		t.Log("还有更多页面可用")
	}
}

func TestGCPWithWebSearch(t *testing.T) {
	// 创建 GCP 客户端
	ai, err := NewGenAI("../../configs/gcp/gcp.json")
	assert.NoError(t, err)

	// 创建生成请求
	ctx := context.Background()
	// 使用需要实时信息的问题来测试 web search
	prompt := "2024年最新的人工智能发展趋势是什么？"

	// 配置 Google Search Retrieval 工具
	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{
				GoogleSearch: &genai.GoogleSearch{},
			},
		},
	}

	// 调用 GenerateContent 方法，传入配置
	resp, err := ai.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", genai.Text(prompt), config)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// 验证响应
	assert.NotEmpty(t, resp.Candidates)
	assert.NotEmpty(t, resp.Candidates[0].Content.Parts)

	// 打印响应内容
	fmt.Printf("=== 使用 Web Search 的模型响应 ===\n")
	fmt.Printf("文本内容: %s\n\n", resp.Text())

	jsonData, err := json.MarshalIndent(resp, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(jsonData))

	// 检查是否有 grounding metadata（搜索来源信息）
	if resp.Candidates[0].GroundingMetadata != nil {
		fmt.Printf("=== Grounding Metadata（搜索来源信息）===\n")
		gm := resp.Candidates[0].GroundingMetadata

		// 打印搜索来源
		if len(gm.SearchEntryPoint.RenderedContent) > 0 {
			fmt.Printf("搜索入口: %s\n", gm.SearchEntryPoint.RenderedContent)
		}

		// 打印每个 grounding chunk（搜索结果片段）
		if len(gm.GroundingChunks) > 0 {
			fmt.Printf("\n找到 %d 个搜索结果片段:\n", len(gm.GroundingChunks))
			for i, chunk := range gm.GroundingChunks {
				fmt.Printf("\n片段 #%d:\n", i+1)
				if chunk.Web != nil {
					fmt.Printf("  标题: %s\n", chunk.Web.Title)
					fmt.Printf("  URL: %s\n", chunk.Web.URI)
				}
			}
		}

		// 打印支持度分数（如果有的话）
		if len(gm.GroundingSupports) > 0 {
			fmt.Printf("\n=== Grounding Supports（支持度信息）===\n")
			for i, support := range gm.GroundingSupports {
				fmt.Printf("支持 #%d: 置信度分数 = %.2f\n", i+1, support.ConfidenceScores)
			}
		}
	}
}
