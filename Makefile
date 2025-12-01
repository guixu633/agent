.PHONY: help install dev backend frontend test clean

# 默认目标
.DEFAULT_GOAL := help

# 颜色输出
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

help: ## 显示帮助信息
	@echo "$(CYAN)可用命令：$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'

install: install-backend install-frontend ## 安装所有依赖
	@echo "$(GREEN)✓ 所有依赖安装完成$(RESET)"

install-backend: ## 安装后端依赖
	@echo "$(CYAN)正在安装后端依赖...$(RESET)"
	@cd backend && go mod download && go mod tidy
	@echo "$(GREEN)✓ 后端依赖安装完成$(RESET)"

install-frontend: ## 安装前端依赖
	@echo "$(CYAN)正在安装前端依赖...$(RESET)"
	@cd frontend && npm install
	@echo "$(GREEN)✓ 前端依赖安装完成$(RESET)"

dev: ## 启动开发环境（前后端）
	@echo "$(CYAN)启动开发环境...$(RESET)"
	@echo "$(YELLOW)提示：使用 Ctrl+C 停止所有服务$(RESET)"
	@trap 'kill 0' EXIT; \
	make backend & \
	make frontend & \
	wait

backend: ## 启动后端服务
	@echo "$(CYAN)启动后端服务 (http://localhost:8080)...$(RESET)"
	@cd backend && go run cmd/server/main.go

frontend: ## 启动前端服务
	@echo "$(CYAN)启动前端服务 (http://localhost:5173)...$(RESET)"
	@cd frontend && npm run dev

test: test-backend ## 运行所有测试
	@echo "$(GREEN)✓ 所有测试完成$(RESET)"

test-backend: ## 运行后端测试
	@echo "$(CYAN)运行后端测试...$(RESET)"
	@cd backend && go test ./... -v

test-backend-coverage: ## 运行后端测试并生成覆盖率报告
	@echo "$(CYAN)运行后端测试（带覆盖率）...$(RESET)"
	@cd backend && go test ./... -v -coverprofile=coverage.out
	@cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ 覆盖率报告已生成：backend/coverage.html$(RESET)"

build: build-backend build-frontend ## 构建前后端
	@echo "$(GREEN)✓ 构建完成$(RESET)"

build-backend: ## 构建后端二进制文件
	@echo "$(CYAN)构建后端...$(RESET)"
	@cd backend && go build -o bin/server cmd/server/main.go
	@echo "$(GREEN)✓ 后端构建完成：backend/bin/server$(RESET)"

build-frontend: ## 构建前端静态文件
	@echo "$(CYAN)构建前端...$(RESET)"
	@cd frontend && npm run build
	@echo "$(GREEN)✓ 前端构建完成：frontend/dist$(RESET)"

clean: ## 清理构建产物和依赖
	@echo "$(CYAN)清理构建产物...$(RESET)"
	@rm -rf backend/bin
	@rm -rf backend/coverage.out backend/coverage.html
	@rm -rf frontend/dist
	@rm -rf frontend/node_modules
	@echo "$(GREEN)✓ 清理完成$(RESET)"

clean-all: clean ## 深度清理（包括 Go 缓存）
	@echo "$(CYAN)深度清理...$(RESET)"
	@cd backend && go clean -cache -modcache -testcache
	@echo "$(GREEN)✓ 深度清理完成$(RESET)"

lint-backend: ## 运行后端代码检查
	@echo "$(CYAN)检查后端代码...$(RESET)"
	@cd backend && go vet ./...
	@cd backend && go fmt ./...
	@echo "$(GREEN)✓ 后端代码检查完成$(RESET)"

lint-frontend: ## 运行前端代码检查
	@echo "$(CYAN)检查前端代码...$(RESET)"
	@cd frontend && npm run lint
	@echo "$(GREEN)✓ 前端代码检查完成$(RESET)"

logs: ## 查看服务日志（需要服务正在运行）
	@echo "$(CYAN)查看服务日志...$(RESET)"
	@echo "$(YELLOW)提示：使用 Ctrl+C 退出$(RESET)"

status: ## 检查服务状态
	@echo "$(CYAN)检查服务状态...$(RESET)"
	@echo -n "后端服务 (8080): "
	@curl -s http://localhost:8080/health > /dev/null && echo "$(GREEN)运行中$(RESET)" || echo "$(RED)未运行$(RESET)"
	@echo -n "前端服务 (5173): "
	@curl -s http://localhost:5173 > /dev/null && echo "$(GREEN)运行中$(RESET)" || echo "$(RED)未运行$(RESET)"

deps-update: ## 更新所有依赖
	@echo "$(CYAN)更新依赖...$(RESET)"
	@cd backend && go get -u ./... && go mod tidy
	@cd frontend && npm update
	@echo "$(GREEN)✓ 依赖更新完成$(RESET)"

info: ## 显示项目信息
	@echo "$(CYAN)项目信息：$(RESET)"
	@echo "  名称: AI Agent"
	@echo "  后端: Go + Gin"
	@echo "  前端: React + TypeScript + Vite"
	@echo ""
	@echo "$(CYAN)目录结构：$(RESET)"
	@echo "  backend/  - 后端服务"
	@echo "  frontend/ - 前端应用"
	@echo ""
	@echo "$(CYAN)服务地址：$(RESET)"
	@echo "  后端 API: http://localhost:8080"
	@echo "  前端页面: http://localhost:5173"
	@echo "  健康检查: http://localhost:8080/health"
	@echo ""
	@echo "$(CYAN)快速开始：$(RESET)"
	@echo "  1. make install  - 安装依赖"
	@echo "  2. make dev      - 启动开发环境"
	@echo ""
	@echo "$(CYAN)更多命令请运行: make help$(RESET)"

