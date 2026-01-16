package publications

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"woragis-posts-service/pkg/middleware"
	"woragis-posts-service/pkg/response"
)

// Handler exposes publication endpoints.
type Handler interface {
	// Publication handlers
	CreatePublication(c *fiber.Ctx) error
	GetPublication(c *fiber.Ctx) error
	ListPublications(c *fiber.Ctx) error
	UpdatePublication(c *fiber.Ctx) error
	DeletePublication(c *fiber.Ctx) error

	// Platform handlers
	ListPlatforms(c *fiber.Ctx) error
	CreatePlatform(c *fiber.Ctx) error

	// Publishing handlers
	PublishToplatform(c *fiber.Ctx) error
	UnpublishFromPlatform(c *fiber.Ctx) error
	ListPublicationPlatforms(c *fiber.Ctx) error
	RetryPublish(c *fiber.Ctx) error
	BulkPublish(c *fiber.Ctx) error

	// Media handlers
	UploadMedia(c *fiber.Ctx) error
	ListPublicationMedia(c *fiber.Ctx) error
}

// NewHandler creates a new publication handler.
func NewHandler(service Service, logger *slog.Logger) Handler {
	return &handler{
		logger:  logger,
		service: service,
	}
}

type handler struct {
	logger  *slog.Logger
	service Service
}

// CreatePublication creates a new publication.
func (h *handler) CreatePublication(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	var req CreatePublicationRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", "error", err)
		return response.Error(c, fiber.StatusBadRequest, fiber.StatusBadRequest, fiber.Map{
			"message": "Invalid request body",
		})
	}

	publication, err := h.service.CreatePublication(c.Context(), userID.String(), &req)
	if err != nil {
		h.logger.Error("failed to create publication", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to create publication",
		})
	}

	return response.Success(c, fiber.StatusCreated, publication)
}

// GetPublication retrieves a publication by ID.
func (h *handler) GetPublication(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")

	publication, err := h.service.GetPublication(c.Context(), userID.String(), publicationID)
	if err != nil {
		h.logger.Error("failed to get publication", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to retrieve publication",
		})
	}

	return response.Success(c, fiber.StatusOK, publication)
}

// ListPublications lists all publications for the user.
func (h *handler) ListPublications(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	filter := PublicationFilter{
		Limit:  limit,
		Offset: offset,
	}

	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	if contentType := c.Query("contentType"); contentType != "" {
		filter.ContentType = contentType
	}

	if archived := c.Query("archived"); archived == "true" {
		t := true
		filter.IsArchived = &t
	}

	publications, total, err := h.service.ListPublications(c.Context(), userID.String(), filter)
	if err != nil {
		h.logger.Error("failed to list publications", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to list publications",
		})
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"publications": publications,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

// UpdatePublication updates a publication.
func (h *handler) UpdatePublication(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")

	var req UpdatePublicationRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", "error", err)
		return response.Error(c, fiber.StatusBadRequest, fiber.StatusBadRequest, fiber.Map{
			"message": "Invalid request body",
		})
	}

	publication, err := h.service.UpdatePublication(c.Context(), userID.String(), publicationID, &req)
	if err != nil {
		h.logger.Error("failed to update publication", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to update publication",
		})
	}

	return response.Success(c, fiber.StatusOK, publication)
}

// DeletePublication deletes a publication.
func (h *handler) DeletePublication(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")

	if err := h.service.DeletePublication(c.Context(), userID.String(), publicationID); err != nil {
		h.logger.Error("failed to delete publication", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to delete publication",
		})
	}

	return response.Success(c, fiber.StatusNoContent, nil)
}

// ListPlatforms lists all active platforms.
func (h *handler) ListPlatforms(c *fiber.Ctx) error {
	platforms, err := h.service.ListPlatforms(c.Context())
	if err != nil {
		h.logger.Error("failed to list platforms", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to list platforms",
		})
	}

	return response.Success(c, fiber.StatusOK, platforms)
}

// CreatePlatform creates a new platform.
func (h *handler) CreatePlatform(c *fiber.Ctx) error {
	var req CreatePlatformRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", "error", err)
		return response.Error(c, fiber.StatusBadRequest, fiber.StatusBadRequest, fiber.Map{
			"message": "Invalid request body",
		})
	}

	platform, err := h.service.CreatePlatform(c.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create platform", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to create platform",
		})
	}

	return response.Success(c, fiber.StatusCreated, platform)
}

// PublishToplatform publishes to a platform.
func (h *handler) PublishToplatform(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")
	platformID := c.Params("platformId")

	var req PublishRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", "error", err)
		return response.Error(c, fiber.StatusBadRequest, fiber.StatusBadRequest, fiber.Map{
			"message": "Invalid request body",
		})
	}

	pubPlatform, err := h.service.PublishToplatform(c.Context(), userID.String(), publicationID, platformID, &req)
	if err != nil {
		h.logger.Error("failed to publish to platform", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to publish to platform",
		})
	}

	return response.Success(c, fiber.StatusCreated, pubPlatform)
}

// UnpublishFromPlatform unpublishes from a platform.
func (h *handler) UnpublishFromPlatform(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")
	platformID := c.Params("platformId")

	if err := h.service.UnpublishFromPlatform(c.Context(), userID.String(), publicationID, platformID); err != nil {
		h.logger.Error("failed to unpublish from platform", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to unpublish from platform",
		})
	}

	return response.Success(c, fiber.StatusNoContent, nil)
}

// ListPublicationPlatforms lists platforms for a publication.
func (h *handler) ListPublicationPlatforms(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")

	platforms, err := h.service.ListPublicationPlatforms(c.Context(), userID.String(), publicationID)
	if err != nil {
		h.logger.Error("failed to list publication platforms", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to list publication platforms",
		})
	}

	return response.Success(c, fiber.StatusOK, platforms)
}

// RetryPublish retries publishing to a platform.
func (h *handler) RetryPublish(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")
	platformID := c.Params("platformId")

	pubPlatform, err := h.service.RetryPublishToplatform(c.Context(), userID.String(), publicationID, platformID)
	if err != nil {
		h.logger.Error("failed to retry publish", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to retry publish",
		})
	}

	return response.Success(c, fiber.StatusOK, pubPlatform)
}

// BulkPublish publishes to multiple platforms.
func (h *handler) BulkPublish(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")

	var req BulkPublishRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", "error", err)
		return response.Error(c, fiber.StatusBadRequest, fiber.StatusBadRequest, fiber.Map{
			"message": "Invalid request body",
		})
	}

	platforms, err := h.service.BulkPublish(c.Context(), userID.String(), publicationID, &req)
	if err != nil {
		h.logger.Error("failed to bulk publish", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to bulk publish",
		})
	}

	return response.Success(c, fiber.StatusCreated, platforms)
}

// UploadMedia uploads media for a publication.
func (h *handler) UploadMedia(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")
	platformID := c.FormValue("platformId")
	mediaType := c.FormValue("mediaType")

	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("failed to get file", "error", err)
		return response.Error(c, fiber.StatusBadRequest, fiber.StatusBadRequest, fiber.Map{
			"message": "No file provided",
		})
	}

	fileReader, err := file.Open()
	if err != nil {
		h.logger.Error("failed to open file", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to read file",
		})
	}
	defer fileReader.Close()

	media, err := h.service.UploadMedia(c.Context(), userID.String(), publicationID, platformID, mediaType, fileReader, file.Filename)
	if err != nil {
		h.logger.Error("failed to upload media", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": fmt.Sprintf("Failed to upload media: %v", err),
		})
	}

	return response.Success(c, fiber.StatusCreated, media)
}

// ListPublicationMedia lists media for a publication.
func (h *handler) ListPublicationMedia(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromFiberContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, fiber.StatusUnauthorized, fiber.Map{
			"message": "authentication required",
		})
	}

	publicationID := c.Params("id")
	platformID := c.Query("platformId")

	var media []*PublicationMedia
	if platformID != "" {
		media, err = h.service.GetPublicationMediaByPlatform(c.Context(), userID.String(), publicationID, platformID)
	} else {
		media, err = h.service.ListPublicationMedia(c.Context(), userID.String(), publicationID)
	}

	if err != nil {
		h.logger.Error("failed to list media", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, fiber.StatusInternalServerError, fiber.Map{
			"message": "Failed to list media",
		})
	}

	return response.Success(c, fiber.StatusOK, media)
}
