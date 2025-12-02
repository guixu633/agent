package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guixu633/agent/backend/internal/model"
	"github.com/guixu633/agent/backend/internal/service/workspace"
	"github.com/guixu633/agent/backend/pkg/response"
)

// Handler 工作区处理器
type Handler struct {
	workspaceService *workspace.Service
}

// NewHandler 创建工作区处理器实例
func NewHandler(workspaceService *workspace.Service) *Handler {
	return &Handler{
		workspaceService: workspaceService,
	}
}

// List 列出所有工作区
// @Summary 列出所有工作区
// @Description 获取所有工作区列表
// @Tags workspace
// @Produce json
// @Success 200 {object} response.Response{data=model.ListWorkspacesResponse}
// @Router /api/workspace [get]
func (h *Handler) List(c *gin.Context) {
	result, err := h.workspaceService.ListWorkspaces(c.Request.Context())
	if err != nil {
		response.Error(c, 500, "获取工作区列表失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// Create 创建工作区
// @Summary 创建工作区
// @Description 创建新的工作区
// @Tags workspace
// @Accept json
// @Produce json
// @Param request body model.CreateWorkspaceRequest true "创建工作区请求"
// @Success 200 {object} response.Response{data=model.CreateWorkspaceResponse}
// @Router /api/workspace [post]
func (h *Handler) Create(c *gin.Context) {
	var req model.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.workspaceService.CreateWorkspace(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, "创建工作区失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// Delete 删除工作区
// @Summary 删除工作区
// @Description 删除指定的工作区及其所有文件
// @Tags workspace
// @Accept json
// @Produce json
// @Param request body model.DeleteWorkspaceRequest true "删除工作区请求"
// @Success 200 {object} response.Response
// @Router /api/workspace [delete]
func (h *Handler) Delete(c *gin.Context) {
	var req model.DeleteWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithStatus(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}

	err := h.workspaceService.DeleteWorkspace(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, "删除工作区失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

