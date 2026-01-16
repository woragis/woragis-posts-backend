package posts

import (
	"gorm.io/gorm"

	"woragis-posts-service/internal/domains/posts"
	"woragis-posts-service/internal/domains/problemsolutions"
	"woragis-posts-service/internal/domains/impactmetrics"
	"woragis-posts-service/internal/domains/technicalwritings"
	"woragis-posts-service/internal/domains/casestudies"
	"woragis-posts-service/internal/domains/systemdesigns"
	"woragis-posts-service/internal/domains/reports"
	"woragis-posts-service/internal/domains/aimlintegrations"
	"woragis-posts-service/internal/domains/publications"
)

// MigratePostsTables runs database migrations for posts service
func MigratePostsTables(db *gorm.DB) error {
	// Enable UUID extension if not already enabled
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return err
	}

	// Enable gen_random_uuid function if not already available
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		return err
	}

	// Migrate posts tables
	if err := db.AutoMigrate(
		&posts.Post{},
	); err != nil {
		return err
	}

	// Migrate problem solutions tables
	if err := db.AutoMigrate(
		&problemsolutions.ProblemSolution{},
	); err != nil {
		return err
	}

	// Migrate impact metrics tables
	if err := db.AutoMigrate(
		&impactmetrics.ImpactMetric{},
	); err != nil {
		return err
	}

	// Migrate technical writings tables
	if err := db.AutoMigrate(
		&technicalwritings.TechnicalWriting{},
	); err != nil {
		return err
	}

	// Migrate case studies tables
	if err := db.AutoMigrate(
		&casestudies.CaseStudy{},
	); err != nil {
		return err
	}

	// Migrate system designs tables
	if err := db.AutoMigrate(
		&systemdesigns.SystemDesign{},
	); err != nil {
		return err
	}

	// Migrate reports tables
	if err := db.AutoMigrate(
		&reports.ReportDefinition{},
		&reports.ReportSchedule{},
		&reports.ReportDelivery{},
		&reports.ReportRun{},
	); err != nil {
		return err
	}

	// Migrate AIML integrations tables
	if err := db.AutoMigrate(
		&aimlintegrations.AIMLIntegration{},
	); err != nil {
		return err
	}

	// Migrate publications tables
	if err := publications.Migrate(db); err != nil {
		return err
	}

	// Seed default platforms
	if err := publications.SeedDefaultPlatforms(db); err != nil {
		return err
	}

	return nil
}
