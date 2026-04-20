package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

// TelegramAuthSetting Telegram 登录配置实体
type TelegramAuthSetting struct {
	Enabled                      bool   `json:"enabled"`
	BotUsername                  string `json:"bot_username"`
	BotToken                     string `json:"bot_token"`
	MiniAppURL                   string `json:"mini_app_url"`
	TelegramUserWhitelistEnabled bool   `json:"telegram_user_whitelist_enabled"`
	TelegramUserWhitelist        string `json:"telegram_user_whitelist"`
	LoginExpireSeconds           int    `json:"login_expire_seconds"`
	ReplayTTLSeconds             int    `json:"replay_ttl_seconds"`
}

// TelegramAuthSettingPatch Telegram 登录配置补丁
type TelegramAuthSettingPatch struct {
	Enabled                      *bool   `json:"enabled"`
	BotUsername                  *string `json:"bot_username"`
	BotToken                     *string `json:"bot_token"`
	MiniAppURL                   *string `json:"mini_app_url"`
	TelegramUserWhitelistEnabled *bool   `json:"telegram_user_whitelist_enabled"`
	TelegramUserWhitelist        *string `json:"telegram_user_whitelist"`
	LoginExpireSeconds           *int    `json:"login_expire_seconds"`
	ReplayTTLSeconds             *int    `json:"replay_ttl_seconds"`
}

// TelegramAuthDefaultSetting 根据运行时配置生成默认设置
func TelegramAuthDefaultSetting(cfg config.TelegramAuthConfig) TelegramAuthSetting {
	return NormalizeTelegramAuthSetting(TelegramAuthSetting{
		Enabled:                      cfg.Enabled,
		BotUsername:                  strings.TrimSpace(cfg.BotUsername),
		BotToken:                     strings.TrimSpace(cfg.BotToken),
		MiniAppURL:                   strings.TrimSpace(cfg.MiniAppURL),
		TelegramUserWhitelistEnabled: cfg.TelegramUserWhitelistEnabled,
		TelegramUserWhitelist:        strings.TrimSpace(cfg.TelegramUserWhitelist),
		LoginExpireSeconds:           cfg.LoginExpireSeconds,
		ReplayTTLSeconds:             cfg.ReplayTTLSeconds,
	})
}

// NormalizeTelegramAuthSetting 归一化 Telegram 配置
func NormalizeTelegramAuthSetting(setting TelegramAuthSetting) TelegramAuthSetting {
	setting.BotUsername = strings.TrimPrefix(strings.TrimSpace(setting.BotUsername), "@")
	setting.BotToken = strings.TrimSpace(setting.BotToken)
	setting.MiniAppURL = strings.TrimSpace(setting.MiniAppURL)
	setting.TelegramUserWhitelist = normalizeTelegramUserWhitelist(setting.TelegramUserWhitelist)

	if setting.LoginExpireSeconds <= 0 {
		setting.LoginExpireSeconds = 300
	}
	if setting.LoginExpireSeconds < 30 {
		setting.LoginExpireSeconds = 30
	}
	if setting.LoginExpireSeconds > 86400 {
		setting.LoginExpireSeconds = 86400
	}

	if setting.ReplayTTLSeconds <= 0 {
		setting.ReplayTTLSeconds = setting.LoginExpireSeconds
	}
	if setting.ReplayTTLSeconds < 60 {
		setting.ReplayTTLSeconds = 60
	}
	if setting.ReplayTTLSeconds > 86400 {
		setting.ReplayTTLSeconds = 86400
	}
	return setting
}

// ValidateTelegramAuthSetting 校验 Telegram 配置合法性
func ValidateTelegramAuthSetting(setting TelegramAuthSetting) error {
	normalized := NormalizeTelegramAuthSetting(setting)

	if normalized.LoginExpireSeconds < 30 || normalized.LoginExpireSeconds > 86400 {
		return fmt.Errorf("%w: 登录有效期需在 30-86400 秒之间", ErrTelegramAuthConfigInvalid)
	}
	if normalized.ReplayTTLSeconds < 60 || normalized.ReplayTTLSeconds > 86400 {
		return fmt.Errorf("%w: 重放保护时长需在 60-86400 秒之间", ErrTelegramAuthConfigInvalid)
	}
	if _, err := parseTelegramUserWhitelist(normalized.TelegramUserWhitelist); err != nil {
		return fmt.Errorf("%w: Telegram 用户白名单格式不合法（应为 id|备注，多个用英文逗号分隔）", ErrTelegramAuthConfigInvalid)
	}
	if !normalized.Enabled {
		return nil
	}
	if normalized.BotUsername == "" {
		return fmt.Errorf("%w: Bot 用户名不能为空", ErrTelegramAuthConfigInvalid)
	}
	if strings.ContainsAny(normalized.BotUsername, " \t\r\n") {
		return fmt.Errorf("%w: Bot 用户名格式无效", ErrTelegramAuthConfigInvalid)
	}
	if normalized.BotToken == "" {
		return fmt.Errorf("%w: Bot Token 不能为空", ErrTelegramAuthConfigInvalid)
	}
	return nil
}

// TelegramAuthSettingToConfig 转换为运行时配置
func TelegramAuthSettingToConfig(setting TelegramAuthSetting) config.TelegramAuthConfig {
	normalized := NormalizeTelegramAuthSetting(setting)
	return config.TelegramAuthConfig{
		Enabled:                      normalized.Enabled,
		BotUsername:                  normalized.BotUsername,
		BotToken:                     normalized.BotToken,
		MiniAppURL:                   normalized.MiniAppURL,
		TelegramUserWhitelistEnabled: normalized.TelegramUserWhitelistEnabled,
		TelegramUserWhitelist:        normalized.TelegramUserWhitelist,
		LoginExpireSeconds:           normalized.LoginExpireSeconds,
		ReplayTTLSeconds:             normalized.ReplayTTLSeconds,
	}
}

// TelegramAuthSettingToMap 转换为 settings 存储结构
func TelegramAuthSettingToMap(setting TelegramAuthSetting) map[string]interface{} {
	normalized := NormalizeTelegramAuthSetting(setting)
	return map[string]interface{}{
		"enabled":                         normalized.Enabled,
		"bot_username":                    normalized.BotUsername,
		"bot_token":                       normalized.BotToken,
		"mini_app_url":                    normalized.MiniAppURL,
		"telegram_user_whitelist_enabled": normalized.TelegramUserWhitelistEnabled,
		"telegram_user_whitelist":         normalized.TelegramUserWhitelist,
		"login_expire_seconds":            normalized.LoginExpireSeconds,
		"replay_ttl_seconds":              normalized.ReplayTTLSeconds,
	}
}

// MaskTelegramAuthSettingForAdmin 返回脱敏配置
func MaskTelegramAuthSettingForAdmin(setting TelegramAuthSetting) models.JSON {
	normalized := NormalizeTelegramAuthSetting(setting)
	return models.JSON{
		"enabled":                         normalized.Enabled,
		"bot_username":                    normalized.BotUsername,
		"bot_token":                       "",
		"has_bot_token":                   normalized.BotToken != "",
		"mini_app_url":                    normalized.MiniAppURL,
		"telegram_user_whitelist_enabled": normalized.TelegramUserWhitelistEnabled,
		"telegram_user_whitelist":         normalized.TelegramUserWhitelist,
		"login_expire_seconds":            normalized.LoginExpireSeconds,
		"replay_ttl_seconds":              normalized.ReplayTTLSeconds,
	}
}

// GetTelegramAuthSetting 获取 Telegram 登录配置
func (s *SettingService) GetTelegramAuthSetting(defaultCfg config.TelegramAuthConfig) (TelegramAuthSetting, error) {
	fallback := TelegramAuthDefaultSetting(defaultCfg)
	value, err := s.GetByKey(constants.SettingKeyTelegramAuthConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	parsed := telegramAuthSettingFromJSON(value, fallback)
	return NormalizeTelegramAuthSetting(parsed), nil
}

// PatchTelegramAuthSetting 基于补丁更新 Telegram 登录配置
func (s *SettingService) PatchTelegramAuthSetting(defaultCfg config.TelegramAuthConfig, patch TelegramAuthSettingPatch) (TelegramAuthSetting, error) {
	current, err := s.GetTelegramAuthSetting(defaultCfg)
	if err != nil {
		return TelegramAuthSetting{}, err
	}

	next := current
	if patch.Enabled != nil {
		next.Enabled = *patch.Enabled
	}
	if patch.BotUsername != nil {
		next.BotUsername = strings.TrimSpace(*patch.BotUsername)
	}
	if patch.BotToken != nil {
		botToken := strings.TrimSpace(*patch.BotToken)
		if botToken != "" {
			next.BotToken = botToken
		}
	}
	if patch.MiniAppURL != nil {
		next.MiniAppURL = strings.TrimSpace(*patch.MiniAppURL)
	}
	if patch.TelegramUserWhitelistEnabled != nil {
		next.TelegramUserWhitelistEnabled = *patch.TelegramUserWhitelistEnabled
	}
	if patch.TelegramUserWhitelist != nil {
		next.TelegramUserWhitelist = strings.TrimSpace(*patch.TelegramUserWhitelist)
	}
	if patch.LoginExpireSeconds != nil {
		next.LoginExpireSeconds = *patch.LoginExpireSeconds
	}
	if patch.ReplayTTLSeconds != nil {
		next.ReplayTTLSeconds = *patch.ReplayTTLSeconds
	}

	normalized := NormalizeTelegramAuthSetting(next)
	if err := ValidateTelegramAuthSetting(normalized); err != nil {
		return TelegramAuthSetting{}, err
	}
	if _, err := s.Update(constants.SettingKeyTelegramAuthConfig, TelegramAuthSettingToMap(normalized)); err != nil {
		return TelegramAuthSetting{}, err
	}
	return normalized, nil
}

func telegramAuthSettingFromJSON(raw models.JSON, fallback TelegramAuthSetting) TelegramAuthSetting {
	next := fallback
	if raw == nil {
		return next
	}
	next.Enabled = readBool(raw, "enabled", next.Enabled)
	next.BotUsername = readString(raw, "bot_username", next.BotUsername)
	next.BotToken = readString(raw, "bot_token", next.BotToken)
	next.MiniAppURL = readString(raw, "mini_app_url", next.MiniAppURL)
	next.TelegramUserWhitelistEnabled = readBool(raw, "telegram_user_whitelist_enabled", next.TelegramUserWhitelistEnabled)
	next.TelegramUserWhitelist = readString(raw, "telegram_user_whitelist", next.TelegramUserWhitelist)
	next.LoginExpireSeconds = readInt(raw, "login_expire_seconds", next.LoginExpireSeconds)
	next.ReplayTTLSeconds = readInt(raw, "replay_ttl_seconds", next.ReplayTTLSeconds)
	return next
}

type telegramUserWhitelistEntry struct {
	TelegramID int64
	Remark     string
}

func normalizeTelegramUserWhitelist(raw string) string {
	normalizedRaw := strings.TrimSpace(strings.ReplaceAll(raw, "，", ","))
	if normalizedRaw == "" {
		return ""
	}
	entries, err := parseTelegramUserWhitelist(normalizedRaw)
	if err != nil {
		return normalizedRaw
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		idText := strconv.FormatInt(entry.TelegramID, 10)
		if entry.Remark == "" {
			parts = append(parts, idText)
			continue
		}
		parts = append(parts, idText+"|"+entry.Remark)
	}
	return strings.Join(parts, ",")
}

func parseTelegramUserWhitelist(raw string) ([]telegramUserWhitelistEntry, error) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(raw, "，", ","))
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	entries := make([]telegramUserWhitelistEntry, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		tokens := strings.SplitN(item, "|", 2)
		idText := strings.TrimSpace(tokens[0])
		if idText == "" {
			return nil, ErrTelegramAuthConfigInvalid
		}
		telegramID, err := strconv.ParseInt(idText, 10, 64)
		if err != nil || telegramID <= 0 {
			return nil, ErrTelegramAuthConfigInvalid
		}
		remark := ""
		if len(tokens) == 2 {
			remark = strings.TrimSpace(tokens[1])
		}
		if _, exists := seen[telegramID]; exists {
			continue
		}
		seen[telegramID] = struct{}{}
		entries = append(entries, telegramUserWhitelistEntry{TelegramID: telegramID, Remark: remark})
	}

	return entries, nil
}
