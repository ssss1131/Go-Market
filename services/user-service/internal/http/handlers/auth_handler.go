package handlers

import (
	"GoMarket/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerReq struct {
	Name     string `json:"name" binding:"required,max=255"`
	Surname  string `json:"surname" binding:"required,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type registerResp struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input := service.RegisterInput{
		Name:     req.Name,
		Surname:  req.Surname,
		Email:    req.Email,
		Password: req.Password,
	}
	out, err := h.auth.Register(input)
	if err != nil {
		if service.IsEmailTaken(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already taken!"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	c.JSON(http.StatusCreated, registerResp{
		UserID: out.UserID.String(),
		Status: string(out.Status),
	})

}
