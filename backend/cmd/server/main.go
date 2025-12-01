package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	imageHandler "github.com/guixu633/agent/backend/internal/handler/image"
	imageService "github.com/guixu633/agent/backend/internal/service/image"
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

	// 初始化服务层
	imgService := imageService.NewService(genaiClient)

	// 初始化处理器层
	imgHandler := imageHandler.NewHandler(imgService)

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
		imageGroup := api.Group("/image")
		{
			imageGroup.POST("/generate", imgHandler.Generate)
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
