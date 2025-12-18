package handlers

import (
	"GoProduct/internal/http/middleware"
	"GoProduct/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartSvc *service.CartService
}

func NewCartHandler(cartSvc *service.CartService) *CartHandler {
	return &CartHandler{cartSvc: cartSvc}
}

type addItemReq struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type updateItemReq struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	var req addItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.cartSvc.AddItem(userID.(uint), req.ProductID, req.Quantity)
	if err != nil {
		if err == service.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item added"})
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil || productID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product_id"})
		return
	}

	var req updateItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.cartSvc.UpdateItem(userID.(uint), uint(productID), req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item updated"})
}

func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil || productID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product_id"})
		return
	}

	err = h.cartSvc.DeleteItem(userID.(uint), uint(productID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, ok := c.Get(middleware.UserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	items, err := h.cartSvc.GetCart(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
