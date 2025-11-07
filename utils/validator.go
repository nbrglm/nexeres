package utils

import (
	"net"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	nonstd_validators "github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/google/uuid"
	"github.com/nbrglm/nexeres/opts"
)

func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("password", ValidatePasswordRequirements)
	v.RegisterValidation("notblank", nonstd_validators.NotBlank)
	v.RegisterValidation("uuidv7", ValidateUUIDV7)
	v.RegisterValidation("domain", ValidateDomain)
	v.RegisterValidation("urlslug", ValidateURLSlug)
}

var Validator *validator.Validate

func InitValidator() {
	if Validator != nil {
		return
	}
	Validator = validator.New()
	RegisterCustomValidators(Validator)
}

// Validates whether the field is a valid UUID v7
func ValidateUUIDV7(fl validator.FieldLevel) bool {
	val := ""
	field := fl.Field()
	switch field.Kind() {
	case reflect.Pointer:
		valType := field.Elem().Kind()
		if valType != reflect.String {
			return false
		}
		val = field.Elem().String()
	case reflect.String:
		val = field.String()
	default:
		return false
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return false
	}

	return id.Version() == 7
}

var allowedSpecialCharactersForPassword = "-_*@."

func ValidatePasswordRequirements(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	length := len(password)

	if length < 8 || length > 32 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecialSymbol bool

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case containsRunes(allowedSpecialCharactersForPassword, c):
			hasSpecialSymbol = true
		case unicode.IsSpace(c):
			// Do not allow spaces in passwords
			return false
		default:
			// We explicitly disallow other characters
			// This helps in preventing any kind of injection attacks
			// as well as increases the performance
			return false
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecialSymbol
}

func containsRunes(runes string, runeToFind rune) bool {
	for _, r := range runes {
		if r == runeToFind {
			return true
		}
	}
	return false
}

// domainRegex is a regex pattern to validate domain names.
var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,63}$`)

// Block internal/K8s suffixes
var blockedSuffixes = []string{
	".svc", ".svc.cluster.local", ".cluster.local", ".local",
}

// ValidateDomain checks if the provided field is a valid domain.
//
// Only valid format is: "example.com" or "sub.example.com"
func ValidateDomain(fl validator.FieldLevel) bool {
	domain1 := fl.Field().String()
	if domain1 == "" {
		return false
	}

	if opts.Debug && domain1 == "localhost" {
		return true
	}

	// A simple check for domain format
	if len(domain1) < 3 || len(domain1) > 253 {
		return false
	}

	// Basic sanity checks
	if strings.Contains(domain1, "://") {
		return false
	}

	// Lowercase for safety
	d := strings.ToLower(strings.TrimSpace(domain1))

	// Reject raw IP addresses
	if ip := net.ParseIP(d); ip != nil {
		// It's an IP address, not a domain
		return false
	}

	// Basic format check: must contain a dot
	if !strings.Contains(d, ".") {
		return false
	}

	for _, suffix := range blockedSuffixes {
		if strings.HasSuffix(d, suffix) {
			return false
		}
	}

	return domainRegex.MatchString(domain1)
}

// ValidateURLSlug checks if the provided field is a valid URL slug.
//
// Only valid format is: "example-slug" or "example_slug" or "example123"
func ValidateURLSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	if slug == "" {
		return false
	}

	// A simple check for slug format
	if len(slug) < 1 || len(slug) > 50 {
		return false
	}

	for _, c := range slug {
		if !(unicode.IsLower(c) || unicode.IsDigit(c) || c == '-' || c == '_') {
			return false
		}
	}

	return true
}
