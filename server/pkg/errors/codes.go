package errors

// Error code format: SERVICE_CATEGORY_NUMBER
// Each code should be used in exactly one place for easy tracking

const (
	// AUTH - Authentication errors (1000-1099)
	AUTH_JWT_INVALID_SIGNATURE    = "POSTS_1001"
	AUTH_JWT_EXPIRED              = "POSTS_1002"
	AUTH_JWT_MISSING_CLAIMS       = "POSTS_1003"
	AUTH_JWT_MALFORMED            = "POSTS_1004"
	AUTH_TOKEN_MISSING            = "POSTS_1020"
	AUTH_TOKEN_INVALID_FORMAT     = "POSTS_1021"
	AUTH_UNAUTHORIZED             = "POSTS_1022"
	AUTH_SERVICE_UNAVAILABLE      = "POSTS_1030"
	AUTH_SERVICE_ERROR            = "POSTS_1031"

	// CSRF - CSRF token errors (2000-2099)
	CSRF_TOKEN_EXPIRED            = "POSTS_2001"
	CSRF_TOKEN_INVALID            = "POSTS_2002"
	CSRF_TOKEN_MISSING            = "POSTS_2003"
	CSRF_TOKEN_GENERATION_FAILED  = "POSTS_2004"
	CSRF_TOKEN_MISMATCH           = "POSTS_2005"

	// DB - Database errors (3000-3099)
	DB_CONNECTION_FAILED          = "POSTS_3001"
	DB_QUERY_FAILED               = "POSTS_3002"
	DB_TRANSACTION_FAILED         = "POSTS_3003"
	DB_RECORD_NOT_FOUND           = "POSTS_3004"
	DB_DUPLICATE_ENTRY            = "POSTS_3005"
	DB_CONSTRAINT_VIOLATION       = "POSTS_3006"

	// VALIDATION - Input validation errors (4000-4099)
	VALIDATION_INVALID_INPUT      = "POSTS_4001"
	VALIDATION_MISSING_FIELD      = "POSTS_4002"
	VALIDATION_FIELD_TOO_LONG     = "POSTS_4003"
	VALIDATION_FIELD_TOO_SHORT    = "POSTS_4004"
	VALIDATION_INVALID_UUID       = "POSTS_4005"
	VALIDATION_INVALID_SLUG       = "POSTS_4006"
	VALIDATION_INVALID_STATUS     = "POSTS_4007"
	VALIDATION_INVALID_URL        = "POSTS_4008"

	// REDIS - Redis/Cache errors (5000-5099)
	REDIS_CONNECTION_FAILED       = "POSTS_5001"
	REDIS_GET_FAILED              = "POSTS_5002"
	REDIS_SET_FAILED              = "POSTS_5003"
	REDIS_DELETE_FAILED           = "POSTS_5004"

	// POST - Blog post specific errors (6000-6099)
	POST_NOT_FOUND                = "POSTS_6001"
	POST_SLUG_EXISTS              = "POSTS_6002"
	POST_ALREADY_PUBLISHED        = "POSTS_6003"
	POST_CANNOT_UNPUBLISH         = "POSTS_6004"
	POST_UPDATE_FAILED            = "POSTS_6005"
	POST_DELETE_FAILED            = "POSTS_6006"

	// TECHNICAL_WRITING - Technical writing errors (6100-6199)
	TECHNICAL_WRITING_NOT_FOUND   = "POSTS_6101"
	TECHNICAL_WRITING_EXISTS      = "POSTS_6102"
	TECHNICAL_WRITING_INVALID_TYPE = "POSTS_6103"
	TECHNICAL_WRITING_INVALID_PLATFORM = "POSTS_6104"

	// CASE_STUDY - Case study errors (6200-6299)
	CASE_STUDY_NOT_FOUND          = "POSTS_6201"
	CASE_STUDY_EXISTS             = "POSTS_6202"

	// COMMENT - Comment errors (6300-6399)
	COMMENT_NOT_FOUND             = "POSTS_6301"
	COMMENT_UPDATE_NOT_ALLOWED    = "POSTS_6302"
	COMMENT_DELETE_NOT_ALLOWED    = "POSTS_6303"

	// TAG - Tag errors (6400-6499)
	TAG_NOT_FOUND                 = "POSTS_6401"
	TAG_ALREADY_EXISTS            = "POSTS_6402"

	// SERVER - Server/System errors (9000-9099)
	SERVER_INTERNAL_ERROR         = "POSTS_9001"
	SERVER_SERVICE_UNAVAILABLE    = "POSTS_9002"
	SERVER_TIMEOUT                = "POSTS_9003"
	SERVER_CONTEXT_CANCELLED      = "POSTS_9004"
	SERVER_RATE_LIMIT_EXCEEDED    = "POSTS_9005"
)

// Error messages - human-readable descriptions
var errorMessages = map[string]string{
	// Authentication
	AUTH_JWT_INVALID_SIGNATURE: "JWT token signature is invalid",
	AUTH_JWT_EXPIRED:           "JWT token has expired",
	AUTH_JWT_MISSING_CLAIMS:    "JWT token is missing required claims",
	AUTH_JWT_MALFORMED:         "JWT token is malformed",
	AUTH_TOKEN_MISSING:         "Authentication token is missing",
	AUTH_TOKEN_INVALID_FORMAT:  "Authentication token has invalid format",
	AUTH_UNAUTHORIZED:          "Unauthorized access",
	AUTH_SERVICE_UNAVAILABLE:   "Authentication service is unavailable",
	AUTH_SERVICE_ERROR:         "Authentication service error",

	// CSRF
	CSRF_TOKEN_EXPIRED:           "CSRF token has expired",
	CSRF_TOKEN_INVALID:           "CSRF token validation failed",
	CSRF_TOKEN_MISSING:           "CSRF token is missing from request",
	CSRF_TOKEN_GENERATION_FAILED: "Failed to generate CSRF token",
	CSRF_TOKEN_MISMATCH:          "CSRF token does not match stored value",

	// Database
	DB_CONNECTION_FAILED:    "Failed to connect to database",
	DB_QUERY_FAILED:         "Database query execution failed",
	DB_TRANSACTION_FAILED:   "Database transaction failed",
	DB_RECORD_NOT_FOUND:     "Requested record not found",
	DB_DUPLICATE_ENTRY:      "Record already exists",
	DB_CONSTRAINT_VIOLATION: "Database constraint violation",

	// Validation
	VALIDATION_INVALID_INPUT:  "Input validation failed",
	VALIDATION_MISSING_FIELD:  "Required field is missing",
	VALIDATION_FIELD_TOO_LONG: "Field value exceeds maximum length",
	VALIDATION_FIELD_TOO_SHORT: "Field value is below minimum length",
	VALIDATION_INVALID_UUID:   "Invalid UUID format",
	VALIDATION_INVALID_SLUG:   "Invalid slug format",
	VALIDATION_INVALID_STATUS: "Invalid status value",
	VALIDATION_INVALID_URL:    "Invalid URL format",

	// Redis
	REDIS_CONNECTION_FAILED: "Failed to connect to Redis",
	REDIS_GET_FAILED:        "Failed to retrieve data from cache",
	REDIS_SET_FAILED:        "Failed to store data in cache",
	REDIS_DELETE_FAILED:     "Failed to delete data from cache",

	// Posts
	POST_NOT_FOUND:           "Blog post not found",
	POST_SLUG_EXISTS:         "Post with this slug already exists",
	POST_ALREADY_PUBLISHED:   "Post is already published",
	POST_CANNOT_UNPUBLISH:    "Cannot unpublish post",
	POST_UPDATE_FAILED:       "Failed to update post",
	POST_DELETE_FAILED:       "Failed to delete post",

	// Technical Writing
	TECHNICAL_WRITING_NOT_FOUND:      "Technical writing not found",
	TECHNICAL_WRITING_EXISTS:         "Technical writing already exists",
	TECHNICAL_WRITING_INVALID_TYPE:   "Invalid technical writing type",
	TECHNICAL_WRITING_INVALID_PLATFORM: "Invalid platform specified",

	// Case Study
	CASE_STUDY_NOT_FOUND: "Case study not found",
	CASE_STUDY_EXISTS:    "Case study already exists",

	// Comments
	COMMENT_NOT_FOUND:          "Comment not found",
	COMMENT_UPDATE_NOT_ALLOWED: "Not allowed to update this comment",
	COMMENT_DELETE_NOT_ALLOWED: "Not allowed to delete this comment",

	// Tags
	TAG_NOT_FOUND:      "Tag not found",
	TAG_ALREADY_EXISTS: "Tag already exists",

	// Server
	SERVER_INTERNAL_ERROR:      "Internal server error occurred",
	SERVER_SERVICE_UNAVAILABLE: "Service is temporarily unavailable",
	SERVER_TIMEOUT:             "Request timeout",
	SERVER_CONTEXT_CANCELLED:   "Request was cancelled",
	SERVER_RATE_LIMIT_EXCEEDED: "Rate limit exceeded",
}

// GetMessage returns the human-readable message for an error code
func GetMessage(code string) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Unknown error occurred"
}
