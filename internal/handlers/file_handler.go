package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxUploadSize = 10 * 1024 * 1024 // 10 MB
)

// Define allowed MIME types for better security and clarity.
var allowedMIMETypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/heic": "heic",
}

// SuccessPayload is the response for a successful file upload.
type SuccessPayload struct {
	Message string `json:"message" example:"Avatar file uploaded successfully"`
	URL     string `json:"url" example:"https://s3.example.com/avatars/01H8XGJWBWBAQ9JDBQWEXXXXXX.jpg"`
}

type FileHandler struct {
	uploadService     ports.FileUploadService
	specialistService ports.SpecialistService
	logger            *zap.Logger
}

var _ ports.FileHandlers = (*FileHandler)(nil)

func NewFileHandler(fileService ports.FileUploadService, spec ports.SpecialistService, logger *zap.Logger) *FileHandler {
	return &FileHandler{
		uploadService:     fileService,
		specialistService: spec,
		logger:            logger.Named("FileHandler"),
	}
}

// UploadAvatar
// @Summary      Upload specialist avatar
// @Description  Uploads a new avatar for the authenticated specialist. The file should be sent as multipart/form-data with the key 'file'. The server validates file size (max 10MB) and type (jpeg, png, webp, heic).
// @Tags         Specialist
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Avatar file to upload"
// @Success      201  {object}  SuccessPayload "Avatar uploaded successfully"
// @Failure      400  {object}  domain.ErrorResponse "Bad Request: file is required, extension mismatch, or other validation errors"
// @Failure      401  {object}  domain.ErrorResponse "Unauthorized: User is not authenticated"
// @Failure      413  {object}  domain.ErrorResponse "Payload Too Large: File size exceeds the 10MB limit"
// @Failure      415  {object}  domain.ErrorResponse "Unsupported Media Type: File type is not allowed"
// @Failure      500  {object}  domain.ErrorResponse "Internal Server Error"
// @Router       /specialist/avatar [post]
// @Security 	 BearerAuth
func (fh *FileHandler) UploadAvatar(c *gin.Context) {

	// Parse the multipart form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "file is required"})
		return
	}

	userID, ok := getUserIDFromContext(c, fh.logger)
	if !ok {
		return
	}

	// Validate file size
	if fileHeader.Size > maxUploadSize {
		fh.logger.Warn("upload attempt with oversized file",
			zap.String("filename", fileHeader.Filename),
			zap.Int64("size", fileHeader.Size),
		)
		c.JSON(http.StatusRequestEntityTooLarge, domain.ErrorResponse{
			Code:    http.StatusRequestEntityTooLarge,
			Message: fmt.Sprintf("file is too large. Maximum size is %d MB", maxUploadSize/1024/1024),
		})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		fh.logger.Error("failed to open uploaded file",
			zap.String("filename", fileHeader.Filename),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to open file",
		})
		return
	}
	defer src.Close()

	// Read the file content into a buffer to detect MIME type and reuse the reader
	// This is more efficient than reading the header and then seeking.
	buf, err := io.ReadAll(src)
	if err != nil {
		fh.logger.Error("failed to read file content", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "could not process file",
		})
		return
	}

	// Detect the MIME type from the file content
	mtype := mimetype.Detect(buf)

	// Validate the detected MIME type against our allowed list
	expectedExtension, isAllowed := allowedMIMETypes[mtype.String()]
	if !isAllowed {
		fh.logger.Warn("upload attempt with disallowed file type",
			zap.String("filename", fileHeader.Filename),
			zap.String("detected_type", mtype.String()),
		)
		c.JSON(http.StatusUnsupportedMediaType, domain.ErrorResponse{
			Code:    http.StatusUnsupportedMediaType,
			Message: fmt.Sprintf("file type '%s' is not allowed", mtype.String())})
		return
	}

	// Validate file extension against detected type for an extra layer of security
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."))
	if ext != expectedExtension && !(ext == "jpeg" && expectedExtension == "jpg") {
		fh.logger.Warn("file extension does not match detected content type",
			zap.String("filename", fileHeader.Filename),
			zap.String("extension", ext),
			zap.String("detected_extension", expectedExtension),
		)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "file extension mismatch",
		})
		return
	}

	file := domain.FileUpload{
		Name:     fileHeader.Filename,
		Size:     fileHeader.Size,
		MIMEType: mtype.String(),
		Content:  bytes.NewReader(buf),
	}

	// Call the core service
	uploadedFile, err := fh.uploadService.UploadAvatar(c.Request.Context(), &file)
	if err != nil {
		fh.logger.Error("failed to save avatar file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to upload file"})
		return
	}

	err = fh.specialistService.UpdateAvatar(c.Request.Context(), userID, uploadedFile.URL)
	if err != nil {
		fh.logger.Error("failed to update specialist avatar", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "failed to update specialist avatar",
		})
	}

	// Return a successful response
	c.JSON(http.StatusCreated, SuccessPayload{
		Message: "Avatar file uploaded successfully",
		URL:     uploadedFile.URL,
	})

}
