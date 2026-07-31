package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"magicbox/internal/auth"
	"magicbox/internal/config"
)

type AuthController struct {
	cfg     *config.Config
	manager *auth.Manager
}

func NewAuthController(cfg *config.Config, manager *auth.Manager) *AuthController {
	return &AuthController{cfg: cfg, manager: manager}
}

// SystemInfoPublic is the unauthenticated "is this a real server" endpoint
// Jellyfin clients (e.g. magicbox-appletv's AddServerView) call before
// showing a login screen.
func (a *AuthController) SystemInfoPublic(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ServerName": "MagicBox",
		"Version":    "1.0.0",
		"Id":         "magicbox",
	})
}

type authenticateByNameRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

// AuthenticateByName is Jellyfin's login endpoint. Username is accepted and
// echoed back but never validated -- MagicBox has one shared password, not
// per-user accounts.
func (a *AuthController) AuthenticateByName(c *gin.Context) {
	var req authenticateByNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and Pw are required"})
		return
	}

	if !auth.CheckPassword(a.cfg.Auth.PasswordHash, req.Pw) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token, _, err := a.manager.IssueToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	username := req.Username
	if username == "" {
		username = "magicbox"
	}

	c.JSON(http.StatusOK, gin.H{
		"AccessToken": token,
		"ServerId":    "magicbox",
		"User": gin.H{
			"Id":   "1",
			"Name": username,
		},
	})
}
