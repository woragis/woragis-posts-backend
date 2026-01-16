package publications

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository defines database operations for publications.
type Repository interface {
	// Publication operations
	CreatePublication(ctx context.Context, pub *Publication) error
	GetPublication(ctx context.Context, id string) (*Publication, error)
	ListPublications(ctx context.Context, userID string, filter PublicationFilter) ([]*Publication, int64, error)
	UpdatePublication(ctx context.Context, id string, updates *Publication) error
	DeletePublication(ctx context.Context, id string) error

	// Platform operations
	ListPlatforms(ctx context.Context) ([]*Platform, error)
	GetPlatformByID(ctx context.Context, id string) (*Platform, error)
	GetPlatformBySlug(ctx context.Context, slug string) (*Platform, error)
	CreatePlatform(ctx context.Context, platform *Platform) (*Platform, error)

	// PublicationPlatform operations
	PublishToplatform(ctx context.Context, pubPlatform *PublicationPlatform) error
	UnpublishFromPlatform(ctx context.Context, publicationID, platformID string) error
	ListPublicationPlatforms(ctx context.Context, publicationID string) ([]*PublicationPlatform, error)
	GetPublicationPlatform(ctx context.Context, publicationID, platformID string) (*PublicationPlatform, error)
	UpdatePublicationPlatform(ctx context.Context, publicationID, platformID string, updates *PublicationPlatform) error

	// Media operations
	UploadMedia(ctx context.Context, media *PublicationMedia) error
	ListPublicationMedia(ctx context.Context, publicationID string) ([]*PublicationMedia, error)
	GetPublicationMediaByPlatform(ctx context.Context, publicationID, platformID string) ([]*PublicationMedia, error)
	DeleteMedia(ctx context.Context, mediaID string) error
}

// GormRepository implements the Repository interface using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository.
func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// CreatePublication creates a new publication.
func (r *GormRepository) CreatePublication(ctx context.Context, pub *Publication) error {
	return r.db.WithContext(ctx).Create(pub).Error
}

// GetPublication retrieves a publication by ID.
func (r *GormRepository) GetPublication(ctx context.Context, id string) (*Publication, error) {
	var pub Publication
	if err := r.db.WithContext(ctx).
		Preload("PublicationPlatforms").
		Preload("PublicationMedia").
		First(&pub, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("publication not found")
		}
		return nil, err
	}
	return &pub, nil
}

// ListPublications retrieves publications with filters.
func (r *GormRepository) ListPublications(ctx context.Context, userID string, filter PublicationFilter) ([]*Publication, int64, error) {
	var publications []*Publication
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ContentType != "" {
		query = query.Where("content_type = ?", filter.ContentType)
	}
	if filter.IsArchived != nil {
		query = query.Where("is_archived = ?", *filter.IsArchived)
	}

	// Count total
	if err := query.Model(&Publication{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination and eager loading
	if err := query.
		Preload("PublicationPlatforms").
		Preload("PublicationMedia").
		Offset(int(filter.Offset)).
		Limit(int(filter.Limit)).
		Order("created_at DESC").
		Find(&publications).Error; err != nil {
		return nil, 0, err
	}

	return publications, total, nil
}

// UpdatePublication updates a publication.
func (r *GormRepository) UpdatePublication(ctx context.Context, id string, updates *Publication) error {
	return r.db.WithContext(ctx).Model(&Publication{}).Where("id = ?", id).Updates(updates).Error
}

// DeletePublication deletes a publication.
func (r *GormRepository) DeletePublication(ctx context.Context, id string) error {
	// Delete in correct order due to foreign keys
	if err := r.db.WithContext(ctx).Where("publication_id = ?", id).Delete(&PublicationMedia{}).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Where("publication_id = ?", id).Delete(&PublicationPlatform{}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&Publication{}, "id = ?", id).Error
}

// ListPlatforms retrieves all active platforms.
func (r *GormRepository) ListPlatforms(ctx context.Context) ([]*Platform, error) {
	var platforms []*Platform
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&platforms).Error; err != nil {
		return nil, err
	}
	return platforms, nil
}

// GetPlatformByID retrieves a platform by ID.
func (r *GormRepository) GetPlatformByID(ctx context.Context, id string) (*Platform, error) {
	var platform Platform
	if err := r.db.WithContext(ctx).First(&platform, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("platform not found")
		}
		return nil, err
	}
	return &platform, nil
}

// GetPlatformBySlug retrieves a platform by slug.
func (r *GormRepository) GetPlatformBySlug(ctx context.Context, slug string) (*Platform, error) {
	var platform Platform
	if err := r.db.WithContext(ctx).First(&platform, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("platform not found")
		}
		return nil, err
	}
	return &platform, nil
}

// CreatePlatform creates a new platform.
func (r *GormRepository) CreatePlatform(ctx context.Context, platform *Platform) (*Platform, error) {
	if err := r.db.WithContext(ctx).Create(platform).Error; err != nil {
		return nil, err
	}
	return platform, nil
}

// PublishToplatform creates a publication platform entry.
func (r *GormRepository) PublishToplatform(ctx context.Context, pubPlatform *PublicationPlatform) error {
	return r.db.WithContext(ctx).Create(pubPlatform).Error
}

// UnpublishFromPlatform removes a publication platform entry.
func (r *GormRepository) UnpublishFromPlatform(ctx context.Context, publicationID, platformID string) error {
	return r.db.WithContext(ctx).
		Where("publication_id = ? AND platform_id = ?", publicationID, platformID).
		Delete(&PublicationPlatform{}).Error
}

// ListPublicationPlatforms retrieves all platforms for a publication.
func (r *GormRepository) ListPublicationPlatforms(ctx context.Context, publicationID string) ([]*PublicationPlatform, error) {
	var pubPlatforms []*PublicationPlatform
	if err := r.db.WithContext(ctx).
		Where("publication_id = ?", publicationID).
		Preload("Publication").
		Order("created_at DESC").
		Find(&pubPlatforms).Error; err != nil {
		return nil, err
	}
	return pubPlatforms, nil
}

// GetPublicationPlatform retrieves a specific publication platform.
func (r *GormRepository) GetPublicationPlatform(ctx context.Context, publicationID, platformID string) (*PublicationPlatform, error) {
	var pubPlatform PublicationPlatform
	if err := r.db.WithContext(ctx).
		First(&pubPlatform, "publication_id = ? AND platform_id = ?", publicationID, platformID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("publication platform not found")
		}
		return nil, err
	}
	return &pubPlatform, nil
}

// UpdatePublicationPlatform updates a publication platform status.
func (r *GormRepository) UpdatePublicationPlatform(ctx context.Context, publicationID, platformID string, updates *PublicationPlatform) error {
	return r.db.WithContext(ctx).
		Model(&PublicationPlatform{}).
		Where("publication_id = ? AND platform_id = ?", publicationID, platformID).
		Updates(updates).Error
}

// UploadMedia creates a new media record.
func (r *GormRepository) UploadMedia(ctx context.Context, media *PublicationMedia) error {
	return r.db.WithContext(ctx).Create(media).Error
}

// ListPublicationMedia retrieves all media for a publication.
func (r *GormRepository) ListPublicationMedia(ctx context.Context, publicationID string) ([]*PublicationMedia, error) {
	var media []*PublicationMedia
	if err := r.db.WithContext(ctx).
		Where("publication_id = ?", publicationID).
		Order("created_at DESC").
		Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

// GetPublicationMediaByPlatform retrieves media for a specific platform.
func (r *GormRepository) GetPublicationMediaByPlatform(ctx context.Context, publicationID, platformID string) ([]*PublicationMedia, error) {
	var media []*PublicationMedia
	if err := r.db.WithContext(ctx).
		Where("publication_id = ? AND platform_id = ?", publicationID, platformID).
		Order("created_at DESC").
		Find(&media).Error; err != nil {
		return nil, err
	}
	return media, nil
}

// DeleteMedia deletes a media record.
func (r *GormRepository) DeleteMedia(ctx context.Context, mediaID string) error {
	return r.db.WithContext(ctx).Delete(&PublicationMedia{}, "id = ?", mediaID).Error
}
