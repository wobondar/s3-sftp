package main

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// processJobs processes all jobs with concurrency
func processJobs(app *App, jobs []FileJob, s3Client *s3.Client) {
	startTime := time.Now()
	app.log.info.Printf("Starting to process %d jobs", len(jobs))

	// Create a channel for jobs with appropriate buffer size
	jobCh := make(chan FileJob, len(jobs))

	// Create a channel for downloaded files ready for upload
	uploadCh := make(chan struct {
		job      FileJob
		tempPath string
	}, len(jobs)) // Buffer all possible uploads

	// Create a channel to track completed jobs
	completedCh := make(chan FileJob, len(jobs))

	// Fill the job channel
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	// Create wait groups for download and upload workers
	var downloadWg, uploadWg sync.WaitGroup
	downloadWg.Add(app.conf.ConcurrentS3Downloads)
	uploadWg.Add(app.conf.ConcurrentSFTPUploads)

	// Track statistics
	var (
		downloadCount       int32
		uploadCount         int32
		failedDownloadCount int32
		failedUploadCount   int32
	)

	// Start download workers
	for i := 0; i < app.conf.ConcurrentS3Downloads; i++ {
		go func(workerID int) {
			defer downloadWg.Done()
			for job := range jobCh {
				tempPath := downloadFile(app, workerID, job, s3Client)
				if tempPath != "" {
					atomic.AddInt32(&downloadCount, 1)
					uploadCh <- struct {
						job      FileJob
						tempPath string
					}{job, tempPath}
				} else {
					atomic.AddInt32(&failedDownloadCount, 1)
				}
			}
		}(i + 1)
	}

	// Start a goroutine to close the upload channel when all downloads are done
	go func() {
		downloadWg.Wait()
		close(uploadCh)
		app.log.info.Printf("All downloads completed. Downloaded: %d, Failed: %d", atomic.LoadInt32(&downloadCount), atomic.LoadInt32(&failedDownloadCount))
	}()

	// Start upload workers
	for i := 0; i < app.conf.ConcurrentSFTPUploads; i++ {
		go func(workerID int) {
			defer uploadWg.Done()

			// Each upload worker creates its own SFTP connection
			sftpClient, sshClient, err := initSFTPClient(app)
			if err != nil {
				app.log.err.Printf("Upload worker %d: Failed to initialize SFTP client: %v", workerID, err)
				return
			}
			defer sshClient.Close()
			defer sftpClient.Close()

			for upload := range uploadCh {
				success := uploadFile(app, workerID, upload.job, upload.tempPath, sftpClient)
				if success {
					os.Remove(upload.tempPath)
					atomic.AddInt32(&uploadCount, 1)
					completedCh <- upload.job
				} else {
					sshClient.Close()
					sftpClient.Close()
					// try to reinitialize the SFTP client once in case of failure
					app.log.info.Printf("Upload worker %d: Reinitialize SFTP client", workerID)
					sftpClient, sshClient, err = initSFTPClient(app)
					if err != nil {
						// Yes, I'm too lazy to use for loop here with attempt counter. It's not a big deal for now.
						// If it becomes a problem, I'll refactor it with configurable retry attempts in the config.
						app.log.info.Printf("Upload worker %d: Failed to initialize SFTP client on second attempt: %v. Waiting 20 seconds before final attempt...", workerID, err)
						time.Sleep(20 * time.Second)
						app.log.info.Printf("Upload worker %d: Making final attempt to initialize SFTP client", workerID)
						sftpClient, sshClient, err = initSFTPClient(app)
						if err != nil {
							app.log.err.Printf("Upload worker %d: Failed to initialize SFTP client on final attempt: %v", workerID, err)
							// Log the failed file information for later retry
							app.log.err.Printf("Upload worker %d: Failed to upload file after 3 attempts to initialize SFTP client", workerID)
							app.log.err.Printf("\nRetry this file manually.\n\tFailed file: %s/%s\n\tKey: %s\n",
								upload.job.Folder,
								upload.job.File,
								upload.job.Key)
							return
						}
					}
					success1 := uploadFile(app, workerID, upload.job, upload.tempPath, sftpClient)
					if success1 {
						atomic.AddInt32(&uploadCount, 1)
						completedCh <- upload.job
					} else {
						atomic.AddInt32(&failedUploadCount, 1)
					}
					os.Remove(upload.tempPath)
				}
			}
		}(i + 1)
	}

	// Start a goroutine to close the completed channel when all uploads are done
	go func() {
		uploadWg.Wait()
		close(completedCh)
		app.log.info.Printf("All uploads completed. Uploaded: %d, Failed: %d", atomic.LoadInt32(&uploadCount), atomic.LoadInt32(&failedUploadCount))
	}()

	// Count completed jobs
	completedCount := 0
	for range completedCh {
		completedCount++

		// Print progress every 10 jobs
		if (completedCount <= 300 && completedCount%10 == 0) || completedCount%100 == 0 {
			elapsed := time.Since(startTime)
			rate := float64(completedCount) / elapsed.Seconds()
			app.log.info.Printf("Progress: %d/%d jobs completed (%.2f jobs/sec)", completedCount, len(jobs), rate)
		}
	}

	// Print final statistics
	elapsed := time.Since(startTime)
	rate := float64(completedCount) / elapsed.Seconds()
	app.log.info.Printf("Completed %d/%d jobs in %.2f seconds (%.2f jobs/sec)", completedCount, len(jobs), elapsed.Seconds(), rate)
}
