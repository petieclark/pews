#!/bin/bash

API_URL="http://localhost:8190"
EMAIL="pco-test@example.com"
PASSWORD="Password123!"
TENANT_NAME="PCO Test Church"

echo "=== Testing Pews PCO Import ==="
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

# 2. Test PCO CSV import (dry run)
echo "2. Testing PCO CSV import (dry_run=true)..."
curl -s -X POST "$API_URL/api/import/pco?dry_run=true" \
  -H "Content-Type: text/csv" \
  -H "Authorization: Bearer $TOKEN" \
  --data-binary @test_data/pco_sample.csv | jq .

echo ""

# 3. Test PCO CSV import (actual)
echo "3. Testing PCO CSV import (actual)..."
curl -s -X POST "$API_URL/api/import/pco" \
  -H "Content-Type: text/csv" \
  -H "Authorization: Bearer $TOKEN" \
  --data-binary @test_data/pco_sample.csv | jq .

echo ""

# 4. Verify people were imported
echo "4. Verifying imported people..."
PEOPLE_COUNT=$(curl -s -X GET "$API_URL/api/people" \
  -H "Authorization: Bearer $TOKEN" | jq '.people | length')

echo "Total people imported: $PEOPLE_COUNT"

# 5. Show first few people
echo ""
echo "5. First 3 imported people:"
curl -s -X GET "$API_URL/api/people" \
  -H "Authorization: Bearer $TOKEN" | jq '.people[0:3] | .[] | {first_name, last_name, email, membership_status}'

echo ""
echo "=== PCO Import Test Complete ==="
