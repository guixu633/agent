package image

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

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

// Upload 上传图片
// @Summary 上传图片
// @Description 上传图片到 OSS，返回路径和 URL
// @Tags image
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Param workspace formData string true "工作区名称"
// @Success 200 {object} response.Response{data=model.ImageUploadResponse}
// @Router /api/image/upload [post]
func (h *Handler) Upload(c *gin.Context) {
	// 获取工作区名称
	workspace := c.PostForm("workspace")
	if workspace == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "工作区名称不能为空")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "获取上传文件失败: "+err.Error())
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "打开文件失败: "+err.Error())
		return
	}
	defer src.Close()

	// 获取文件名
	filename := file.Filename
	if filename == "" {
		// 如果没有文件名，生成一个基于时间戳的文件名
		filename = fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	} else {
		// 使用原始文件名，但确保只使用文件名部分（去除路径）
		filename = filepath.Base(filename)
	}

	// 调用服务层上传
	result, err := h.imageService.UploadImage(c.Request.Context(), src, filename, workspace)
	if err != nil {
		response.Error(c, 500, "上传图片失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// Generate 生成图片
// @Summary 生成图片
// @Description 根据提示词和输入图片路径生成新图片
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

// List 列出工作区的所有图片
// @Summary 列出工作区图片
// @Description 获取指定工作区的所有图片列表
// @Tags image
// @Produce json
// @Param workspace query string true "工作区名称"
// @Success 200 {object} response.Response{data=model.ListWorkspaceImagesResponse}
// @Router /api/image/list [get]
func (h *Handler) List(c *gin.Context) {
	workspace := c.Query("workspace")
	if workspace == "" {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "工作区名称不能为空")
		return
	}

	// 调用服务层
	result, err := h.imageService.ListWorkspaceImages(c.Request.Context(), workspace)
	if err != nil {
		response.Error(c, 500, "获取图片列表失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// Delete 删除图片
// @Summary 删除图片
// @Description 删除指定的图片
// @Tags image
// @Accept json
// @Produce json
// @Param request body model.DeleteImageRequest true "删除图片请求"
// @Success 200 {object} response.Response
// @Router /api/image [delete]
func (h *Handler) Delete(c *gin.Context) {
	var req model.DeleteImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}

	// 调用服务层
	err := h.imageService.DeleteImage(c.Request.Context(), req.Path)
	if err != nil {
		response.Error(c, 500, "删除图片失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// Rename 重命名图片
// @Summary 重命名图片
// @Description 重命名指定的图片
// @Tags image
// @Accept json
// @Produce json
// @Param request body model.RenameImageRequest true "重命名图片请求"
// @Success 200 {object} response.Response{data=model.RenameImageResponse}
// @Router /api/image/rename [post]
func (h *Handler) Rename(c *gin.Context) {
	var req model.RenameImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}

	// 调用服务层
	result, err := h.imageService.RenameImage(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, "重命名图片失败: "+err.Error())
		return
	}

	response.Success(c, result)
}
