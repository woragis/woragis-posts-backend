package publications

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PublicationStatus represents the status of a publication.
type PublicationStatus string

const (
	PublicationStatusSkeleton   PublicationStatus = "skeleton"   // Just title/outline
	PublicationStatusDraft      PublicationStatus = "draft"      // In progress
	PublicationStatusScheduled  PublicationStatus = "scheduled"  // Ready to post
	PublicationStatusPublished  PublicationStatus = "published"  // Live on platforms
	PublicationStatusArchived   PublicationStatus = "archived"   // No longer promoting
)

// ContentType represents the type of content being published.
type ContentType string

const (
	ContentTypePost              ContentType = "post"
	ContentTypeCaseStudy         ContentType = "case_study"
	ContentTypeProblemSolution   ContentType = "problem_solution"
	ContentTypeTechnicalWriting  ContentType = "technical_writing"
	ContentTypeSystemDesign      ContentType = "system_design"
	ContentTypeReport            ContentType = "report"
	ContentTypeImpactMetric      ContentType = "impact_metric"
	ContentTypeAIMLIntegration   ContentType = "aiml_integration"
)

// Publication represents a content publication plan.
// It can be a full content reference or just a skeleton/draft.
type Publication struct {
	ID           uuid.UUID           `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID           `gorm:"column:user_id;type:uuid;index;not null" json:"userId"`
	ContentID    *uuid.UUID          `gorm:"column:content_id;type:uuid;index" json:"contentId,omitempty"`
	ContentType  ContentType         `gorm:"column:content_type;type:varchar(32);index" json:"contentType"`
	Title        string              `gorm:"column:title;size:255;not null" json:"title"`
	Outline      string              `gorm:"column:outline;type:text" json:"outline,omitempty"`
	Status       PublicationStatus   `gorm:"column:status;type:varchar(32);not null;default:'skeleton';index" json:"status"`
	IsArchived   bool                `gorm:"column:is_archived;not null;default:false;index" json:"isArchived"`
	CreatedAt    time.Time           `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time           `gorm:"column:updated_at" json:"updatedAt"`

	// Relationships (loaded separately)
	Platforms []*PublicationPlatform `gorm:"foreignKey:PublicationID" json:"platforms,omitempty"`
	Media     []*PublicationMedia    `gorm:"foreignKey:PublicationID" json:"media,omitempty"`
}

// PublicationPlatform represents publishing to a specific platform.
type PublicationPlatform struct {
	ID              uuid.UUID                      `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	PublicationID   uuid.UUID                      `gorm:"column:publication_id;type:uuid;index;not null" json:"publicationId"`
	PlatformID      uuid.UUID                      `gorm:"column:platform_id;type:uuid;index;not null" json:"platformId"`
	PublishedAt     *time.Time                     `gorm:"column:published_at;index" json:"publishedAt,omitempty"`
	PublishedURL    string                         `gorm:"column:published_url;size:512" json:"publishedUrl,omitempty"`
	Status          PublicationPlatformStatus      `gorm:"column:status;type:varchar(32);not null;default:'scheduled';index" json:"status"`
	Metadata        PublicationPlatformMetadata    `gorm:"column:metadata;type:jsonb;serializer:json" json:"metadata,omitempty"`
	FailureReason   string                         `gorm:"column:failure_reason;size:512" json:"failureReason,omitempty"`
	RetryCount      int                            `gorm:"column:retry_count;not null;default:0" json:"retryCount"`
	LastRetryAt     *time.Time                     `gorm:"column:last_retry_at" json:"lastRetryAt,omitempty"`
	CreatedAt       time.Time                      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time                      `gorm:"column:updated_at" json:"updatedAt"`

	// Relationships
	Platform *Platform `gorm:"foreignKey:PlatformID" json:"platform,omitempty"`
}

// PublicationPlatformStatus represents the status on a specific platform.
type PublicationPlatformStatus string

const (
	PublicationPlatformStatusScheduled PublicationPlatformStatus = "scheduled"
	PublicationPlatformStatusPublished PublicationPlatformStatus = "published"
	PublicationPlatformStatusFailed    PublicationPlatformStatus = "failed"
	PublicationPlatformStatusArchived  PublicationPlatformStatus = "archived"
)

// PublicationPlatformMetadata stores optional metadata for the platform post.
type PublicationPlatformMetadata struct {
	PostID       string `json:"postId,omitempty"`        // Platform-specific post ID
	ScheduledFor *time.Time `json:"scheduledFor,omitempty"` // For scheduled posts
}

// Scan implements the Scanner interface for GORM.
func (m *PublicationPlatformMetadata) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	return json.Unmarshal(bytes, &m)
}

// Value implements the Valuer interface for GORM.
func (m PublicationPlatformMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// PublicationMedia represents media/evidence of a publication.
type PublicationMedia struct {
	ID              uuid.UUID     `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	PublicationID   uuid.UUID     `gorm:"column:publication_id;type:uuid;index;not null" json:"publicationId"`
	PlatformID      *uuid.UUID    `gorm:"column:platform_id;type:uuid;index" json:"platformId,omitempty"`
	MediaType       MediaType     `gorm:"column:media_type;type:varchar(32);not null" json:"mediaType"`
	FilePath        string        `gorm:"column:file_path;size:512;not null" json:"filePath"`
	FileSize        int64         `gorm:"column:file_size" json:"fileSize,omitempty"`
	MimeType        string        `gorm:"column:mime_type;size:128" json:"mimeType,omitempty"`
	UploadedAt      time.Time     `gorm:"column:uploaded_at" json:"uploadedAt"`
	CreatedAt       time.Time     `gorm:"column:created_at" json:"createdAt"`
}

// MediaType represents the type of media stored.
type MediaType string

const (
	MediaTypeScreenshot  MediaType = "screenshot"  // Screenshot of posted content
	MediaTypeArchive     MediaType = "archive"     // Archived HTML/PDF of content
	MediaTypeThumbnail   MediaType = "thumbnail"   // Thumbnail/preview image
	MediaTypeAttachment  MediaType = "attachment"  // Media attached to post
	MediaTypeMetadata    MediaType = "metadata"    // JSON metadata snapshot
)

// Platform represents a social media platform or distribution channel.
type Platform struct {
	ID          uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"column:name;size:128;not null;uniqueIndex" json:"name"`
	Slug        string    `gorm:"column:slug;size:128;not null;uniqueIndex" json:"slug"`
	Description string    `gorm:"column:description;type:text" json:"description,omitempty"`
	Icon        string    `gorm:"column:icon;size:255" json:"icon,omitempty"`
	Color       string    `gorm:"column:color;size:32" json:"color,omitempty"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true;index" json:"isActive"`
	APIEndpoint string    `gorm:"column:api_endpoint;size:512" json:"apiEndpoint,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// Default platforms
var (
	PlatformLinkedIn = Platform{
		Name:    "LinkedIn",
		Slug:    "linkedin",
		Color:   "#0A66C2",
		IsActive: true,
	}
	PlatformTwitter = Platform{
		Name:    "Twitter/X",
		Slug:    "twitter",
		Color:   "#000000",
		IsActive: true,
	}
	PlatformInstagram = Platform{
		Name:    "Instagram",
		Slug:    "instagram",
		Color:   "#E1306C",
		IsActive: true,
	}
	PlatformNewsletter = Platform{
		Name:    "Newsletter",
		Slug:    "newsletter",
		Color:   "#6B46C1",
		IsActive: true,
	}
)
