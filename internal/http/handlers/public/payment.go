package public

import (
	"errors"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderNo    string `json:"order_no" binding:"required"`
	ChannelID  uint   `json:"channel_id"`
	UseBalance bool   `json:"use_balance"`
}

// LatestPaymentQuery 查询最新待支付记录
type LatestPaymentQuery struct {
	OrderNo string `form:"order_no" binding:"required"`
}

// PaypalWebhookQuery PayPal webhook 查询参数。
type PaypalWebhookQuery struct {
	ChannelID uint `form:"channel_id" binding:"required"`
}

// WechatCallbackQuery 微信支付回调查询参数。
type WechatCallbackQuery struct {
	ChannelID uint `form:"channel_id"`
}

// StripeWebhookQuery Stripe webhook 查询参数。
type StripeWebhookQuery struct {
	ChannelID uint `form:"channel_id"`
}

const callbackLogValueLimit = 4096

// CreatePayment 创建支付单
func (h *Handler) CreatePayment(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	order, err := h.OrderService.GetOrderByUserOrderNo(req.OrderNo, uid)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.order_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}

	result, err := h.PaymentService.CreatePayment(service.CreatePaymentInput{
		OrderID:    order.ID,
		ChannelID:  req.ChannelID,
		UseBalance: req.UseBalance,
		ClientIP:   c.ClientIP(),
		Context:    c.Request.Context(),
	})
	if err != nil {
		respondPaymentCreateError(c, err)
		return
	}

	response.Success(c, dto.NewCreatePaymentResp(result))
}

// CapturePayment 用户捕获支付。
func (h *Handler) CapturePayment(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	paymentID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.payment_invalid", nil)
		return
	}
	payment, err := h.PaymentService.GetPayment(paymentID)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.payment_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.payment_fetch_failed", err)
		return
	}
	if _, err := h.OrderService.GetOrderByUser(payment.OrderID, uid); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.order_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	updated, err := h.PaymentService.CapturePayment(service.CapturePaymentInput{
		PaymentID: paymentID,
		Context:   c.Request.Context(),
	})
	if err != nil {
		respondPaymentCaptureError(c, err)
		return
	}
	response.Success(c, gin.H{
		"payment_id": updated.ID,
		"status":     updated.Status,
	})
}

// GetLatestPayment 获取用户最新待支付记录
func (h *Handler) GetLatestPayment(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	var query LatestPaymentQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}

	order, err := h.OrderService.GetOrderByUserOrderNo(query.OrderNo, uid)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.order_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}

	if order.ParentID != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.payment_invalid", nil)
		return
	}
	if order.Status != constants.OrderStatusPendingPayment {
		shared.RespondError(c, response.CodeBadRequest, "error.order_status_invalid", nil)
		return
	}
	if order.ExpiresAt != nil && !order.ExpiresAt.After(time.Now()) {
		shared.RespondError(c, response.CodeBadRequest, "error.order_status_invalid", nil)
		return
	}

	payment, err := h.PaymentRepo.GetLatestPendingByOrder(order.ID, time.Now())
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.payment_fetch_failed", err)
		return
	}
	if payment == nil {
		shared.RespondError(c, response.CodeNotFound, "error.payment_not_found", nil)
		return
	}

	response.Success(c, dto.NewLatestPaymentResp(payment, order.OrderNo))
}

func respondPaymentCreateError(c *gin.Context, err error) {
	respondWithMappedError(c, err, paymentCreateErrorRules, response.CodeInternal, "error.payment_create_failed")
}

func respondPaymentCaptureError(c *gin.Context, err error) {
	respondWithMappedError(c, err, paymentCaptureErrorRules, response.CodeInternal, "error.payment_callback_failed")
}

// getAvailablePaymentChannels 获取过滤后的可用支付渠道
func (h *Handler) getAvailablePaymentChannels(targetAmount *models.Money, user *models.User, paymentType string) ([]map[string]interface{}, error) {
	channels, _, err := h.PaymentService.ListChannels(repository.PaymentChannelListFilter{
		Page:       1,
		PageSize:   200,
		ActiveOnly: true,
	})
	if err != nil {
		return nil, err
	}

	publicChannels := make([]map[string]interface{}, 0, len(channels))
	for _, channel := range channels {
		// Filter by hide_amount_out_range
		if targetAmount != nil && channel.HideAmountOutRange {
			minAmt := channel.MinAmount.Decimal
			maxAmt := channel.MaxAmount.Decimal
			amtDec := targetAmount.Decimal

			if minAmt.IsPositive() && amtDec.LessThan(minAmt) {
				continue
			}
			if maxAmt.IsPositive() && amtDec.GreaterThan(maxAmt) {
				continue
			}
		}

		// Filter by PaymentRoles
		if len(channel.PaymentRoles) > 0 {
			targetRole := constants.PaymentRoleGuest
			if user != nil {
				targetRole = constants.PaymentRoleMember
			}
			matched := false
			for _, role := range channel.PaymentRoles {
				if role == targetRole {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Filter by MemberLevels
		if len(channel.MemberLevels) > 0 {
			if user == nil || user.MemberLevelID == 0 {
				continue
			}
			matched := false
			for _, ml := range channel.MemberLevels {
				if ml == user.MemberLevelID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Filter by PaymentTypes
		if len(channel.PaymentTypes) > 0 && paymentType != "" {
			matched := false
			for _, pt := range channel.PaymentTypes {
				if pt == paymentType {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		ch := map[string]interface{}{
			"id":                    channel.ID,
			"name":                  channel.Name,
			"provider_type":         channel.ProviderType,
			"channel_type":          channel.ChannelType,
			"interaction_mode":      channel.InteractionMode,
			"fee_rate":              channel.FeeRate,
			"fixed_fee":             channel.FixedFee,
			"min_amount":            channel.MinAmount,
			"max_amount":            channel.MaxAmount,
			"hide_amount_out_range": channel.HideAmountOutRange,
		}
		if channel.Icon != "" {
			ch["icon"] = channel.Icon
		}
		publicChannels = append(publicChannels, ch)
	}

	return publicChannels, nil
}
