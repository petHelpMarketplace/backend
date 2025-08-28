#!/bin/bash

BASE_URL=${BASE_URL:-http://localhost:3000}

echo "Creating first unauth appointment..."
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/public-appointment-request \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": 1,
    "city_id": 1,
    "district_id": 1,
    "street": "Main Street",
    "location_type": "at home",
    "unit": "12B",
    "apt": "34",
    "animal_size_id": 2,
    "description": "Routine check-up",
    "specialist_id": 1,
    "date": "2025-09-01",
    "start_time": "14:30",
    "end_time": "15:00",
    "amount": 75.50,
    "email": "user@example.com"
  }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 201 ]; then
    echo "✅ First appointment created successfully"
else
    echo "❌ Failed to create first appointment: $http_code"
    echo "Response: $body"
    exit 1
fi

echo -e "\nTesting duplicate appointment (should fail with 409)..."
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/public-appointment-request \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": 1,
    "city_id": 1,
    "district_id": 1,
    "street": "Main Street",
    "location_type": "home",
    "unit": "12B",
    "apt": "34",
    "animal_size_id": 2,
    "description": "Routine check-up",
    "specialist_id": 123,
    "date": "2025-09-01",
    "start_time": "14:30",
    "end_time": "15:00",
    "amount": 75.50,
    "email": "user@example.com"
  }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 409 ]; then
    echo "✅ Duplicate appointment test passed"
else
    echo "❌ Duplicate appointment test failed: expected 409, got $http_code"
    echo "Response: $body"
    exit 1
fi
