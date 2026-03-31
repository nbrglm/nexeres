package notifications

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/nbrglm/nexeres/internal/logging"
	"go.uber.org/zap"
)

func NewSMTPEmailSender(host, port, fromAddress, fromName, password string) *SMTPEmailSender {
	return &SMTPEmailSender{
		Host:        host,
		Port:        port,
		FromAddress: fromAddress,
		FromName:    fromName,
		Password:    password,
	}
}

type SMTPEmailSender struct {
	Host        string
	Port        string
	FromAddress string
	FromName    string
	Password    string
}

// SendEmail sends an email using SMTP.
func (s *SMTPEmailSender) SendEmail(to string, subject, htmlContent, plainTextContent string) error {
	// Create the authentication for the SMTP server
	auth := smtp.PlainAuth("", s.FromAddress, s.Password, s.Host)

	// Boundary for the multipart message
	boundary := "----=_NextPart_000_0000_01DA1234.56789ABC"

	// Create the email headers
	headers := map[string]string{
		"From":         fmt.Sprintf("%s <%s>", s.FromName, s.FromAddress),
		"To":           strings.Join([]string{to}, ","),
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": fmt.Sprintf("multipart/alternative; boundary=\"%s\"", boundary),
	}

	// Create the email body
	var message strings.Builder

	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// End of headers
	message.WriteString("\r\n")

	// Add the plain text part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(plainTextContent)
	message.WriteString("\r\n")

	// Add the HTML part
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlContent)
	message.WriteString("\r\n")

	// End of the multipart message
	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	logging.Logger.Debug("Sending email via SMTP",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("html_body", htmlContent),
		zap.String("plain_text_body", plainTextContent),
		zap.String("from", s.FromAddress),
		zap.String("host", s.Host),
		zap.String("port", s.Port),
		zap.String("message", message.String()),
	)

	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", s.Host, s.Port),
		auth, s.FromAddress,
		[]string{to},
		[]byte(message.String()),
	)
	if err != nil {
		return &SMTPSenderError{
			ErrorString: err.Error(),
			StatusCode:  500, // SMTP errors typically don't have a status code, but we can use 500 for general server error
		}
	}
	return nil
}

type SMTPSenderError struct {
	ErrorString string
	StatusCode  int
}

func (e *SMTPSenderError) Error() string {
	return fmt.Sprintf("SMTP error: %s (status code: %d)", e.ErrorString, e.StatusCode)
}
