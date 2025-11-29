package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestClaudeSimpleChat(t *testing.T) {
	// 创建 Claude 客户端
	ctx := context.Background()
	llm, err := NewClaude(ctx, "../../configs/gcp/gcp.json")
	assert.NoError(t, err)

	// 创建消息请求
	prompt := "你好，请用一句话介绍你自己。"

	// 调用 Claude 生成消息
	message, err := llm.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-5",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{{
			Content: []anthropic.ContentBlockParamUnion{{
				OfText: &anthropic.TextBlockParam{Text: prompt},
			}},
			Role: anthropic.MessageParamRoleUser,
		}},
	},)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	// 验证响应
	assert.NotEmpty(t, message.Content)

	jsonData, err := json.MarshalIndent(message, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(jsonData))

	for i, content := range message.Content {
		switch content.Type {
		case "text":
			fmt.Printf("内容 #%d: %s\n", i+1, content.Text)
		}
	}
}

func TestClaudeWithThinking(t *testing.T) {
	// 创建 Claude 客户端（已启用 thinking 功能）
	ctx := context.Background()
	llm, err := NewClaude(ctx, "../../configs/gcp/gcp.json")
	assert.NoError(t, err)

	// 测试需要思考的问题
	prompt := "计算 123 * 456 等于多少？请展示你的思考过程。"

	// 调用 Claude 生成消息，启用扩展思考
	message, err := llm.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-3-5-sonnet-v2@20241022",
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{{
			Content: []anthropic.ContentBlockParamUnion{{
				OfText: &anthropic.TextBlockParam{Text: prompt},
			}},
			Role: anthropic.MessageParamRoleUser,
		}},
		// 启用思考模式
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfEnabled: &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: 1024,
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, message)

	// 打印所有内容块
	fmt.Printf("模型: %s\n", message.Model)
	fmt.Printf("内容块数量: %d\n\n", len(message.Content))

	for i, content := range message.Content {
		switch content.Type {
		case "text":
			fmt.Printf("文本块 #%d:\n%s\n\n", i+1, content.Text)
		case "thinking":
			fmt.Printf("思考块 #%d:\n%s\n\n", i+1, content.Thinking)
		case "redacted_thinking":
			fmt.Printf("隐藏思考块 #%d\n\n", i+1)
		default:
			fmt.Printf("其他类型块 #%d: %s\n\n", i+1, content.Type)
		}
	}

	// 打印 token 使用情况
	fmt.Printf("Token 使用情况:\n")
	fmt.Printf("  输入 tokens: %d\n", message.Usage.InputTokens)
	fmt.Printf("  输出 tokens: %d\n", message.Usage.OutputTokens)
}

func TestClaudeStream(t *testing.T) {
	// 创建 Claude 客户端
	ctx := context.Background()
	llm, err := NewClaude(ctx, "../../configs/gcp/gcp.json")
	assert.NoError(t, err)

	// 创建流式请求
	prompt := "请用三句话介绍人工智能的发展历史。"

	fmt.Printf("问题: %s\n\n流式响应:\n", prompt)

	// 调用流式 API
	stream := llm.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     "claude-3-5-sonnet-v2@20241022",
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{{
			Content: []anthropic.ContentBlockParamUnion{{
				OfText: &anthropic.TextBlockParam{Text: prompt},
			}},
			Role: anthropic.MessageParamRoleUser,
		}},
	})

	// 使用 Accumulate 收集完整消息
	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		message.Accumulate(event)

		// 实时打印文本增量
		if event.Type == "content_block_delta" {
			if event.Delta.Type == "text_delta" {
				fmt.Print(event.Delta.Text)
			}
		}
	}

	// 检查流式处理错误
	if err := stream.Err(); err != nil {
		t.Fatalf("流式处理出错: %v", err)
	}

	fmt.Println()

	// 打印完整消息内容
	var fullText string
	for _, content := range message.Content {
		if content.Type == "text" {
			fullText += content.Text
		}
	}

	fmt.Printf("完整文本: %s\n", fullText)
	fmt.Printf("\nToken 使用情况:\n")
	fmt.Printf("  输入 tokens: %d\n", message.Usage.InputTokens)
	fmt.Printf("  输出 tokens: %d\n", message.Usage.OutputTokens)

	assert.NotEmpty(t, fullText, "流式响应应该包含文本")
}
