# Weather Service

A simple HTTP server that provides weather forecasts using the National Weather Service API. This service accepts latitude and longitude coordinates and returns the current weather forecast along with a temperature characterization.

## Features

- **Weather Endpoint**: Accepts `lat` and `lon` query parameters
- **Forecast Data**: Returns the short forecast description (e.g., "Partly Cloudy")
- **Temperature Characterization**: Categorizes temperature as "hot", "cold", or "moderate"
- **Data Source**: Uses the National Weather Service API (free, no API key required)
- **Input Validation**: Validates coordinate ranges and formats
- **Error Handling**: Comprehensive error handling with meaningful messages

## Important: Coverage Limitations

**⚠️ The National Weather Service API only covers the United States and its territories.**

This service will return errors for coordinates outside:
- **Continental US**: 25°N to 50°N, 65°W to 125°W
- **Alaska**: 50°N to 75°N, 140°W to 180°E
- **Hawaii**: 19°N to 23°N, 154°W to 162°W
- **Puerto Rico & Caribbean**: 15°N to 20°N, 68°W to 80°W

## Temperature Classification

- **Hot**: 80°F and above
- **Cold**: 40°F and below  
- **Moderate**: Between 41°F and 79°F

## API Endpoints

### GET /weather
Returns weather information for specified coordinates.

**Query Parameters:**
- `lat` (required): Latitude (-90 to 90)
- `lon` (required): Longitude (-180 to 180)

**Example Request:**
```
GET /weather?lat=40.7128&lon=-74.0060
```

**Example Response:**
```json
{
  "forecast": "Partly Cloudy",
  "temperature": "moderate",
  "coordinates": "40.7128, -74.0060"
}
```

### GET /health
Health check endpoint to verify service is running.

### GET /
Root endpoint with usage instructions and example links.

## Building and Running

### Prerequisites
- Go 1.21 or later

### Build Instructions

1. **Clone and navigate to the project:**
   ```bash
   cd weather-api
   ```

2. **Download dependencies:**
   ```bash
   go mod tidy
   ```

3. **Build the service:**
   ```bash
   go build -o weather-service
   ```

4. **Run the service:**
   ```bash
   ./weather-service
   ```

   Or run directly with Go:
   ```bash
   go run main.go
   ```

The service will start on port 8080.

### Testing the Service

1. **Open your browser and navigate to:** `http://localhost:8080`
   - This will show the main page with usage instructions and example links

2. **Test the weather endpoint:**
   - Philadelphia: `http://localhost:8080/weather?lat=32.771496&lon=-89.118347` 🦅 
   - New York City: `http://localhost:8080/weather?lat=40.7128&lon=-74.0060`
   - Los Angeles: `http://localhost:8080/weather?lat=34.0522&lon=-118.2437`
   - Chicago: `http://localhost:8080/weather?lat=41.8781&lon=-87.6298`
   - Miami: `http://localhost:8080/weather?lat=25.7617&lon=-80.1918`
   - Seattle: `http://localhost:8080/weather?lat=47.6062&lon=-122.3321`

3. **Health check:**
   - `http://localhost:8080/health`

## Project Structure

```
weather-api/
├── main.go          # Main application code
├── go.mod           # Go module dependencies
├── README.md        # This file
└── .gitignore      # Git ignore file
```

## Implementation Details

### National Weather Service API Integration

The service uses the NWS API in two steps:
1. **Grid Points API**: Converts coordinates to grid points
2. **Forecast API**: Retrieves actual weather data for those grid points

### Error Handling

- Input validation for coordinates
- HTTP error responses with appropriate status codes
- Logging of internal errors for debugging
- User-friendly error messages

### Shortcuts and Limitations

**Note: This is a sample project, not production-ready. Key shortcuts include:**

- No authentication or rate limiting
- No caching of weather data
- No retry logic for failed API calls
- Basic error handling without detailed logging
- Single temperature characterization algorithm (could be more sophisticated)
- No unit tests
- No configuration file for customization
- Hardcoded port (8080)

## Dependencies

- **gorilla/mux**: HTTP router and URL matcher for clean routing
- **Standard library**: net/http, encoding/json, io, etc.

## Troubleshooting

### Common Issues

1. **Port already in use**: Change the port in `main.go` or stop other services using port 8080
2. **API errors**: The NWS API may have rate limits or temporary outages
3. **Invalid coordinates**: Ensure coordinates are within valid ranges

### Debug Mode

The service logs all API calls and errors to stdout. Check the console output for debugging information.

## Future Enhancements

If this were a production service, consider adding:
- Configuration management
- Caching layer (Redis)
- Rate limiting
- Authentication
- Comprehensive logging
- Metrics and monitoring
- Unit and integration tests
- Docker containerization
- Kubernetes deployment files
- API documentation (Swagger/OpenAPI)
