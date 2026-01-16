package publications

import (
	"gorm.io/gorm"
)

// Migrate creates the publications tables.
func Migrate(db *gorm.DB) error {
	// Create platforms table
	if err := db.AutoMigrate(&Platform{}); err != nil {
		return err
	}

	// Create publications table
	if err := db.AutoMigrate(&Publication{}); err != nil {
		return err
	}

	// Create publication_platforms junction table
	if err := db.AutoMigrate(&PublicationPlatform{}); err != nil {
		return err
	}

	// Create publication_media table
	if err := db.AutoMigrate(&PublicationMedia{}); err != nil {
		return err
	}

	// Create indexes for common queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_publications_user_id_status 
		ON publications(user_id, status);
		
		CREATE INDEX IF NOT EXISTS idx_publications_user_id_archived
		ON publications(user_id, is_archived);
		
		CREATE INDEX IF NOT EXISTS idx_publications_content_id_type
		ON publications(content_id, content_type);
		
		CREATE INDEX IF NOT EXISTS idx_publication_platforms_published_at
		ON publication_platforms(published_at);
		
		CREATE INDEX IF NOT EXISTS idx_publication_media_platform_id
		ON publication_media(platform_id);
	`).Error; err != nil {
		// Indexes might already exist, so we don't fail
		_ = err
	}

	return nil
}

// SeedDefaultPlatforms creates the default platforms.
func SeedDefaultPlatforms(db *gorm.DB) error {
	platforms := []Platform{
		{
			Name:     "LinkedIn",
			Slug:     "linkedin",
			Color:    "#0A66C2",
			IsActive: true,
		},
		{
			Name:     "Twitter/X",
			Slug:     "twitter",
			Color:    "#000000",
			IsActive: true,
		},
		{
			Name:     "Instagram",
			Slug:     "instagram",
			Color:    "#E1306C",
			IsActive: true,
		},
		{
			Name:     "Newsletter",
			Slug:     "newsletter",
			Color:    "#6B46C1",
			IsActive: true,
		},
		{
			Name:     "Medium",
			Slug:     "medium",
			Color:    "#000000",
			IsActive: true,
		},
		{
			Name:     "Hashnode",
			Slug:     "hashnode",
			Color:    "#2962FF",
			IsActive: true,
		},
		{
			Name:     "Dev.to",
			Slug:     "devto",
			Color:    "#0A0A0A",
			IsActive: true,
		},
		{
			Name:     "Substack",
			Slug:     "substack",
			Color:    "#FF6719",
			IsActive: true,
		},
	}

	for i := range platforms {
		// Use firstOrCreate pattern
		var existing Platform
		db.Where("slug = ?", platforms[i].Slug).FirstOrCreate(&existing, platforms[i])
	}

	return nil
}
