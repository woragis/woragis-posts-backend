package publications

import "fmt"

// PublicationError represents a publication-specific error.
type PublicationError struct {
	Code    string
	Message string
	Err     error
}

func (e *PublicationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodePublicationNotFound     = "PUBLICATION_NOT_FOUND"
	ErrCodePlatformNotFound        = "PLATFORM_NOT_FOUND"
	ErrCodePublicationPlatformDup  = "PUBLICATION_PLATFORM_EXISTS"
	ErrCodeMediaNotFound           = "MEDIA_NOT_FOUND"
	ErrCodeInvalidContentType      = "INVALID_CONTENT_TYPE"
	ErrCodeInvalidStatus           = "INVALID_STATUS"
	ErrCodeInvalidMediaType        = "INVALID_MEDIA_TYPE"
	ErrCodeUnauthorized            = "UNAUTHORIZED"
	ErrCodeFileUploadFailed        = "FILE_UPLOAD_FAILED"
	ErrCodeDirectoryCreationFailed = "DIRECTORY_CREATION_FAILED"
	ErrCodeValidationFailed        = "VALIDATION_FAILED"
	ErrCodeDatabaseError           = "DATABASE_ERROR"
	ErrCodePublishFailed           = "PUBLISH_FAILED"
	ErrCodeStateTransitionInvalid  = "INVALID_STATE_TRANSITION"
)

// PublicationNotFoundError returns an error for publication not found.
func PublicationNotFoundError(id string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodePublicationNotFound,
		Message: fmt.Sprintf("publication with id '%s' not found", id),
	}
}

// PlatformNotFoundError returns an error for platform not found.
func PlatformNotFoundError(id string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodePlatformNotFound,
		Message: fmt.Sprintf("platform with id '%s' not found", id),
	}
}

// PublicationPlatformExistsError returns an error for duplicate publication platform.
func PublicationPlatformExistsError(publicationID, platformID string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodePublicationPlatformDup,
		Message: fmt.Sprintf("publication '%s' is already published to platform '%s'", publicationID, platformID),
	}
}

// MediaNotFoundError returns an error for media not found.
func MediaNotFoundError(id string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeMediaNotFound,
		Message: fmt.Sprintf("media with id '%s' not found", id),
	}
}

// InvalidContentTypeError returns an error for invalid content type.
func InvalidContentTypeError(contentType string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeInvalidContentType,
		Message: fmt.Sprintf("'%s' is not a valid content type", contentType),
	}
}

// InvalidStatusError returns an error for invalid status.
func InvalidStatusError(status string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeInvalidStatus,
		Message: fmt.Sprintf("'%s' is not a valid publication status", status),
	}
}

// InvalidMediaTypeError returns an error for invalid media type.
func InvalidMediaTypeError(mediaType string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeInvalidMediaType,
		Message: fmt.Sprintf("'%s' is not a valid media type", mediaType),
	}
}

// UnauthorizedError returns an error for unauthorized access.
func UnauthorizedError(reason string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeUnauthorized,
		Message: reason,
	}
}

// FileUploadFailedError returns an error for file upload failure.
func FileUploadFailedError(fileName string, err error) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeFileUploadFailed,
		Message: fmt.Sprintf("failed to upload file '%s'", fileName),
		Err:     err,
	}
}

// DirectoryCreationFailedError returns an error for directory creation failure.
func DirectoryCreationFailedError(path string, err error) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeDirectoryCreationFailed,
		Message: fmt.Sprintf("failed to create directory '%s'", path),
		Err:     err,
	}
}

// ValidationFailedError returns an error for validation failure.
func ValidationFailedError(reason string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeValidationFailed,
		Message: reason,
	}
}

// DatabaseError returns an error for database operations.
func DatabaseError(reason string, err error) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeDatabaseError,
		Message: reason,
		Err:     err,
	}
}

// PublishFailedError returns an error for publish failure.
func PublishFailedError(reason string, err error) *PublicationError {
	return &PublicationError{
		Code:    ErrCodePublishFailed,
		Message: reason,
		Err:     err,
	}
}

// InvalidStateTransitionError returns an error for invalid status transition.
func InvalidStateTransitionError(from, to string) *PublicationError {
	return &PublicationError{
		Code:    ErrCodeStateTransitionInvalid,
		Message: fmt.Sprintf("cannot transition from '%s' to '%s'", from, to),
	}
}
