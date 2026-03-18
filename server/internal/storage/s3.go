package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config holds configuration for the S3 storage provider.
type S3Config struct {
	Bucket         string
	Region         string
	Endpoint       string // optional, for S3-compatible stores
	Prefix         string // optional key prefix
	ForcePathStyle bool
}

// S3Provider stores objects in an S3-compatible bucket.
type S3Provider struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewS3Provider creates an S3Provider from the given config.
// Uses the default AWS credential chain (env vars, shared config, EC2 metadata).
func NewS3Provider(ctx context.Context, cfg S3Config) (*S3Provider, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("storage: S3 bucket is required")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("storage: S3 region is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("storage: load AWS config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}
	if cfg.ForcePathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	prefix := strings.Trim(cfg.Prefix, "/")
	return &S3Provider{client: client, bucket: cfg.Bucket, prefix: prefix}, nil
}

// buildKey prepends the optional prefix to the object key.
func (p *S3Provider) buildKey(key string) string {
	if p.prefix == "" {
		return key
	}
	return p.prefix + "/" + key
}

// Put uploads r to S3 at key. Note: S3 PutObject requires content length,
// so we read the full body into memory first.
func (p *S3Provider) Put(ctx context.Context, key string, r io.Reader) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("storage.S3.Put: read body: %w", err)
	}

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(p.buildKey(key)),
		Body:          strings.NewReader(string(body)),
		ContentLength: aws.Int64(int64(len(body))),
	})
	if err != nil {
		return fmt.Errorf("storage.S3.Put: %w", err)
	}
	return nil
}

// Get downloads the object at key and returns a ReadCloser.
func (p *S3Provider) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(p.buildKey(key)),
	})
	if err != nil {
		return nil, fmt.Errorf("storage.S3.Get: %w", err)
	}
	return out.Body, nil
}

// Delete removes the object at key. Idempotent.
func (p *S3Provider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(p.buildKey(key)),
	})
	if err != nil {
		return fmt.Errorf("storage.S3.Delete: %w", err)
	}
	return nil
}

// List returns all keys with the given prefix from S3.
func (p *S3Provider) List(ctx context.Context, prefix string) ([]string, error) {
	s3Prefix := p.buildKey(prefix)
	paginator := s3.NewListObjectsV2Paginator(p.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucket),
		Prefix: aws.String(s3Prefix),
	})

	var keys []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("storage.S3.List: %w", err)
		}
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			k := *obj.Key
			// Strip the provider prefix to return clean keys.
			if p.prefix != "" && strings.HasPrefix(k, p.prefix+"/") {
				k = k[len(p.prefix)+1:]
			}
			keys = append(keys, k)
		}
	}
	return keys, nil
}
