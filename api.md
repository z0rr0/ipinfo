# IPInfo API Documentation

## Endpoints

### GET /
Returns detailed IP information in text format.

### GET /short
Returns concise IP information in text format.

### GET /compact
Returns minimal IP information in text format.

### GET /json
Returns IP information in JSON format.

### GET /xml
Returns IP information in XML format.

### GET /html
Returns IP information in HTML format.

### GET /full
Returns IP information in enhanced HTML format.

### GET /version
Returns application version information.

### GET /health
Returns application health status.

## Response Format

### JSON Response
```json
{
  "ip": "192.168.1.1",
  "country": "United States",
  "city": "New York",
  "longitude": -74.0060,
  "latitude": 40.7128,
  "utc_time": "2023-01-01T12:00:00Z",
  "time_zone": "America/New_York",
  "language": "en"
}
