#!/bin/bash

API_URL="http://localhost:8190"
EMAIL="test-import@example.com"
PASSWORD="Password123!"
TENANT_NAME="Test Church Import"

echo "=== Testing Pews CSV Import API ==="
echo ""

# 1. Register account
echo "1. Registering test account..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"tenant_name\":\"$TENANT_NAME\"}")

echo "$REGISTER_RESPONSE" | jq . || echo "$REGISTER_RESPONSE"

# Extract token
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to get token, trying login..."
  LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
  
  TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
fi

echo ""
echo "Token: $TOKEN"
echo ""

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to get authentication token!"
  exit 1
fi

# 2. Test JSON import (dry run)
echo "2. Testing JSON import (dry_run=true)..."
curl -s -X POST "$API_URL/api/import/people?dry_run=true" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @test_import_people.json | jq .

echo ""

# 3. Test JSON import (actual)
echo "3. Testing JSON import (actual)..."
curl -s -X POST "$API_URL/api/import/people" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @test_import_people.json | jq .

echo ""

# 4. Test CSV import
echo "4. Testing CSV import..."
curl -s -X POST "$API_URL/api/import/people" \
  -H "Content-Type: text/csv" \
  -H "Authorization: Bearer $TOKEN" \
  --data-binary @test_import_people.csv | jq .

echo ""

# 5. Verify people were imported
echo "5. Verifying imported people..."
curl -s -X GET "$API_URL/api/people" \
  -H "Authorization: Bearer $TOKEN" | jq '.people | length'

echo ""
echo "=== Test Complete ==="
