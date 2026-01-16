package posts

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"woragis-posts-service/internal/domains/aimlintegrations"
	"woragis-posts-service/internal/domains/casestudies"
	"woragis-posts-service/internal/domains/impactmetrics"
	"woragis-posts-service/internal/domains/posts"
	postcomments "woragis-posts-service/internal/domains/posts/comments"
	"woragis-posts-service/internal/domains/problemsolutions"
	"woragis-posts-service/internal/domains/publications"
	"woragis-posts-service/internal/domains/reports"
	"woragis-posts-service/internal/domains/systemdesigns"
	"woragis-posts-service/internal/domains/technicalwritings"
	"woragis-posts-service/pkg/authservice"
	"woragis-posts-service/pkg/middleware"
)

// SetupRoutes sets up all posts service routes
func SetupRoutes(api fiber.Router, db *gorm.DB, authServiceURL string, logger *slog.Logger) {
	// Initialize Auth Service client
	authClient := authservice.NewClient(authServiceURL)

	// Apply auth validation middleware to all routes
	api.Use(middleware.AuthValidationMiddleware(middleware.DefaultAuthValidationConfig(authClient)))

	// Initialize repositories
	postRepo := posts.NewGormRepository(db)
	problemSolutionRepo := problemsolutions.NewGormRepository(db)
	impactMetricRepo := impactmetrics.NewGormRepository(db)
	technicalWritingRepo := technicalwritings.NewGormRepository(db)
	caseStudyRepo := casestudies.NewGormRepository(db)
	systemDesignRepo := systemdesigns.NewGormRepository(db)
	reportRepo := reports.NewGormRepository(db)
	aimlIntegrationRepo := aimlintegrations.NewGormRepository(db)
	publicationRepo := publications.NewGormRepository(db)

	// Initialize services
	postService := posts.NewService(postRepo, logger)
	problemSolutionService := problemsolutions.NewService(problemSolutionRepo) // No logger parameter
	impactMetricService := impactmetrics.NewService(impactMetricRepo, logger)
	technicalWritingService := technicalwritings.NewService(technicalWritingRepo, logger)
	caseStudyService := casestudies.NewService(caseStudyRepo, logger)
	systemDesignService := systemdesigns.NewService(systemDesignRepo) // No logger parameter
	// TODO: Initialize repositories when services are implemented
	var ideasRepo reports.IdeaRepository = nil
	var projectsRepo reports.ProjectRepository = nil
	var financeRepo reports.FinanceRepository = nil
	var chatsRepo reports.ChatRepository = nil
	var publisher reports.Publisher = nil
	reportService := reports.NewService(reportRepo, ideasRepo, projectsRepo, financeRepo, chatsRepo, publisher, logger)
	aimlIntegrationService := aimlintegrations.NewService(aimlIntegrationRepo, logger)
	publicationService := publications.NewService(publicationRepo)

	// Initialize handlers (simplified - without translation enricher for now)
	postHandler := posts.NewHandler(postService, nil, nil, nil, logger) // enricher, translationService, creativeAssetsService
	problemSolutionHandler := problemsolutions.NewHandler(problemSolutionService, nil, nil, logger) // enricher, translationService
	impactMetricHandler := impactmetrics.NewHandler(impactMetricService, nil, nil, logger) // enricher, translationService
	technicalWritingHandler := technicalwritings.NewHandler(technicalWritingService, nil, nil, logger) // enricher, translationService
	caseStudyHandler := casestudies.NewHandler(caseStudyService, nil, nil, logger) // enricher, translationService
	systemDesignHandler := systemdesigns.NewHandler(systemDesignService, nil, nil, logger) // enricher, translationService
	reportHandler := reports.NewHandler(reportService, logger)
	aimlIntegrationHandler := aimlintegrations.NewHandler(aimlIntegrationService, nil, nil, logger) // enricher, translationService
	publicationHandler := publications.NewHandler(publicationService, logger)

	// Initialize subdomain handlers for posts
	commentRepo := postcomments.NewGormRepository(db)
	commentService := postcomments.NewService(commentRepo, logger)
	commentHandler := postcomments.NewHandler(commentService, logger)

	// Setup routes
	postsGroup := api.Group("/posts")
	posts.SetupRoutes(postsGroup, postHandler)
	postcomments.SetupRoutes(postsGroup.Group("/:postId/comments"), commentHandler)
	problemsolutions.SetupRoutes(api.Group("/problem-solutions"), problemSolutionHandler)
	impactmetrics.SetupRoutes(api.Group("/impact-metrics"), impactMetricHandler)
	technicalwritings.SetupRoutes(api.Group("/technical-writings"), technicalWritingHandler)
	casestudies.SetupRoutes(api.Group("/case-studies"), caseStudyHandler)
	systemdesigns.SetupRoutes(api.Group("/system-designs"), systemDesignHandler)
	reports.SetupRoutes(api.Group("/reports"), reportHandler)
	aimlintegrations.SetupRoutes(api.Group("/aiml-integrations"), aimlIntegrationHandler)
	publications.SetupRoutes(api.Group("/publications"), publicationHandler)
}
