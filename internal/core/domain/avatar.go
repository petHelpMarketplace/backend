package domain

import (
	"io"
)

// FileUpload represents the core domain model for a upload file.
type FileUpload struct {
	ID       string
	Name     string
	Size     int64
	MIMEType string
	Content  io.Reader // The actual file content
	URL      string    // URL after upload
}
