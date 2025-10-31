// internal/handlers/helpers.go (New File)
package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"pethelp-backend/internal/core/domain"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// getUserIDFromContext extracts, validates, and parses the userID from the Gin context.
// It handles errors and writes the appropriate JSON response if any step fails.
// It returns the userID and a boolean indicating success.
func getUserIDFromContext(c *gin.Context, logger *zap.Logger) (int64, bool) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		logger.Warn("userID not found in context, middleware might not have run or failed")
		c.JSON(http.StatusUnauthorized, domain.UnauthorizedError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized user access attempt",
		})
		return 0, false
	}

	userIDStr, ok := userIDRaw.(string)
	if !ok {
		logger.Error("userID in context is not a string", zap.Any("type", fmt.Sprintf("%T", userIDRaw)))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
		})
		return 0, false
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Error("failed to parse userID from context", zap.String("userID", userIDStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error"})
		return 0, false
	}
	if userID < 0 {
		logger.Error("invalid userID in context: negative value not allowed", zap.String("userID", userIDStr), zap.Int64("parsed_userID", userID))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error"})
		return 0, false
	}

	return userID, true
}

func validateUploadFile(c *gin.Context, logger *zap.Logger, fileH *multipart.FileHeader) ([]byte, *mimetype.MIME, bool) {

	src, err := fileH.Open()
	if err != nil {
		logger.Error("failed to open uploaded file",
			zap.String("filename", fileH.Filename),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "failed to open file",
		})
		return nil, nil, false
	}
	defer src.Close()

	// Read the file content into a buffer to detect MIME type and reuse the reader
	limited := io.LimitReader(src, maxUploadSize+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		logger.Error("failed to read file content", zap.Error(err))
		c.JSON(http.StatusInternalServerError, domain.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "could not process file",
		})
		c.Abort()
		return nil, nil, false
	}

	if int64(len(buf)) > maxUploadSize {
		logger.Warn("upload attempt exceeded server read cap",
			zap.String("filename", fileH.Filename),
			zap.Int("read_bytes", len(buf)),
		)
		c.JSON(http.StatusRequestEntityTooLarge, domain.PayloadTooLargeError{
			Code:    http.StatusRequestEntityTooLarge,
			Message: fmt.Sprintf("file is too large. Maximum size is %d MB", maxUploadSize/1024/1024),
		})
		c.Abort()
		return nil, nil, false
	}

	// Detect the MIME type from the file content
	mtype := mimetype.Detect(buf)

	// Validate the detected MIME type against allowed list
	expectedExtension, isAllowed := allowedMIMETypes[mtype.String()]
	if !isAllowed {
		logger.Warn("upload attempt with disallowed file type",
			zap.String("filename", fileH.Filename),
			zap.String("detected_type", mtype.String()),
		)
		c.JSON(http.StatusUnsupportedMediaType, domain.UnsupportedMediaTypeError{
			Code:    http.StatusUnsupportedMediaType,
			Message: fmt.Sprintf("file type '%s' is not allowed", mtype.String())})
		c.Abort()
		return nil, mtype, false
	}

	// Validate file extension against detected type for an extra layer of security
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileH.Filename), "."))
	if ext != expectedExtension && !(ext == "jpeg" && expectedExtension == "jpg") {
		logger.Warn("file extension does not match detected content type",
			zap.String("filename", fileH.Filename),
			zap.String("extension", ext),
			zap.String("detected_extension", expectedExtension),
		)
		c.JSON(http.StatusBadRequest, domain.BadRequestError{
			Code:    http.StatusBadRequest,
			Message: "file extension mismatch",
		})
		c.Abort()
		return nil, mtype, false
	}

	return buf, mtype, true
}
