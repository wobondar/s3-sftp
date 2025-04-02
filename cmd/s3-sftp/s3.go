package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// initS3Client initializes and returns an S3 client
func initS3Client(app *App) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(app.conf.S3Region),
		config.WithSharedConfigProfile(app.conf.AWSProfile),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with checksum validation disabled
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.DisableLogOutputChecksumValidationSkipped = true
	})

	return client, nil
}

// downloadFile downloads a file from S3 with retry mechanism
func downloadFile(app *App, workerID int, job FileJob, s3Client *s3.Client) string {
	app.debug("Download worker %d: Processing job for file %s", workerID, job.Key)

	// Create a temporary file to store the downloaded content
	tempFile, err := os.CreateTemp("", "s3-download-*")
	if err != nil {
		app.log.err.Printf("Download worker %d: Failed to create temp file: %v", workerID, err)
		return ""
	}
	tempFilePath := tempFile.Name()

	// Download file from S3 with retry
	var downloadSuccess bool
	for retry := range app.conf.MaxRetries {
		if retry > 0 {
			app.log.info.Printf("Download worker %d: Retrying download for %s (attempt %d/%d)", workerID, job.Key, retry+1, app.conf.MaxRetries)
			time.Sleep(app.conf.RetryDelay)
		}

		err = downloadFromS3(s3Client, app.conf.S3BucketName, job.Key, tempFile)
		if err == nil {
			downloadSuccess = true
			break
		}
		app.log.err.Printf("Download worker %d: Download attempt %d failed for %s: %v", workerID, retry+1, job.Key, err)
	}

	tempFile.Close()

	if !downloadSuccess {
		app.log.err.Printf("Download worker %d: Failed to download %s after %d attempts", workerID, job.Key, app.conf.MaxRetries)
		os.Remove(tempFilePath)
		return ""
	}

	app.debug("Download worker %d: Successfully downloaded %s to %s", workerID, job.Key, tempFilePath)
	return tempFilePath
}

// downloadFromS3 downloads a file from S3 to a local file
func downloadFromS3(client *s3.Client, bucket string, key string, file *os.File) error {
	ctx := context.Background()
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write S3 object to file: %w", err)
	}

	return nil
}
