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
		"ServerName":             "MagicBoxie",
		"Version":                "10.11.6",
		"ProductName":            "Jellyfin Server",
		"OperatingSystem":        "Linux",
		"StartupWizardCompleted": true,
		"Id":                     "magicbox",
	})
}

// PublicUsers returns the single shared account in Jellyfin's UserDto shape.
// Clients use this endpoint to populate the login screen.
func (a *AuthController) PublicUsers(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{jellyfinUser("admin")})
}

// CurrentUser returns the authenticated shared account. A complete-enough
// UserDto is important here: generated Jellyfin clients decode several of
// these fields as non-optional values.
func (a *AuthController) CurrentUser(c *gin.Context) {
	c.JSON(http.StatusOK, jellyfinUser("admin"))
}

func jellyfinUser(name string) gin.H {
	return gin.H{
		"Name":                      name,
		"ServerId":                  "magicbox",
		"Id":                        "1",
		"HasPassword":               true,
		"HasConfiguredPassword":     true,
		"HasConfiguredEasyPassword": false,
		"EnableAutoLogin":           false,
		"Configuration": gin.H{
			"PlayDefaultAudioTrack":      true,
			"SubtitleMode":               "Default",
			"DisplayMissingEpisodes":     false,
			"GroupedFolders":             []string{},
			"OrderedViews":               []string{},
			"LatestItemsExcludes":        []string{},
			"MyMediaExcludes":            []string{},
			"HidePlayedInLatest":         true,
			"RememberAudioSelections":    true,
			"RememberSubtitleSelections": true,
		},
		"Policy": gin.H{
			"IsAdministrator":                true,
			"IsHidden":                       false,
			"IsDisabled":                     false,
			"EnableAllFolders":               true,
			"EnableAllDevices":               true,
			"EnableContentDeletion":          true,
			"EnableContentDownloading":       true,
			"EnableMediaPlayback":            true,
			"EnableAudioPlaybackTranscoding": true,
			"EnableVideoPlaybackTranscoding": true,
			"EnablePlaybackRemuxing":         true,
		},
	}
}

type authenticateByNameRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

// AuthenticateByName is Jellyfin's login endpoint. Username is accepted and
// echoed back but never validated -- MagicBoxie has one shared password, not
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
		"User":        jellyfinUser(username),
	})
}
