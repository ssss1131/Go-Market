package handlers

import (
	"GoCart/internal/service"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

type AddRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req AddRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	err := h.cartService.AddItem(userID, req.ProductID, req.Quantity)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "item added"})
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID := c.GetUint("user_id")
	productID := uint(mustUint(c.Param("product_id")))

	var req AddRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
	}
	err = h.cartService.UpdateItem(userID, productID, req.Quantity)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "item updated"})
}

func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID := c.GetUint("user_id")
	productID := uint(mustUint(c.Param("product_id")))

	err := h.cartService.DeleteItem(userID, productID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "item deleted"})
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	items, _ := h.cartService.GetCart(userID)

	c.JSON(200, gin.H{"items": items})
}

func mustUint(s string) uint {
	var n uint
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		log.Fatal(err)
	}
	return n
}
