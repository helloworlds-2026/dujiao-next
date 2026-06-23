package router

import (
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/i18n"
	"github.com/dujiao-next/internal/service"
	"github.com/gin-gonic/gin"
)

func denyWhenAccessDisabled(settingService *service.SettingService, disabled func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if settingService != nil && disabled != nil && disabled() {
			msg := i18n.T(i18n.ResolveLocale(c), "error.forbidden")
			response.Forbidden(c, msg)
			c.Abort()
			return
		}
		c.Next()
	}
}

func DenyWhenAffiliateDisabled(settingService *service.SettingService) gin.HandlerFunc {
	return denyWhenAccessDisabled(settingService, settingService.GetDisableAffiliate)
}

func DenyWhenApiDisabled(settingService *service.SettingService) gin.HandlerFunc {
	return denyWhenAccessDisabled(settingService, settingService.GetDisableAPI)
}
