# Publications Domain - Integration Guide

## Quick Start

The Publications domain is fully integrated into the Posts service backend.

### Database Setup

Migrations are automatically run on service startup. Default platforms are seeded:

```go
// In main.go or initialization code
if err := posts.MigratePostsTables(db); err != nil {
    return err
}
```

This creates:
- `publications` table
- `publication_platforms` table (junction)
- `publication_media` table
- `platforms` table (with 8 default platforms)

### API Access

The Publications domain is available at:

```
Base URL: /api/v1/publications
Port: 3013 (Posts service)
Auth: JWT Bearer token required for all endpoints
```

### Examples

#### Create a Publication (Draft)

```bash
curl -X POST http://localhost:3013/api/v1/publications \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contentType": "post",
    "title": "My Great Blog Post",
    "outline": "This post covers..."
  }'

# Response
{
  "success": true,
  "data": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "userId": "user-uuid",
    "contentId": null,
    "contentType": "post",
    "title": "My Great Blog Post",
    "outline": "This post covers...",
    "status": "skeleton",
    "isArchived": false,
    "createdAt": "2024-01-20T10:00:00Z",
    "updatedAt": "2024-01-20T10:00:00Z"
  }
}
```

#### Link Content and Publish

```bash
# Update publication with content reference
curl -X PUT http://localhost:3013/api/v1/publications/f47ac10b-58cc-4372-a567-0e02b2c3d479 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft"
  }'

# Publish to LinkedIn
curl -X POST http://localhost:3013/api/v1/publications/f47ac10b-58cc-4372-a567-0e02b2c3d479/publish/linkedin-uuid \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "publishedUrl": "https://linkedin.com/feed/update/urn:li:activity:...",
    "metadata": {
      "postId": "12345"
    }
  }'
```

#### Upload Proof of Publishing

```bash
curl -X POST http://localhost:3013/api/v1/publications/f47ac10b-58cc-4372-a567-0e02b2c3d479/media \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@screenshot.png" \
  -F "platformId=linkedin-uuid" \
  -F "mediaType=screenshot"
```

#### Bulk Publish

```bash
curl -X POST http://localhost:3013/api/v1/publications/f47ac10b-58cc-4372-a567-0e02b2c3d479/publish/bulk \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "platformIds": ["linkedin-uuid", "twitter-uuid", "newsletter-uuid"],
    "urls": {
      "linkedin-uuid": "https://linkedin.com/...",
      "twitter-uuid": "https://twitter.com/...",
      "newsletter-uuid": "https://substack.com/..."
    }
  }'
```

## Frontend Integration

The Publications domain is exposed via REST API and can be integrated into:

### SvelteKit Frontend (posts/frontend)

Add API client:

```typescript
// src/lib/api/publications/client.ts
import { fetch } from '$app/utils';

export const PublicationsAPI = {
  async create(req: CreatePublicationRequest) {
    return fetch('/api/v1/publications', {
      method: 'POST',
      body: JSON.stringify(req)
    });
  },

  async list(filters: PublicationFilter) {
    const params = new URLSearchParams({
      limit: filters.limit?.toString() || '20',
      offset: filters.offset?.toString() || '0',
      ...(filters.status && { status: filters.status }),
      ...(filters.contentType && { contentType: filters.contentType }),
    });
    return fetch(`/api/v1/publications?${params}`);
  },

  async publish(publicationId: string, platformId: string, req: PublishRequest) {
    return fetch(`/api/v1/publications/${publicationId}/publish/${platformId}`, {
      method: 'POST',
      body: JSON.stringify(req)
    });
  },

  async uploadMedia(publicationId: string, formData: FormData) {
    return fetch(`/api/v1/publications/${publicationId}/media`, {
      method: 'POST',
      body: formData
    });
  }
};
```

Create pages:

```
src/routes/publications/
├── +page.svelte          # List publications
├── [id]/
│   ├── +page.svelte      # View publication
│   ├── edit/+page.svelte # Edit publication
│   └── publish/+page.svelte # Publish UI
└── new/+page.svelte      # Create publication
```

## Domain Integration Points

### With Posts Domain

Publications can link to Posts content:

```typescript
// Link a publication to a specific post
{
  "contentType": "post",
  "contentId": "post-uuid",  // Reference to actual Post
  "title": "..."
}
```

### With Case Studies Domain

```typescript
{
  "contentType": "case_study",
  "contentId": "case-study-uuid",
  "title": "..."
}
```

### With All Other Domains

Same pattern for:
- `problem_solution`
- `technical_writing`
- `system_design`
- `report`
- `impact_metric`
- `aiml_integration`

## Development Workflow

### 1. Create Publication (Skeleton)

User creates a bare publication with just title and outline as a brain-dump/note.

### 2. Add Content

User either:
- Links to existing content via `contentId`
- Writes outline for future content

### 3. Draft State

User moves publication to "draft" when content is ready for publishing.

### 4. Multi-platform Publishing

User publishes to multiple platforms:
- LinkedIn for professional network
- Twitter for quick share
- Newsletter for subscribers
- Blog link in each

### 5. Track Evidence

User uploads screenshots/archives as proof of publication.

### 6. Archive

When done promoting, user archives publication to clean up active list.

## Error Scenarios

### Unauthorized Access
```json
{
  "success": false,
  "code": 401,
  "data": {
    "message": "authentication required"
  }
}
```

### Publication Not Found
```json
{
  "success": false,
  "code": 500,
  "data": {
    "message": "publication with id 'xxx' not found"
  }
}
```

### Invalid State Transition
```json
{
  "success": false,
  "code": 500,
  "data": {
    "message": "[INVALID_STATE_TRANSITION] cannot transition from 'published' to 'draft'"
  }
}
```

### Already Published to Platform
```json
{
  "success": false,
  "code": 500,
  "data": {
    "message": "[PUBLICATION_PLATFORM_EXISTS] publication 'xxx' is already published to platform 'yyy'"
  }
}
```

## Performance Considerations

### Queries

List queries are indexed:
- By `user_id` + `status` for quick filtering
- By `content_id` + `content_type` for content linking
- By `published_at` for timeline views

### Pagination

Default limit: 20
Maximum limit: 100

```
GET /api/v1/publications?limit=50&offset=100
```

### Media Storage

- Local filesystem (not S3)
- Files stored in: `uploads/publications/{publicationId}/`
- Each file prefixed with UUID for uniqueness
- Automatic cleanup on publication deletion

## Troubleshooting

### Publications not appearing

1. Check user_id matches authenticated user
2. Verify database migrations ran
3. Check publication status filters

### File upload fails

1. Ensure `uploads/` directory exists and is writable
2. Check file size limits
3. Verify multipart form data formatting

### Platform not found

1. Verify platform exists: `GET /api/v1/publications/platforms`
2. Use correct platform UUID (not slug)
3. Check platform is active (`isActive: true`)

## Next Steps

1. **Frontend Implementation** - Build Publications UI in SvelteKit
2. **Social Media Integration** - Add automatic posting via platform APIs
3. **Scheduling** - Support scheduled publishing times
4. **Analytics** - Track engagement metrics per platform
5. **Webhooks** - Receive platform updates (likes, comments, shares)
6. **Approval Workflow** - Add multi-step review process

## Support

For issues or questions:
1. Check the Publications README.md for detailed API documentation
2. Review error codes in errors.go
3. Check existing tests in publications_test.go (when added)
4. Review service implementation logic in service_impl.go
