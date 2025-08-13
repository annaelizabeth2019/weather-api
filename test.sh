#!/bin/bash

echo "Weather Service Test Script"
echo "=========================="
echo ""

# Check if service is running
echo "1. Testing health endpoint..."
curl -s http://localhost:8080/health
echo ""
echo ""

# Test weather endpoint with New York coordinates
echo "2. Testing weather endpoint (New York City)..."
curl -s "http://localhost:8080/weather?lat=40.7128&lon=-74.0060" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/weather?lat=40.7128&lon=-74.0060"
echo ""
echo ""

# Test weather endpoint with Los Angeles coordinates
echo "3. Testing weather endpoint (Los Angeles)..."
curl -s "http://localhost:8080/weather?lat=34.0522&lon=-118.2437" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/weather?lat=34.0522&lon=-118.2437"
echo ""
echo ""

# Test weather endpoint with Chicago coordinates
echo "4. Testing weather endpoint (Chicago)..."
curl -s "http://localhost:8080/weather?lat=41.8781&lon=-87.6298" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/weather?lat=41.8781&lon=-87.6298"
echo ""
echo ""

# Test error handling - missing parameters
echo "5. Testing error handling (missing parameters)..."
curl -s "http://localhost:8080/weather" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/weather"
echo ""
echo ""

# Test error handling - invalid coordinates
echo "6. Testing error handling (invalid coordinates)..."
curl -s "http://localhost:8080/weather?lat=100&lon=200" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/weather?lat=100&lon=200"
echo ""
echo ""

# Test error handling - international coordinates (should fail with coverage error)
echo "7. Testing error handling (international coordinates - London)..."
curl -s "http://localhost:8080/weather?lat=51.5074&lon=-0.1278" | jq '.' 2>/dev/null || curl -s "http://localhost:8080/weather?lat=51.5074&lon=-0.1278"
echo ""
echo ""

echo "Test completed!"
echo ""
echo "Note: If you see JSON responses, the service is working correctly."
echo "If you see connection refused errors, make sure the service is running on port 8080."
echo "International coordinates should now return clear error messages about coverage limitations."
