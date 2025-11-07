package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type ConfigFinder struct {
	Paths []string
	ToUse string
}

// Find searches for configuration files in the specified Paths.
func (cf *ConfigFinder) Find() ([]string, error) {

	var found []string
	for _, p := range cf.Paths {
		fmt.Printf("[%s]\n", p)
		_, err := os.Stat(p)
		if err == nil {
			found = append(found, p)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return found, nil
}

// Use sets the configuration file to be used.
func (cf *ConfigFinder) Use(path string) {
	cf.ToUse = path
}

// Read reads the configuration file specified in ToUse.
//
// If ToUse is empty, it returns an error.
//
// Supports env variables in the config file by expanding the ${var} or $var syntax using os.ExpandEnv.
func (cf *ConfigFinder) Read() ([]byte, error) {
	if cf.ToUse == "" {
		return nil, fs.ErrNotExist
	}

	data, err := os.ReadFile(cf.ToUse)
	if err != nil {
		return nil, err
	}
	expandedData := os.ExpandEnv(string(data))
	return []byte(expandedData), nil
}
