package public

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"
)

var userResellerFinanceErrorRules = []mappedHandlerError{
	{target: service.ErrResellerNotOpened, code: response.CodeBadRequest, key: "error.bad_request"},
	{target: service.ErrResellerProfileInactive, code: response.CodeBadRequest, key: "error.forbidden"},
	{target: service.ErrResellerSettlementUnavailable, code: response.CodeBadRequest, key: "error.forbidden"},
	{target: service.ErrResellerWithdrawAmountInvalid, code: response.CodeBadRequest, key: "error.bad_request"},
	{target: service.ErrResellerWithdrawCurrencyUnavailable, code: response.CodeBadRequest, key: "error.bad_request"},
	{target: service.ErrResellerWithdrawInsufficient, code: response.CodeBadRequest, key: "error.bad_request"},
	{target: service.ErrResellerBalanceAccountFrozen, code: response.CodeBadRequest, key: "error.forbidden"},
}

func respondUserResellerFinanceError(c *gin.Context, err error, fallbackKey string) {
	respondWithMappedError(c, err, userResellerFinanceErrorRules, response.CodeInternal, fallbackKey)
}

// GetResellerDashboard 获取当前用户的分销商财务看板。
func (h *Handler) GetResellerDashboard(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.ResellerAccountingService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	data, err := h.ResellerAccountingService.GetUserFinanceDashboard(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.Success(c, dto.NewResellerDashboardResp(data.Opened, data.Profile, data.Balances))
}

// ListResellerBalanceAccounts 查询当前用户的分销余额账户。
func (h *Handler) ListResellerBalanceAccounts(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.ResellerAccountingService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.ResellerAccountingService.ListUserBalanceAccounts(uid, service.ResellerUserBalanceAccountListFilter{
		Page:     page,
		PageSize: pageSize,
		Currency: strings.TrimSpace(c.Query("currency")),
		Status:   strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		respondUserResellerFinanceError(c, err, "error.user_fetch_failed")
		return
	}
	response.SuccessWithPage(c, dto.NewResellerBalanceRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ListResellerLedgerEntries 查询当前用户的分销账务流水。
func (h *Handler) ListResellerLedgerEntries(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.ResellerAccountingService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	page, pageSize := shared.ParsePagination(c)
	orderID, err := shared.ParseQueryUint(c.Query("order_id"), false)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	rows, total, err := h.ResellerAccountingService.ListUserLedgerEntries(uid, service.ResellerUserLedgerListFilter{
		Page:     page,
		PageSize: pageSize,
		Currency: strings.TrimSpace(c.Query("currency")),
		Type:     strings.TrimSpace(c.Query("type")),
		Status:   strings.TrimSpace(c.Query("status")),
		OrderID:  orderID,
	})
	if err != nil {
		respondUserResellerFinanceError(c, err, "error.user_fetch_failed")
		return
	}
	response.SuccessWithPage(c, dto.NewResellerLedgerRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ListResellerWithdraws 查询当前用户的分销提现申请。
func (h *Handler) ListResellerWithdraws(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.ResellerAccountingService == nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", nil)
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.ResellerAccountingService.ListUserWithdrawRequests(uid, service.ResellerUserWithdrawListFilter{
		Page:     page,
		PageSize: pageSize,
		Currency: strings.TrimSpace(c.Query("currency")),
		Status:   strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		respondUserResellerFinanceError(c, err, "error.user_fetch_failed")
		return
	}
	response.SuccessWithPage(c, dto.NewResellerWithdrawRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ResellerWithdrawApplyRequest 分销商提现申请请求。
type ResellerWithdrawApplyRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Currency string `json:"currency" binding:"required"`
	Channel  string `json:"channel" binding:"required"`
	Account  string `json:"account" binding:"required"`
}

// ApplyResellerWithdraw 提交当前用户的分销提现申请。
func (h *Handler) ApplyResellerWithdraw(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	if h.ResellerAccountingService == nil {
		shared.RespondError(c, response.CodeInternal, "error.save_failed", nil)
		return
	}

	var req ResellerWithdrawApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(req.Amount))
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	row, err := h.ResellerAccountingService.ApplyUserWithdraw(uid, service.ResellerWithdrawApplyInput{
		Amount:   amount,
		Currency: strings.TrimSpace(req.Currency),
		Channel:  strings.TrimSpace(req.Channel),
		Account:  strings.TrimSpace(req.Account),
	})
	if err != nil {
		respondUserResellerFinanceError(c, err, "error.save_failed")
		return
	}
	response.Success(c, dto.NewResellerWithdrawResp(row))
}
