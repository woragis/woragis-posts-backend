package publications

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// Service defines publication operations.
type Service interface {
	// Publication operations
	CreatePublication(ctx context.Context, userID string, req *CreatePublicationRequest) (*Publication, error)
	GetPublication(ctx context.Context, userID, publicationID string) (*Publication, error)
	ListPublications(ctx context.Context, userID string, filter PublicationFilter) ([]*Publication, int64, error)
	UpdatePublication(ctx context.Context, userID, publicationID string, req *UpdatePublicationRequest) (*Publication, error)
	DeletePublication(ctx context.Context, userID, publicationID string) error

	// Platform operations
	ListPlatforms(ctx context.Context) ([]*Platform, error)
	GetOrCreateDefaultPlatforms(ctx context.Context) ([]*Platform, error)
	CreatePlatform(ctx context.Context, req *CreatePlatformRequest) (*Platform, error)

	// Publication-Platform operations
	PublishToplatform(ctx context.Context, userID, publicationID, platformID string, req *PublishRequest) (*PublicationPlatform, error)
	UnpublishFromPlatform(ctx context.Context, userID, publicationID, platformID string) error
	ListPublicationPlatforms(ctx context.Context, userID, publicationID string) ([]*PublicationPlatform, error)
	RetryPublishToplatform(ctx context.Context, userID, publicationID, platformID string) (*PublicationPlatform, error)
	BulkPublish(ctx context.Context, userID, publicationID string, req *BulkPublishRequest) ([]*PublicationPlatform, error)

	// Media operations
	UploadMedia(ctx context.Context, userID, publicationID, platformID, mediaType string, file io.Reader, filename string) (*PublicationMedia, error)
	ListPublicationMedia(ctx context.Context, userID, publicationID string) ([]*PublicationMedia, error)
	GetPublicationMediaByPlatform(ctx context.Context, userID, publicationID, platformID string) ([]*PublicationMedia, error)
}

// PublicationFilter represents filters for listing publications.
type PublicationFilter struct {
	Status      string
	ContentType string
	IsArchived   *bool
	Limit        int
	Offset       int
}

// CreatePublicationRequest is the request to create a publication.
type CreatePublicationRequest struct {
	ContentID   *uuid.UUID  `json:"contentId"`
	ContentType ContentType `json:"contentType"`
	Title       string      `json:"title"`
	Outline     string      `json:"outline"`
}

// UpdatePublicationRequest is the request to update a publication.
type UpdatePublicationRequest struct {
	Title       *string              `json:"title"`
	Outline     *string              `json:"outline"`
	Status      *PublicationStatus   `json:"status"`
	IsArchived  *bool                `json:"isArchived"`
	PlatformIDs *[]uuid.UUID         `json:"platformIds"` // For bulk platform assignment
}

// CreatePlatformRequest is the request to create a platform.
type CreatePlatformRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// PublishRequest is the request to publish to a platform.
type PublishRequest struct {
	PublishedURL string                          `json:"publishedUrl"`
	Metadata     *PublicationPlatformMetadata    `json:"metadata,omitempty"`
}

// BulkPublishRequest is the request to publish to multiple platforms.
type BulkPublishRequest struct {
	PlatformIDs []string          `json:"platformIds"`
	URLs        map[string]string `json:"urls"` // platformId -> URL
}
