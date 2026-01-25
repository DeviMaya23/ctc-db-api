# Cacheability Implementation

This API implements RESTful cacheability using ETags and Cache-Control headers.

## Overview

- ETag generation logic lives in the domain layer
- HTTP caching headers are managed in the handler layer
- Cache validation is timestamp-based using `UpdatedAt` field

## Features Implemented

### 1. **ETag Support**
- Automatic ETag generation based on resource `UpdatedAt` timestamp
- `If-None-Match` header support for GET requests (returns 304 Not Modified)
- `If-Match` header support for PUT requests (optimistic locking)

### 2. **Cache-Control Headers**
- Individual resources: `Cache-Control: public, max-age=600` (10 minutes)
- List endpoints: `Cache-Control: public, max-age=300` (5 minute)

### 3. **Last-Modified Headers**
- Included in GET responses for additional cache validation

### 4. **Optimistic Locking**
- Prevents lost updates using `If-Match` header on PUT requests
- Returns `412 Precondition Failed` if resource was modified

## API Usage Examples

### GET Request with Cache Validation

```bash
# First request - Get resource with ETag
curl -i http://localhost:8080/api/v1/travellers/1

HTTP/1.1 200 OK
ETag: "1737820800"
Cache-Control: public, max-age=600
Last-Modified: Sat, 25 Jan 2026 12:00:00 GMT
Content-Type: application/json

{
  "message": "success",
  "data": {
    "name": "Fiore",
    "rarity": 5,
    ...
  }
}

# Second request - Use If-None-Match (resource unchanged)
curl -i -H 'If-None-Match: "1737820800"' http://localhost:8080/api/v1/travellers/1

HTTP/1.1 304 Not Modified
ETag: "1737820800"
Cache-Control: private, max-age=300
(no body - saves bandwidth!)
```

### PUT Request with Optimistic Locking

```bash
# Update with correct ETag - Success
curl -i -X PUT \
  -H 'If-Match: "1737820800"' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Updated Name","rarity":5,...}' \
  http://localhost:8080/api/v1/travellers/1

HTTP/1.1 200 OK
ETag: "1737824400"
...

# Update with stale ETag - Conflict
curl -i -X PUT \
  -H 'If-Match: "1737820800"' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Another Update",...}' \
  http://localhost:8080/api/v1/travellers/1

HTTP/1.1 412 Precondition Failed
{
  "error": "Resource has been modified by another request. Please refresh and try again."
}
```

## Benefits

✅ **Bandwidth Savings**: 304 responses have no body, saving network traffic  
✅ **Reduced Server Load**: Clients use cached data when valid  
✅ **Prevent Lost Updates**: Optimistic locking detects concurrent modifications  
✅ **Better Performance**: Faster responses when cache is valid  
✅ **REST Compliant**: Follows HTTP caching standards (RFC 7232)

## Implementation Details

### Domain Layer
```go
// pkg/domain/traveller.go
func (t TravellerResponse) ETag() string {
    return fmt.Sprintf(`"%d"`, t.updatedAt.Unix())
}

func (t TravellerResponse) LastModified() time.Time {
    return t.updatedAt
}
```

### Handler Layer
```go
// internal/rest/traveller_handler.go
response := domain.ToTravellerResponse(traveller)

// Set cache headers
etag := response.ETag()
ctx.Response().Header().Set("ETag", etag)
ctx.Response().Header().Set("Cache-Control", "public, max-age=600")
ctx.Response().Header().Set("Last-Modified", response.LastModified().UTC().Format(http.TimeFormat))

// Check cache
if ctx.Request().Header.Get("If-None-Match") == etag {
    return ctx.NoContent(http.StatusNotModified)
}
```

## Testing

Run the API and test with curl:

```bash
# Start the server
make run

# Test ETag on GET
curl -i http://localhost:8080/api/v1/travellers/1

# Test cache hit
ETAG=$(curl -s -I http://localhost:8080/api/v1/travellers/1 | grep -i etag | cut -d' ' -f2)
curl -i -H "If-None-Match: $ETAG" http://localhost:8080/api/v1/travellers/1

# Test optimistic locking
curl -i -X PUT -H "If-Match: $ETAG" -H "Content-Type: application/json" \
  -d '{"name":"Test",...}' http://localhost:8080/api/v1/travellers/1
```

## Future Enhancements

- [ ] Add ETag support to accessory endpoints
- [ ] Implement Vary headers for content negotiation
- [ ] Add Redis caching layer for frequently accessed resources
- [ ] Support `If-Modified-Since` header validation
- [ ] Add cache warming strategies for popular resources
