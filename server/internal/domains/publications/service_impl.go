package publications

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// ServiceImpl implements the Service interface.
type ServiceImpl struct {
	repo Repository
}

// NewService creates a new publication service.
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// CreatePublication creates a new publication.
func (s *ServiceImpl) CreatePublication(ctx context.Context, userID string, req *CreatePublicationRequest) (*Publication, error) {
	// Validate content type
	if !isValidContentType(string(req.ContentType)) {
		return nil, InvalidContentTypeError(string(req.ContentType))
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, ValidationFailedError("invalid user ID")
	}

	pub := &Publication{
		ID:          uuid.New(),
		UserID:      userUUID,
		ContentID:   req.ContentID,
		ContentType: req.ContentType,
		Title:       req.Title,
		Outline:     req.Outline,
		Status:      PublicationStatusSkeleton,
		IsArchived:  false,
		CreatedAt:     time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreatePublication(ctx, pub); err != nil {
		return nil, DatabaseError("failed to create publication", err)
	}

	return pub, nil
}

// GetPublication retrieves a publication.
func (s *ServiceImpl) GetPublication(ctx context.Context, userID, publicationID string) (*Publication, error) {
	pub, err := s.repo.GetPublication(ctx, publicationID)
	if err != nil {
		return nil, PublicationNotFoundError(publicationID)
	}

	// Verify ownership
	if pub.UserID.String() != userID {
		return nil, UnauthorizedError("you do not own this publication")
	}

	return pub, nil
}

// ListPublications lists publications for a user.
func (s *ServiceImpl) ListPublications(ctx context.Context, userID string, filter PublicationFilter) ([]*Publication, int64, error) {
	// Validate filter
	if filter.Status != "" && !isValidStatus(filter.Status) {
		return nil, 0, InvalidStatusError(filter.Status)
	}
	if filter.ContentType != "" && !isValidContentType(filter.ContentType) {
		return nil, 0, InvalidContentTypeError(filter.ContentType)
	}

	// Set defaults
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	publications, total, err := s.repo.ListPublications(ctx, userID, filter)
	if err != nil {
		return nil, 0, DatabaseError("failed to list publications", err)
	}

	return publications, total, nil
}

// UpdatePublication updates a publication.
func (s *ServiceImpl) UpdatePublication(ctx context.Context, userID, publicationID string, req *UpdatePublicationRequest) (*Publication, error) {
	pub, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil && *req.Title != "" {
		pub.Title = *req.Title
	}
	if req.Outline != nil && *req.Outline != "" {
		pub.Outline = *req.Outline
	}
	if req.Status != nil {
		statusStr := string(*req.Status)
		if !isValidStatus(statusStr) {
			return nil, InvalidStatusError(statusStr)
		}
		// Validate state transition
		if !isValidStateTransition(string(pub.Status), statusStr) {
			return nil, InvalidStateTransitionError(string(pub.Status), statusStr)
		}
		pub.Status = *req.Status
	}
	if req.IsArchived != nil {
		pub.IsArchived = *req.IsArchived
	}

	pub.UpdatedAt = time.Now()

	if err := s.repo.UpdatePublication(ctx, publicationID, pub); err != nil {
		return nil, DatabaseError("failed to update publication", err)
	}

	return pub, nil
}

// DeletePublication deletes a publication.
func (s *ServiceImpl) DeletePublication(ctx context.Context, userID, publicationID string) error {
	pub, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return err
	}

	// Clean up media files
	media, err := s.repo.ListPublicationMedia(ctx, pub.ID.String())
	if err == nil {
		for _, m := range media {
			_ = os.Remove(m.FilePath) // Best effort cleanup
		}
	}

	if err := s.repo.DeletePublication(ctx, publicationID); err != nil {
		return DatabaseError("failed to delete publication", err)
	}

	return nil
}

// ListPlatforms lists all active platforms.
func (s *ServiceImpl) ListPlatforms(ctx context.Context) ([]*Platform, error) {
	platforms, err := s.repo.ListPlatforms(ctx)
	if err != nil {
		return nil, DatabaseError("failed to list platforms", err)
	}
	return platforms, nil
}

// GetOrCreateDefaultPlatforms ensures default platforms exist.
func (s *ServiceImpl) GetOrCreateDefaultPlatforms(ctx context.Context) ([]*Platform, error) {
	// Try to list platforms first
	platforms, err := s.repo.ListPlatforms(ctx)
	if err == nil && len(platforms) > 0 {
		return platforms, nil
	}

	// Create default platforms
	defaults := []Platform{
		{
			ID:        uuid.New(),
			Name:      "LinkedIn",
			Slug:      "linkedin",
			Color:     "#0A66C2",
			IsActive:  true,
		CreatedAt:     time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Twitter/X",
			Slug:      "twitter",
			Color:     "#000000",
			IsActive:  true,
		CreatedAt:     time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Instagram",
			Slug:      "instagram",
			Color:     "#E1306C",
			IsActive:  true,
		CreatedAt:     time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Newsletter",
			Slug:      "newsletter",
			Color:     "#6B46C1",
			IsActive:  true,
		CreatedAt:     time.Now(),
		},
	}

	result := make([]*Platform, 0, len(defaults))
	for i := range defaults {
		if platform, err := s.repo.CreatePlatform(ctx, &defaults[i]); err == nil {
			result = append(result, platform)
		}
	}

	return result, nil
}

// CreatePlatform creates a new platform.
func (s *ServiceImpl) CreatePlatform(ctx context.Context, req *CreatePlatformRequest) (*Platform, error) {
	platform := &Platform{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		Color:     req.Color,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	platform, err := s.repo.CreatePlatform(ctx, platform)
	if err != nil {
		return nil, DatabaseError("failed to create platform", err)
	}

	return platform, nil
}

// PublishToplatform publishes to a platform.
func (s *ServiceImpl) PublishToplatform(ctx context.Context, userID, publicationID, platformID string, req *PublishRequest) (*PublicationPlatform, error) {
	// Verify publication ownership
	pub, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	// Verify platform exists
	platform, err := s.repo.GetPlatformByID(ctx, platformID)
	if err != nil {
		return nil, PlatformNotFoundError(platformID)
	}

	// Check if already published
	existing, _ := s.repo.GetPublicationPlatform(ctx, publicationID, platformID)
	if existing != nil && existing.Status != PublicationPlatformStatusFailed {
		return nil, PublicationPlatformExistsError(publicationID, platformID)
	}

	// Create publication platform entry
	pubPlatform := &PublicationPlatform{
		ID:            uuid.New(),
		PublicationID: pub.ID,
		PlatformID:    platform.ID,
		Status:        PublicationPlatformStatusScheduled,
		PublishedAt:   nil,
		PublishedURL:  req.PublishedURL,
		RetryCount:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if req.Metadata != nil {
		pubPlatform.Metadata = *req.Metadata
	}

	if err := s.repo.PublishToplatform(ctx, pubPlatform); err != nil {
		return nil, DatabaseError("failed to publish to platform", err)
	}

	// Update publication status if needed
	if pub.Status == PublicationStatusSkeleton || pub.Status == PublicationStatusDraft {
		pub.Status = PublicationStatusScheduled
		pub.UpdatedAt = time.Now()
		_ = s.repo.UpdatePublication(ctx, publicationID, pub)
	}

	return pubPlatform, nil
}

// UnpublishFromPlatform unpublishes from a platform.
func (s *ServiceImpl) UnpublishFromPlatform(ctx context.Context, userID, publicationID, platformID string) error {
	// Verify ownership
	_, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return err
	}

	if err := s.repo.UnpublishFromPlatform(ctx, publicationID, platformID); err != nil {
		return DatabaseError("failed to unpublish from platform", err)
	}

	return nil
}

// ListPublicationPlatforms lists platforms for a publication.
func (s *ServiceImpl) ListPublicationPlatforms(ctx context.Context, userID, publicationID string) ([]*PublicationPlatform, error) {
	// Verify ownership
	_, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	platforms, err := s.repo.ListPublicationPlatforms(ctx, publicationID)
	if err != nil {
		return nil, DatabaseError("failed to list publication platforms", err)
	}

	return platforms, nil
}

// RetryPublishToplatform retries publishing to a platform.
func (s *ServiceImpl) RetryPublishToplatform(ctx context.Context, userID, publicationID, platformID string) (*PublicationPlatform, error) {
	// Verify ownership
	_, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	pubPlatform, err := s.repo.GetPublicationPlatform(ctx, publicationID, platformID)
	if err != nil {
		return nil, PublicationNotFoundError(publicationID)
	}

	// Increment retry count
	pubPlatform.RetryCount++
	pubPlatform.Status = PublicationPlatformStatusScheduled
	pubPlatform.UpdatedAt = time.Now()

	if err := s.repo.UpdatePublicationPlatform(ctx, publicationID, platformID, pubPlatform); err != nil {
		return nil, DatabaseError("failed to retry publish", err)
	}

	return pubPlatform, nil
}

// BulkPublish publishes to multiple platforms.
func (s *ServiceImpl) BulkPublish(ctx context.Context, userID, publicationID string, req *BulkPublishRequest) ([]*PublicationPlatform, error) {
	// Verify ownership
	pub, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	result := make([]*PublicationPlatform, 0)

	for _, platformID := range req.PlatformIDs {
		publishReq := &PublishRequest{
			PublishedURL: req.URLs[platformID],
			Metadata:     nil,
		}

		pubPlatform, err := s.PublishToplatform(ctx, userID, publicationID, platformID, publishReq)
		if err != nil {
			continue // Skip failed platforms
		}
		result = append(result, pubPlatform)
	}

	// Update publication status
	if len(result) > 0 && (pub.Status == PublicationStatusSkeleton || pub.Status == PublicationStatusDraft) {
		pub.Status = PublicationStatusScheduled
		pub.UpdatedAt = time.Now()
		_ = s.repo.UpdatePublication(ctx, publicationID, pub)
	}

	return result, nil
}

// UploadMedia uploads media for a publication.
func (s *ServiceImpl) UploadMedia(ctx context.Context, userID, publicationID, platformID, mediaType string, file io.Reader, filename string) (*PublicationMedia, error) {
	// Verify ownership
	pub, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	// Validate media type
	if !isValidMediaType(mediaType) {
		return nil, InvalidMediaTypeError(mediaType)
	}

	// Create upload directory
	uploadDir := filepath.Join("uploads/publications", pub.ID.String())
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, DirectoryCreationFailedError(uploadDir, err)
	}

	// Save file
	filePath := filepath.Join(uploadDir, fmt.Sprintf("%s_%s", uuid.New(), filename))
	f, err := os.Create(filePath)
	if err != nil {
		return nil, FileUploadFailedError(filename, err)
	}
	defer f.Close()

	written, err := io.Copy(f, file)
	if err != nil {
		os.Remove(filePath)
		return nil, FileUploadFailedError(filename, err)
	}

	// Create media record
	media := &PublicationMedia{
		ID:            uuid.New(),
		PublicationID: pub.ID,
		MediaType:     MediaType(mediaType),
		FilePath:      filePath,
		FileSize:      written,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.UploadMedia(ctx, media); err != nil {
		os.Remove(filePath)
		return nil, DatabaseError("failed to save media record", err)
	}

	return media, nil
}

// ListPublicationMedia lists media for a publication.
func (s *ServiceImpl) ListPublicationMedia(ctx context.Context, userID, publicationID string) ([]*PublicationMedia, error) {
	// Verify ownership
	_, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	media, err := s.repo.ListPublicationMedia(ctx, publicationID)
	if err != nil {
		return nil, DatabaseError("failed to list media", err)
	}

	return media, nil
}

// GetPublicationMediaByPlatform gets media for a specific platform.
func (s *ServiceImpl) GetPublicationMediaByPlatform(ctx context.Context, userID, publicationID, platformID string) ([]*PublicationMedia, error) {
	// Verify ownership
	_, err := s.GetPublication(ctx, userID, publicationID)
	if err != nil {
		return nil, err
	}

	media, err := s.repo.GetPublicationMediaByPlatform(ctx, publicationID, platformID)
	if err != nil {
		return nil, DatabaseError("failed to get media", err)
	}

	return media, nil
}

// Helper functions

func isValidContentType(contentType string) bool {
	validTypes := map[string]bool{
		string(ContentTypePost):              true,
		string(ContentTypeCaseStudy):         true,
		string(ContentTypeProblemSolution):   true,
		string(ContentTypeTechnicalWriting):  true,
		string(ContentTypeSystemDesign):      true,
		string(ContentTypeReport):            true,
		string(ContentTypeImpactMetric):      true,
		string(ContentTypeAIMLIntegration):   true,
	}
	return validTypes[contentType]
}

func isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		string(PublicationStatusSkeleton):   true,
		string(PublicationStatusDraft):      true,
		string(PublicationStatusScheduled):  true,
		string(PublicationStatusPublished):  true,
		string(PublicationStatusArchived):   true,
	}
	return validStatuses[status]
}

func isValidMediaType(mediaType string) bool {
	validTypes := map[string]bool{
		string(MediaTypeScreenshot): true,
		string(MediaTypeArchive):    true,
		string(MediaTypeThumbnail):  true,
		string(MediaTypeAttachment): true,
		string(MediaTypeMetadata):   true,
	}
	return validTypes[mediaType]
}

func isValidStateTransition(from, to string) bool {
	// Define valid transitions
	validTransitions := map[string]map[string]bool{
		string(PublicationStatusSkeleton): {
			string(PublicationStatusDraft):     true,
			string(PublicationStatusScheduled): true,
			string(PublicationStatusArchived):  true,
		},
		string(PublicationStatusDraft): {
			string(PublicationStatusScheduled): true,
			string(PublicationStatusArchived):  true,
		},
		string(PublicationStatusScheduled): {
			string(PublicationStatusPublished): true,
			string(PublicationStatusArchived):  true,
		},
		string(PublicationStatusPublished): {
			string(PublicationStatusArchived): true,
		},
		string(PublicationStatusArchived): {
			string(PublicationStatusSkeleton):   true,
			string(PublicationStatusDraft):      true,
			string(PublicationStatusScheduled):  true,
		},
	}

	if transitions, ok := validTransitions[from]; ok {
		return transitions[to]
	}
	return false
}
