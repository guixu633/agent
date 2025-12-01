# Agent 项目

一个前后端分离的 AI Agent 应用，支持多种 LLM 模型集成。

## 项目结构

```
agent/
├── backend/              # 后端服务 (Go)
│   ├── cmd/
│   │   └── server/       # 主程序入口
│   ├── internal/
│   │   ├── handler/      # HTTP 处理层
│   │   │   └── image/    # Image API handlers
│   │   ├── service/      # 业务逻辑层
│   │   │   └── image/    # Image 服务
│   │   ├── model/        # 数据模型
│   │   └── llm/          # LLM 客户端
│   │       ├── claude/   # Anthropic Claude
│   │       └── gemini/   # Google Gemini
│   ├── pkg/
│   │   └── response/     # 统一响应格式
│   ├── configs/          # 配置文件
│   ├── go.mod
│   └── go.sum
├── frontend/             # 前端应用 (React + TypeScript)
│   ├── src/
│   │   ├── services/     # API 服务层
│   │   │   └── image/    # Image API 客户端
│   │   ├── pages/        # 页面组件
│   │   │   └── ImageGenerator/
│   │   ├── types/        # TypeScript 类型定义
│   │   └── utils/        # 工具函数
│   ├── package.json
│   └── vite.config.ts
├── DEPLOYMENT.md         # 部署指南
└── README.md             # 项目说明文档
```

## 技术栈

### 后端
- **语言**: Go 1.24+
- **LLM SDK**: 
  - Google GenAI SDK (`google.golang.org/genai`)
  - Anthropic SDK (`github.com/anthropics/anthropic-sdk-go`)
- **测试框架**: testify

### 前端
- **框架**: React 18
- **语言**: TypeScript
- **构建工具**: Vite
- **HTTP 客户端**: Axios
- **样式**: CSS Modules

## 快速开始

### 使用 Makefile（推荐）

```bash
# 查看所有可用命令
make help

# 安装所有依赖
make install

# 启动开发环境（前后端同时运行）
make dev

# 或者分别启动
make backend   # 启动后端服务 (http://localhost:8080)
make frontend  # 启动前端服务 (http://localhost:5173)

# 运行测试
make test

# 构建项目
make build
```

> 💡 详细的使用指南请参考 [QUICKSTART.md](./QUICKSTART.md)

### 手动启动

**后端：**
```bash
cd backend
go run cmd/server/main.go
# 服务运行在 http://localhost:8080
```

**前端：**
```bash
cd frontend
npm install
npm run dev
# 服务运行在 http://localhost:5173
```

### 后端开发

```bash
# 进入后端目录
cd backend

# 安装依赖
go mod download

# 运行测试
go test ./...

# 运行服务器
go run cmd/server/main.go
```

### 配置文件

后端需要配置 GCP 服务账号密钥：

```bash
# 将 GCP 密钥文件放在以下位置
backend/configs/gcp/gcp.json
```

## 功能特性

### 已实现

#### Image 服务 🎨
- ✅ **AI 图片生成**：基于 Gemini 3 Pro Image Preview 模型
- ✅ **图片编辑**：上传参考图片 + 提示词进行风格转换
- ✅ **前后端分离架构**：RESTful API + React 前端
- ✅ **实时预览**：生成结果即时展示和下载

#### LLM 集成
- ✅ Google Gemini 2.5 Pro 文本生成
- ✅ Google Gemini 3 Pro 文本生成
- ✅ Google Gemini 3 Pro Image Preview 图片处理
- ✅ 图片风格转换（如羊毛毡风格）
- ✅ Google Search 集成
- ✅ Anthropic Claude 集成

### 计划中
- ⏳ 更多应用服务（聊天、文档处理等）
- ⏳ 用户认证和权限管理
- ⏳ 历史记录和收藏功能
- ⏳ 更多 LLM 模型支持

## 开发指南

### 添加新的 LLM 集成

1. 在 `backend/internal/llm/` 下创建新目录
2. 实现 LLM 客户端封装
3. 编写单元测试
4. 更新文档

## 许可证

待定

