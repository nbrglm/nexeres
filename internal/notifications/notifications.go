// Package notifications provides functionality to manage and send notifications.
// It includes support for email and SMS notifications, with configurations for each type.
//
// Email and SMS senders are implemented as interfaces, allowing for different implementations.
// Every implementation must have a config object, which is passed during initialization.
package notifications

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/internal/logging"
	"github.com/nbrglm/nexeres/internal/notifications/templates"
	"go.uber.org/zap"
)

type EmailSenderInterface interface {
	SendEmail(to string, subject string, htmlContent, plainTextContent string) error // SendEmail sends an email to the specified recipient with the given subject and body.
}

type SMSSenderInterface interface {
	SendSMS(to string, message string) error // SendSMS sends an SMS to the specified recipient with the given message.
}

var EmailSender EmailSenderInterface // EmailSender is the global email sender instance.
var SMSSender SMSSenderInterface     // SMSSender is the global SMS sender instance.

// EmailEnabled indicates whether email notifications are enabled.
var EmailEnabled = false

// SMSEnabled indicates whether SMS notifications are enabled.
var SMSEnabled = false

// InitEmail initializes the email sender based on the configuration.
// If the email sender is not configured, it logs a warning and skips initialization.
//
// Sets EmailEnabled to true if the email sender is configured.
func InitEmail() {
	switch config.C.Notifications.Email.Provider {
	case "smtp":
		EmailSender = NewSMTPEmailSender(config.C.Notifications.Email.SMTP.Host, strconv.Itoa(config.C.Notifications.Email.SMTP.Port), config.C.Notifications.Email.SMTP.FromAddress, config.C.Notifications.Email.SMTP.FromName, config.C.Notifications.Email.SMTP.Password)
		EmailEnabled = true
	case "sendgrid":
		EmailSender = NewSendGridEmailSender(config.C.Notifications.Email.SendGrid.APIKey, config.C.Notifications.Email.SendGrid.FromAddress, &config.C.Notifications.Email.SendGrid.FromName)
		EmailEnabled = true
	case "ses":
		EmailSender = NewSESEmailSender(config.C.Notifications.Email.SES.Region, config.C.Notifications.Email.SES.AccessKeyID, config.C.Notifications.Email.SES.SecretAccessKey, config.C.Notifications.Email.SES.FromAddress, &config.C.Notifications.Email.SES.FromName)
		EmailEnabled = true
	default:
		EmailEnabled = false
		logging.Logger.Warn("Unknown email provider, skipping email sender initialization", zap.String("provider", config.C.Notifications.Email.Provider))
	}
}

func InitSMS() {
	if config.C.Notifications.SMS == nil {
		logging.Logger.Warn("SMS notifications are not configured, skipping SMS sender initialization... this will result in an error every time an SMS is sent")
		return
	}
}

var ErrEmailSenderNotSet = fmt.Errorf("email sender is not set, please set it using the config file! notifications.email.provider and the respective provider config")

type SendAdminLoginEmailParams struct {
	Email     string
	Code      string
	ExpiresAt time.Time
}

// SendAdminLoginEmail sends an admin login email to the specified recipient.
// It uses the global EmailSender instance to send the email.
// The email includes the OTP code and its expiration time.
func SendAdminLoginEmail(ctx context.Context, params SendAdminLoginEmailParams) error {
	rendered, err := templates.RenderEmailTemplate(templates.TemplateData{
		AppName:     config.C.Branding.AppName,
		UserName:    "User",
		UserEmail:   params.Email,
		ActionURL:   params.Code,
		ExpiresAt:   params.ExpiresAt,
		CompanyName: config.C.Branding.CompanyNameShort,
		SupportURL:  config.C.Branding.SupportURL,
	}, *templates.AdminLoginTemplate)
	if err != nil {
		return err
	}

	logging.Logger.Debug("Sending admin login email", zap.String("to", params.Email), zap.String("subject", rendered.Subject), zap.String("html_body", rendered.HTMLBody), zap.String("plain_text_body", rendered.PlainTextBody))

	err = sendEmail(params.Email, rendered.Subject, rendered.HTMLBody, rendered.PlainTextBody)
	if err != nil {
		return err
	}

	return nil
}

type SendWelcomeEmailParams struct {
	User struct {
		Email string
		Name  string
	}
	VerificationToken string
	ExpiresAt         time.Time
}

// SendWelcomeEmail sends a welcome email to the specified recipient.
// It uses the global EmailSender instance to send the email.
// The email also includes a link to verify the email address.
func SendWelcomeEmail(ctx context.Context, params SendWelcomeEmailParams) error {
	verificationUrl := fmt.Sprintf("%s?token=%s", config.C.Notifications.Email.Endpoints.Verification, params.VerificationToken)
	rendered, err := templates.RenderEmailTemplate(templates.TemplateData{
		AppName:     config.C.Branding.AppName,
		UserName:    params.User.Name,
		UserEmail:   params.User.Email,
		ActionURL:   verificationUrl,
		ExpiresAt:   params.ExpiresAt,
		CompanyName: config.C.Branding.CompanyNameShort,
		SupportURL:  config.C.Branding.SupportURL,
	}, *templates.VerifyEmailTemplate)
	if err != nil {
		return err
	}

	logging.Logger.Debug("Sending welcome email", zap.String("to", params.User.Email), zap.String("subject", rendered.Subject), zap.String("html_body", rendered.HTMLBody), zap.String("plain_text_body", rendered.PlainTextBody))
	err = sendEmail(params.User.Email, rendered.Subject, rendered.HTMLBody, rendered.PlainTextBody)
	if err != nil {
		return err
	}

	return nil
}

// sendEmail is a helper function to send an email using the global EmailSender instance.
func sendEmail(to string, subject string, htmlContent, plainTextContent string) error {
	if EmailSender == nil {
		return ErrEmailSenderNotSet
	}
	return EmailSender.SendEmail(to, subject, htmlContent, plainTextContent)
}

// getUserName constructs a user name from the provided first and last names.
func getUserName(firstName, lastName *string) string {
	if firstName == nil && lastName == nil {
		return "User"
	}
	if firstName != nil && lastName != nil {
		return *firstName + " " + *lastName
	}
	if firstName != nil {
		return *firstName
	}
	return *lastName
}
