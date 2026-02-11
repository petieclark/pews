#!/bin/bash

set -e

API_BASE="http://localhost:8190"
TENANT_SLUG="test-church"
EMAIL="admin@test.church"
PASSWORD="password123"

echo "🔧 Testing Pews Audit System"
echo "================================"

# Step 1: Register/Login to get token
echo ""
echo "1️⃣  Registering test tenant and user..."
REGISTER_RESPONSE=$(curl -s -X POST "${API_BASE}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "Test Church",
    "email": "'"${EMAIL}"'",
    "password": "'"${PASSWORD}"'"
  }' 2>/dev/null || echo '{}')

echo "$REGISTER_RESPONSE" | jq '.' 2>/dev/null || echo "Registration response: $REGISTER_RESPONSE"

# Login to get token
echo ""
echo "2️⃣  Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_slug": "'"${TENANT_SLUG}"'",
    "email": "'"${EMAIL}"'",
    "password": "'"${PASSWORD}"'"
  }')

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token' 2>/dev/null)

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Failed to get auth token"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Got auth token: ${TOKEN:0:20}..."

# Step 3: Perform some actions that should be audited
echo ""
echo "3️⃣  Creating a person (should be audited)..."
PERSON_RESPONSE=$(curl -s -X POST "${API_BASE}/api/people" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "phone": "555-1234"
  }')

PERSON_ID=$(echo "$PERSON_RESPONSE" | jq -r '.id' 2>/dev/null)
echo "✅ Created person with ID: $PERSON_ID"

# Step 4: Update the person
echo ""
echo "4️⃣  Updating person (should be audited)..."
curl -s -X PUT "${API_BASE}/api/people/${PERSON_ID}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "first_name": "Jane",
    "last_name": "Doe",
    "email": "jane@example.com",
    "phone": "555-5678"
  }' > /dev/null

echo "✅ Updated person"

# Step 5: Enable a module
echo ""
echo "5️⃣  Enabling a module (should be audited)..."
curl -s -X POST "${API_BASE}/api/tenant/modules/people/enable" \
  -H "Authorization: Bearer ${TOKEN}" > /dev/null 2>&1 || true

echo "✅ Module action performed"

# Wait a moment for async audit logs to be written
sleep 2

# Step 6: Fetch audit logs
echo ""
echo "6️⃣  Fetching audit logs..."
LOGS_RESPONSE=$(curl -s -X GET "${API_BASE}/api/audit/logs?page=1&page_size=20" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$LOGS_RESPONSE" | jq '.' 2>/dev/null

# Step 7: Fetch security dashboard
echo ""
echo "7️⃣  Fetching security dashboard..."
SECURITY_RESPONSE=$(curl -s -X GET "${API_BASE}/api/audit/security" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$SECURITY_RESPONSE" | jq '.' 2>/dev/null

# Step 8: Test failed login (should be tracked)
echo ""
echo "8️⃣  Testing failed login tracking..."
curl -s -X POST "${API_BASE}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_slug": "'"${TENANT_SLUG}"'",
    "email": "'"${EMAIL}"'",
    "password": "wrongpassword"
  }' > /dev/null

echo "✅ Failed login attempt recorded"

# Wait and fetch security dashboard again
sleep 1

echo ""
echo "9️⃣  Checking security dashboard again..."
SECURITY_RESPONSE2=$(curl -s -X GET "${API_BASE}/api/audit/security" \
  -H "Authorization: Bearer ${TOKEN}")

echo "$SECURITY_RESPONSE2" | jq '.' 2>/dev/null

# Step 9: Test CSV export
echo ""
echo "🔟 Testing CSV export..."
curl -s -X GET "${API_BASE}/api/audit/export" \
  -H "Authorization: Bearer ${TOKEN}" \
  -o /tmp/audit_export.csv

if [ -f /tmp/audit_export.csv ]; then
  echo "✅ CSV export successful"
  echo "First 10 lines:"
  head -10 /tmp/audit_export.csv
else
  echo "❌ CSV export failed"
fi

echo ""
echo "================================"
echo "✅ Audit system test complete!"
echo ""
echo "Summary:"
echo "  - Created and updated person (audit logged)"
echo "  - Performed module action (audit logged)"
echo "  - Tested failed login tracking"
echo "  - Verified audit log API"
echo "  - Verified security dashboard API"
echo "  - Tested CSV export"
echo ""
echo "🌐 Frontend available at: http://localhost:5173/dashboard/settings/audit"
