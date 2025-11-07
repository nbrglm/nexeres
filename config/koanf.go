package config

import (
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/nbrglm/nexeres/opts"
	"github.com/nbrglm/nexeres/utils"
)

var (
	k      = koanf.New(".")
	parser = yaml.Parser()
)

func InitKoanf(cfgPath string) error {
	var cf *ConfigFinder

	if cfgPath != "" {
		cf = &ConfigFinder{
			Paths: []string{cfgPath},
		}
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		home = path.Join(home, string(os.PathSeparator), ".nexeres", "nexeres.yaml")

		cf = &ConfigFinder{
			Paths: []string{
				"./nexeres.yaml",
				path.Join(string(os.PathSeparator), "etc", "nbrglm", "workspace", "nexeres", "nexeres.yaml"),
				home,
			},
		}
	}

	found, err := cf.Find()
	if err != nil {
		return err
	}

	if len(found) == 0 {
		return fmt.Errorf("no configuration file found in paths: %v", cf.Paths)
	}

	cf.Use(found[0])

	bytes, err := cf.Read()
	if err != nil {
		return err
	}

	if err := k.Load(rawbytes.Provider(bytes), parser); err != nil {
		return err
	}

	k, err = defaultConfig.KoanfMerge(k)
	if err != nil {
		return err
	}

	if err := k.UnmarshalWithConf("", &C, koanf.UnmarshalConf{
		Tag: "yaml",
	}); err != nil {
		return err
	}

	setPostDefaults()

	// Set the debug mode value before validation
	opts.Debug = C.Debug

	fmt.Printf("Using configuration file: \n%#+v\n", C)

	if err := utils.Validator.Struct(C); err != nil {
		return ConfigError{Message: "Configuration validation failed", UnderlyingError: err}
	}

	return nil
}

func setPostDefaults() {
	// The config has NOT been VALIDATED yet. So we have to be careful about what we set here.
	// We can only set values that are not validated or are optional.

	// Public URL
	if strings.TrimSpace(C.Public.DebugBaseURL) == "" {
		if C.Debug {
			scheme := "http"
			if C.Server.TLSConfig {
				scheme = "https"
			}
			C.Public.DebugBaseURL = fmt.Sprintf("%s://%s:%s", scheme, C.Server.Host, C.Server.Port)
		}
		// No debug base URL in production mode
	}

	jwtAudiences := C.JWT.Audiences
	if len(jwtAudiences) == 0 {
		jwtAudiences = []string{C.Public.Domain, fmt.Sprintf("%s.%s", C.Public.SubDomain, C.Public.Domain)}
	} else {
		if !slices.Contains(jwtAudiences, C.Public.Domain) {
			jwtAudiences = append(jwtAudiences, C.Public.Domain)
		}
		subdomain := fmt.Sprintf("%s.%s", C.Public.SubDomain, C.Public.Domain)
		if !slices.Contains(jwtAudiences, subdomain) {
			jwtAudiences = append(jwtAudiences, subdomain)
		}
	}
	C.JWT.Audiences = jwtAudiences

	// Logs configuration
	if strings.TrimSpace(C.Observability.Logs.Level) == "" {
		if C.Debug {
			C.Observability.Logs.Level = "debug" // Default to debug in debug mode
		} else {
			C.Observability.Logs.Level = "info" // Default to info in production mode
		}
	}

	// Branding defaults
	if C.Branding.AppName != "" && C.Branding.CompanyNameShort == "" {
		C.Branding.CompanyNameShort = C.Branding.AppName
	}
}
