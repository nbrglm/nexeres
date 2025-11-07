package notifications

import (
	"fmt"

	"github.com/nbrglm/nexeres/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func NewSendGridEmailSender(apiKey string, fromAddress string, fromName *string) *SendGridEmailSender {
	if fromName == nil {
		// Default to the application name if not provided
		fromName = &config.C.Branding.AppName
	}
	return &SendGridEmailSender{
		Client:      sendgrid.NewSendClient(apiKey),
		FromAddress: fromAddress,
		FromName:    *fromName,
	}
}

// SendGridEmailSender implements the EmailSenderInterface using SendGrid for sending emails.
type SendGridEmailSender struct {
	Client *sendgrid.Client
	// FromAddress is the email address from which notifications are sent.
	FromAddress string
	// FromName is the name associated with the FromAddress. If not given, it defaults to the `config.Application.Name`.
	FromName string
}

// SendEmail sends an email using the SendGrid API.
func (s *SendGridEmailSender) SendEmail(to string, subject, htmlContent, plainTextContent string) error {
	message := mail.NewSingleEmail(
		mail.NewEmail(s.FromName, s.FromAddress), // From
		subject,                                  // Subject
		mail.NewEmail("", to),                    // To
		plainTextContent,                         // Plain text content
		htmlContent,                              // HTML content
	)
	resp, err := s.Client.Send(message)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return &SendGridSenderError{
			Body:       resp.Body,
			StatusCode: resp.StatusCode,
		}
	}
	return nil
}

// SendGridSenderError implements the error interface for SendGrid errors.
type SendGridSenderError struct {
	Body       string
	StatusCode int
}

func (e *SendGridSenderError) Error() string {
	return fmt.Sprintf("SendGrid error: %s (status code: %d)", e.Body, e.StatusCode)
}
