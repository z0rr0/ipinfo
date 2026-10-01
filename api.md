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

### GET /ip
Returns only the client IP address as a single text line.

### GET /version
Returns application version information.

### GET /health
Returns application health status. It is a liveness check: no GeoIP lookup,
the `ip_header` request header is not required, requests are not written to the access log.

## Query Parameters

### ip
Any endpoint except `/health` accepts `?ip=<address>` (IPv4 or IPv6) and returns info about
this address instead of the client IP, e.g. `/json?ip=8.8.8.8`.

- The address is normalized: `/ip?ip=::ffff:8.8.8.8` returns `8.8.8.8`.
- It takes priority over the client IP: the `ip_header` request header is not required,
  `ip_itself` is not applied, so a private or loopback address returns empty geo fields.
- An empty value (`?ip=`) is ignored and the client IP is used.
- An invalid value returns `400 Bad Request` with the body `invalid ip parameter`.
- A query pair with an invalid escape is dropped entirely and the client IP is used,
  e.g. an unescaped `%` in an IPv6 zone `?ip=fe80::1%eth0`.

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
```

### Health Response
```
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Cache-Control: no-cache, no-store, must-revalidate

OK
```

### Invalid ip Parameter Response
```
HTTP/1.1 400 Bad Request
Content-Type: text/plain; charset=utf-8

invalid ip parameter
```
