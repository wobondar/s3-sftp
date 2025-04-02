package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	Version   string = "dev"
	Commit    string = "dev"
	BuildTime string = "dev"
)

type App struct {
	conf *Config
	log  struct {
		info      *log.Logger
		err       *log.Logger
		debug     *log.Logger
		debugMode bool
	}
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "S3 to SFTP %s\n\n", Version)
		fmt.Fprintf(flag.CommandLine.Output(), "Transfer files from AWS S3 to SFTP servers.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "Example:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -csv /path/to/input.csv\n", os.Args[0])
	}

	// Define command line flags
	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, "Show version information")
	flag.BoolVar(&versionFlag, "v", false, "")
	var (
		debugFlag    = flag.Bool("debug", false, "Enable debug mode")
		configPath   = flag.String("c", ConfigFilename, "Path to config file")
		csvFilePath  = flag.String("csv", "", "Path to input CSV file (required)")
		csvDelimiter = flag.String("delim", ",", "CSV delimiter")
	)

	// Parse command line flags
	flag.Parse()

	if versionFlag {
		fmt.Printf("s3-sftp %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		os.Exit(0)
	}

	app := &App{
		log: struct {
			info      *log.Logger
			err       *log.Logger
			debug     *log.Logger
			debugMode bool
		}{
			debugMode: *debugFlag,
		},
	}

	// Set up logging
	setupAppLogger(app)

	// Check if CSV file path is provided
	if *csvFilePath == "" {
		app.log.err.Printf("CSV file path is required. Use -csv flag to specify the path.\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	conf, err := LoadConfig(*configPath)
	if err != nil {
		app.log.err.Fatalf("Failed to load configuration: %v", err)
	}

	app.conf = conf
	app.debug("Configuration loaded: %+v", app.conf)

	// Parse CSV file
	jobs, err := parseCSV(*csvDelimiter, *csvFilePath)
	if err != nil {
		app.log.err.Fatalf("Failed to parse CSV file: %v", err)
	}
	app.log.info.Printf("Loaded %d jobs from CSV file", len(jobs))

	// Initialize AWS S3 client
	s3Client, err := initS3Client(app)
	if err != nil {
		app.log.err.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// Process jobs with concurrency
	processJobs(app, jobs, s3Client)

	app.log.info.Println("All jobs completed successfully")
}
