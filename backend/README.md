# Backend

Go 语言实现的后端服务，集成多种 LLM 模型。

## 项目结构

```
backend/
├── configs/           # 配置文件目录
│   └── gcp/
│       └── gcp.json   # GCP 服务账号密钥 (gitignore)
├── internal/          # 内部包
│   └── llm/          # LLM 集成模块
│       ├── claude/   # Anthropic Claude 集成
│       │   ├── claude.go
│       │   └── claude_test.go
│       └── gemini/   # Google Gemini 集成
│           ├── gcp.go
│           ├── gcp_test.go
│           ├── banana_test.go
│           └── data/  # 测试数据 (图片等)
├── go.mod
└── go.sum
```

## 依赖

- Go 1.24+
- google.golang.org/genai v1.36.0
- github.com/anthropics/anthropic-sdk-go v1.17.0
- github.com/stretchr/testify v1.9.0

## 配置

### GCP 认证

创建 `configs/gcp/gcp.json` 文件，内容为 GCP 服务账号密钥：

```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "...",
  "client_email": "...",
  ...
}
```

## 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test -v ./internal/llm/gemini

# 运行特定测试用例
go test -v ./internal/llm/gemini -run TestNanoBananaImageStyle
```

## API 文档

待实现 - 将会提供 RESTful API 接口。

## 开发说明

### 添加新的 LLM 模型

1. 在 `internal/llm/` 创建新目录
2. 实现客户端初始化和 API 调用
3. 编写测试用例
4. 更新文档

