package shared

import (
	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

// BuildWalletRechargePaymentPayload 构造钱包充值支付响应载荷。
func BuildWalletRechargePaymentPayload(recharge *models.WalletRechargeOrder, payment *models.Payment, account *models.WalletAccount) gin.H {
	payload := gin.H{
		"recharge": recharge,
		"payment":  payment,
	}
	if account != nil {
		payload["account"] = account
	}
	if payment != nil {
		payload["payment_id"] = payment.ID
		payload["provider_type"] = payment.ProviderType
		payload["channel_type"] = payment.ChannelType
		payload["interaction_mode"] = payment.InteractionMode
		payload["pay_url"] = payment.PayURL
		payload["qr_code"] = payment.QRCode
		payload["expires_at"] = payment.ExpiredAt
		payload["status"] = payment.Status
	}
	if recharge != nil {
		payload["recharge_no"] = recharge.RechargeNo
		payload["recharge_status"] = recharge.Status
	}
	return payload
}
