# Publications Domain

The Publications domain provides a comprehensive publishing control and aggregation system for managing content across multiple platforms and distribution channels.

## Overview

Publications is a separate domain that acts as a **publishing control system** and **content archive**, allowing you to:

1. **Manage Publishing**: Track which content gets published where and when
2. **Multi-platform Distribution**: Publish the same content to multiple social media platforms
3. **State Management**: Support publication lifecycle (skeleton → draft → scheduled → published → archived)
4. **Media Management**: Store evidence of publications (screenshots, archives, metadata)
5. **Content Aggregation**: Link content from all 8 domain types in one place

## Domain Architecture

### Entities

#### Publication
The main publication entity that represents a publishing plan for any content type.

```go
type Publication struct {
    ID           uuid.UUID            // Unique identifier
    UserID       uuid.UUID            // Owner of the publication
    ContentID    *uuid.UUID           // Optional: Link to source content
    ContentType  ContentType          // Type of content (post, case_study, etc.)
    Title        string               // Publication title
    Outline      string               // Optional description/outline
    Status       PublicationStatus    // skeleton, draft, scheduled, published, archived
    IsArchived   bool                 // Soft-delete flag
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Status Transitions:**
- `skeleton` → `draft` → `scheduled` → `published` → `archived`
- `skeleton` → `draft` → `archived` (skip publishing)
- `archived` → any state (restore from archive)

#### PublicationPlatform
Junction entity linking publications to specific platforms with publishing metadata.

```go
type PublicationPlatform struct {
    ID           uuid.UUID    // Unique identifier
    PublicationID uuid.UUID   // FK to Publication
    PlatformID    uuid.UUID   // FK to Platform
    Status       PublicationPlatformStatus  // scheduled, published, failed, archived
    PublishedAt  *time.Time   // When it was published
    PublishedURL string       // URL to the published post
    Metadata     PublicationPlatformMetadata // Platform-specific data
    RetryCount   int          // Number of publish attempts
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### PublicationMedia
Stores media evidence of publications (screenshots, archives, etc.)

```go
type PublicationMedia struct {
    ID             uuid.UUID
    PublicationID  uuid.UUID
    PlatformID     *uuid.UUID  // Optional: Link to specific platform
    MediaType      MediaType   // screenshot, archive, thumbnail, attachment, metadata
    FilePath       string      // Local filesystem path
    FileSize       int64
    UploadedAt     time.Time
    CreatedAt      time.Time
}
```

#### Platform
Registry of distribution platforms/channels.

```go
type Platform struct {
    ID        uuid.UUID
    Name      string    // Display name (e.g., "LinkedIn")
    Slug      string    // Unique identifier (e.g., "linkedin")
    Color     string    // Brand color for UI
    IsActive  bool      // Can new publications be sent to this platform?
    CreatedAt time.Time
}
```

**Default Platforms:**
- LinkedIn
- Twitter/X
- Instagram
- Newsletter
- Medium
- Hashnode
- Dev.to
- Substack

## API Endpoints

All endpoints require JWT authentication.

### Publication Management

**Create Publication**
```
POST /api/v1/publications
Content-Type: application/json

{
    "contentId": "uuid",           // Optional
    "contentType": "post",         // Required: post, case_study, problem_solution, etc.
    "title": "Publication Title",  // Required
    "outline": "Optional outline"
}
```

**Get Publication**
```
GET /api/v1/publications/:id
```

**List Publications**
```
GET /api/v1/publications?limit=20&offset=0&status=scheduled&contentType=post&archived=false
```

**Update Publication**
```
PUT /api/v1/publications/:id
Content-Type: application/json

{
    "title": "Updated Title",
    "outline": "Updated outline",
    "status": "published",
    "isArchived": false
}
```

**Delete Publication**
```
DELETE /api/v1/publications/:id
```

### Platform Management

**List Platforms**
```
GET /api/v1/publications/platforms
```

**Create Platform**
```
POST /api/v1/publications/platforms
Content-Type: application/json

{
    "name": "Custom Platform",
    "slug": "custom",
    "description": "Optional description",
    "icon": "https://example.com/icon.png",
    "color": "#FF5733"
}
```

### Publishing Operations

**Publish to Single Platform**
```
POST /api/v1/publications/:publicationId/publish/:platformId
Content-Type: application/json

{
    "publishedUrl": "https://platform.com/post/123",
    "metadata": {
        "postId": "123",
        "scheduledFor": "2024-01-20T10:00:00Z"
    }
}
```

**Unpublish from Platform**
```
DELETE /api/v1/publications/:publicationId/publish/:platformId
```

**List Platforms for Publication**
```
GET /api/v1/publications/:publicationId/publish
```

**Retry Publish**
```
POST /api/v1/publications/:publicationId/publish/:platformId/retry
```

**Bulk Publish**
```
POST /api/v1/publications/:publicationId/publish/bulk
Content-Type: application/json

{
    "platformIds": ["uuid1", "uuid2", "uuid3"],
    "urls": {
        "uuid1": "https://linkedin.com/post/...",
        "uuid2": "https://twitter.com/post/...",
        "uuid3": "https://instagram.com/post/..."
    }
}
```

### Media Management

**Upload Media**
```
POST /api/v1/publications/:publicationId/media
Content-Type: multipart/form-data

- file: (binary)
- platformId: "uuid"          // Optional
- mediaType: "screenshot"     // Required: screenshot, archive, thumbnail, attachment, metadata
```

**List Media**
```
GET /api/v1/publications/:publicationId/media?platformId=uuid
```

## Content Types Supported

The Publications domain supports all 8 content types:

1. **post** - Blog posts
2. **case_study** - Case studies
3. **problem_solution** - Problem solutions
4. **technical_writing** - Technical articles
5. **system_design** - System design documents
6. **report** - Reports
7. **impact_metric** - Impact metrics
8. **aiml_integration** - AI/ML integrations

## Media Types Supported

1. **screenshot** - Screenshot of posted content
2. **archive** - Archived HTML/PDF version
3. **thumbnail** - Thumbnail/preview image
4. **attachment** - Media attached to post
5. **metadata** - JSON metadata snapshot

## Use Cases

### 1. Multi-Platform Publishing Workflow
```
User creates publication in "draft" state
    ↓
User adds content from Posts domain
    ↓
User publishes to LinkedIn (moves to "scheduled")
    ↓
System publishes to platform (moves to "published")
    ↓
User uploads screenshot as evidence
    ↓
Publication remains active for future republishing to other platforms
```

### 2. Content Archive as Drafts
```
User creates "skeleton" publication with just a title
    ↓
User later writes full outline (moves to "draft")
    ↓
User can link to content when ready
    ↓
User publishes or archives without publishing
```

### 3. Bulk Multi-platform Campaign
```
User creates publication for Case Study
    ↓
User publishes to 3 platforms simultaneously via bulk publish
    ↓
System tracks status on each platform independently
    ↓
User can unpublish from one platform while keeping others active
```

### 4. Publishing Evidence & Audit Trail
```
Each publication stores:
- URLs where content was posted
- Timestamps of publishing
- Screenshots of published posts
- Platform-specific metadata (post IDs, engagement)
- Retry history if publishing failed
```

## File Storage

Media files are stored locally in the filesystem:

```
uploads/publications/
├── {publicationId}/
│   ├── {uuid}_screenshot.png
│   ├── {uuid}_archive.pdf
│   └── {uuid}_metadata.json
└── {publicationId}/
    └── {uuid}_thumbnail.jpg
```

## State Machine

Publication status transitions are validated:

```
skeleton ──→ draft ──────→ scheduled ──→ published
  ↓           ↓              ↓
  └─────────→ archived ←─────┘
              ↑________________|
```

**Key Rules:**
- Can only move forward through states (mostly)
- Can archive from any state
- Can restore from archived back to skeleton/draft/scheduled
- Each platform has independent status tracking

## Database Schema

### Tables
- `publications` - Main publication records
- `publication_platforms` - Junction table for platform assignments
- `publication_media` - Media/evidence records
- `platforms` - Platform registry

### Indexes
- `idx_publications_user_id_status` - Quick user filtering
- `idx_publications_user_id_archived` - Archive querying
- `idx_publications_content_id_type` - Content linkage
- `idx_publication_platforms_published_at` - Timeline queries
- `idx_publication_media_platform_id` - Media retrieval

## Error Handling

Custom error types provide semantic error codes:

```
[PUBLICATION_NOT_FOUND]     - Publication doesn't exist
[PLATFORM_NOT_FOUND]        - Platform doesn't exist
[PUBLICATION_PLATFORM_EXISTS] - Already published to this platform
[INVALID_STATUS]            - Invalid status value
[INVALID_STATE_TRANSITION]  - Can't move to this state
[UNAUTHORIZED]              - User doesn't own this publication
[FILE_UPLOAD_FAILED]        - Media upload failed
[INVALID_CONTENT_TYPE]      - Invalid content type
```

## Service Interface

The Service layer handles business logic:

```go
type Service interface {
    // Publication CRUD
    CreatePublication(ctx, userID, req) (*Publication, error)
    GetPublication(ctx, userID, pubID) (*Publication, error)
    ListPublications(ctx, userID, filter) ([]*Publication, int64, error)
    UpdatePublication(ctx, userID, pubID, req) (*Publication, error)
    DeletePublication(ctx, userID, pubID) error
    
    // Platform operations
    ListPlatforms(ctx) ([]*Platform, error)
    GetOrCreateDefaultPlatforms(ctx) ([]*Platform, error)
    CreatePlatform(ctx, req) (*Platform, error)
    
    // Publishing
    PublishToplatform(ctx, userID, pubID, platformID, req) (*PublicationPlatform, error)
    UnpublishFromPlatform(ctx, userID, pubID, platformID) error
    ListPublicationPlatforms(ctx, userID, pubID) ([]*PublicationPlatform, error)
    RetryPublishToplatform(ctx, userID, pubID, platformID) (*PublicationPlatform, error)
    BulkPublish(ctx, userID, pubID, req) ([]*PublicationPlatform, error)
    
    // Media
    UploadMedia(ctx, userID, pubID, platformID, mediaType, file, filename) (*PublicationMedia, error)
    ListPublicationMedia(ctx, userID, pubID) ([]*PublicationMedia, error)
    GetPublicationMediaByPlatform(ctx, userID, pubID, platformID) ([]*PublicationMedia, error)
}
```

## Future Enhancements

1. **Scheduled Publishing** - Publish at specific times
2. **Social Media API Integration** - Auto-post to platforms
3. **Analytics** - Track engagement metrics per platform
4. **Content Templates** - Different content for different platforms
5. **Approval Workflow** - Multi-step publishing process
6. **Notifications** - Alert on publish success/failure
7. **Platform Webhooks** - Receive updates from platforms
8. **A/B Testing** - Test different versions on different platforms

## Related Domains

The Publications domain works with:
- **Posts** - Blog post content
- **Case Studies** - Project case study content
- **Problem Solutions** - Solution content
- **Technical Writings** - Technical documentation
- **System Designs** - Architecture documentation
- **Reports** - Business reports
- **Impact Metrics** - Metric dashboards
- **AIML Integrations** - AI/ML integration documentation

## Testing

Run tests:
```bash
go test ./internal/domains/publications/...
```

Run with coverage:
```bash
go test -cover ./internal/domains/publications/...
```
