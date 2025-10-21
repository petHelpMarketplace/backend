package handlers

import (
	"bytes"
	"errors"
	"maps"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxUploadSize = 8 * 1024 * 1024 // 8 MB
)

// Define allowed MIME types for better security and clarity.
var allowedMIMETypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/heic": "heic",
}

// SuccessPayload is the response for a successful file upload.
type successAvatarPayload struct {
	Message string `json:"message" example:"Avatar file uploaded successfully"`
	URL     string `json:"url" example:"https://s3.example.com/avatars/01H8XGJWBWBAQ9JDBQWEXXXXXX.jpg"`
}

type successPortfolioPayload struct {
	Message string            `json:"message" example:"Portfolio files uploaded successfully"`
	URLs    map[string]string `json:"urls_map" example:"{\"original_cv.webp\":\"https://storage-provider.com/bucket/user-id/a1b2c3d4-cv.webp\",\"photo.jpg\":\"https://storage-provider.com/bucket/user-id/e5f6g7h8-thumbnail.jpg\"}"`
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
// @Success      201  {object}  successAvatarPayload "Avatar uploaded successfully"
// @Failure      400  {object}  domain.BadRequestError "Bad Request: file is required, extension mismatch, or other validation errors"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized: User is not authenticated"
// @Failure      413  {object}  domain.PayloadTooLargeError "Payload Too Large: File size exceeds the 10MB limit"
// @Failure      415  {object}  domain.UnsupportedMediaTypeError "Unsupported Media Type: File type is not allowed"
// @Failure      500  {object}  domain.InternalServerError "Internal Server Error"
// @Router       /specialist/avatar [post]
// @Security 	 BearerAuth
func (fh *FileHandler) UploadAvatar(c *gin.Context) {

	// Parse the multipart form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "file is required"})
		return
	}

	userID, ok := getUserIDFromContext(c, fh.logger)
	if !ok {
		return
	}

	// --- Start of validation ---
	buf, mtype, ok := validateUploadFile(c, fh.logger, fileHeader)
	if !ok {
		return
	}

	file := domain.FileUpload{
		Name:     fileHeader.Filename,
		Size:     fileHeader.Size,
		MIMEType: mtype.String(),
		Content:  bytes.NewReader(buf),
	}

	specDTO, err := fh.specialistService.ShowByID(c.Request.Context(), userID)
	if err != nil || specDTO.ID == 0 {
		fh.logger.Error("failed to get specialist by ID", zap.Int64("userID", userID), zap.Error(err))
	}

	if specDTO.AvatarURL != "" {
		err = fh.uploadService.DeleteAvatar(c.Request.Context(), specDTO.AvatarURL)
		if err != nil {
			fh.logger.Error("failed to delete old avatar file from S3", zap.Error(err))
		}
	}

	// Call the core service
	uploadedFile, err := fh.uploadService.UploadAvatar(c.Request.Context(), strconv.FormatInt(userID, 10), &file)
	if err != nil {
		fh.logger.Error("failed to save avatar file in repository", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "failed to upload avatar file",
		})
		return
	}

	err = fh.specialistService.UpdateAvatar(c.Request.Context(), userID, uploadedFile.URL)
	if err != nil {
		fh.logger.Error("failed to update specialist avatar URL in DB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "failed to update specialist avatar file in DB",
		})
		return
	}

	// Return a successful response
	c.JSON(http.StatusCreated, successAvatarPayload{
		Message: "Avatar file uploaded successfully",
		URL:     uploadedFile.URL,
	})

}

// UploadPortfolio handles uploading multiple files for a specialist's portfolio.
// @Summary      Upload specialist portfolio images
// @Description  Uploads multiple images for the authenticated specialist's portfolio. Files should be sent as multipart/form-data with the key 'files[]'. The server validates file size (max 8MB each) and type (jpeg, png, webp, heic).
// @Tags         Specialist
// @Accept       multipart/form-data
// @Produce      json
// @Param        files[] formData file true "Portfolio image files to upload"
// @Success      201  {object}  successPortfolioPayload "Portfolio files uploaded successfully"
// @Failure      400  {object}  domain.BadRequestError "Bad Request: file is required, extension mismatch, or other validation errors"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized: User is not authenticated"
// @Failure      413  {object}  domain.PayloadTooLargeError "Payload Too Large: File size exceeds the 10MB limit"
// @Failure      415  {object}  domain.UnsupportedMediaTypeError "Unsupported Media Type: File type is not allowed"
// @Failure      500  {object}  domain.InternalServerError "Internal Server Error"
// @Router       /specialist/portfolio [post]
// @Security 	 BearerAuth
func (fh *FileHandler) UploadPortfolio(c *gin.Context) {
	// Use MultipartForm to handle multiple files
	form, err := c.MultipartForm()
	if err != nil {
		fh.logger.Error("failed to parse multipart form", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "Invalid form data",
		})
		return
	}

	// "files[]" is the key for the file array in the form-data
	fileHeaders := form.File["files[]"]
	if len(fileHeaders) == 0 {
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "No files were uploaded"})
		return
	}

	userID, ok := getUserIDFromContext(c, fh.logger)
	if !ok {
		return
	}

	filesToUpload := make([]*domain.FileUpload, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		// --- Start file validation ---
		buf, mtype, ok := validateUploadFile(c, fh.logger, fileHeader)
		if !ok {
			return
		}

		filesToUpload = append(filesToUpload, &domain.FileUpload{
			Name:     fileHeader.Filename,
			Size:     fileHeader.Size,
			MIMEType: mtype.String(),
			Content:  bytes.NewBuffer(buf),
		})
	}

	// Call the service with the list of files
	uploadedFiles, err := fh.uploadService.UploadPortfolio(c.Request.Context(), strconv.FormatInt(userID, 10), filesToUpload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to upload portfolio files to storage",
		})
		return
	}

	urls := make(map[string]string, len(uploadedFiles))
	for _, file := range uploadedFiles {
		urls[filepath.Base(file.ID)] = file.URL
	}

	err = fh.specialistService.AddImages(c.Request.Context(), userID, slices.Sorted(maps.Values(urls)))
	if err != nil {
		fh.logger.Error("failed to update specialist avatar URL in DB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update specialist portfolio files in DB",
		})
		return
	}

	c.JSON(http.StatusCreated, successPortfolioPayload{
		Message: "Portfolio files uploaded successfully",
		URLs:    urls,
	})
}

// DeletePortfolioImage handles deleting a single image from a specialist's portfolio.
// @Summary      Delete specialist portfolio image
// @Description  Deletes a specific image from the authenticated specialist's portfolio. The full URL of the image to be deleted must be provided as a query parameter.
// @Tags         Specialist
// @Produce      json
// @Param        url query string true "Full URL of the image to delete"
// @Success      200  {object}  domain.SuccessResponse "Image deleted successfully"
// @Failure      400  {object}  domain.BadRequestError "Bad Request: file is required, extension mismatch, or other validation errors"
// @Failure      401  {object}  domain.UnauthorizedError "Unauthorized: User is not authenticated"
// @Failure      404  {object}  domain.NotFoundError "Not Found: Specialist account not found"
// @Failure      500  {object}  domain.InternalServerError "Internal Server Error"
// @Router       /specialist/portfolio/image [delete]
// @Security 	 BearerAuth
func (fh *FileHandler) DeletePortfolioImage(c *gin.Context) {

	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "url query parameter is required",
		})
		return
	}

	userID, ok := getUserIDFromContext(c, fh.logger)
	if !ok {
		return
	}

	if !strings.Contains(imageURL, strconv.FormatInt(userID, 10)) {
		fh.logger.Error("URL doesn't contain user ID", zap.String("url", imageURL), zap.Int64("userID", userID))
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "URL doesn't contain user ID",
		})
		return
	}

	// remove the image URL from the specialist's record in the database.
	err := fh.specialistService.DeleteImage(c.Request.Context(), userID, imageURL)
	if err != nil {
		fh.logger.Error("failed to delete specialist portfolio files from DB", zap.String("url", imageURL), zap.Int64("userID", userID), zap.Error(err))
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrAccountNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    status,
			Message: "Failed to remove portfolio image from DB",
		})
		return
	}

	// If the DB update was successful, proceed to delete the file from storage.
	if err := fh.uploadService.DeletePortfolioImage(c.Request.Context(), imageURL); err != nil {
		fh.logger.Error("failed to delete portfolio image from storage after DB update", zap.String("url", imageURL), zap.Error(err))
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Code:    http.StatusOK,
		Message: "Image deleted successfully.",
	})
}
