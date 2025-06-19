package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Repository implements the Repository interface for AWS S3 repositories
type S3Repository struct {
	*BaseRepository

	// client is the S3 API client
	client *s3.Client

	// bucketName is the S3 bucket name
	bucketName string

	// prefix is the prefix for all objects in the bucket
	prefix string

	// region is the AWS region
	region string

	// auditLogger is the audit logger for repository operations
	auditLogger *RepositoryAuditLogger
}

// NewS3Repository creates a new S3 repository
func NewS3Repository(config *Config) (Repository, error) {
	// Create base repository
	base := NewBaseRepository(config)

	// Parse S3 URL to extract bucket name and prefix
	bucketName, prefix, region, err := parseS3URL(config.URL)
	if err != nil {
		return nil, err
	}

	// Create audit logger if audit logging is enabled
	var auditLogger *RepositoryAuditLogger
	if config.AuditLogger != nil {
		auditLogger = NewRepositoryAuditLogger(config.AuditLogger, "S3", config.URL)
	}

	return &S3Repository{
		BaseRepository: base,
		bucketName:     bucketName,
		prefix:         prefix,
		region:         region,
		auditLogger:    auditLogger,
	}, nil
}

// parseS3URL parses an S3 URL to extract bucket name, prefix, and region
// Format: s3://bucket-name/prefix?region=us-east-1
func parseS3URL(urlStr string) (string, string, string, error) {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid S3 URL: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "s3" {
		return "", "", "", fmt.Errorf("invalid S3 URL scheme: %s", parsedURL.Scheme)
	}

	// Extract bucket name (host)
	bucketName := parsedURL.Host
	if bucketName == "" {
		return "", "", "", fmt.Errorf("missing bucket name in S3 URL")
	}

	// Extract prefix (path)
	prefix := strings.TrimPrefix(parsedURL.Path, "/")

	// Extract region from query parameters
	region := "us-east-1" // Default region
	if parsedURL.Query().Has("region") {
		region = parsedURL.Query().Get("region")
	}

	return bucketName, prefix, region, nil
}

// Connect establishes a connection to the S3 repository
func (r *S3Repository) Connect(ctx context.Context) error {
	// Log repository connection attempt
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryConnect(ctx, r.config.URL)
	}

	// Check if already connected
	if r.IsConnected() {
		return nil
	}

	// Create AWS config
	var cfg aws.Config
	var err error

	// Create config options
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(r.region),
	}

	// Add credentials if provided
	if r.config.Username != "" && r.config.Password != "" {
		// Use username as access key ID and password as secret access key
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			r.config.Username,
			r.config.Password,
			"", // Session token (optional)
		)))
	}

	// Load AWS config
	cfg, err = config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	r.client = s3.NewFromConfig(cfg)

	// Test connection by checking if bucket exists
	_, err = r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucketName),
	})
	if err != nil {
		r.setLastError(err)
		return fmt.Errorf("failed to connect to S3 bucket: %w", err)
	}

	// Set connected flag
	r.setConnected(true)

	return nil
}

// Disconnect closes the connection to the S3 repository
func (r *S3Repository) Disconnect() error {
	// Log repository disconnection
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryDisconnect(context.Background(), r.config.URL)
	}

	// Set connected flag to false
	r.setConnected(false)

	// Clear client
	r.client = nil

	return nil
}

// ListFiles lists files in the S3 repository matching the pattern
func (r *S3Repository) ListFiles(ctx context.Context, pattern string) ([]FileInfo, error) {
	// Log file listing operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryListFiles(ctx, r.config.URL, pattern)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return nil, err
	}
	defer r.ReleaseConnection()

	// Create result slice
	var result []FileInfo

	// Use WithRetry for the operation
	err := r.WithRetry(ctx, func() error {
		// Create list objects input
		input := &s3.ListObjectsV2Input{
			Bucket: aws.String(r.bucketName),
		}

		// Add prefix if specified
		if r.prefix != "" {
			input.Prefix = aws.String(r.prefix)
		}

		// List objects
		paginator := s3.NewListObjectsV2Paginator(r.client, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return err
			}

			// Process objects
			for _, obj := range page.Contents {
				// Get object key without prefix
				key := *obj.Key
				if r.prefix != "" {
					key = strings.TrimPrefix(key, r.prefix)
					key = strings.TrimPrefix(key, "/")
				}

				// Skip if empty
				if key == "" {
					continue
				}

				// Extract file name
				name := filepath.Base(key)

				// Skip if not matching pattern
				if pattern != "" && !matchPattern(name, pattern) {
					continue
				}

				// Create file info
				fileInfo := FileInfo{
					Path:         key,
					Name:         name,
					Size:         *obj.Size,
					LastModified: *obj.LastModified,
					IsDirectory:  strings.HasSuffix(key, "/"),
				}

				// Add to result
				result = append(result, fileInfo)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetFile retrieves a file from the S3 repository
func (r *S3Repository) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	// Log file retrieval operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryGetFile(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return nil, err
	}

	// Create full object key
	key := path
	if r.prefix != "" {
		key = r.prefix + "/" + path
	}

	// Create a pipe for streaming the content
	pr, pw := io.Pipe()

	// Fetch and write content in a goroutine
	go func() {
		defer pw.Close()

		// Use WithRetry for the operation
		err := r.WithRetry(ctx, func() error {
			// Get object
			resp, err := r.client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(r.bucketName),
				Key:    aws.String(key),
			})
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Copy content to pipe
			_, err = io.Copy(pw, resp.Body)
			return err
		})

		if err != nil {
			pw.CloseWithError(err)
			r.ReleaseConnection()
		}
	}()

	// Create a wrapper for the pipe reader that releases the connection when closed
	return &connectionCloser{
		ReadCloser: pr,
		release: func() {
			r.ReleaseConnection()
		},
		ctx:         ctx,
		auditLogger: r.auditLogger,
		filePath:    path,
		baseURL:     r.config.URL,
	}, nil
}

// FileExists checks if a file exists in the S3 repository
func (r *S3Repository) FileExists(ctx context.Context, path string) (bool, error) {
	// Log file existence check operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryFileExists(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return false, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return false, err
	}
	defer r.ReleaseConnection()

	// Create full object key
	key := path
	if r.prefix != "" {
		key = r.prefix + "/" + path
	}

	// Use WithRetry for the operation
	var exists bool
	err := r.WithRetry(ctx, func() error {
		// Check if object exists
		_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(r.bucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			// Check if error is because object doesn't exist
			if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				exists = false
				return nil
			} else if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "403") {
				// Object might exist but we don't have permission to access it
				exists = true
				return nil
			} else {
				// Some other error
				return err
			}
		}

		// Object exists
		exists = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetBranch returns the branch of the repository
// S3 repositories don't have branches, so this returns an empty string
func (r *S3Repository) GetBranch() string {
	return ""
}

// GetLastModified gets the last modified time of a file in the S3 repository
func (r *S3Repository) GetLastModified(ctx context.Context, path string) (time.Time, error) {
	// Log last modified time retrieval operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryGetLastModified(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return time.Time{}, err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return time.Time{}, err
	}
	defer r.ReleaseConnection()

	// Create full object key
	key := path
	if r.prefix != "" {
		key = r.prefix + "/" + path
	}

	// Use WithRetry for the operation
	var lastModified time.Time
	err := r.WithRetry(ctx, func() error {
		// Get object metadata
		resp, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(r.bucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			return err
		}

		// Extract last modified time
		lastModified = *resp.LastModified
		return nil
	})

	if err != nil {
		return time.Time{}, err
	}

	return lastModified, nil
}

// StoreFile stores a file in the S3 repository
func (r *S3Repository) StoreFile(ctx context.Context, path string, content []byte) error {
	// Log file storage operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryStoreFile(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return err
	}
	defer r.ReleaseConnection()

	// Create full object key
	key := path
	if r.prefix != "" {
		key = r.prefix + "/" + path
	}

	// Determine content type
	contentType := "application/octet-stream"
	if strings.HasSuffix(path, ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		contentType = "application/yaml"
	} else if strings.HasSuffix(path, ".txt") {
		contentType = "text/plain"
	} else if strings.HasSuffix(path, ".md") {
		contentType = "text/markdown"
	}

	// Use WithRetry for the operation
	return r.WithRetry(ctx, func() error {
		// Put object
		_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(r.bucketName),
			Key:         aws.String(key),
			Body:        bytes.NewReader(content),
			ContentType: aws.String(contentType),
		})
		return err
	})
}

// DeleteFile deletes a file from the S3 repository
func (r *S3Repository) DeleteFile(ctx context.Context, path string) error {
	// Log file deletion operation
	if r.auditLogger != nil {
		r.auditLogger.LogRepositoryDeleteFile(ctx, r.config.URL, path)
	}

	// Ensure connected
	if err := r.Connect(ctx); err != nil {
		return err
	}

	// Acquire connection
	if err := r.AcquireConnection(ctx); err != nil {
		return err
	}
	defer r.ReleaseConnection()

	// Create full object key
	key := path
	if r.prefix != "" {
		key = r.prefix + "/" + path
	}

	// Use WithRetry for the operation
	return r.WithRetry(ctx, func() error {
		// Delete object
		_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(r.bucketName),
			Key:    aws.String(key),
		})
		return err
	})
}

// init registers the S3 repository type with the default factory
func init() {
	// Register the S3 repository type
	DefaultFactory.Register("s3", NewS3Repository)
}
