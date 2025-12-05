package handlers

import (
	"GoProduct/internal/http/middleware"
	"GoProduct/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

type createProductReq struct {
	Name        string  `json:"name" binding:"required,max=255"`
	Description string  `json:"description" binding:"max=1000"`
	Price       float64 `json:"price" binding:"required,gt=0"`
}

type productResp struct {
	ID          uint    `json:"id"`
	UserID      uint    `json:"user_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func (h *ProductHandler) Create(c *gin.Context) {
	userIDRaw, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id type"})
		return
	}

	var req createProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.CreateProductInput{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	p, err := h.svc.CreateProduct(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.JSON(http.StatusCreated, productResp{
		ID:          p.ID,
		UserID:      p.UserID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	})
}

func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.svc.ListProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	resp := make([]productResp, 0, len(products))
	for _, p := range products {
		resp = append(resp, productResp{
			ID:          p.ID,
			UserID:      p.UserID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProductHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	p, err := h.svc.GetProduct(uint(id))
	if err != nil {
		if service.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.JSON(http.StatusOK, productResp{
		ID:          p.ID,
		UserID:      p.UserID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	})
}

type updateProductReq struct {
	Name        string  `json:"name" binding:"required,max=255"`
	Description string  `json:"description" binding:"max=1000"`
	Price       float64 `json:"price" binding:"required,gt=0"`
}

func (h *ProductHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := service.UpdateProductInput{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	p, err := h.svc.UpdateProduct(input)
	if err != nil {
		if service.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.JSON(http.StatusOK, productResp{
		ID:          p.ID,
		UserID:      p.UserID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteProduct(uint(id)); err != nil {
		if service.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.Status(http.StatusNoContent)
}
