package personinfo

import (
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"strings"

	sqlinit "backend_go/SQLinit"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, status int, point string, message string, err error) {
	if err != nil {
		log.Printf("[%s] %v", point, err)
	} else {
		log.Printf("[%s] %s", point, message)
	}
	c.JSON(status, gin.H{"error": message, "error_point": point})
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateAccountRequest struct {
	UserID       uint   `json:"user_id" binding:"required"`
	Username     string `json:"username"`
	AvatarBase64 string `json:"avatar_base64"`
}

type ResetPasswordRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func Register(c *gin.Context) {
	log.Println("[AUTH][REGISTER][S1] bind request")
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "E_AUTH_REGISTER_BIND", "参数错误", err)
		return
	}

	request.Username = strings.TrimSpace(request.Username)
	request.Password = strings.TrimSpace(request.Password)
	if request.Username == "" || request.Password == "" {
		respondError(c, http.StatusBadRequest, "E_AUTH_REGISTER_EMPTY_FIELD", "用户名和密码不能为空", nil)
		return
	}
	log.Printf("[AUTH][REGISTER][S2] create user username=%s", request.Username)

	user := sqlinit.User{
		Username: request.Username,
		Password: request.Password,
	}
	if err := sqlinit.DB.Create(&user).Error; err != nil {
		respondError(c, http.StatusBadRequest, "E_AUTH_REGISTER_CREATE", "用户名已存在", err)
		return
	}
	log.Printf("[AUTH][REGISTER][S3] register success user_id=%d", user.ID)

	c.JSON(http.StatusOK, gin.H{"user_id": user.ID})
}

func Login(c *gin.Context) {
	log.Println("[AUTH][LOGIN][S1] bind request")
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "E_AUTH_LOGIN_BIND", "参数错误", err)
		return
	}

	request.Username = strings.TrimSpace(request.Username)
	request.Password = strings.TrimSpace(request.Password)
	if request.Username == "" || request.Password == "" {
		respondError(c, http.StatusBadRequest, "E_AUTH_LOGIN_EMPTY_FIELD", "用户名和密码不能为空", nil)
		return
	}
	log.Printf("[AUTH][LOGIN][S2] query user username=%s", request.Username)

	var user sqlinit.User
	if err := sqlinit.DB.Where("username = ? AND password = ?", request.Username, request.Password).First(&user).Error; err != nil {
		respondError(c, http.StatusUnauthorized, "E_AUTH_LOGIN_QUERY", "用户名或密码错误", err)
		return
	}
	log.Printf("[AUTH][LOGIN][S3] login success user_id=%d", user.ID)

	c.JSON(http.StatusOK, gin.H{
		"user_id": user.ID,
		"token":   "mock-jwt-token",
	})
}

func GetAccount(c *gin.Context) {
	log.Println("[AUTH][GET_ACCOUNT][S1] parse query user_id")
	userIDText := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDText, 10, 64)
	if err != nil || userID == 0 {
		respondError(c, http.StatusBadRequest, "E_AUTH_GET_ACCOUNT_USER_ID", "user_id参数错误", err)
		return
	}
	log.Printf("[AUTH][GET_ACCOUNT][S2] query user user_id=%d", userID)

	var user sqlinit.User
	if err := sqlinit.DB.First(&user, uint(userID)).Error; err != nil {
		respondError(c, http.StatusNotFound, "E_AUTH_GET_ACCOUNT_QUERY", "用户不存在", err)
		return
	}

	avatar := ""
	if len(user.AvatarData) > 0 {
		log.Printf("[AUTH][GET_ACCOUNT][S3] encode avatar user_id=%d bytes=%d", userID, len(user.AvatarData))
		avatar = "data:image/png;base64," + base64.StdEncoding.EncodeToString(user.AvatarData)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       user.ID,
		"username":      user.Username,
		"avatar_base64": avatar,
	})
}

func UpdateAccount(c *gin.Context) {
	log.Println("[AUTH][UPDATE_ACCOUNT][S1] bind request")
	var request UpdateAccountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "E_AUTH_UPDATE_ACCOUNT_BIND", "参数错误", err)
		return
	}
	log.Printf("[AUTH][UPDATE_ACCOUNT][S2] query user user_id=%d", request.UserID)

	var user sqlinit.User
	if err := sqlinit.DB.First(&user, request.UserID).Error; err != nil {
		respondError(c, http.StatusNotFound, "E_AUTH_UPDATE_ACCOUNT_QUERY", "用户不存在", err)
		return
	}

	updates := map[string]interface{}{}
	username := strings.TrimSpace(request.Username)
	if username != "" {
		updates["username"] = username
	}

	avatarText := strings.TrimSpace(request.AvatarBase64)
	if avatarText != "" {
		if idx := strings.Index(avatarText, ","); idx >= 0 {
			avatarText = avatarText[idx+1:]
		}
		decoded, err := base64.StdEncoding.DecodeString(avatarText)
		if err != nil {
			respondError(c, http.StatusBadRequest, "E_AUTH_UPDATE_ACCOUNT_AVATAR_DECODE", "头像格式错误", err)
			return
		}
		updates["avatar_data"] = decoded
	}

	if len(updates) == 0 {
		respondError(c, http.StatusBadRequest, "E_AUTH_UPDATE_ACCOUNT_EMPTY_UPDATES", "没有可更新的内容", nil)
		return
	}
	log.Printf("[AUTH][UPDATE_ACCOUNT][S3] apply updates user_id=%d fields=%d", request.UserID, len(updates))

	if err := sqlinit.DB.Model(&user).Updates(updates).Error; err != nil {
		respondError(c, http.StatusBadRequest, "E_AUTH_UPDATE_ACCOUNT_DB_UPDATE", "更新失败，用户名可能已存在", err)
		return
	}
	log.Printf("[AUTH][UPDATE_ACCOUNT][S4] update success user_id=%d", request.UserID)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ResetPassword(c *gin.Context) {
	log.Println("[AUTH][RESET_PASSWORD][S1] bind request")
	var request ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "E_AUTH_RESET_PASSWORD_BIND", "参数错误", err)
		return
	}

	if len(strings.TrimSpace(request.NewPassword)) < 6 {
		respondError(c, http.StatusBadRequest, "E_AUTH_RESET_PASSWORD_NEW_PASSWORD", "新密码至少6位", nil)
		return
	}
	log.Printf("[AUTH][RESET_PASSWORD][S2] query user user_id=%d", request.UserID)

	var user sqlinit.User
	if err := sqlinit.DB.First(&user, request.UserID).Error; err != nil {
		respondError(c, http.StatusNotFound, "E_AUTH_RESET_PASSWORD_QUERY", "用户不存在", err)
		return
	}

	if user.Password != request.OldPassword {
		respondError(c, http.StatusBadRequest, "E_AUTH_RESET_PASSWORD_OLD_PASSWORD", "旧密码错误", nil)
		return
	}

	user.Password = request.NewPassword
	if err := sqlinit.DB.Save(&user).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_AUTH_RESET_PASSWORD_SAVE", "重设密码失败", err)
		return
	}
	log.Printf("[AUTH][RESET_PASSWORD][S3] reset success user_id=%d", request.UserID)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
