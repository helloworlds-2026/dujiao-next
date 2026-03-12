package shared

import (
	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

// BuildTelegramBindingResponse 构造 Telegram 绑定信息响应载荷。
func BuildTelegramBindingResponse(identity *models.UserOAuthIdentity) gin.H {
	if identity == nil {
		return gin.H{"bound": false}
	}
	return gin.H{
		"bound":            true,
		"provider":         identity.Provider,
		"provider_user_id": identity.ProviderUserID,
		"username":         identity.Username,
		"avatar_url":       identity.AvatarURL,
		"auth_at":          identity.AuthAt,
		"updated_at":       identity.UpdatedAt,
	}
}

// BuildUserProfilePayload 构造用户资料响应载荷。
func BuildUserProfilePayload(user *models.User, emailMode, passwordMode string) gin.H {
	if user == nil {
		return gin.H{}
	}
	return gin.H{
		"id":                   user.ID,
		"email":                user.Email,
		"nickname":             user.DisplayName,
		"email_verified_at":    user.EmailVerifiedAt,
		"locale":               user.Locale,
		"email_change_mode":    emailMode,
		"password_change_mode": passwordMode,
	}
}
