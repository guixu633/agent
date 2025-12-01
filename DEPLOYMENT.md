# 部署指南

## 快速开始

### 1. 后端启动

```bash
# 进入后端目录
cd backend

# 确保配置文件存在
# 将 GCP 服务账号密钥放在: backend/configs/gcp/gcp.json

# 安装依赖（如果还没安装）
go mod download

# 启动后端服务
go run cmd/server/main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 2. 前端启动

```bash
# 打开新终端，进入前端目录
cd frontend

# 安装依赖（如果还没安装）
npm install

# 启动开发服务器
npm run dev
```

前端服务将在 `http://localhost:5173` 启动。

## 环境变量

### 后端环境变量

创建 `backend/.env` 文件（可选）：

```bash
# 服务器端口
PORT=8080

# GCP 配置文件路径
GOOGLE_APPLICATION_CREDENTIALS=configs/gcp/gcp.json
```

### 前端环境变量

创建 `frontend/.env` 文件（可选）：

```bash
# API 基础 URL（开发环境使用代理，生产环境需要配置实际地址）
VITE_API_BASE_URL=/api
```

## 访问应用

1. 打开浏览器访问 `http://localhost:5173`
2. 输入提示词（例如：帮我把图片修改为羊毛毡的可爱风格）
3. 可选：上传参考图片
4. 点击"生成图片"按钮
5. 等待生成完成（通常需要 10-30 秒）

## 生产部署

### 后端部署

```bash
cd backend

# 编译
go build -o server cmd/server/main.go

# 运行
PORT=8080 ./server
```

### 前端部署

```bash
cd frontend

# 构建
npm run build

# dist 目录包含构建产物，可以部署到任何静态服务器
```

## 故障排查

### 后端问题

1. **GCP 认证失败**
   - 确保 `configs/gcp/gcp.json` 文件存在且有效
   - 检查 GCP 项目 ID 和区域配置是否正确

2. **端口被占用**
   - 修改 `PORT` 环境变量为其他端口

### 前端问题

1. **无法连接后端**
   - 确保后端服务已启动
   - 检查 Vite 代理配置（`vite.config.ts`）
   - 确认 CORS 配置正确

2. **图片上传失败**
   - 检查图片格式（支持 PNG、JPG、JPEG）
   - 确认图片大小合理（建议 < 5MB）

