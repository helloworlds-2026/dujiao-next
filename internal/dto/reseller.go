package dto

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

type ResellerProfileSummaryResp struct {
	ID               uint      `json:"id"`
	Status           string    `json:"status"`
	SettlementStatus string    `json:"settlement_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type ResellerBalanceResp struct {
	ID              uint      `json:"id"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	AvailableAmount string    `json:"available_amount"`
	LockedAmount    string    `json:"locked_amount"`
	NegativeAmount  string    `json:"negative_amount"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ResellerLedgerResp struct {
	ID                uint       `json:"id"`
	OrderID           *uint      `json:"order_id,omitempty"`
	Type              string     `json:"type"`
	Amount            string     `json:"amount"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	AvailableAt       *time.Time `json:"available_at,omitempty"`
	WithdrawRequestID *uint      `json:"withdraw_request_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type ResellerWithdrawResp struct {
	ID           uint       `json:"id"`
	Amount       string     `json:"amount"`
	Currency     string     `json:"currency"`
	Channel      string     `json:"channel"`
	Account      string     `json:"account"`
	Status       string     `json:"status"`
	RejectReason string     `json:"reject_reason,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ResellerDashboardResp struct {
	Opened   bool                        `json:"opened"`
	Profile  *ResellerProfileSummaryResp `json:"profile,omitempty"`
	Balances []ResellerBalanceResp       `json:"balances,omitempty"`
}

func NewResellerProfileSummaryResp(profile *models.ResellerProfile) *ResellerProfileSummaryResp {
	if profile == nil {
		return nil
	}
	return &ResellerProfileSummaryResp{
		ID:               profile.ID,
		Status:           profile.Status,
		SettlementStatus: profile.SettlementStatus,
		CreatedAt:        profile.CreatedAt,
	}
}

func NewResellerBalanceResp(row *models.ResellerBalanceAccount) ResellerBalanceResp {
	if row == nil {
		return ResellerBalanceResp{}
	}
	return ResellerBalanceResp{
		ID:              row.ID,
		Currency:        row.Currency,
		Status:          row.Status,
		AvailableAmount: row.AvailableAmountCache.String(),
		LockedAmount:    row.LockedAmountCache.String(),
		NegativeAmount:  row.NegativeAmountCache.String(),
		UpdatedAt:       row.UpdatedAt,
	}
}

func NewResellerBalanceRespList(rows []models.ResellerBalanceAccount) []ResellerBalanceResp {
	result := make([]ResellerBalanceResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewResellerBalanceResp(&rows[i]))
	}
	return result
}

func NewResellerLedgerResp(row *models.ResellerLedgerEntry) ResellerLedgerResp {
	if row == nil {
		return ResellerLedgerResp{}
	}
	return ResellerLedgerResp{
		ID:                row.ID,
		OrderID:           row.OrderID,
		Type:              row.Type,
		Amount:            row.Amount.String(),
		Currency:          row.Currency,
		Status:            row.Status,
		AvailableAt:       row.AvailableAt,
		WithdrawRequestID: row.WithdrawRequestID,
		CreatedAt:         row.CreatedAt,
	}
}

func NewResellerLedgerRespList(rows []models.ResellerLedgerEntry) []ResellerLedgerResp {
	result := make([]ResellerLedgerResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewResellerLedgerResp(&rows[i]))
	}
	return result
}

func NewResellerWithdrawResp(row *models.ResellerWithdrawRequest) ResellerWithdrawResp {
	if row == nil {
		return ResellerWithdrawResp{}
	}
	return ResellerWithdrawResp{
		ID:           row.ID,
		Amount:       row.Amount.String(),
		Currency:     row.Currency,
		Channel:      row.Channel,
		Account:      row.Account,
		Status:       row.Status,
		RejectReason: row.RejectReason,
		ProcessedAt:  row.ProcessedAt,
		CreatedAt:    row.CreatedAt,
	}
}

func NewResellerWithdrawRespList(rows []models.ResellerWithdrawRequest) []ResellerWithdrawResp {
	result := make([]ResellerWithdrawResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewResellerWithdrawResp(&rows[i]))
	}
	return result
}

func NewResellerDashboardResp(opened bool, profile *models.ResellerProfile, balances []models.ResellerBalanceAccount) ResellerDashboardResp {
	if !opened {
		return ResellerDashboardResp{Opened: false}
	}
	return ResellerDashboardResp{
		Opened:   true,
		Profile:  NewResellerProfileSummaryResp(profile),
		Balances: NewResellerBalanceRespList(balances),
	}
}
