package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// initSFTPClient initializes and returns an SFTP client
func initSFTPClient(app *App) (*sftp.Client, *ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: app.conf.SFTPUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(app.conf.SFTPPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         app.conf.SFTPConnectTimeout,
		ClientVersion:   "SSH-2.0-Cpp",
	}

	addr := fmt.Sprintf("%s:%d", app.conf.SFTPHost, app.conf.SFTPPort)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to SFTP server: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return sftpClient, sshClient, nil
}

// uploadFile uploads a file to SFTP with retry mechanism
func uploadFile(app *App, workerID int, job FileJob, tempFilePath string, sftpClient *sftp.Client) bool {
	app.debug("Upload worker %d: Processing upload for file %s", workerID, job.Key)

	// Open the temporary file for reading
	tempFile, err := os.Open(tempFilePath)
	if err != nil {
		app.log.err.Printf("Upload worker %d: Failed to open temp file: %v", workerID, err)
		return false
	}
	defer tempFile.Close()

	// Create destination directory on SFTP if it doesn't exist
	destDir := filepath.Join(app.conf.SFTPDirectory, job.Folder)
	err = ensureSFTPDirectory(app, sftpClient, destDir)
	if err != nil {
		app.log.err.Printf("Upload worker %d: Failed to create directory %s: %v", workerID, destDir, err)
		return false
	}

	// Upload file to SFTP with retry
	destPath := filepath.Join(destDir, job.File)
	var uploadSuccess bool
	for retry := range app.conf.MaxRetries {
		if retry > 0 {
			app.log.info.Printf("Upload worker %d: Retrying upload for %s (attempt %d/%d)", workerID, destPath, retry+1, app.conf.MaxRetries)
			time.Sleep(app.conf.RetryDelay)
			tempFile.Seek(0, 0) // Reset file pointer for retry
		}

		err = uploadToSFTP(sftpClient, tempFile, destPath)
		if err == nil {
			uploadSuccess = true
			break
		}
		app.log.err.Printf("Upload worker %d: Upload attempt %d failed for %s: %v", workerID, retry+1, destPath, err)
	}

	if !uploadSuccess {
		app.log.err.Printf("Upload worker %d: Failed to upload %s after %d attempts", workerID, destPath, app.conf.MaxRetries)
		return false
	}

	app.debug("Upload worker %d: Successfully uploaded %s to %s", workerID, job.Key, destPath)
	return true
}

// uploadToSFTP uploads a file to the SFTP server
func uploadToSFTP(client *sftp.Client, file *os.File, destPath string) error {
	// For Azure Blob Storage SFTP, we need to ensure the parent directory exists
	dirPath := filepath.Dir(destPath)

	// Try to create the directory one by one
	parts := strings.Split(dirPath, "/")
	currentPath := ""
	for _, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = filepath.Join(currentPath, part)
		}

		// Try to create directory, ignore errors as it might already exist
		client.Mkdir(currentPath)
	}

	// Use OpenFile with specific flags instead of Create
	// This is similar to what's needed for AWS Transfer or Azure Blob Storage SFTP
	destFile, err := client.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to create file on SFTP server: %w", err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, file)
	if err != nil {
		return fmt.Errorf("failed to write file to SFTP server: %w", err)
	}

	return nil
}

// ensureSFTPDirectory ensures that a directory exists on the SFTP server
// Modified to work with Azure Blob Storage SFTP
func ensureSFTPDirectory(app *App, client *sftp.Client, dirPath string) error {
	// For Azure Blob Storage, create directories one by one
	parts := strings.Split(dirPath, "/")
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = filepath.Join(currentPath, part)
		}

		// Try to create directory, ignore errors if it already exists
		err := client.Mkdir(currentPath)
		if err != nil {
			// Check if directory exists
			info, statErr := client.Stat(currentPath)
			if statErr != nil || !info.IsDir() {
				app.log.err.Printf("Warning: Could not create directory %s: %v", currentPath, err)
				// Continue anyway, as the file creation might still work
			}
		}
	}

	return nil
}
