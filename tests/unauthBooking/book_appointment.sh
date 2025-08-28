#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:3000/api/v1}

echo "Creating first unauth appointment..."
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/public-appointment-request" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "service_id": 1,
  "city_id": 1,
  "area_id": 10,
  "street": "вул. Володимирська",
  "location_type": "at home",
  "unit": "3",
  "apt": "32",
  "animal_size_id": 2,
  "description": "Routine check-up",
  "appointment_date": "2025-09-01T00:00:00Z",
  "start_time": "2025-09-01T14:30:00Z",
  "end_time": "2025-09-01T15:00:00Z",
  "amount": 75.50,
  "email": "user@example.com",
  "specialist_id": 1,
  "status": "pending"
}
EOF
)

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
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/public-appointment-request" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  service_id": 1,
  "city_id": 1,
  "area_id": 10,
  "street": "вул. Володимирська",
  "location_type": "at home",
  "unit": "3",
  "apt": "32",
  "animal_size_id": 2,
  "description": "Routine check-up",
  "appointment_date": "2025-09-01T00:00:00Z",
  "start_time": "2025-09-01T14:30:00Z",
  "end_time": "2025-09-01T15:00:00Z",
  "amount": 75.50,
  "email": "user@example.com",
  "specialist_id": 1,
  "status": "pending"
}
EOF
)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 409 ]; then
    echo "✅ Duplicate appointment test passed"
else
    echo "❌ Duplicate appointment test failed: expected 409, got $http_code"
    echo "Response: $body"
    exit 1
fi
