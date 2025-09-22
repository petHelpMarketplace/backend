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
		fs.logger.Error("File repository failed to delete avatar", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to delete avatar from repositor: %s - %w", avatarURL, err)
	}

	fs.logger.Info("Successfully deleted avatar from S3", zap.String("key", key))
	return nil
}
