package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guixu633/agent/backend/internal/config"
	"github.com/guixu633/agent/backend/internal/database"
	imageHandler "github.com/guixu633/agent/backend/internal/handler/image"
	workspaceHandler "github.com/guixu633/agent/backend/internal/handler/workspace"
	"github.com/guixu633/agent/backend/internal/oss"
	"github.com/guixu633/agent/backend/internal/repository"
	imageService "github.com/guixu633/agent/backend/internal/service/image"
	workspaceService "github.com/guixu633/agent/backend/internal/service/workspace"
	"google.golang.org/genai"
)

const (
	DefaultRegion  = "global"
	DefaultProject = "startup-x-465606"
)

func main() {
	// 初始化 GenAI 客户端
	genaiClient, err := initGenAIClient()
	if err != nil {
		log.Fatalf("初始化 GenAI 客户端失败: %v", err)
	}

	// 加载统一配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.json"
	}
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化 OSS 客户端（从统一配置加载）
	ossClient, err := oss.NewClient(configPath)
	if err != nil {
		log.Fatalf("初始化 OSS 客户端失败: %v", err)
	}

	// 初始化数据库连接
	dsn := appConfig.Postgres.GetDSN()
	if err := database.InitDB(dsn); err != nil {
		log.Fatalf("初始化数据库连接失败: %v", err)
	}
	defer database.CloseDB()

	// 运行数据库迁移
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移完成")

	// 初始化 Repository 层
	workspaceRepo := repository.NewWorkspaceRepository()
	imageRepo := repository.NewImageRepository()

	// 初始化服务层
	imgService := imageService.NewService(genaiClient, ossClient, imageRepo, workspaceRepo)
	wsService := workspaceService.NewService(ossClient, workspaceRepo)

	// 初始化处理器层
	imgHandler := imageHandler.NewHandler(imgService)
	wsHandler := workspaceHandler.NewHandler(wsService)

	// 创建 Gin 路由
	r := gin.Default()

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 注册路由
	api := r.Group("/api")
	{
		// 工作区相关接口
		api.GET("/workspace", wsHandler.List)               // 列出所有工作区
		api.POST("/workspace", wsHandler.Create)            // 创建工作区
		api.DELETE("/workspace", wsHandler.Delete)          // 删除工作区
		api.GET("/workspace/current", wsHandler.GetCurrent) // 获取当前工作区
		api.PUT("/workspace/current", wsHandler.SetCurrent) // 设置当前工作区（切换工作区）
		api.POST("/workspace/switch", wsHandler.SetCurrent) // 切换工作区（别名，与 PUT /workspace/current 相同）

		// 图片相关接口
		imageGroup := api.Group("/image")
		{
			imageGroup.GET("/list", imgHandler.List)          // 列出工作区图片接口
			imageGroup.POST("/upload", imgHandler.Upload)     // 图片上传接口
			imageGroup.POST("/generate", imgHandler.Generate) // 图片生成接口
			imageGroup.DELETE("", imgHandler.Delete)          // 删除图片接口
			imageGroup.POST("/rename", imgHandler.Rename)     // 重命名图片接口
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// initGenAIClient 初始化 GenAI 客户端
func initGenAIClient() (*genai.Client, error) {
	configPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if configPath == "" {
		configPath = "configs/gcp/gcp.json"
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", configPath)

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  DefaultProject,
		Location: DefaultRegion,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
