package dto

import (
	"testing"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

func TestResellerLedgerRespOmitsSensitiveSnapshotFields(t *testing.T) {
	orderID := uint(10)
	entry := models.ResellerLedgerEntry{
		ID:             1,
		ResellerID:     99,
		OrderID:        &orderID,
		Type:           models.ResellerLedgerTypeOrderProfit,
		Amount:         models.NewMoneyFromDecimal(decimal.RequireFromString("12.34")),
		Currency:       "USD",
		IdempotencyKey: "order_profit:10",
		MetadataJSON:   models.JSON{"pricing_snapshot_json": "hidden"},
		Status:         models.ResellerLedgerStatusAvailable,
	}

	resp := NewResellerLedgerResp(&entry)
	if resp.ID != 1 || resp.OrderID == nil || *resp.OrderID != 10 || resp.Amount != "12.34" {
		t.Fatalf("unexpected ledger response: %+v", resp)
	}
}

func TestResellerDashboardRespNotOpened(t *testing.T) {
	resp := NewResellerDashboardResp(false, nil, nil)
	if resp.Opened {
		t.Fatalf("expected unopened dashboard, got %+v", resp)
	}
	if resp.Profile != nil || len(resp.Balances) != 0 {
		t.Fatalf("unopened dashboard should not include profile or balances: %+v", resp)
	}
}
