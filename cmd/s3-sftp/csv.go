package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// FileJob represents a file to be moved from S3 to SFTP
type FileJob struct {
	Folder       string
	File         string
	Key          string
	LastModified string
	ETag         string
	Size         int64
}

// parseCSV reads the CSV file and returns a slice of FileJob
func parseCSV(delim, filePath string) ([]FileJob, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = rune(delim[0])

	// Read header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var jobs []FileJob
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		size, err := strconv.ParseInt(record[5], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file size: %w", err)
		}

		job := FileJob{
			Folder:       record[0],
			File:         record[1],
			Key:          record[2],
			LastModified: record[3],
			ETag:         strings.Trim(record[4], "\""),
			Size:         size,
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}
