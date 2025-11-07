package notifications

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/internal/logging"
	"go.uber.org/zap"
)

func NewSESEmailSender(region, accessKeyId, secretAccessKey, fromAddress string, fromName *string) *SESEmailSender {
	cfg := aws.Config{
		Region: region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""),
		),
	}
	if fromName == nil {
		fromName = &config.C.Branding.AppName
	}
	return &SESEmailSender{
		Client:      sesv2.NewFromConfig(cfg),
		FromAddress: fromAddress,
		FromName:    *fromName,
	}
}

type SESEmailSender struct {
	Client      *sesv2.Client
	FromAddress string
	FromName    string
}

// SendEmail sends an email using AWS SES.
func (s *SESEmailSender) SendEmail(to string, subject, htmlContent, plainTextContent string) error {
	input := &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		FromEmailAddress: aws.String(s.FromAddress),
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String(subject),
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: aws.String(plainTextContent),
					},
					Html: &types.Content{
						Data: aws.String(htmlContent),
					},
				},
			},
		},
	}

	_, err := s.Client.SendEmail(context.Background(), input)
	if err != nil {
		// Log the error
		logging.Logger.Debug("Sending email using SES failed", zap.Error(err), zap.String("to", to), zap.String("subject", subject))
		// Record the error in metrics
		logging.Logger.Error("Failed to send email using SES", zap.Error(err), zap.String("subject", subject))
		return err
	}
	return nil
}
