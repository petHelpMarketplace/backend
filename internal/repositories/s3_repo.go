package repositories

import (
	"context"
	"fmt"
	"pethelp-backend/internal/config"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Repository struct {
	client     *s3.Client
	endpoint   string
	bucketName string
}

var _ ports.FileRepository = (*s3Repository)(nil)

// NewS3Repository creates a new S3 repository adapter.
func NewS3Repository(cfg config.S3Config) (*s3Repository, error) {

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithBaseEndpoint(cfg.EndpointURL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// Path‑style is required for many S3‑compatible providers and dotted bucket names over TLS
		if cfg.EndpointURL != "" {
			o.UsePathStyle = true
		}
	})

	return &s3Repository{
		client:     s3Client,
		endpoint:   cfg.EndpointURL,
		bucketName: cfg.BucketName,
	}, nil
}

func (r *s3Repository) Save(ctx context.Context, file *domain.FileUpload) (string, error) {

	inputObj := s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(file.ID), // Use the unique ID as the object key
		Body:          file.Content,
		ContentType:   aws.String(file.MIMEType),
		ContentLength: aws.Int64(file.Size),
		// Grant READ permission to a specific AWS account via its Canonical User ID
		// GrantRead: aws.String("uri=\"http://acs.amazonaws.com/groups/global/AllUsers\""),
		// You could also use a canned ACL like:
		// ACL: types.ObjectCannedACLPublicRead, // Makes object publicly readable (DANGEROUS without good reason)
	}

	uploader := manager.NewUploader(r.client)
	_, err := uploader.Upload(ctx, &inputObj)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct the URL. Use the custom endpoint if provided, otherwise default to AWS S3 format.
	var url string
	if r.endpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", r.endpoint, r.bucketName, file.ID)
	} else {
		region := r.client.Options().Region
		if region == "" {
			region, err = manager.GetBucketRegion(ctx, r.client, r.bucketName)
			if err != nil {
				return "", fmt.Errorf("failed to get bucket region: %w", err)
			}
		}
		url = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", r.bucketName, region, file.ID)
	}

	return url, nil

}

// Delete removes an object from the S3 bucket by its key.
func (r *s3Repository) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	_, err := r.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object with key '%s' from S3: %w", key, err)
	}

	return nil
}

// Bucket method return bucket name
func (r *s3Repository) Bucket() string {
	return r.bucketName
}
