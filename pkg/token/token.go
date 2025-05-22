package token

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// TokenDir is the directory where the token file is stored
	TokenDir = ".telegraphcl"
	// TokenFile is the name of the token file
	TokenFile = "telegraph.token"
)

// GetTokenPath returns the path to the token file
func GetTokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	tokenDir := filepath.Join(home, TokenDir)
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create token directory: %v", err)
	}

	return filepath.Join(tokenDir, TokenFile), nil
}

// SaveToken saves the token to the token file
func SaveToken(token string) error {
	tokenPath, err := GetTokenPath()
	if err != nil {
		return err
	}

	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %v", err)
	}

	return nil
}

// GetToken returns the token from the token file
func GetToken() (string, error) {
	tokenPath, err := GetTokenPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		return "", fmt.Errorf("token file doesn't exist, please create a user first")
	}

	token, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %v", err)
	}

	return string(token), nil
}
