# 快速开始指南

## 🚀 30秒启动项目

```bash
# 1. 克隆项目（如果还没有）
git clone <your-repo-url>
cd agent

# 2. 安装依赖
make install

# 3. 配置 GCP 密钥
# 将你的 GCP 服务账号密钥文件放到：
# backend/configs/gcp/gcp.json

# 4. 启动开发环境
make dev
```

访问 http://localhost:5173 开始使用！

## 📋 常用命令

### 基础命令
```bash
make help      # 查看所有命令
make info      # 查看项目信息
make status    # 检查服务状态
```

### 开发命令
```bash
make dev       # 启动前后端（推荐）
make backend   # 只启动后端
make frontend  # 只启动前端
```

### 测试命令
```bash
make test                    # 运行所有测试
make test-backend           # 运行后端测试
make test-backend-coverage  # 生成测试覆盖率报告
```

### 构建命令
```bash
make build         # 构建前后端
make build-backend # 构建后端
make build-frontend # 构建前端
```

### 维护命令
```bash
make clean         # 清理构建产物
make clean-all     # 深度清理（包括缓存）
make deps-update   # 更新依赖
make lint-backend  # 检查后端代码
make lint-frontend # 检查前端代码
```

## 🎯 使用场景

### 场景1：首次启动项目
```bash
make install  # 安装依赖
make dev      # 启动服务
```

### 场景2：只开发前端
```bash
make backend &  # 后台启动后端
make frontend   # 前台启动前端（可以看到日志）
```

### 场景3：只开发后端
```bash
make backend    # 启动后端，使用 Postman 测试 API
```

### 场景4：运行测试
```bash
make test                    # 快速测试
make test-backend-coverage   # 详细测试 + 覆盖率
open backend/coverage.html   # 查看覆盖率报告
```

### 场景5：清理项目
```bash
make clean      # 清理构建产物
make clean-all  # 深度清理（释放更多空间）
```

## 🔧 故障排查

### 后端无法启动
```bash
# 检查 GCP 配置文件
ls -la backend/configs/gcp/gcp.json

# 检查端口是否被占用
lsof -i :8080

# 查看详细错误
cd backend && go run cmd/server/main.go
```

### 前端无法启动
```bash
# 重新安装依赖
cd frontend
rm -rf node_modules package-lock.json
npm install

# 检查端口是否被占用
lsof -i :5173
```

### 依赖安装失败
```bash
# Go 依赖问题
cd backend
go clean -modcache
go mod download

# Node 依赖问题
cd frontend
npm cache clean --force
npm install
```

## 📝 开发工作流

### 日常开发
1. `make dev` - 启动开发环境
2. 修改代码
3. 浏览器自动刷新（前端）
4. `make test` - 运行测试
5. `git add .` 和 `git commit`

### 提交前检查
```bash
make lint-backend   # 检查后端代码
make lint-frontend  # 检查前端代码
make test          # 运行所有测试
```

### 定期维护
```bash
make deps-update   # 更新依赖
make clean-all     # 深度清理
make install       # 重新安装
```

## 🎨 Image 服务使用

1. 访问 http://localhost:5173
2. 输入提示词，例如：
   - "一只可爱的羊毛毡小猫咪"
   - "帮我把图片修改为羊毛毡的可爱风格"
3. （可选）上传参考图片
4. 点击"生成图片"
5. 等待 10-30 秒
6. 下载生成的图片

## 💡 提示

- 使用 `make dev` 时，按 `Ctrl+C` 会同时停止前后端服务
- 建议在两个终端分别运行 `make backend` 和 `make frontend`，方便查看各自的日志
- 使用 `make status` 快速检查服务是否正常运行
- 首次启动可能需要下载依赖，请耐心等待

## 🔗 更多信息

- [项目 README](./README.md) - 项目总览和技术栈
- [后端文档](./backend/README.md) - 后端开发说明
- [前端文档](./frontend/README.md) - 前端开发说明
- [Makefile](./Makefile) - 完整的命令列表

