package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/nbrglm/nexeres/opts"
	"github.com/nbrglm/nexeres/utils"
)

// NexeresConfig represents the configuration structure for the Nexeres application.
//
// All the sub-configurations are nested within this main configuration struct.
// Each configuration has it's own fields and validation rules.
type NexeresConfig struct {
	Debug          bool                 `yaml:"debug"`
	Admins         AdminsConfig         `yaml:"admins"`
	PublicEndpoint PublicEndpointConfig `yaml:"publicEndpoint"`
	Multitenancy   MultitenancyConfig   `yaml:"multitenancy"`
	Server         ServerConfig         `yaml:"server"`
	Observability  ObservabilityConfig  `yaml:"observability"`
	Password       PasswordConfig       `yaml:"password"`
	JWT            JWTConfig            `yaml:"jwt"`
	Notifications  NotificationsConfig  `yaml:"notifications"`
	Branding       BrandingConfig       `yaml:"branding"`
	Security       SecurityConfig       `yaml:"security"`
	Stores         StoresConfig         `yaml:"stores"`
}

func (c *NexeresConfig) Validate() error {
	if err := c.Admins.Validate(); err != nil {
		return err
	}
	if err := c.PublicEndpoint.Validate(); err != nil {
		return err
	}
	if err := c.Server.Validate(); err != nil {
		return err
	}
	if err := c.Observability.Validate(); err != nil {
		return err
	}
	if err := c.Password.Validate(); err != nil {
		return err
	}
	if err := c.JWT.Validate(); err != nil {
		return err
	}
	if err := c.Notifications.Validate(); err != nil {
		return err
	}
	if err := c.Branding.Validate(); err != nil {
		return err
	}
	if err := c.Security.Validate(); err != nil {
		return err
	}
	if err := c.Stores.Validate(); err != nil {
		return err
	}
	return nil
}

// Currently doesn't require Validation
type MultitenancyConfig struct {
	Enabled bool `yaml:"enabled"`
}

type AdminsConfig struct {
	Emails                []string `yaml:"emails"`
	SessionTimeoutSeconds int      `yaml:"sessionTimeoutSeconds"`
}

func (c *AdminsConfig) Validate() error {
	if c.SessionTimeoutSeconds < 300 {
		return errors.New("AdminsConfig.SessionTimeoutSeconds should atleast be 300")
	}
	if len(c.Emails) == 0 {
		return errors.New("AdminsConfig.Emails: Atleast provide one email address to enable the admin APIs")
	}
	return nil
}

type PublicEndpointConfig struct {
	Scheme       string `yaml:"scheme"`
	Domain       string `yaml:"domain"`
	Subdomain    string `yaml:"subdomain"`
	DebugBaseURL string `yaml:"debugBaseURL"`
}

func (c *PublicEndpointConfig) Validate() error {
	if c.Scheme == "" || (c.Scheme != "http" && c.Scheme != "https") {
		return errors.New("PublicEndpointConfig.Scheme is required and must be either 'http' or 'https'")
	}
	if c.Domain == "" {
		return errors.New("PublicEndpointConfig.Domain is required")
	}
	if c.Subdomain == "" {
		return errors.New("PublicEndpointConfig.Subdomain is required")
	}
	return nil
}

func (p *PublicEndpointConfig) GetBaseURL() string {
	if opts.Debug && strings.TrimSpace(p.DebugBaseURL) != "" {
		return p.DebugBaseURL
	}
	return fmt.Sprintf("%s://%s.%s", p.Scheme, p.Subdomain, p.Domain)
}

func (p *PublicEndpointConfig) GetTenantBaseURL(tenant string) string {
	return fmt.Sprintf("%s://%s.%s.%s", p.Scheme, tenant, p.Subdomain, p.Domain)
}

type ServerConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	InstanceId string `yaml:"instanceId"`
	EnableTLS  bool   `yaml:"enableTLS"`
}

func (c *ServerConfig) Validate() error {
	if c.Host == "" {
		return errors.New("ServerConfig.Host is required")
	}
	if c.Port < 1024 || c.Port > 65535 {
		return errors.New("ServerConfig.Port is required to be between 1024 and 65535")
	}
	if c.InstanceId == "" {
		return errors.New("ServerConfig.InstanceId is required")
	}
	return nil
}

type ObservabilityConfig struct {
	Logs    LogsConfig      `yaml:"logs"`
	Metrics MetricsConfig   `yaml:"metrics"`
	Tracing ObsCommonConfig `yaml:"tracing"`
}

func (c *ObservabilityConfig) Validate() error {
	err := c.Logs.Validate()
	if err != nil {
		return err
	}
	err = c.Metrics.Validate()
	if err != nil {
		return err
	}
	err = c.Tracing.Validate()
	if err != nil {
		return err
	}
	return nil
}

type ObsCommonConfig struct {
	Endpoint string  `yaml:"endpoint"`
	Path     *string `yaml:"path,omitempty"`
	// must be one of "http/protobuf", "grpc", "stdout"
	//
	// stdout means no exporter, just log to stdout
	Protocol     string            `yaml:"protocol"`
	Headers      map[string]string `yaml:"headers"`
	WithInsecure bool              `yaml:"withInsecure"`
}

func (c *ObsCommonConfig) Validate() error {
	if c.Protocol == "" {
		c.Protocol = "stdout"
	}
	if c.Protocol != "http/protobuf" && c.Protocol != "grpc" && c.Protocol != "stdout" {
		return errors.New("ObsCommonConfig.Protocol must be one of 'http/protobuf', 'grpc', 'stdout'")
	}
	if c.Protocol != "stdout" && c.Endpoint == "" {
		return errors.New("ObsCommonConfig.Endpoint is required when Protocol is not 'stdout'")
	}
	if c.Path != nil && *c.Path == "" {
		return errors.New("ObsCommonConfig.Path cannot be empty string if set")
	}
	if C.Debug && strings.Contains(c.Endpoint, "localhost:") {
		// In debug mode, allow localhost endpoints
		c.WithInsecure = true
	}
	// default path is used when not set
	return nil
}

type LogsConfig struct {
	// must be one of "debug" | "info" | "warn" | "error" | "dpanic" | "panic" | "fatal"
	//
	// Logs at and above the specified level will be logged.
	Level           string `yaml:"level"`
	Batch           bool   `yaml:"batch"`
	ObsCommonConfig `yaml:",inline"`
}

func (c *LogsConfig) Validate() error {
	if c.Level == "" {
		return errors.New("LogsConfig.Level is required")
	}
	switch c.Level {
	case "debug", "info", "warn", "error", "dpanic", "panic", "fatal":
	default:
		return errors.New("LogsConfig.Level must be one of 'debug', 'info', 'warn', 'error', 'dpanic', 'panic', 'fatal'")
	}
	if err := c.ObsCommonConfig.Validate(); err != nil {
		return err
	}
	return nil
}

type MetricsConfig struct {
	CollectionIntervalSeconds int `yaml:"collectionIntervalSeconds"` // in seconds
	ObsCommonConfig           `yaml:",inline"`
}

func (c *MetricsConfig) Validate() error {
	if c.CollectionIntervalSeconds < 1 {
		return errors.New("MetricsConfig.CollectionIntervalSeconds must be at least 1")
	}
	if err := c.ObsCommonConfig.Validate(); err != nil {
		return err
	}
	return nil
}

type PasswordHashingAlgorithm string

const (
	Argon2idPasswordHashingAlgorithm PasswordHashingAlgorithm = "argon2id"
	BCryptPasswordHashingAlgorithm   PasswordHashingAlgorithm = "bcrypt"
)

type PasswordConfig struct {
	// One of "argon2id", "bcrypt"
	Algorithm PasswordHashingAlgorithm `yaml:"algorithm"`

	// Parameters for Argon2id
	Argon2idParams *Argon2idParams `yaml:"argon2idParams,omitempty"`

	// Cost for bcrypt
	BcryptCost *int `yaml:"bcryptCost,omitempty"`
}

func (c *PasswordConfig) Validate() error {
	if c.Algorithm == "" {
		return errors.New("PasswordConfig.Algorithm is required")
	}
	if c.Algorithm != "argon2id" && c.Algorithm != "bcrypt" {
		return errors.New("PasswordConfig.Algorithm must be one of 'argon2id', 'bcrypt'")
	}
	switch c.Algorithm {
	case "argon2id":
		if c.Argon2idParams == nil {
			return errors.New("PasswordConfig.Argon2idParams is required when Algorithm is 'argon2id'")
		}
		if err := c.Argon2idParams.Validate(); err != nil {
			return err
		}
	case "bcrypt":
		if c.BcryptCost == nil {
			// Default cost is 10
			c.BcryptCost = new(int)
			*c.BcryptCost = 10
			return nil
		}
		if *c.BcryptCost < 4 || *c.BcryptCost > 31 {
			return errors.New("PasswordConfig.BcryptCost must be between 4 and 31")
		}
	}
	return nil
}

type Argon2idParams struct {
	Memory      int `yaml:"memory"`
	Iterations  int `yaml:"iterations"`
	Parallelism int `yaml:"parallelism"`
	SaltLength  int `yaml:"saltLength"`
	KeyLength   int `yaml:"keyLength"`
}

func (c *Argon2idParams) Validate() error {
	if c.Memory == 0 && c.Iterations == 0 && c.Parallelism == 0 && c.SaltLength == 0 && c.KeyLength == 0 {
		// All default values
		c.Memory = 16384 // 16 MB
		c.Iterations = 2
		c.Parallelism = 1
		c.SaltLength = 16
		c.KeyLength = 32
		return nil
	}
	if c.Memory < 65536 {
		return errors.New("Argon2idParams.Memory must be at least 65536")
	}
	if c.Iterations < 1 {
		return errors.New("Argon2idParams.Iterations must be at least 1")
	}
	if c.Parallelism < 1 {
		return errors.New("Argon2idParams.Parallelism must be at least 1")
	}
	if c.SaltLength < 16 {
		return errors.New("Argon2idParams.SaltLength must be at least 16")
	}
	if c.KeyLength < 32 {
		return errors.New("Argon2idParams.KeyLength must be at least 32")
	}
	return nil
}

type JWTConfig struct {
	SessionTokenExpirySeconds int `yaml:"sessionTokenExpirySeconds"`
	RefreshTokenExpirySeconds int `yaml:"refreshTokenExpirySeconds"`

	// NOTE: Do not add the subdomain or domain you added in PublicConfig here.
	//
	// This is only for additional audiences.
	Audiences []string `yaml:"audiences"`
}

func (c *JWTConfig) Validate() error {
	if c.SessionTokenExpirySeconds < 900 {
		return errors.New("JWTConfig.SessionTokenExpirySeconds must be greater than 900")
	}
	if c.RefreshTokenExpirySeconds < 86400 {
		return errors.New("JWTConfig.RefreshTokenExpirySeconds must be greater than 86400")
	}
	return nil
}

type NotificationsConfig struct {
	Email EmailConfig `yaml:"email"`
	SMS   *SMSConfig  `yaml:"sms,omitempty"`
}

func (c *NotificationsConfig) Validate() error {
	if err := c.Email.Validate(); err != nil {
		return err
	}
	if err := c.SMS.Validate(); err != nil {
		return err
	}
	return nil
}

type EmailConfig struct {
	// One of "smtp", "sendgrid", "ses"
	Provider string `yaml:"provider"`

	// SMTP specific settings
	SMTP *SMTPConfig `yaml:"smtp,omitempty"`

	// SendGrid specific settings
	SendGrid *SendGridConfig `yaml:"sendgrid,omitempty"`

	// AWS SES specific settings
	SES *SESConfig `yaml:"ses,omitempty"`

	// Endpoints used in email templates
	Endpoints EmailTemplateEndpoints `yaml:"endpoints"`
}

func (c *EmailConfig) Validate() error {
	if c.Provider == "" {
		return errors.New("EmailConfig.Provider is required")
	}
	if c.Provider != "smtp" && c.Provider != "sendgrid" && c.Provider != "ses" {
		return errors.New("EmailConfig.Provider must be one of 'smtp', 'sendgrid', 'ses'")
	}
	switch c.Provider {
	case "smtp":
		if c.SMTP == nil {
			return errors.New("EmailConfig.SMTP is required when Provider is 'smtp'")
		}
		if err := c.SMTP.Validate(); err != nil {
			return err
		}
	case "sendgrid":
		if c.SendGrid == nil {
			return errors.New("EmailConfig.SendGrid is required when Provider is 'sendgrid'")
		}
		if err := c.SendGrid.Validate(); err != nil {
			return err
		}
	case "ses":
		if c.SES == nil {
			return errors.New("EmailConfig.SES is required when Provider is 'ses'")
		}
		if err := c.SES.Validate(); err != nil {
			return err
		}
	}
	if err := c.Endpoints.Validate(); err != nil {
		return err
	}
	return nil
}

type SMTPConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"fromAddress"`
	FromName    string `yaml:"fromName"`
}

func (c *SMTPConfig) Validate() error {
	if c.Host == "" {
		return errors.New("SMTPConfig.Host is required")
	}
	if c.Port < 1024 || c.Port > 65535 {
		return errors.New("SMTPConfig.Port must be between 1024 and 65535")
	}
	if c.Password == "" {
		return errors.New("SMTPConfig.Password is required")
	}
	if c.FromAddress == "" {
		return errors.New("SMTPConfig.FromAddress is required")
	}
	return nil
}

type SendGridConfig struct {
	APIKey      string `yaml:"apiKey"`
	FromAddress string `yaml:"fromAddress"`
	FromName    string `yaml:"fromName"`
}

func (c *SendGridConfig) Validate() error {
	if c.APIKey == "" {
		return errors.New("SendGridConfig.APIKey is required")
	}
	if c.FromAddress == "" {
		return errors.New("SendGridConfig.FromAddress is required")
	}
	if c.FromName == "" {
		return errors.New("SendGridConfig.FromName is required")
	}
	return nil
}

type SESConfig struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	FromAddress     string `yaml:"fromAddress"`
	FromName        string `yaml:"fromName"`
}

func (c *SESConfig) Validate() error {
	if c.Region == "" {
		return errors.New("SESConfig.Region is required")
	}
	if c.AccessKeyID == "" {
		return errors.New("SESConfig.AccessKeyID is required")
	}
	if c.SecretAccessKey == "" {
		return errors.New("SESConfig.SecretAccessKey is required")
	}
	if c.FromAddress == "" {
		return errors.New("SESConfig.FromAddress is required")
	}
	if c.FromName == "" {
		return errors.New("SESConfig.FromName is required")
	}
	return nil
}

type EmailTemplateEndpoints struct {
	Verification  string `yaml:"verification"`
	PasswordReset string `yaml:"passwordReset"`
}

func (c *EmailTemplateEndpoints) Validate() error {
	if c.Verification == "" {
		return errors.New("EmailTemplateEndpoints.Verification is required")
	}
	if c.PasswordReset == "" {
		return errors.New("EmailTemplateEndpoints.PasswordReset is required")
	}
	return nil
}

type SMSConfig struct {
	// One of "twilio", "aws"
	Provider string `yaml:"provider"`
}

func (c *SMSConfig) Validate() error {
	if c == nil {
		return nil
	}
	if c.Provider == "" {
		return nil
		// TODO: Enable validation when SMS is implemented
		// return errors.New("SMSConfig.Provider is required")
	}
	if c.Provider != "twilio" && c.Provider != "aws" {
		return errors.New("SMSConfig.Provider must be one of 'twilio', 'aws'")
	}
	return nil
}

type BrandingConfig struct {
	AppName          string `yaml:"appName"`
	CompanyName      string `yaml:"companyName"`
	CompanyNameShort string `yaml:"companyNameShort"`
	SupportURL       string `yaml:"supportURL"`
}

func (c *BrandingConfig) Validate() error {
	if c.AppName == "" {
		return errors.New("BrandingConfig.AppName is required")
	}
	if c.CompanyName == "" {
		return errors.New("BrandingConfig.CompanyName is required")
	}
	if c.CompanyNameShort == "" {
		// CompanyNameShort can be same as AppName if no short version is available
		c.CompanyNameShort = c.AppName
	}
	if c.SupportURL == "" {
		return errors.New("BrandingConfig.SupportURL is required")
	}
	return nil
}

type SecurityConfig struct {
	AuditLogs    AuditLogsConfig    `yaml:"auditLogs"`
	RateLimiting RateLimitingConfig `yaml:"rateLimiting"`
	APIKeys      []APIKeyConfig     `yaml:"apiKeys"`
	CORS         CORSConfig         `yaml:"cors"`
}

func (c *SecurityConfig) Validate() error {
	for _, apiKey := range c.APIKeys {
		if err := apiKey.Validate(); err != nil {
			return err
		}
	}
	if err := c.RateLimiting.Validate(); err != nil {
		return err
	}
	return nil
}

// No Validation required yet for AuditLogsConfig
type AuditLogsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type RateLimitingConfig struct {
	// Number of requests allowed
	Rate int `yaml:"rate"`

	// Window duration, say if rate is set to 120, then duration defines whether its 120/s 120/m or 120/h
	//
	// Accepted values: s, m, h
	Duration string `yaml:"duration"`
}

func (c *RateLimitingConfig) Validate() error {
	if c.Rate == 0 {
		return errors.New("RateLimitingConfig.Rate is required")
	}
	if strings.TrimSpace(c.Duration) == "" {
		return errors.New("RateLimitingConfig.Duration is required")
	}
	validUnits := map[string]bool{
		"s": true,
		"m": true,
		"h": true,
	}
	if !validUnits[c.Duration] {
		return errors.New("RateLimitingConfig.Duration must be one of 's', 'm', 'h'")
	}
	return nil
}

type APIKeyConfig struct {
	Name string `yaml:"name"`

	// The key itself. No restriction on the length, but keep it sensible please.
	// Recommended format: nexeres_[environment:oneof=live,test,dev,staging]_[random alphanumeric string]
	Key string `yaml:"key"`

	Description string `yaml:"description"`
}

func (c *APIKeyConfig) Validate() error {
	if c.Name == "" {
		return errors.New("APIKeyConfig.Name is required")
	}
	if c.Key == "" {
		return errors.New("APIKeyConfig.Key is required")
	}
	if c.Description == "" {
		return errors.New("APIKeyConfig.Description is required")
	}
	return nil
}

// No Validation required yet for CORSConfig
type CORSConfig struct {
	// Allowed origins for CORS requests.
	// Note: Using "*" WILL NOT work.
	//
	// The following origins are allowed by default:
	//
	// In debug mode, you can set the public.debugBaseURL to allow any url.
	//
	// In production mode, this is allowed:
	// - https://{public.domain}
	// - https://{public.subdomain}.{public.domain}
	//
	// Any other domains, even other subdomains of {public.domain} will not be allowed by default.
	// Allowing a domain WOULD NOT allow all its subdomains.
	// You need to explicitly allow subdomains by specifying them here.
	// The following formats work:
	// allowedOrigins:
	// - "https://*.example.com"
	// - "https://xyz.example.com"
	// - "https://example.com"
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

type StoresConfig struct {
	PostgresDSN string      `yaml:"postgresDSN"`
	Redis       RedisConfig `yaml:"redis"`
	S3          S3Config    `yaml:"s3"`
}

func (c *StoresConfig) Validate() error {
	if c.PostgresDSN == "" {
		return errors.New("StoresConfig.PostgresDSN is required")
	}
	if err := c.Redis.Validate(); err != nil {
		return err
	}
	if err := c.S3.Validate(); err != nil {
		return err
	}
	return nil
}

type RedisConfig struct {
	Address  string  `yaml:"address"`
	Password *string `yaml:"password,omitempty"`
	DB       int     `yaml:"db"`
}

func (c *RedisConfig) Validate() error {
	if c.Address == "" {
		return errors.New("RedisConfig.Address is required")
	}
	return nil
}

type S3Config struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	Region          string `yaml:"region"`
	UseSSL          bool   `yaml:"useSSL"`
	Type            string `yaml:"type"` // one of "aws", "minio", "seaweedfs"
}

func (c *S3Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("S3Config.Endpoint is required")
	}
	if c.AccessKeyID == "" {
		return errors.New("S3Config.AccessKeyID is required")
	}
	if c.SecretAccessKey == "" {
		return errors.New("S3Config.SecretAccessKey is required")
	}
	if c.Region == "" {
		return errors.New("S3Config.Region is required")
	}
	if c.Type == "" {
		return errors.New("S3Config.Type is required")
	}
	validTypes := map[string]bool{
		"aws":       true,
		"minio":     true,
		"seaweedfs": true,
	}
	if !validTypes[c.Type] {
		return errors.New("S3Config.Type must be one of 'aws', 'minio', 'seaweedfs'")
	}
	return nil
}

// Finds configuration files and reports which files exists.
//
// If file exists at provided cfg path, it returns that path only.
// Otherwise, the following paths are checked, and the file that's found first is returned.
//   - /etc/nbrglm/workspace/nexeres/.nexeres.yaml
//   - /home/user/.nexeres.yaml
func findConfigFile(cfg string) (string, error) {
	if cfg != "" && utils.FileExists(cfg) {
		return cfg, nil
	}
	if utils.FileExists("/etc/nbrglm/workspace/nexeres/.nexeres.yaml") {
		return "/etc/nbrglm/workspace/nexeres/.nexeres.yaml", nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	homeFile := path.Join(home, ".nexeres.yaml")
	if utils.FileExists(homeFile) {
		return homeFile, nil
	}
	return "", errors.New("no config file found")
}

var C *NexeresConfig

func Environment() string {
	if opts.Debug {
		return "development"
	}
	return "production"
}

func LoadConfig(cfg string) error {
	C = new(NexeresConfig)
	if v, err := findConfigFile(cfg); err != nil {
		return err
	} else {
		cfg = v
	}
	abs, err := filepath.Abs(cfg)
	if err != nil {
		return err
	}
	bytes, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	expanded := os.ExpandEnv(string(bytes))
	err = yaml.Unmarshal([]byte(expanded), C)
	if err != nil {
		return err
	}
	err = C.Validate()
	if err != nil {
		return err
	}
	opts.ConfigPath = &abs
	opts.Debug = C.Debug
	return nil
}
