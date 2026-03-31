// Package templates provides the templates for notifications.
//
// This package contains the templates used for sending notifications like SignUP, Password Reset, and Email Verification, etc.
package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"path"
	"strings"
	"time"

	"github.com/nbrglm/nexeres/convention"
	"github.com/nbrglm/nexeres/utils"
)

//go:embed templs
var templateFs embed.FS

type TemplateData struct {
	AppName     string
	UserName    string
	UserEmail   string
	ActionURL   string
	ExpiresAt   time.Time
	IPAddress   string
	UserAgent   string
	Location    string
	SupportURL  string
	CompanyName string
}

// EmailTemplate represents the structure of an email template.
//
// It contains fields for the template name, subject, HTML body, and plain text body.
// The TemplateName is used for identification purposes, while the Subject, HTMLBody, and PlainTextBody
// are used to define the content of the email. Only these three fields are rendered for the user's email.
type EmailTemplate struct {
	// TemplateName is the name of the template, used for identification.
	TemplateName  string
	Subject       *template.Template
	HTMLBody      *template.Template
	PlainTextBody *template.Template
}

// RenderedEmailTemplate represents the rendered email template.
//
// It contains the template name, subject, HTML body, and plain text body.
// This struct is used to hold the final rendered content after processing the email template with data.
// It is useful for sending the email with the actual content filled in.
type RenderedEmailTemplate struct {
	TemplateName  string
	Subject       string
	HTMLBody      string
	PlainTextBody string
}

type MessageTemplate struct {
}

// The following variables store the parsed email and sms templates.
//
// Any template that needs to be used for sending notifications should be defined here.
var (
	// VerifyEmailTemplate is the template used for verifying email addresses.
	VerifyEmailTemplate *EmailTemplate
	AdminLoginTemplate  *EmailTemplate
)

// Must be called to parse all email templates at application startup.
// This function initializes the email templates used for notifications.
func ParseEmailTemplates() (err error) {
	VerifyEmailTemplate, err = newVerifyEmailTemplate()
	if err != nil {
		return err
	}
	AdminLoginTemplate, err = newAdminLoginTemplate()
	if err != nil {
		return err
	}
	return nil
}

// Must be called to parse all message templates at application startup.
// This function initializes the sms templates used for notifications.
func ParseMessageTemplates() (err error) {
	return nil
}

// RenderEmailTemplate renders the email template with the provided data.
// It takes a TemplateData struct and an EmailTemplate struct as input,
// and returns a RenderedEmailTemplate struct with the rendered content.
// The function is expected to replace placeholders in the email template with actual data from TemplateData.
// If the rendering fails, it returns an error and the original template.
func RenderEmailTemplate(data TemplateData, tmpl EmailTemplate) (*RenderedEmailTemplate, error) {
	var htmlBody, plainTextBody, subject bytes.Buffer

	htmlTmplName := fmt.Sprintf("%sHTML", tmpl.TemplateName)
	plainTextTmplName := fmt.Sprintf("%sText", tmpl.TemplateName)
	subjectTmplName := fmt.Sprintf("%sSubject", tmpl.TemplateName)

	if err := tmpl.HTMLBody.ExecuteTemplate(&htmlBody, htmlTmplName, data); err != nil {
		return nil, err
	}
	if err := tmpl.PlainTextBody.ExecuteTemplate(&plainTextBody, plainTextTmplName, data); err != nil {
		return nil, err
	}
	if err := tmpl.Subject.ExecuteTemplate(&subject, subjectTmplName, data); err != nil {
		return nil, err
	}
	return &RenderedEmailTemplate{
		TemplateName:  tmpl.TemplateName,
		Subject:       strings.TrimSpace(subject.String()),
		HTMLBody:      strings.TrimSpace(htmlBody.String()),
		PlainTextBody: strings.TrimSpace(plainTextBody.String()),
	}, nil
}

// findAndParseTemplates is a utility function that searches for a template file in the embedded filesystem or a specified directory in the config.
func findAndParseTemplates(htmlTmplSubPath, plainTextTmplSubPath, subjectTmplSubPath string) (subjectTemplate, htmlTemplate, plainTextTemplate *template.Template, err error) {
	templatesDir := convention.FilePaths[convention.EMAIL_TEMPLATES]
	if utils.DirExists(templatesDir) {
		htmlTemplate, err = template.ParseFiles(path.Join(templatesDir, htmlTmplSubPath))
		if err != nil {
			return
		}
		plainTextTemplate, err = template.ParseFiles(path.Join(templatesDir, plainTextTmplSubPath))
		if err != nil {
			return
		}
		subjectTemplate, err = template.ParseFiles(path.Join(templatesDir, subjectTmplSubPath))
		if err != nil {
			return
		}
	} else {
		subjectTemplate, err = template.ParseFS(templateFs, subjectTmplSubPath)
		if err != nil {
			return
		}
		htmlTemplate, err = template.ParseFS(templateFs, htmlTmplSubPath)
		if err != nil {
			return
		}
		plainTextTemplate, err = template.ParseFS(templateFs, plainTextTmplSubPath)
		if err != nil {
			return
		}
	}
	return subjectTemplate, htmlTemplate, plainTextTemplate, nil
}
