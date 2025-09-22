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
type FileUploadService interface {
	// UploadAvatar processes and saves a specialist's avatar.
	// It takes a FileUpload object, assigns it a unique ID, persists it via the repository,
	// and returns the updated FileUpload object containing the persistent URL.
	UploadAvatar(ctx context.Context, userID string, file *domain.FileUpload) (*domain.FileUpload, error)

	// DeleteAvatar removes a specialist's avatar from storage.
	// It takes the public URL of the avatar, extracts the object key, and calls the repository to delete it.
	DeleteAvatar(ctx context.Context, avatar_url string) error
}

// FileRepository defines the outbound port for file persistence.
// This interface abstracts the underlying storage mechanism (e.g., a local filesystem,
// AWS S3, or Google Cloud Storage), allowing the core application to remain storage-agnostic.
type FileRepository interface {
	// Save persists the given file to the underlying storage.
	// It returns the publicly accessible URL of the stored file or an error if the operation fails.
	Save(ctx context.Context, file *domain.FileUpload) (string, error)

	// Delete removes a file from the underlying storage using its unique key.
	// The key corresponds to the object's identifier in the storage system (e.g., the object key in S3).
	Delete(ctx context.Context, key string) error

	// Bucket returns the name of the bucket where the files are stored.
	Bucket() string
}
