// Package config handles the configuration for the Nexeres authentication server.
//
// Note: The Config structs have two tags for json and yaml.
// JSON tags: Used to mark what data is visible to admins via the config json endpoint.
// YAML tags: Used to actually configure this auth server.
package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/nbrglm/nexeres/opts"
)

// Type aliases for enums
type PasswordHashingAlgorithm string

const (
	// Bcrypt is the default password hashing algorithm
	BcryptPasswordHashingAlgorithm PasswordHashingAlgorithm = "bcrypt"
	// Argon2id is a more secure password hashing algorithm
	Argon2idPasswordHashingAlgorithm PasswordHashingAlgorithm = "argon2id"
)

// Variables for different configurations
var C *CompleteConfig

func Environment() string {
	if opts.Debug {
		return "development"
	}
	return "production"
}

// ObservabilityConfig holds the configuration for observability features
//
// Used for logs, traces and metrics (Recommended: SigNoz)
type ObservabilityConfig struct {
	// Logs configuration
	Logs LogsConfig `json:"logs" yaml:"logs" validate:"required"`

	// Traces configuration
	Traces TracesConfig `json:"traces" yaml:"traces" validate:"required"`

	// Metrics configuration
	Metrics MetricsConfig `json:"metrics" yaml:"metrics" validate:"required"`
}

// LogsConfig holds the configuration for logs
type LogsConfig struct {
	Level string `json:"level" yaml:"level" validate:"required,oneof=debug info warn error dpanic panic fatal"`

	// Endpoint is the endpoint for the OpenTelemetry logs exporter
	// If not set, logs will be printed to stdout only.
	Endpoint string `json:"endpoint" yaml:"endpoint" validate:"required_unless=Protocol stdout,domain"`

	// EndpointPath is the URL path for the OpenTelemetry logs exporter
	//
	// If not set, it defaults to /v1/logs
	EndpointPath string `json:"endpointPath" yaml:"endpointPath" validate:"excluded_unless=Protocol http/protobuf"`

	// Protocol is the protocol for the OpenTelemetry logs exporter
	// Supported values: grpc, http/protobuf, stdout
	Protocol string `json:"protocol" yaml:"protocol" validate:"required_if=Enable true,oneof=grpc http/protobuf stdout"`

	// Headers are the headers to be sent with Otel exporter requests,
	// You can add API keys or authentication tokens here as needed by your Otel collector/platform.
	Headers map[string]string `json:"headers" yaml:"headers"`

	// WithInsecure indicates whether to use insecure connection for Otel exporter
	WithInsecure bool `json:"withInsecure" yaml:"withInsecure"`
}

// TracesConfig holds the configuration for traces
type TracesConfig struct {
	// Endpoint is the endpoint for the OpenTelemetry traces exporter
	// If not set, traces will be printed to stdout only.
	Endpoint string `json:"endpoint" yaml:"endpoint" validate:"required_unless=Protocol stdout,omitempty,domain"`

	// EndpointPath is the URL path for the OpenTelemetry traces exporter
	//
	// If not set, it defaults to /v1/traces
	EndpointPath string `json:"endpointPath" yaml:"endpointPath" validate:"required_if=Protocol http/protobuf"`

	// Protocol is the protocol for the OpenTelemetry traces exporter
	// Supported values: grpc, http/protobuf, stdout
	Protocol string `json:"protocol" yaml:"protocol" validate:"required,oneof=grpc http/protobuf stdout"`

	// Headers are the headers to be sent with Otel exporter requests,
	// You can add API keys or authentication tokens here as needed by your Otel collector/platform.
	Headers map[string]string `json:"headers" yaml:"headers"`

	// WithInsecure indicates whether to use insecure connection for Otel exporter
	WithInsecure bool `json:"withInsecure" yaml:"withInsecure"`
}

// MetricsConfig holds the configuration for metrics
type MetricsConfig struct {
	// Endpoint is the endpoint for the OpenTelemetry metrics exporter
	// If not set, metrics will be printed to stdout only.
	Endpoint string `json:"endpoint" yaml:"endpoint" validate:"required_unless=Protocol stdout,domain"`

	// EndpointPath is the URL path for the OpenTelemetry metrics exporter
	//
	// If not set, it defaults to /v1/metrics
	EndpointPath string `json:"endpointPath" yaml:"endpointPath" validate:"required_if=Protocol http/protobuf"`

	// Protocol is the protocol for the OpenTelemetry metrics exporter
	// Supported values: grpc, http/protobuf, stdout
	Protocol string `json:"protocol" yaml:"protocol" validate:"required,oneof=grpc http/protobuf stdout"`

	// Headers are the headers to be sent with Otel exporter requests,
	// You can add API keys or authentication tokens here as needed by your Otel collector/platform.
	Headers map[string]string `json:"headers" yaml:"headers"`

	// WithInsecure indicates whether to use insecure connection for Otel exporter
	WithInsecure bool `json:"withInsecure" yaml:"withInsecure"`
}

// PublicConfig holds the public server configuration options
//
// Note: If hosting Nexeres at "auth.example.com", then set the values as follows:
//
// domain: example.com
// subDomain: auth
//
// Since, the OIDC issuer URLs will be like:
// https://auth.example.com/.well-known/openid-configuration
//
// Your DNS should be configured to point to the server's IP address,
// for example:
// *.auth.example.com should point to the server's IP address.
type PublicConfig struct {
	// Scheme for public URLs (http/https)
	//
	// Use "https" even if TLS termination is handled by reverse proxy
	Scheme string `json:"scheme" yaml:"scheme" validate:"required,oneof=http https"`

	// Domain is the domain at which Nexeres UI is being hosted.
	//
	// Eg. if hosting Nexeres at auth.example.com, provide "example.com" here.
	//
	// This setting is used as the Domain for all Cookies (except Refresh Token Cookie, that is only set at "subdomain.domain")
	Domain string `json:"domain" yaml:"domain" validate:"required"`

	// Subdomain at which Nexeres UI is being hosted.
	//
	// Eg. If hosting Nexeres at auth.example.com, provide "auth" here.
	SubDomain string `json:"subDomain" yaml:"subDomain" validate:"required"`

	DebugBaseURL string `json:"debugBaseURL,omitempty" yaml:"debugBaseURL,omitempty" validate:"omitempty,url"`
}

func (p *PublicConfig) GetBaseURL() string {
	if opts.Debug && strings.TrimSpace(p.DebugBaseURL) != "" {
		return p.DebugBaseURL
	}
	return fmt.Sprintf("%s://%s.%s", p.Scheme, p.SubDomain, p.Domain)
}

func (p *PublicConfig) GetTenantBaseURL(tenant string) string {
	return fmt.Sprintf("%s://%s.%s.%s", p.Scheme, tenant, p.SubDomain, p.Domain)
}

// ServerConfig holds the configuration for the server
type ServerConfig struct {
	// Host interface to listen on, Default: localhost
	Host string `json:"host" yaml:"host" validate:"required"`
	// Port to listen on, Default: 3360, must be between 1024 and 65535
	Port string `json:"port" yaml:"port" validate:"required"`

	// Instance ID for the metrics, logs, traces....
	InstanceID string `json:"instanceID" yaml:"instanceID" validate:"required"`

	// TLSConfig contains the TLS configuration for the server
	//
	// If TLS is enabled, the server will listen on HTTPS, else HTTP.
	TLSConfig bool `json:"tls" yaml:"tls"`
}

type AdminConfig struct {
	// Emails is a list of admin emails.
	//
	// For login, an email will be sent to the user's address if it exists in this list.
	// The user can then use the code to login.
	Emails []string `json:"-" yaml:"emails" validate:"required,dive,email"`

	// The session, without any calls to an admin api, will expire in this duration.
	//
	// In seconds, default 15 minutes
	SessionTimeout int `json:"-" yaml:"sessionTimeout" validate:"required,min=300"`
}

// PasswordConfig holds the configuration for password policies
type PasswordConfig struct {
	// The algorithm used for hashing passwords, default: bcrypt
	Algorithm PasswordHashingAlgorithm `json:"-" yaml:"algorithm" validate:"required,oneof=bcrypt argon2id"`

	// Configuration for Bcrypt password hashing
	Bcrypt BcryptConfig `json:"-" yaml:"bcrypt" validate:"required_if=Algorithm bcrypt"`

	// Configuration for Argon2id password hashing
	Argon2id Argon2idConfig `json:"-" yaml:"argon2id" validate:"required_if=Algorithm argon2id"`
}

// BcryptConfig holds the configuration for Bcrypt password hashing
type BcryptConfig struct {
	// The cost factor for Bcrypt hashing (default is 10)
	Cost int `json:"-" yaml:"cost" validate:"required,min=10,max=31"`
}

// Argon2idConfig holds the configuration for Argon2id password hashing
type Argon2idConfig struct {
	// Memory in KiB (16MB minimum, and default)
	Memory int `json:"-" yaml:"memory" validate:"required,min=16384"`

	// Number of iterations (3-4 recommended, default 3)
	Iterations int `json:"-" yaml:"iterations" validate:"required,min=1"`

	// Number of parallel threads (1-4 recommended, default 1)
	Parallelism int `json:"-" yaml:"parallelism" validate:"required,min=1"`

	// Salt length in bytes (16-32 recommended, default 16)
	SaltLength int `json:"-" yaml:"saltLength" validate:"required,min=16,max=32"`

	// Key length in bytes (32-64 recommended, default 32)
	KeyLength int `json:"-" yaml:"keyLength" validate:"required,min=32,max=64"`
}

// JWTConfig holds the configuration for JWT tokens
type JWTConfig struct {
	// Session token expiration time in seconds (default: 1hr, 3600)
	SessionTokenExpiration int `json:"sessionTokenExpiration" yaml:"sessionTokenExpiration" validate:"required,min=60"`

	// Refresh token expiration time in seconds (default: 30d, 2592000)
	RefreshTokenExpiration int `json:"refreshTokenExpiration" yaml:"refreshTokenExpiration" validate:"required,min=86400"`

	// Audiences claim for the JWT.
	//
	// NOTE: This is NOT for OIDC. This is for the session tokens! OIDC configuration is stored in the DB per tenant.
	//
	// NOTE: Do not add the subdomain or domain you added in the Public Config here, it is automatically added and will be duplicated if you add.
	// Eg. If domain has value example.com and subdomain has value auth,
	// then auth.example.com, and example.com are automatically added as audience claims.
	Audiences []string `json:"-" yaml:"audiences" validate:"required,dive,required"`
}

type NotificationsConfig struct {
	// Email configuration for sending notifications
	Email EmailNotificationConfig `json:"email" yaml:"email" validate:"required"`

	// SMS Configuration for sending notifications
	SMS *SMSNotificationConfig `json:"sms,omitempty" yaml:"sms,omitempty" validate:"omitempty"`
}

// EmailNotificationConfig holds the configuration for email notifications.
type EmailNotificationConfig struct {
	// Provider is the email provider to use for sending emails.
	Provider string `json:"provider" yaml:"provider" validate:"required,oneof=ses sendgrid smtp"`

	// SendGridConfig holds the configuration for SendGrid email provider.
	SendGrid *SendGridProviderConfig `json:"-" yaml:"sendgrid,omitempty" validate:"omitempty,required_if=Provider sendgrid"`

	// SMTPConfig holds the configuration for SMTP email provider.
	SMTP *SMTPProviderConfig `json:"-" yaml:"smtp,omitempty" validate:"omitempty,required_if=Provider smtp"`

	// SESConfig holds the configuration for AWS SES email provider.
	SES *SESProviderConfig `json:"-" yaml:"ses,omitempty" validate:"omitempty,required_if=Provider ses"`

	// Endpoints holds the configuration for URLs inside emails.
	Endpoints EmailEndpointsConfig `json:"endpoints" yaml:"endpoints" validate:"required"`
}

// EmailEndpointsConfig holds the configuration for URLs inside emails.
type EmailEndpointsConfig struct {
	// VerificationEmail is the endpoint for the email verification link.
	// A `token` parameter will be passed to this URL.
	// Pass a full url, eg. https://auth.example.com/verify-email
	VerificationEmail string `json:"verificationEmail" yaml:"verificationEmail" validate:"required,url"`

	// PasswordReset is the endpoint for the password reset link.
	// A `token` parameter will be passed to this URL.
	// Pass a full url, eg. https://auth.example.com/password-reset
	PasswordReset string `json:"passwordReset" yaml:"passwordReset" validate:"required,url"`
}

// SendGridProviderConfig holds the configuration for SendGrid email provider.
type SendGridProviderConfig struct {
	// API key for SendGrid
	APIKey string `json:"-" yaml:"apiKey" validate:"required"`

	// Email address from which notifications are sent
	FromAddress string `json:"-" yaml:"fromAddress" validate:"required,email"`

	// Name associated with the FromAddress, optional
	FromName *string `json:"-" yaml:"fromName,omitempty"`
}

// SMTPProviderConfig holds the configuration for SMTP email provider.
//
// Note: SMTP is a generic email provider that can be used with any SMTP server.
// IT DOES NOT SUPPORT RETRIES, OR ANY ADVANCED FEATURES.
// It is recommended to use a more robust provider like SendGrid or AWS SES for production use.
type SMTPProviderConfig struct {
	// SMTP server host
	Host string `json:"-" yaml:"host" validate:"required"`

	// SMTP server port
	Port string `json:"-" yaml:"port"`

	// Email address from which notifications are sent
	FromAddress string `json:"-" yaml:"fromAddress" validate:"required,email"`

	// Password for SMTP authentication
	Password string `json:"-" yaml:"password" validate:"required"`
}

// SESProviderConfig holds the configuration for AWS SES email provider.
type SESProviderConfig struct {
	// AWS region where SES is configured
	Region string `json:"-" yaml:"region" validate:"required"`

	// AWS Access Key ID for SES
	AccessKeyID string `json:"-" yaml:"accessKeyID" validate:"required"`

	// AWS Secret Access Key for SES
	SecretAccessKey string `json:"-" yaml:"secretAccessKey" validate:"required"`

	// Email address from which notifications are sent
	FromAddress string `json:"-" yaml:"fromAddress" validate:"required,email"`

	// Optional: Name associated with the FromAddress, used in email headers, if not provided, AppName is used.
	FromName *string `json:"-" yaml:"fromName,omitempty"`
}

// SMSNotificationConfig holds the configuration for SMS notifications.
type SMSNotificationConfig struct {
	// TODO: Implement SMSNotificationConfig
	Provider string `json:"provider" yaml:"provider" validate:"required,oneof=twilio"`
}

// BrandingConfig holds the configuration for branding elements such as names.
type BrandingConfig struct {
	// AppName is the name of the application, used in various places like email templates, UI, etc.
	AppName string `json:"appName" yaml:"appName" validate:"required"`

	// CompanyName is the name of the company, used in emails, UI, etc.
	CompanyName string `json:"companyName" yaml:"companyName" validate:"required"`

	// CompanyNameShort is a short version of the company name, used in places where space is limited.
	CompanyNameShort string `json:"companyNameShort" yaml:"companyNameShort" validate:"required"`

	// SupportURL is the URL for support, used in emails, UI, etc.
	SupportURL string `json:"supportURL" yaml:"supportURL" validate:"required,url"`
}

// SecurityConfig holds the security-related configurations for the application.
type SecurityConfig struct {
	// Enable or disable audit logs.
	AuditLogs AuditLogsConfig `json:"auditLogs" yaml:"auditLogs" validate:"required"`

	// The list of API Keys which are allowed to access the API endpoints.
	// Requests without an API key, or with a key not specified here, will be denied with 401.
	APIKeys []APIKeyConfig `json:"apiKeys" yaml:"apiKeys" validate:"required,dive"`

	// CORS configuration for the application.
	CORS CORSConfig `json:"-" yaml:"cors" validate:"required"`

	// Rate limiting configuration.
	RateLimit RateLimitConfig `json:"rateLimit" yaml:"rateLimit" validate:"required"`
}

type AuditLogsConfig struct {
	// Enable or disable audit logs.
	Enable bool `json:"enable" yaml:"enable"`
}

type APIKeyConfig struct {
	// The name of the API Key
	Name string `json:"name" yaml:"name" validate:"required"`

	// Description of the key
	Description string `json:"description" yaml:"description" validate:"required"`

	// The key itself. No restriction on the length, but keep it sensible please.
	Key string `json:"-" yaml:"key" validate:"required"`
}

// CORSConfig holds the configuration for Cross-Origin Resource Sharing (CORS).
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to access the resources.
	AllowedOrigins []string `json:"-" yaml:"allowedOrigins" validate:"dive,url"`

	// AllowedMethods is a list of HTTP methods that are allowed.
	AllowedMethods []string `json:"-" yaml:"allowedMethods" validate:"required,dive,oneof=GET POST PUT DELETE OPTIONS PATCH HEAD"`

	// AllowedHeaders is a list of headers that are allowed in requests.
	AllowedHeaders []string `json:"-" yaml:"allowedHeaders" validate:"required"`
}

// RateLimitConfig holds the configuration for rate limiting the API.
type RateLimitConfig struct {
	// Rate limit for API requests.
	// Format: "R-U", where R is requests and U is the time unit (s - per second, m - per minute, h - per hour, d - per day)
	Rate string `json:"rate" yaml:"rate" validate:"required"`
}

// StoresConfig holds the configuration for the different stores like postgres,redis, s3-like.
type StoresConfig struct {
	// Postgres configuration
	Postgres PostgreSQLConfig `json:"-" yaml:"postgres" validate:"required"`

	// Redis configuration
	Redis RedisConfig `json:"-" yaml:"redis" validate:"required"`

	S3 S3Config `json:"-" yaml:"s3" validate:"required"`
}

// S3Config holds the configuration for S3-like object storage.
type S3Config struct {
	// Endpoint is the S3-compatible storage endpoint.
	Endpoint string `json:"-" yaml:"endpoint" validate:"required,url"`

	// AccessKeyID is the access key ID for the S3 bucket.
	AccessKeyID string `json:"-" yaml:"accessKeyID" validate:"required"`

	// SecretAccessKey is the secret access key for the S3 bucket.
	SecretAccessKey string `json:"-" yaml:"secretAccessKey" validate:"required"`

	// Region is the region where the S3 bucket is located.
	Region string `json:"-" yaml:"region" validate:"required"`

	// UseSSL indicates whether to use SSL for S3 requests.
	// Set to true if using HTTPS
	UseSSL bool `json:"-" yaml:"useSSL"`
}

// PostgreSQLConfig holds the configuration for connecting to a PostgreSQL database
type PostgreSQLConfig struct {
	// Data Source Name for the database connection, in pgx format
	//
	// For connection pooling arguments, take a look at https://pkg.go.dev/github.com/jackc/pgx/v5@v5.7.5/pgxpool#ParseConfig
	DSN string `json:"-" yaml:"dsn" validate:"required"`
}

// RedisConfig holds the configuration for connecting to a Redis database
type RedisConfig struct {
	// Address of the Redis server, e.g., "localhost:6379"
	Address string `json:"-" yaml:"address" validate:"required"`

	// Password for the Redis server, if any
	Password *string `json:"-" yaml:"password,omitempty"`

	// Database index to use (default is 0)
	DB int `json:"-" yaml:"db" validate:"min=0"`
}

// This represents a temporary struct for configuration extraction from the config file.
type CompleteConfig struct {
	// Debug mode for the application
	Debug bool `json:"debug" yaml:"debug"`
	// Admins is a list of credentials
	Admins        AdminConfig         `json:"-" yaml:"admins" validate:"required"`
	Public        PublicConfig        `json:"public" yaml:"public" validate:"required"`
	Multitenancy  bool                `json:"multitenancy" yaml:"multitenancy"`
	Server        ServerConfig        `json:"-" yaml:"server" validate:"required"`
	Observability ObservabilityConfig `json:"-" yaml:"observability" validate:"required"`
	Password      PasswordConfig      `json:"-" yaml:"password" validate:"required"`
	JWT           JWTConfig           `json:"jwt" yaml:"jwt" validate:"required"`
	Notifications NotificationsConfig `json:"notifications" yaml:"notifications" validate:"required"`
	Branding      BrandingConfig      `json:"branding" yaml:"branding" validate:"required"`
	Security      SecurityConfig      `json:"security" yaml:"security" validate:"required"`
	Stores        StoresConfig        `json:"-" yaml:"stores" validate:"required"`
}

func (c *CompleteConfig) KoanfMerge(other *koanf.Koanf) (*koanf.Koanf, error) {
	kf := koanf.New(".")
	if err := kf.Load(structs.Provider(defaultConfig, "yaml"), nil); err != nil {
		return nil, err
	}

	if err := other.Merge(kf); err != nil {
		return nil, err
	}

	return other, nil
}

// ConfigError represents an error that occurs during configuration initialization/reinitialization
type ConfigError struct {
	Message         string
	UnderlyingError error
}

func (c ConfigError) Error() string {
	if c.UnderlyingError != nil {
		return fmt.Sprintf("%v, UnderlyingError: %v", c.Message, c.UnderlyingError.Error())
	}
	return c.Message
}

func LoadConfigOptions(configFile string) (err error) {
	C = new(CompleteConfig)
	filePath, err := filepath.Abs(configFile)
	if err != nil {
		return ConfigError{
			Message: fmt.Sprintf("Unable to get convert given path (%s) to absolute path, Error: %v", configFile, err.Error()),
		}
	}

	if err := InitKoanf(filePath); err != nil {
		return err
	}

	return nil
}

var defaultConfig = CompleteConfig{
	Debug:        true,
	Multitenancy: false,
	Admins: AdminConfig{
		Emails:         []string{},
		SessionTimeout: 900,
	},
	Public: PublicConfig{
		Scheme:       "http",
		Domain:       "localhost",
		SubDomain:    "auth",
		DebugBaseURL: "http://localhost:3360",
	},
	Server: ServerConfig{
		Host:       "localhost",
		Port:       "3360",
		InstanceID: "instance.dev.1",
		TLSConfig:  false,
	},
	Observability: ObservabilityConfig{
		Logs: LogsConfig{
			Protocol: "stdout",
		},
		Traces: TracesConfig{
			Protocol: "stdout",
		},
		Metrics: MetricsConfig{
			Protocol: "stdout",
		},
	},
	Password: PasswordConfig{
		Algorithm: BcryptPasswordHashingAlgorithm,
		Bcrypt: BcryptConfig{
			Cost: 10,
		},
		Argon2id: Argon2idConfig{
			Memory:      64 * 1024, // 64MB in KiB
			Iterations:  1,
			Parallelism: 1,
			SaltLength:  16,
			KeyLength:   32,
		},
	},
	JWT: JWTConfig{
		SessionTokenExpiration: 3600,
		RefreshTokenExpiration: 2592000,
	},
	Notifications: NotificationsConfig{}, // No defaults for notifications
	Branding:      BrandingConfig{},      // No defaults for branding
	Security: SecurityConfig{
		AuditLogs: AuditLogsConfig{
			Enable: false,
		},
		APIKeys: []APIKeyConfig{},
		CORS: CORSConfig{
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders: []string{"Content-Type", "X-Requested-With", "Accept", "Origin", "User-Agent", "X-NEXERES-Refresh-Token", "X-NEXERES-API-Key", "X-NEXERES-Session-Token", "X-NEXERES-Admin-Token"},
		},
		RateLimit: RateLimitConfig{},
	},
	Stores: StoresConfig{}, // No defaults for stores
}
