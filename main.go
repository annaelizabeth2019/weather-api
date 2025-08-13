package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// WeatherResponse represents the response structure for our weather endpoint
type WeatherResponse struct {
	Forecast     string `json:"forecast"`
	Temperature  string `json:"temperature"`
	Coordinates  string `json:"coordinates"`
	ErrorMessage string `json:"error,omitempty"`
}

// NWSResponse represents the National Weather Service API response structure
type NWSResponse struct {
	Properties struct {
		Periods []struct {
			ShortForecast   string `json:"shortForecast"`
			Temperature     int    `json:"temperature"`
			TemperatureUnit string `json:"temperatureUnit"`
		} `json:"periods"`
	} `json:"properties"`
}

func main() {
	r := mux.NewRouter()

	// Weather endpoint
	r.HandleFunc("/weather", getWeather).Methods("GET")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Weather service is running"))
	}).Methods("GET")

	// Root endpoint with instructions
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
		<!DOCTYPE html>
		<html>
		<head><title>Weather Service</title></head>
		<body>
			<h1>Weather Service</h1>
			<p>Use the /weather endpoint with latitude and longitude parameters:</p>
			<p><code>/weather?lat=40.7128&lon=-74.0060</code></p>
			
			<h2>Example US Cities:</h2>
			<p>Example: <a href="/weather?lat=40.7128&lon=-74.0060">New York City</a></p>
			<p>Example: <a href="/weather?lat=34.0522&lon=-118.2437">Los Angeles</a></p>
			<p>Example: <a href="/weather?lat=41.8781&lon=-87.6298">Chicago</a></p>
			<p>Example: <a href="/weather?lat=25.7617&lon=-80.1918">Miami</a></p>
			<p>Example: <a href="/weather?lat=47.6062&lon=-122.3321">Seattle</a></p>
			
			<h2>⚠️ Important Note:</h2>
			<p><strong>This service only works for US locations.</strong> The National Weather Service API covers the United States and its territories only.</p>
			<p>For international locations, coordinates outside the US will return an error.</p>
		</body>
		</html>
		`
		w.Write([]byte(html))
	}).Methods("GET")

	log.Println("Starting weather service on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	if lat == "" || lon == "" {
		respondWithError(w, "Missing required parameters: lat and lon", http.StatusBadRequest)
		return
	}

	// Validate coordinates
	latFloat, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		respondWithError(w, "Invalid latitude format", http.StatusBadRequest)
		return
	}

	lonFloat, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		respondWithError(w, "Invalid longitude format", http.StatusBadRequest)
		return
	}

	if latFloat < -90 || latFloat > 90 {
		respondWithError(w, "Latitude must be between -90 and 90", http.StatusBadRequest)
		return
	}

	if lonFloat < -180 || lonFloat > 180 {
		respondWithError(w, "Longitude must be between -180 and 180", http.StatusBadRequest)
		return
	}

	// Get weather data from National Weather Service
	forecast, temp, err := getNWSWeather(latFloat, lonFloat)
	if err != nil {
		log.Printf("Error getting weather data: %v", err)

		// Provide more specific error messages based on error type
		if strings.Contains(err.Error(), "outside NWS coverage area") {
			respondWithError(w, err.Error(), http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "not found in NWS grid system") {
			respondWithError(w, err.Error(), http.StatusBadRequest)
		} else {
			respondWithError(w, "Failed to retrieve weather data", http.StatusInternalServerError)
		}
		return
	}

	// Determine temperature characterization
	tempChar := characterizeTemperature(temp)

	response := WeatherResponse{
		Forecast:    forecast,
		Temperature: tempChar,
		Coordinates: fmt.Sprintf("%.4f, %.4f", latFloat, lonFloat),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// isNWSCoverageArea checks if coordinates are within the National Weather Service coverage area
// NWS covers the United States and its territories
func isNWSCoverageArea(lat, lon float64) bool {
	// NWS coverage roughly covers:
	// Continental US: 25°N to 50°N, 65°W to 125°W
	// Alaska: 50°N to 75°N, 140°W to 180°E
	// Hawaii: 19°N to 23°N, 154°W to 162°W
	// Puerto Rico & Caribbean: 15°N to 20°N, 68°W to 80°W

	// Main continental US (most restrictive bounds)
	if lat >= 25 && lat <= 50 && lon >= -125 && lon <= -65 {
		return true
	}

	// Alaska (roughly 50°N to 75°N, 140°W to 180°E)
	if lat >= 50 && lat <= 75 && lon >= -180 && lon <= -140 {
		return true
	}

	// Hawaii (roughly 19°N to 23°N, 154°W to 162°W)
	if lat >= 19 && lat <= 23 && lon >= -162 && lon <= -154 {
		return true
	}

	// Puerto Rico and Caribbean (roughly 15°N to 20°N, 68°W to 80°W)
	if lat >= 15 && lat <= 20 && lon >= -80 && lon <= -68 {
		return true
	}

	return false
}

func getNWSWeather(lat, lon float64) (string, int, error) {
	// Check if coordinates are within NWS coverage area
	if !isNWSCoverageArea(lat, lon) {
		return "", 0, fmt.Errorf("coordinates (%.4f, %.4f) are outside NWS coverage area (US and territories only)", lat, lon)
	}

	log.Printf("Fetching weather for coordinates: %.4f, %.4f", lat, lon)

	// First, get the grid points for the coordinates
	gridURL := fmt.Sprintf("https://api.weather.gov/points/%.4f,%.4f", lat, lon)
	log.Printf("Calling NWS grid points API: %s", gridURL)

	resp, err := http.Get(gridURL)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get grid points: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Grid points API response status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		return "", 0, fmt.Errorf("coordinates (%.4f, %.4f) not found in NWS grid system - may be outside coverage area", lat, lon)
	} else if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("grid points API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read grid points response: %w", err)
	}

	// Parse grid response to get forecast URL
	var gridResp struct {
		Properties struct {
			Forecast string `json:"forecast"`
		} `json:"properties"`
	}

	if err := json.Unmarshal(body, &gridResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse grid response: %w", err)
	}

	if gridResp.Properties.Forecast == "" {
		return "", 0, fmt.Errorf("no forecast URL found in grid response")
	}

	log.Printf("Forecast URL: %s", gridResp.Properties.Forecast)

	// Get the actual forecast
	forecastResp, err := http.Get(gridResp.Properties.Forecast)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get forecast: %w", err)
	}
	defer forecastResp.Body.Close()

	log.Printf("Forecast API response status: %d", forecastResp.StatusCode)

	if forecastResp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("forecast API returned status: %d", forecastResp.StatusCode)
	}

	forecastBody, err := io.ReadAll(forecastResp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read forecast response: %w", err)
	}

	var nwsResp NWSResponse
	if err := json.Unmarshal(forecastBody, &nwsResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse forecast response: %w", err)
	}

	if len(nwsResp.Properties.Periods) == 0 {
		return "", 0, fmt.Errorf("no forecast periods found")
	}

	// Get today's forecast (first period)
	period := nwsResp.Properties.Periods[0]
	log.Printf("Retrieved forecast: %s, Temperature: %d°%s", period.ShortForecast, period.Temperature, period.TemperatureUnit)

	// Convert temperature to Fahrenheit if it's in Celsius
	temperature := period.Temperature
	if strings.ToUpper(period.TemperatureUnit) == "C" {
		temperature = int(float64(period.Temperature)*9/5 + 32)
		log.Printf("Converted temperature from %d°C to %d°F", period.Temperature, temperature)
	}

	return period.ShortForecast, temperature, nil
}

func characterizeTemperature(tempF int) string {
	switch {
	case tempF >= 80:
		return "hot"
	case tempF <= 40:
		return "cold"
	default:
		return "moderate"
	}
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	response := WeatherResponse{
		ErrorMessage: message,
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
