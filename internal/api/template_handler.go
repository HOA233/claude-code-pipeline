package api

import (
	"net/http"

	"github.com/company/claude-pipeline/internal/service"
	"github.com/gin-gonic/gin"
)

// TemplateHandler 模板 API 处理器
type TemplateHandler struct {
	templateSvc *service.TemplateService
}

// NewTemplateHandler 创建模板 Handler
func NewTemplateHandler(templateSvc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateSvc: templateSvc}
}

// GetBuiltInTemplates 获取内置模板
func (h *TemplateHandler) GetBuiltInTemplates(c *gin.Context) {
	templates := h.templateSvc.GetBuiltInTemplates()
	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     len(templates),
	})
}

// ListCustomTemplates 列出自定义模板
func (h *TemplateHandler) ListCustomTemplates(c *gin.Context) {
	templates, err := h.templateSvc.ListCustomTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     len(templates),
	})
}

// InstantiateTemplate 实例化模板
func (h *TemplateHandler) InstantiateTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	workflow, err := h.templateSvc.InstantiateTemplate(c.Request.Context(), templateID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if workflow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

// SaveCustomTemplate 保存自定义模板
func (h *TemplateHandler) SaveCustomTemplate(c *gin.Context) {
	var template service.WorkflowTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if template.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if err := h.templateSvc.SaveCustomTemplate(c.Request.Context(), &template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// DeleteCustomTemplate 删除自定义模板
func (h *TemplateHandler) DeleteCustomTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := h.templateSvc.DeleteCustomTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}