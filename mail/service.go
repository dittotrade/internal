package mail

import (
	"context"
	"fmt"
	"github.com/dmitrymomot/go-env"

	"github.com/google/uuid"

	"github.com/keighl/postmark"
)

// Predefined email templates
var (
	VerificationCodeTmpl   = "verification_code"
	PasswordResetTmpl      = "password_reset"
	DestroyAccountCodeTmpl = "destroy_account"
	StopLossTmpl           = "stop_loss"
)

type (
	// Service struct
	Service struct {
		client postmarkClient
		config Config
	}

	// Config struct
	Config struct {
		ProductName    string
		ProductURL     string
		SupportURL     string
		SupportEmail   string
		CompanyName    string
		CompanyAddress string
		FromEmail      string
		FromName       string
	}

	postmarkClient interface {
		SendTemplatedEmail(email postmark.TemplatedEmail) (postmark.EmailResponse, error)
	}
)

func GetMailer() *Service {
	postmarkServerToken := env.MustString("POSTMARK_SERVER_TOKEN")
	postmarkAccountToken := env.MustString("POSTMARK_ACCOUNT_TOKEN")
	// Product
	productName := env.GetString("PRODUCT_NAME", "Ditto Trade")
	productURL := env.GetString("PRODUCT_URL", "https://ditto.trade")
	supportURL := env.GetString("SUPPORT_URL", "https://ditto.trade/support")
	supportEmail := env.GetString("SUPPORT_EMAIL", "support@ditto.trade")
	companyName := env.GetString("COMPANY_NAME", "Ditto Trade Pty Limited")
	companyAddress := env.GetString("COMPANY_ADDRESS", "Level 27, 25 Bligh Street, Sydney NSW 2000")
	// Mailer
	notificationFromName := env.GetString("NOTIFICATION_FROM_NAME", "Ditto Trade")
	notificationFromEmail := env.GetString("NOTIFICATION_FROM_EMAIL", "notifications@ditto.trade")
	client := postmark.NewClient(postmarkServerToken, postmarkAccountToken)
	config := Config{
		ProductName:    productName,
		ProductURL:     productURL,
		SupportURL:     supportURL,
		SupportEmail:   supportEmail,
		CompanyName:    companyName,
		CompanyAddress: companyAddress,
		FromEmail:      notificationFromEmail,
		FromName:       notificationFromName,
	}
	return &Service{client: client, config: config}
}

// SendVerificationCode ...
func (s *Service) SendVerificationCode(_ context.Context, email, otp string) error {
	if err := s.send(VerificationCodeTmpl, "verification", email, map[string]interface{}{
		"otp": otp,
	}); err != nil {
		return fmt.Errorf("could not send verification code: %w", err)
	}
	return nil
}

// SendResetPasswordCode ...
func (s *Service) SendResetPasswordCode(_ context.Context, email, otp string) error {
	if err := s.send(PasswordResetTmpl, "reset_password", email, map[string]interface{}{
		"otp": otp,
	}); err != nil {
		return fmt.Errorf("could not send reset password code: %w", err)
	}
	return nil
}

// SendDestroyAccountCode ...
func (s *Service) SendDestroyAccountCode(_ context.Context, email, otp string) error {
	if err := s.send(DestroyAccountCodeTmpl, "destroy_account", email, map[string]interface{}{
		"otp": otp,
	}); err != nil {
		return fmt.Errorf("could not send verification code: %w", err)
	}
	return nil
}

func (s *Service) SendNotificationStopLoss(_ context.Context, email, strategyName string, strategyID uuid.UUID, currentEquity, stopLoss float64) error {
	if err := s.send(StopLossTmpl, "investment_stop_loss", email, map[string]interface{}{
		"strategy_name": strategyName,
		"strategy_id":   strategyID,
		"equity":        currentEquity,
		"stopLoss":      stopLoss,
	}); err != nil {
		return fmt.Errorf("could not send verification code: %w", err)
	}
	return nil
}

// send email
func (s *Service) send(tpl, tag, email string, data map[string]interface{}) error {
	// Default model data
	payload := map[string]interface{}{
		"product_url":     s.config.ProductURL,
		"product_name":    s.config.ProductName,
		"support_url":     s.config.SupportURL,
		"company_name":    s.config.CompanyName,
		"company_address": s.config.CompanyAddress,
		"email":           email,
	}

	// Merge custom data with default fields
	for k, v := range data {
		payload[k] = v
	}

	if _, err := s.client.SendTemplatedEmail(postmark.TemplatedEmail{
		TemplateAlias: tpl,
		InlineCss:     true,
		TrackOpens:    true,
		From:          s.config.FromEmail,
		To:            email,
		Tag:           tag,
		ReplyTo:       s.config.SupportEmail,
		TemplateModel: payload,
	}); err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}

	return nil
}
