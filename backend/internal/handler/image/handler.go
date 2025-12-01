package image

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guixu633/agent/backend/internal/model"
	"github.com/guixu633/agent/backend/internal/service/image"
	"github.com/guixu633/agent/backend/pkg/response"
)

// Handler 图片处理器
type Handler struct {
	imageService *image.Service
}

// NewHandler 创建图片处理器实例
func NewHandler(imageService *image.Service) *Handler {
	return &Handler{
		imageService: imageService,
	}
}

// Generate 生成图片
// @Summary 生成图片
// @Description 根据提示词和输入图片生成新图片
// @Tags image
// @Accept json
// @Produce json
// @Param request body model.ImageGenerateRequest true "生成请求"
// @Success 200 {object} response.Response{data=model.ImageGenerateResponse}
// @Router /api/image/generate [post]
func (h *Handler) Generate(c *gin.Context) {
	var req model.ImageGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}

	// 调用服务层
	result, err := h.imageService.GenerateImage(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, "生成图片失败: "+err.Error())
		return
	}

	response.Success(c, result)
}
