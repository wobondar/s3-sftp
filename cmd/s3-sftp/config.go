package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	ConfigFilename = "config.json"
)

// Configuration for the application
type Config struct {
	// S3 Configuration
	S3BucketName string `json:"s3_bucket_name"`
	S3Region     string `json:"s3_region"`
	AWSProfile   string `json:"aws_profile"`

	// SFTP Configuration
	SFTPHost           string        `json:"sftp_host"`
	SFTPPort           int           `json:"sftp_port"`
	SFTPUsername       string        `json:"sftp_username"`
	SFTPPassword       string        `json:"sftp_password"`
	SFTPDirectory      string        `json:"sftp_directory"`
	SFTPConnectTimeout time.Duration `json:"sftp_connect_timeout"`

	// Concurrency Configuration
	ConcurrentS3Downloads int `json:"concurrent_s3_downloads"`
	ConcurrentSFTPUploads int `json:"concurrent_sftp_uploads"`

	// Retry Configuration
	MaxRetries int           `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`
}

func LoadConfig(filename string) (*Config, error) {
	var config Config

	// Read the config file
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config file
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Convert timeout to seconds
	config.SFTPConnectTimeout = config.SFTPConnectTimeout * time.Second
	config.RetryDelay = config.RetryDelay * time.Second

	return &config, nil
}

// ValidateConfig checks if the configuration is valid
func (c *Config) Validate() error {
	// Check S3 configuration
	if c.S3BucketName == "" {
		return fmt.Errorf("s3_bucket_name cannot be empty")
	}
	if c.S3Region == "" {
		return fmt.Errorf("s3_region cannot be empty")
	}

	// Check SFTP configuration
	if c.SFTPHost == "" {
		return fmt.Errorf("sftp_host cannot be empty")
	}
	if c.SFTPPort <= 0 || c.SFTPPort > 65535 {
		return fmt.Errorf("invalid sftp_port: must be between 1 and 65535")
	}
	if c.SFTPUsername == "" {
		return fmt.Errorf("sftp_username cannot be empty")
	}
	if c.SFTPPassword == "" {
		return fmt.Errorf("sftp_password cannot be empty")
	}
	if c.SFTPConnectTimeout <= 0 {
		return fmt.Errorf("sftp_connect_timeout must be positive")
	}

	// Check concurrency configuration
	if c.ConcurrentS3Downloads <= 0 {
		return fmt.Errorf("concurrent_s3_downloads must be positive")
	}
	if c.ConcurrentSFTPUploads <= 0 {
		return fmt.Errorf("concurrent_sftp_uploads must be positive")
	}

	// Check retry configuration
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	if c.RetryDelay <= 0 {
		return fmt.Errorf("retry_delay must be positive")
	}

	return nil
}
