package services

import (
	"context"
	"fmt"
	"path/filepath"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
)

type FileUploadServiceImpl struct {
	uploadRepo ports.FileRepository
	logger     *zap.Logger
}

var _ ports.FileUploadService = (*FileUploadServiceImpl)(nil)

func NewFileUploadService(repo ports.FileRepository, logger *zap.Logger) *FileUploadServiceImpl {

	return &FileUploadServiceImpl{
		uploadRepo: repo,
		logger:     logger.Named("FileUploadService"),
	}
}

func (s *FileUploadServiceImpl) UploadAvatar(ctx context.Context, file *domain.FileUpload) (*domain.FileUpload, error) {

	// Generate a unique ID for the file
	file.ID = "avatars/" + ulid.MustNew(ulid.Timestamp(time.Now()), ulid.DefaultEntropy()).String() + filepath.Ext(file.Name)

	// Call the repository to save the file
	url, err := s.uploadRepo.Save(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file in repository: %w", err)
	}

	// Update the file model with the URL from the repository
	file.URL = url
	file.Content = nil // We don't need to hold the content in memory anymore

	return file, nil

}
