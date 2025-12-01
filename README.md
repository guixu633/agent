# Agent 项目

一个前后端分离的 AI Agent 应用，支持多种 LLM 模型集成。

## 项目结构

```
agent/
├── backend/          # 后端服务 (Go)
│   ├── configs/      # 配置文件
│   ├── internal/     # 内部包
│   │   └── llm/      # LLM 集成模块
│   │       ├── claude/   # Anthropic Claude 集成
│   │       └── gemini/   # Google Gemini 集成
│   ├── go.mod
│   └── go.sum
├── frontend/         # 前端应用 (待实现)
└── README.md         # 项目说明文档
```

## 技术栈

### 后端
- **语言**: Go 1.24+
- **LLM SDK**: 
  - Google GenAI SDK (`google.golang.org/genai`)
  - Anthropic SDK (`github.com/anthropics/anthropic-sdk-go`)
- **测试框架**: testify

### 前端
- 待定 (可选 React / Vue / Next.js 等)

## 快速开始

### 后端开发

```bash
# 进入后端目录
cd backend

# 安装依赖
go mod download

# 运行测试
go test ./...

# 运行特定测试
go test -v ./internal/llm/gemini -run TestNanoBananaImageStyle
```

### 配置文件

后端需要配置 GCP 服务账号密钥：

```bash
# 将 GCP 密钥文件放在以下位置
backend/configs/gcp/gcp.json
```

## 功能特性

### 已实现
- ✅ Google Gemini 2.5 Pro 文本生成
- ✅ Google Gemini 3 Pro 文本生成
- ✅ Google Gemini 3 Pro Image Preview 图片处理
- ✅ 图片风格转换（如羊毛毡风格）
- ✅ Google Search 集成
- ✅ Anthropic Claude 集成

### 计划中
- ⏳ RESTful API 服务
- ⏳ 前端界面
- ⏳ 更多 LLM 模型支持

## 开发指南

### 添加新的 LLM 集成

1. 在 `backend/internal/llm/` 下创建新目录
2. 实现 LLM 客户端封装
3. 编写单元测试
4. 更新文档

## 许可证

待定

