package services

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"strings"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
)

type FileUploadServiceImpl struct {
	fileRepo ports.FileRepository
	logger   *zap.Logger
}

var _ ports.FileUploadService = (*FileUploadServiceImpl)(nil)

func NewFileUploadService(repo ports.FileRepository, logger *zap.Logger) *FileUploadServiceImpl {

	return &FileUploadServiceImpl{
		fileRepo: repo,
		logger:   logger.Named("FileUploadService"),
	}
}

func (fs *FileUploadServiceImpl) UploadAvatar(ctx context.Context, user_id string, file *domain.FileUpload) (*domain.FileUpload, error) {
	fs.logger.Info("Saving object to S3", zap.String("name", file.Name))
	// Generate a unique ID for the file
	var idBuilder strings.Builder
	idBuilder.WriteString("avatars/")

	hash := ulid.Make().String()

	fmt.Fprintf(&idBuilder, "%s/%s%s", user_id, hash, filepath.Ext(file.Name))
	file.ID = idBuilder.String()

	// Call the repository to save the file
	url, err := fs.fileRepo.Save(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file in repository: %w", err)
	}

	// Update the file model with the URL from the repository
	file.URL = url
	file.Content = nil

	fs.logger.Info("Uploaded object to S3", zap.String("id", file.ID))

	return file, nil

}

// DeleteAvatar removes an object from the S3 bucket.
func (fs *FileUploadServiceImpl) DeleteAvatar(ctx context.Context, avatarURL string) error {
	fs.logger.Info("Deleting object from S3", zap.String("url", avatarURL))

	// Parse the URL to safely access its components.
	parsedURL, err := url.Parse(avatarURL)
	if err != nil {
		fs.logger.Error("Failed to parse avatar URL", zap.String("url", avatarURL), zap.Error(err))
		return fmt.Errorf("invalid avatar URL: %w", err)
	}

	cutPrefix := fmt.Sprintf("%s/", fs.fileRepo.Bucket())
	// Extract the key from the URL path.
	// The path for this URL style is `/{bucket-name}/{key}`.
	pathWithoutSlash := strings.TrimPrefix(parsedURL.Path, "/")
	key := strings.TrimPrefix(pathWithoutSlash, cutPrefix)

	if key == "" {
		fs.logger.Error("Extracted an empty key from avatar URL", zap.String("url", avatarURL))
		return fmt.Errorf("could not determine object key from URL: %s", avatarURL)
	}

	// Call the repository's Delete method with the extracted key.
	err = fs.fileRepo.Delete(ctx, key)
	if err != nil {
		fs.logger.Error("failed to delete avatar image from storage", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to delete avatar image from storage: %s - %w", avatarURL, err)
	}

	fs.logger.Info("Successfully deleted avatar image from S3", zap.String("key", key))
	return nil
}

// UploadPortfolio prepares and saves multiple portfolio files.
func (fs *FileUploadServiceImpl) UploadPortfolio(ctx context.Context, user_id string, files []*domain.FileUpload) ([]domain.FileUpload, error) {
	if len(files) == 0 {
		return []domain.FileUpload{}, nil
	}

	var idBuilder strings.Builder
	// Assign a unique ID to each file before saving.
	for _, file := range files {
		idBuilder.WriteString("portfolios/")
		hash := ulid.Make().String()
		fmt.Fprintf(&idBuilder, "%s/%s%s", user_id, hash, filepath.Ext(file.Name))
		file.ID = idBuilder.String()
		idBuilder.Reset()
	}

	urls, err := fs.fileRepo.SaveBatch(ctx, files)
	if err != nil {
		fs.logger.Error("failed to save portfolio batch", zap.Error(err))
		return nil, err
	}

	uploadedFiles := make([]domain.FileUpload, len(urls))
	for i, url := range urls {
		uploadedFiles[i] = domain.FileUpload{
			ID:   files[i].ID,
			Name: files[i].Name,
			Size: files[i].Size,
			URL:  url,
		}
	}

	fs.logger.Info("Successfully uploaded portfolio files", zap.Int("count", len(urls)))
	return uploadedFiles, nil
}

// DeletePortfolioImage removes an object from the S3 bucket.
func (fs *FileUploadServiceImpl) DeletePortfolioImage(ctx context.Context, imageURL string) error {
	fs.logger.Info("Deleting portfolio image from S3", zap.String("url", imageURL))

	// Parse the URL to safely access its components.
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		fs.logger.Error("Failed to parse portfolio image URL", zap.String("url", imageURL), zap.Error(err))
		return fmt.Errorf("invalid image URL: %w", err)
	}

	cutPrefix := fmt.Sprintf("%s/", fs.fileRepo.Bucket())
	// Extract the key from the URL path.
	pathWithoutSlash := strings.TrimPrefix(parsedURL.Path, "/")
	key := strings.TrimPrefix(pathWithoutSlash, cutPrefix)

	if key == "" {
		fs.logger.Error("Extracted an empty key from portfolio image URL", zap.String("url", imageURL))
		return fmt.Errorf("could not determine object key from URL: %s", imageURL)
	}

	if err := fs.fileRepo.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete portfolio image from storage: %w", err)
	}

	return nil
}
