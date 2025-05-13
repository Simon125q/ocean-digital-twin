# API Documentation

This document outlines the available API endpoints and their supported query parameters.

## Endpoints

### `/health`

Returns the current health status of the database.

**Method:** GET  
**Response:** Status of the database connection

### `/chlorophyll`

Provides chlorophyll data in GeoJSON format.

**Method:** GET  
**Response:** GeoJSON containing chlorophyll measurements

#### Response Fields

| Field              | Description                                          |
| ------------------ | ---------------------------------------------------- |
| `id`               | Unique identifier of the data record in the database |
| `measurement_time` | Timestamp when the data was measured                 |
| `chlor_a`          | Measured chlorophyll-a value                         |

#### Query Parameters

| Parameter    | Description                                           |
| ------------ | ----------------------------------------------------- |
| `start_time` | Filter for records with measurement time ≥ this value |
| `end_time`   | Filter for records with measurement time ≤ this value |
| `min_lat`    | Filter for records with latitude ≥ this value         |
| `min_lon`    | Filter for records with longitude ≥ this value        |
| `max_lat`    | Filter for records with latitude ≤ this value         |
| `max_lon`    | Filter for records with longitude ≤ this value        |

## Examples

### Get chlorophyll data for a specific time range

```
GET /chlorophyll?start_time=2025-01-01T00:00:00Z&end_time=2025-01-31T23:59:59Z
```

### Get chlorophyll data for a specific geographic area

```
GET /chlorophyll?min_lat=40.0&min_lon=-75.0&max_lat=42.0&max_lon=-72.0
```
