package convention

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type FileType string

const (
	RS256_PRIVATE_KEY FileType = "RS256_PRIVATE_KEY"
	RS256_PUBLIC_KEY  FileType = "RS256_PUBLIC_KEY"
	TLS_CERT          FileType = "TLS_CERT"
	TLS_KEY           FileType = "TLS_KEY"
	TLC_CA            FileType = "TLS_CA"
	EMAIL_TEMPLATES   FileType = "EMAIL_TEMPLATES"
	SMS_TEMPLATES     FileType = "SMS_TEMPLATES"
)

var FilePaths = map[FileType]string{
	RS256_PRIVATE_KEY: "/etc/nbrglm/workspace/nexeres/jwt/private.key",
	RS256_PUBLIC_KEY:  "/etc/nbrglm/workspace/nexeres/jwt/public.key",
	TLS_CERT:          "/etc/nbrglm/workspace/nexeres/tls/tls.crt",
	TLS_KEY:           "/etc/nbrglm/workspace/nexeres/tls/tls.key",
	TLC_CA:            "/etc/nbrglm/workspace/nexeres/tls/tls.ca",
	EMAIL_TEMPLATES:   "/etc/nbrglm/workspace/nexeres/templates/email",
	SMS_TEMPLATES:     "/etc/nbrglm/workspace/nexeres/templates/sms",
}

var CORSAllowedHeaders = []string{"Content-Type", "X-Requested-With", "Accept", "Origin", "User-Agent", "X-NEXERES-Refresh-Token", "X-NEXERES-API-Key", "X-NEXERES-Session-Token", "X-NEXERES-Admin-Token"}

var CORSAllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}

var HandlerAttrSetSuccess = attribute.NewSet(attribute.KeyValue{
	Key:   attribute.Key("status"),
	Value: attribute.StringValue("success"),
})

var HandlerAttrSetError = attribute.NewSet(attribute.KeyValue{
	Key:   attribute.Key("status"),
	Value: attribute.StringValue("error"),
})

var HandlerAttrSetTotal = attribute.NewSet(attribute.KeyValue{
	Key:   attribute.Key("status"),
	Value: attribute.StringValue("total"),
})

var OptAttrSetSuccess = metric.WithAttributeSet(HandlerAttrSetSuccess)
var OptAttrSetError = metric.WithAttributeSet(HandlerAttrSetError)
var OptAttrSetTotal = metric.WithAttributeSet(HandlerAttrSetTotal)

func OptAttrSet(status string) metric.MeasurementOption {
	return metric.WithAttributeSet(attribute.NewSet(attribute.KeyValue{
		Key:   attribute.Key("status"),
		Value: attribute.StringValue(status),
	}))
}
