# AWS S3 to SFTP File Mover

[![Release](https://github.com/wobondar/s3-sftp/actions/workflows/release.yml/badge.svg)](https://github.com/wobondar/s3-sftp/actions/workflows/release.yml)
![Code style: gofmt](https://img.shields.io/badge/code_style-gofmt-00ADD8.svg)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/wobondar/s3-sftp)

A high-performance CLI utility for transferring a large number of files from AWS S3 to SFTP servers with concurrency, retry mechanisms, and robust logging.

## Features

- **Concurrent Processing**: Configurable concurrent downloads from S3 and uploads to SFTP
- **Flexible Configuration**: JSON-based configuration for easy customization
- **Robust Error Handling**: Automatic retries for failed transfers
- **Comprehensive Logging**: Detailed logs with debug mode for troubleshooting
- **CSV Input**: Process file transfers defined in CSV format

### Prerequisites

- Go 1.23 or higher
- AWS credentials configured (via profile or environment variables)
- Access to both S3 bucket and SFTP server

## Installation

```bash
go install github.com/wobondar/s3-sftp/cmd/s3-sftp@latest
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/wobondar/s3-sftp.git
cd s3-sftp

# Build the application
make all
```

## Usage

```bash
# Basic usage
s3-sftp -csv input.csv

# With debug mode
s3-sftp -d -csv input.csv

# With custom config file
s3-sftp -csv input.csv -c custom_config.json
```

### Command Line Options

- `-c <path>`: Path to the configuration file (default: `config.json`)
- `-csv <path>`: Path to the CSV file containing file transfer definitions (required)
- `-debug`: Enable debug mode for verbose logging
- `-delim <char>`: Delimiter for the CSV file (default: `;`)
- `-version`: Show version information
- `-help`: Show help information


### CSV Format

The CSV file should be semicolon-separated with the following columns:

```csv
folder;file;key;lastmodified;etag;size
```

Example:
```csv
test_1;some_file.pdf;tests/test_1/some_file.pdf;2023-11-17T17:15:25.000Z;"c2760c8065ee374bb92c36a6899f2016";28500
test_1;another_file.pdf;tests/test_1/another_file.pdf;2023-11-17T17:15:25.000Z;"8e985792cda4d1adba490c77a3cb76db";22113
```

### Configuration

The application uses a JSON configuration file (`config.json` by default):

```json
{
    "s3_bucket_name": "your_s3_bucket_name",
    "s3_region": "your_s3_region",
    "aws_profile": "your_aws_profile",
    "sftp_host": "your_sftp_host",
    "sftp_port": 22,
    "sftp_username": "your_sftp_username",
    "sftp_password": "your_sftp_password",
    "sftp_directory": "your_sftp_directory",
    "sftp_connect_timeout": 30,
    "concurrent_s3_downloads": 8,
    "concurrent_sftp_uploads": 16,
    "max_retries": 5,
    "retry_delay": 1
}
```

## Development

### Available Make Targets

```bash
$ make help
Available targets:
  all                - Check dependencies, clean, and build
  check-deps         - Check for required dependencies
  build              - Build the application
  build-all          - Build for multiple platforms (Linux, macOS, Windows)
  clean              - Remove build artifacts
  vet                - Run go vet
  release-check      - Check goreleaser release without publishing
  fmt                - Format code
  mod-tidy           - Tidy Go modules
  run                - Build and run the application
  help               - Show this help message

Example usage:
  make run ARGS="-csv input.csv -c config.json"
  make build-all ARGS="v0.0.1"
```

## Performance Tuning

For optimal performance, adjust the following configuration parameters:

- `concurrent_s3_downloads`: Number of concurrent S3 download operations
- `concurrent_sftp_uploads`: Number of concurrent SFTP upload operations
- `max_retries`: Maximum number of retry attempts for failed operations
- `retry_delay`: Delay in seconds between retry attempts

In my personal experience, the optimal configuration for 4 million files was:

```text
"concurrent_s3_downloads": 80,
"concurrent_sftp_uploads": 128,
"max_retries": 5,
"retry_delay": 2
```

Also, to improve performance, you might consider increasing:
1. maximum number of open file descriptors: `ulimit -n 30000`
2. maximum number of pending connections (backlog)
3. TCP SYN backlog

## Troubleshooting

If you encounter issues:

1. Enable debug mode with the `-d` flag
2. Check AWS credentials and permissions
3. Verify SFTP server connectivity and credentials
4. Ensure the CSV file is properly formatted

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.