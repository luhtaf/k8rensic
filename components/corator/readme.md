# Corator

**Coraza-based WAF with Transparent File Interception for Digital Forensics**

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/luhtaf/corator)](https://goreportcard.com/report/github.com/luhtaf/corator)

---

## 📋 Table of Contents

- [Overview](#overview)
- [The Problem](#the-problem)
- [The Solution](#the-solution)
- [Architecture](#architecture)
- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Deployment](#deployment)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## 🎯 Overview

**Corator** is a high-performance reverse proxy that combines Web Application Firewall (WAF) protection with transparent file interception capabilities. Built on top of the powerful Coraza WAF engine, it provides enterprise-grade security while automatically capturing and preserving digital evidence from HTTP requests.

## 🚨 The Problem: Digital Evidence in Stateless Environments

In modern, stateless environments like **Kubernetes**, collecting digital evidence for forensic analysis is a significant challenge:

- **Data Loss**: When pods are terminated or applications crash, uploaded files are lost forever
- **Development Overhead**: Forcing developers to redesign applications for persistent storage is costly and time-consuming
- **Compliance Gaps**: Many organizations struggle to meet digital forensics requirements in cloud-native environments
- **Security Blind Spots**: Traditional WAF solutions don't capture the actual files being processed

## 💡 The Solution: Corator

Corator solves these challenges by providing:

1. **🕵️ Transparent File Interception**: Automatically detects and intercepts files from HTTP requests without code changes
2. **🛡️ Enterprise WAF Protection**: Full integration with Coraza WAF and OWASP Core Rule Set
3. **💾 Persistent Evidence Storage**: Secure upload to S3-compatible storage or local filesystem
4. **📊 Comprehensive Logging**: Structured logging for SIEM integration and audit trails

### How It Works

```
Client Request → Corator (WAF + Interceptor) → Backend Application
                      ↓
              [S3/Local Storage]
```

Corator acts as a transparent proxy that:
- Inspects all incoming traffic with WAF rules
- Detects file uploads (multipart/form-data and Base64-encoded)
- Uploads intercepted files to persistent storage
- Forwards clean requests to your backend application
- Logs all activities for forensic analysis

---

## 🏗️ Architecture

Corator is built with a modular, factory-based architecture:

```
corator/
├── cmd/main.go              # Application entry point
├── config/config.go         # Configuration management
├── detector/                # File detection modules
│   ├── factory.go          # Detector factory
│   ├── file_detector.go    # Multipart file detection
│   ├── base64_detector.go  # Base64 file detection
│   └── type.go             # Detector interfaces
├── waf/                    # WAF integration
│   └── coraza_waf.go      # Coraza WAF wrapper
├── uploader/               # Storage modules
│   ├── factory.go         # Uploader factory
│   ├── local_uploader.go  # Local filesystem storage
│   └── s3_uploader.go     # S3-compatible storage
├── logger/                 # Logging modules
│   ├── factory.go         # Logger factory
│   ├── file_logger.go     # File-based logging
│   └── elastic_logger.go  # Elasticsearch logging
└── handler/               # HTTP request handling
    └── request_handler.go # Main request processor
```

### Key Components

- **Detectors**: Identify files in HTTP requests (multipart, Base64)
- **WAF Engine**: Coraza-based security inspection
- **Uploaders**: Handle file storage (local/S3)
- **Loggers**: Structured logging for different outputs
- **Config**: Environment-based configuration management

---

## ✨ Features

### 🔍 File Detection
- **Multipart Detection**: Automatically detects `multipart/form-data` file uploads
- **Base64 Detection**: Identifies Base64-encoded files in any request field
- **Configurable**: Enable/disable detectors via environment variables
- **Non-intrusive**: Works without modifying your application code

### 🛡️ Security
- **Coraza WAF**: Enterprise-grade Web Application Firewall
- **OWASP Rules**: Built-in protection against common web attacks
- **Custom Rules**: Support for custom Coraza rule sets
- **Real-time Inspection**: All traffic inspected before reaching backend

### 💾 Storage
- **Local Storage**: File system-based storage for development/testing
- **S3 Compatible**: Support for AWS S3 and S3-compatible services
- **Metadata Preservation**: Maintains request context and timestamps
- **Async Upload**: Non-blocking file upload for optimal performance

### 📊 Logging
- **Structured JSON**: Machine-readable log format
- **Multiple Outputs**: File, Elasticsearch, or both
- **SIEM Ready**: Compatible with security information systems
- **Audit Trail**: Complete request and file processing logs

### ⚙️ Configuration
- **Environment Variables**: All configuration via environment variables
- **Hot Reload**: Configuration changes without restart
- **Validation**: Automatic configuration validation
- **Defaults**: Sensible defaults for quick setup

---

## 🚀 Installation

### Prerequisites

- **Go 1.25+**: [Download Go](https://golang.org/dl/)
- **Git**: For cloning the repository

### Quick Start

1. **Clone the repository**:
   ```bash
   git clone https://github.com/luhtaf/corator.git
   cd corator
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build the application**:
   ```bash
   go build -o corator ./cmd/main.go
   ```

4. **Run with default configuration**:
   ```bash
   ./corator
   ```

### Docker Installation

```bash
# Build Docker image
docker build -t corator .

# Run with environment variables
docker run -p 8080:8080 \
  -e SERVER_BACKEND_URL=http://your-backend:3000 \
  -e DETECTORS_ENABLE_FILE=true \
  -e UPLOADER_TYPE=local \
  corator
```

---

## ⚙️ Configuration

Corator uses environment variables for all configuration. Create a `.env` file or set environment variables directly.

### Server Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_LISTEN_ADDRESS` | Address and port to listen on | `:8080` | No |
| `SERVER_BACKEND_URL` | Backend application URL | `http://localhost:3000` | Yes |

### WAF Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `WAF_CORAZA_CONFIG_PATH` | Path to Coraza configuration file | - | No |

### Detector Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DETECTORS_ENABLE_FILE` | Enable multipart file detection | `false` | No |
| `DETECTORS_ENABLE_BASE64` | Enable Base64 file detection | `false` | No |

### Uploader Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `UPLOADER_TYPE` | Storage type (`local` or `s3`) | `local` | No |
| `UPLOADER_LOCAL_PATH` | Local storage directory | `/tmp/uploads` | No |
| `UPLOADER_S3_ENDPOINT` | S3 endpoint URL | - | Yes (if S3) |
| `UPLOADER_S3_BUCKET` | S3 bucket name | - | Yes (if S3) |
| `UPLOADER_S3_REGION` | S3 region | - | Yes (if S3) |
| `UPLOADER_S3_ACCESS_KEY` | S3 access key | - | Yes (if S3) |
| `UPLOADER_S3_SECRET_KEY` | S3 secret key | - | Yes (if S3) |

### Logger Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LOGGER_ENABLE_FILE` | Enable file logging | `false` | No |
| `LOGGER_ENABLE_ELASTIC` | Enable Elasticsearch logging | `false` | No |
| `LOGGER_FILE_PATH` | Log file path | `/tmp/interceptor.log` | No |
| `LOGGER_ELASTIC_URLS` | Elasticsearch URLs (comma-separated) | - | Yes (if Elastic) |
| `LOGGER_ELASTIC_INDEX` | Elasticsearch index name | `coraza-interceptor` | No |

### Example Configuration

```bash
# Server
SERVER_LISTEN_ADDRESS=:8080
SERVER_BACKEND_URL=http://localhost:3000

# WAF
WAF_CORAZA_CONFIG_PATH=/etc/coraza/coraza.conf

# Detectors
DETECTORS_ENABLE_FILE=true
DETECTORS_ENABLE_BASE64=true

# Uploader (Local)
UPLOADER_TYPE=local
UPLOADER_LOCAL_PATH=/tmp/corator_uploads

# Uploader (S3)
# UPLOADER_TYPE=s3
# UPLOADER_S3_ENDPOINT=https://s3.amazonaws.com
# UPLOADER_S3_BUCKET=my-forensics-bucket
# UPLOADER_S3_REGION=us-east-1
# UPLOADER_S3_ACCESS_KEY=your-access-key
# UPLOADER_S3_SECRET_KEY=your-secret-key

# Logger
LOGGER_ENABLE_FILE=true
LOGGER_FILE_PATH=/tmp/corator.log
LOGGER_ENABLE_ELASTIC=true
LOGGER_ELASTIC_URLS=http://localhost:9200
LOGGER_ELASTIC_INDEX=corator-logs
```

---

## 📖 Usage

### Basic Usage

1. **Start Corator**:
   ```bash
   export SERVER_BACKEND_URL=http://your-app:3000
   export DETECTORS_ENABLE_FILE=true
   export UPLOADER_TYPE=local
   ./corator
   ```

2. **Test file upload**:
   ```bash
   # Create test file
   echo "test content" > test.txt
   
   # Upload via Corator
   curl -X POST http://localhost:8080/upload \
     -F "file=@test.txt" \
     -F "user=testuser"
   ```

3. **Check results**:
   ```bash
   # Check uploaded files
   ls -la /tmp/uploads/
   
   # Check logs
   tail -f /tmp/corator.log
   ```

### Advanced Usage

#### Base64 File Detection

```bash
# Enable Base64 detection
export DETECTORS_ENABLE_BASE64=true

# Test with Base64-encoded file
echo "test content" | base64 | curl -X POST http://localhost:8080/api/data \
  -H "Content-Type: application/json" \
  -d '{"file": "'$(echo "test content" | base64)'", "filename": "test.txt"}'
```

#### S3 Storage

```bash
# Configure S3 storage
export UPLOADER_TYPE=s3
export UPLOADER_S3_ENDPOINT=https://s3.amazonaws.com
export UPLOADER_S3_BUCKET=my-forensics-bucket
export UPLOADER_S3_REGION=us-east-1
export UPLOADER_S3_ACCESS_KEY=your-access-key
export UPLOADER_S3_SECRET_KEY=your-secret-key

./corator
```

#### Elasticsearch Logging

```bash
# Configure Elasticsearch logging
export LOGGER_ENABLE_ELASTIC=true
export LOGGER_ELASTIC_URLS=http://localhost:9200
export LOGGER_ELASTIC_INDEX=corator-logs

./corator
```

---

## 🔌 API Reference

Corator acts as a transparent proxy and doesn't expose its own API endpoints. All requests are forwarded to the backend application after processing.

### Request Flow

1. **Client** → **Corator** (port 8080)
2. **Corator** processes request:
   - WAF inspection
   - File detection and interception
   - Logging
3. **Corator** → **Backend** (configured URL)
4. **Backend** → **Corator** → **Client**

### Supported File Types

#### Multipart Form Data
```http
POST /upload HTTP/1.1
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="document.pdf"
Content-Type: application/pdf

[file content]
------WebKitFormBoundary--
```

#### Base64 Encoded
```http
POST /api/data HTTP/1.1
Content-Type: application/json

{
  "file": "JVBERi0xLjQKJcOkw7zDtsO...",
  "filename": "document.pdf"
}
```

### Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "uuid-1234-5678",
  "method": "POST",
  "url": "/upload",
  "client_ip": "192.168.1.100",
  "files": [
    {
      "filename": "document.pdf",
      "size": 1024,
      "content_type": "application/pdf",
      "storage_path": "/tmp/uploads/uuid-1234-5678/document.pdf",
      "detection_type": "multipart"
    }
  ],
  "waf_actions": [],
  "processing_time_ms": 150
}
```

---

## 🚀 Deployment

### Docker Deployment

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o corator ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/corator .
EXPOSE 8080
CMD ["./corator"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: corator
spec:
  replicas: 2
  selector:
    matchLabels:
      app: corator
  template:
    metadata:
      labels:
        app: corator
    spec:
      containers:
      - name: corator
        image: corator:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVER_BACKEND_URL
          value: "http://backend-service:3000"
        - name: DETECTORS_ENABLE_FILE
          value: "true"
        - name: DETECTORS_ENABLE_BASE64
          value: "true"
        - name: UPLOADER_TYPE
          value: "s3"
        - name: UPLOADER_S3_ENDPOINT
          value: "https://s3.amazonaws.com"
        - name: UPLOADER_S3_BUCKET
          value: "forensics-bucket"
        - name: UPLOADER_S3_REGION
          value: "us-east-1"
        - name: UPLOADER_S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: s3-credentials
              key: access-key
        - name: UPLOADER_S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: s3-credentials
              key: secret-key
        - name: LOGGER_ENABLE_ELASTIC
          value: "true"
        - name: LOGGER_ELASTIC_URLS
          value: "http://elasticsearch:9200"
---
apiVersion: v1
kind: Service
metadata:
  name: corator-service
spec:
  selector:
    app: corator
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### Docker Compose

```yaml
version: '3.8'
services:
  corator:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SERVER_BACKEND_URL=http://backend:3000
      - DETECTORS_ENABLE_FILE=true
      - DETECTORS_ENABLE_BASE64=true
      - UPLOADER_TYPE=local
      - UPLOADER_LOCAL_PATH=/tmp/uploads
      - LOGGER_ENABLE_FILE=true
      - LOGGER_FILE_PATH=/tmp/corator.log
    volumes:
      - ./uploads:/tmp/uploads
      - ./logs:/tmp
    depends_on:
      - backend

  backend:
    image: your-backend-app:latest
    ports:
      - "3000:3000"
```

---

## 🛠️ Development

### Project Structure

```
corator/
├── cmd/                    # Application entry points
├── config/                 # Configuration management
├── detector/               # File detection modules
├── waf/                    # WAF integration
├── uploader/               # Storage modules
├── logger/                 # Logging modules
├── handler/                # HTTP request handling
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── README.md              # This file
```

### Building from Source

```bash
# Clone repository
git clone https://github.com/luhtaf/corator.git
cd corator

# Install dependencies
go mod tidy

# Build
go build -o corator ./cmd/main.go

# Run tests
go test ./...

# Run with race detection
go test -race ./...
```

### Adding New Detectors

1. Implement the `Detector` interface in `detector/type.go`
2. Add your detector to the factory in `detector/factory.go`
3. Update configuration in `config/config.go`
4. Add tests for your detector

### Adding New Uploaders

1. Implement the `Uploader` interface in `uploader/type.go`
2. Add your uploader to the factory in `uploader/factory.go`
3. Update configuration in `config/config.go`
4. Add tests for your uploader

### Code Style

- Follow Go conventions and `gofmt`
- Use meaningful variable and function names
- Add comments for exported functions
- Write tests for new functionality
- Use error wrapping with `fmt.Errorf` and `%w`

---

## 🔧 Troubleshooting

### Common Issues

#### Corator Won't Start

**Error**: `Failed to load configuration`
```bash
# Check environment variables
env | grep -E "(SERVER|WAF|DETECTORS|UPLOADER|LOGGER)"

# Verify backend URL is accessible
curl -I http://your-backend:3000
```

**Error**: `Failed to create uploader`
```bash
# Check S3 credentials
aws s3 ls s3://your-bucket

# Check local directory permissions
ls -la /tmp/uploads
```

#### Files Not Being Intercepted

**Check detector configuration**:
```bash
# Verify detectors are enabled
echo $DETECTORS_ENABLE_FILE
echo $DETECTORS_ENABLE_BASE64
```

**Check request format**:
```bash
# Test multipart upload
curl -X POST http://localhost:8080/upload \
  -F "file=@test.txt" \
  -v

# Test Base64 upload
echo "test" | base64 | curl -X POST http://localhost:8080/api \
  -H "Content-Type: application/json" \
  -d '{"file": "'$(echo "test" | base64)'"}'
```

#### WAF Blocking Requests

**Check WAF configuration**:
```bash
# Verify Coraza config file exists
ls -la /etc/coraza/coraza.conf

# Check WAF logs
tail -f /tmp/corator.log | grep -i waf
```

#### Storage Issues

**S3 Upload Failures**:
```bash
# Test S3 connectivity
aws s3 ls s3://your-bucket

# Check credentials
aws sts get-caller-identity
```

**Local Storage Issues**:
```bash
# Check directory permissions
ls -la /tmp/uploads

# Check disk space
df -h /tmp
```

### Log Analysis

#### Enable Debug Logging

```bash
export LOG_LEVEL=debug
./corator
```

#### Common Log Patterns

```bash
# View all intercepted files
grep "file intercepted" /tmp/corator.log

# View WAF actions
grep "waf_action" /tmp/corator.log

# View errors
grep "ERROR" /tmp/corator.log

# View request processing times
grep "processing_time" /tmp/corator.log
```

### Performance Tuning

#### Memory Usage

```bash
# Monitor memory usage
ps aux | grep corator

# Check for memory leaks
go tool pprof http://localhost:6060/debug/pprof/heap
```

#### Throughput Optimization

```bash
# Increase worker threads
export GOMAXPROCS=4

# Monitor request latency
grep "processing_time" /tmp/corator.log | awk '{sum+=$NF; count++} END {print sum/count}'
```

---

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines

- Write tests for new functionality
- Update documentation for API changes
- Follow Go coding conventions
- Add meaningful commit messages
- Include issue numbers in commit messages

### Reporting Issues

When reporting issues, please include:

- Corator version
- Operating system
- Configuration (sanitized)
- Steps to reproduce
- Expected vs actual behavior
- Logs (if applicable)

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Coraza WAF](https://coraza.io/) - The powerful WAF engine that powers Corator
- [OWASP](https://owasp.org/) - For the Core Rule Set that protects against web attacks
- [Go Community](https://golang.org/) - For the excellent ecosystem and tools

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/luhtaf/corator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/luhtaf/corator/discussions)
- **Documentation**: [This README](README.md)

---

**Made with ❤️ for the security community**