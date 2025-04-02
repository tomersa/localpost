package util

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// LoadCookies reads runtime cookies from .ephemeral (YAML).
func LoadCookies() (map[string]string, error) {
	data, err := os.ReadFile(EphemeralFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("error reading %s: %v", EphemeralFilePath, err)
	}
	var cookies map[string]string
	if err := yaml.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", EphemeralFilePath, err)
	}
	return cookies, nil
}

// SaveCookies writes runtime cookies to .ephemeral (YAML).
func SaveCookies(cookies map[string]string) error {
	data, err := yaml.Marshal(cookies)
	if err != nil {
		return fmt.Errorf("error marshaling cookies: %v", err)
	}
	return os.WriteFile(EphemeralFilePath, data, 0644)
}

// ClearCookies removes the .cookies file.
func ClearCookies() error {
	err := os.Remove(EphemeralFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
	}

	return err
}

// SetCookie adds or updates a single ephemeral in .cookies.
func SetCookie(name, value string) error {
	cookies, err := LoadCookies()
	if err != nil {
		return fmt.Errorf("error loading cookies: %v", err)
	}
	cookies[name] = value
	return SaveCookies(cookies)
}
