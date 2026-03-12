package shared

import (
	"testing"
	"time"

	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

func TestBuildChannelIdentityResponse(t *testing.T) {
	user := &models.User{
		ID:                    12,
		Email:                 "telegram_12@login.local",
		DisplayName:           "TG Buyer",
		Status:                "active",
		Locale:                "zh-CN",
		PasswordSetupRequired: true,
	}
	identity := &models.UserOAuthIdentity{
		Provider:       "telegram",
		ProviderUserID: "12",
		Username:       "buyer12",
		AvatarURL:      "https://example.com/avatar.png",
	}

	payload := BuildChannelIdentityResponse(true, true, user, identity)
	if payload["bound"] != true {
		t.Fatalf("bound flag mismatch: %#v", payload["bound"])
	}
	if payload["created"] != true {
		t.Fatalf("created flag mismatch: %#v", payload["created"])
	}
	identityPayload, ok := payload["identity"].(gin.H)
	if !ok {
		t.Fatalf("identity payload type mismatch: %#v", payload["identity"])
	}
	if identityPayload["provider_user_id"] != "12" {
		t.Fatalf("provider user id mismatch: %#v", identityPayload["provider_user_id"])
	}
	userPayload, ok := payload["user"].(gin.H)
	if !ok {
		t.Fatalf("user payload type mismatch: %#v", payload["user"])
	}
	if userPayload["display_name"] != "TG Buyer" {
		t.Fatalf("display name mismatch: %#v", userPayload["display_name"])
	}
}

func TestBuildWalletRechargePaymentPayload(t *testing.T) {
	expiresAt := time.Now().UTC()
	recharge := &models.WalletRechargeOrder{
		RechargeNo: "RCG-001",
		Status:     "success",
	}
	payment := &models.Payment{
		ID:              9,
		ProviderType:    "epusdt",
		ChannelType:     "telegram_bot",
		InteractionMode: "redirect",
		PayURL:          "https://pay.example.com",
		QRCode:          "qr-code-data",
		ExpiredAt:       &expiresAt,
		Status:          "pending",
	}
	account := &models.WalletAccount{UserID: 7}

	payload := BuildWalletRechargePaymentPayload(recharge, payment, account)
	if payload["payment_id"] != uint(9) {
		t.Fatalf("payment id mismatch: %#v", payload["payment_id"])
	}
	if payload["recharge_no"] != "RCG-001" {
		t.Fatalf("recharge no mismatch: %#v", payload["recharge_no"])
	}
	if payload["provider_type"] != "epusdt" {
		t.Fatalf("provider type mismatch: %#v", payload["provider_type"])
	}
	if payload["account"] != account {
		t.Fatalf("account payload mismatch: %#v", payload["account"])
	}
}

func TestBuildTelegramBindingResponse(t *testing.T) {
	if payload := BuildTelegramBindingResponse(nil); payload["bound"] != false {
		t.Fatalf("nil identity should be unbound: %#v", payload)
	}

	authAt := time.Now().UTC()
	identity := &models.UserOAuthIdentity{
		Provider:       "telegram",
		ProviderUserID: "9988",
		Username:       "buyer9988",
		AuthAt:         &authAt,
	}
	payload := BuildTelegramBindingResponse(identity)
	if payload["bound"] != true {
		t.Fatalf("bound flag mismatch: %#v", payload["bound"])
	}
	if payload["provider_user_id"] != "9988" {
		t.Fatalf("provider user id mismatch: %#v", payload["provider_user_id"])
	}
}

func TestBuildUserProfilePayload(t *testing.T) {
	user := &models.User{
		ID:          3,
		Email:       "buyer@example.com",
		DisplayName: "Buyer",
		Locale:      "en-US",
	}
	payload := BuildUserProfilePayload(user, "bind_only", "set_without_old")
	if payload["email"] != "buyer@example.com" {
		t.Fatalf("email mismatch: %#v", payload["email"])
	}
	if payload["email_change_mode"] != "bind_only" {
		t.Fatalf("email mode mismatch: %#v", payload["email_change_mode"])
	}
	if payload["password_change_mode"] != "set_without_old" {
		t.Fatalf("password mode mismatch: %#v", payload["password_change_mode"])
	}
}
