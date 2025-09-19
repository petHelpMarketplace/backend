package ports

import (
	"context"
	"pethelp-backend/internal/core/domain"

	"github.com/gin-gonic/gin"
)

// FileHandlers defines the inbound port for handling file-related HTTP requests.
type FileHandlers interface {
	UploadAvatar(c *gin.Context)
}

// FileUploadService defines the application's core business logic for handling file uploads.
// It orchestrates the process of validating and persisting files, acting as an intermediary
// between the transport layer (handlers) and the persistence layer (repositories).
type FileUploadService interface {
	// UploadAvatar processes and saves a specialist's avatar.
	// It takes a FileUpload object, assigns it a unique ID, persists it via the repository,
	// and returns the updated FileUpload object containing the persistent URL.
	UploadAvatar(ctx context.Context, file *domain.FileUpload) (*domain.FileUpload, error)
}

// FileRepository defines the outbound port for file persistence.
// This interface abstracts the underlying storage mechanism (e.g., a local filesystem,
// AWS S3, or Google Cloud Storage), allowing the core application to remain storage-agnostic.
type FileRepository interface {
	// Save persists the given file to the underlying storage.
	// It returns the publicly accessible URL of the stored file or an error if the operation fails.
	Save(ctx context.Context, file *domain.FileUpload) (string, error)
}
