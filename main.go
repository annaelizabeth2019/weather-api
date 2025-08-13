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
			<p>Example: <a href="/weather?lat=40.7128&lon=-74.0060">New York City</a></p>
			<p>Example: <a href="/weather?lat=34.0522&lon=-118.2437">Los Angeles</a></p>
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
		respondWithError(w, "Failed to retrieve weather data", http.StatusInternalServerError)
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

func getNWSWeather(lat, lon float64) (string, int, error) {
	// First, get the grid points for the coordinates
	gridURL := fmt.Sprintf("https://api.weather.gov/points/%.4f,%.4f", lat, lon)

	resp, err := http.Get(gridURL)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get grid points: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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

	// Get the actual forecast
	forecastResp, err := http.Get(gridResp.Properties.Forecast)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get forecast: %w", err)
	}
	defer forecastResp.Body.Close()

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

	// Convert temperature to Fahrenheit if it's in Celsius
	temperature := period.Temperature
	if strings.ToUpper(period.TemperatureUnit) == "C" {
		temperature = int(float64(period.Temperature)*9/5 + 32)
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
