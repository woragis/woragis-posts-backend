package publications

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up publication routes.
// Note: Auth middleware is already applied at parent level in domains/routes.go
func SetupRoutes(router fiber.Router, handler Handler) {
	// Publication routes
	router.Post("/", handler.CreatePublication)
	router.Get("/", handler.ListPublications)
	router.Get("/:id", handler.GetPublication)
	router.Put("/:id", handler.UpdatePublication)
	router.Delete("/:id", handler.DeletePublication)

	// Platform routes
	router.Get("/platforms", handler.ListPlatforms)
	router.Post("/platforms", handler.CreatePlatform)

	// Publishing routes
	router.Post("/:publicationId/publish/:platformId", handler.PublishToplatform)
	router.Delete("/:publicationId/publish/:platformId", handler.UnpublishFromPlatform)
	router.Get("/:publicationId/publish", handler.ListPublicationPlatforms)
	router.Post("/:publicationId/publish/:platformId/retry", handler.RetryPublish)
	router.Post("/:publicationId/publish/bulk", handler.BulkPublish)

	// Media routes
	router.Post("/:publicationId/media", handler.UploadMedia)
	router.Get("/:publicationId/media", handler.ListPublicationMedia)
}
