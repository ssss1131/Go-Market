package handlers

import (
	"GoProduct/internal/domain"
	"GoProduct/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

type createCategoryReq struct {
	Name string `json:"name" binding:"required,max=255"`
	Slug string `json:"slug" binding:"max=255"`
}

type updateCategoryReq struct {
	Name string `json:"name" binding:"required,max=255"`
	Slug string `json:"slug" binding:"max=255"`
}

type categoryResp struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req createCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.CreateCategoryInput{
		Name: req.Name,
		Slug: req.Slug,
	}

	cat, err := h.svc.CreateCategory(input)
	if err != nil {
		if service.IsSlugTaken(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "slug already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.JSON(http.StatusCreated, toCategoryResp(*cat))
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.svc.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	resp := make([]categoryResp, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, toCategoryResp(cat))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CategoryHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	cat, err := h.svc.GetCategory(uint(id))
	if err != nil {
		if service.IsCategoryNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.JSON(http.StatusOK, toCategoryResp(*cat))
}

func (h *CategoryHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.UpdateCategoryInput{
		ID:   uint(id),
		Name: req.Name,
		Slug: req.Slug,
	}

	cat, err := h.svc.UpdateCategory(input)
	if err != nil {
		if service.IsCategoryNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if service.IsSlugTaken(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "slug already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.JSON(http.StatusOK, toCategoryResp(*cat))
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteCategory(uint(id)); err != nil {
		if service.IsCategoryNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.Status(http.StatusNoContent)
}

func toCategoryResp(c domain.Category) categoryResp {
	return categoryResp{
		ID:   c.ID,
		Name: c.Name,
		Slug: c.Slug,
	}
}
