#!/bin/bash
# Demo Seed Data Script for Pews
# Populates the demo database with realistic church data
# Usage: ./scripts/seed-demo.sh [base_url] [slug] [email] [password]

set -e

BASE=${1:-https://demo.pews.app}
SLUG=${2:-demo-church}
EMAIL=${3:-demo@pews.app}
PASS=${4:-demo1234}

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Pews Demo Seed Data Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Base URL: $BASE"
echo -e "Church: $SLUG"
echo -e "Email: $EMAIL"
echo ""

# Login to get JWT token
echo -e "${YELLOW}Logging in...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
  echo -e "${RED}Login failed. Please check your credentials.${NC}"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✓ Logged in successfully${NC}"
echo ""

# Helper function to make authenticated API calls
api_call() {
  local method=$1
  local endpoint=$2
  local data=$3
  
  if [ -z "$data" ]; then
    curl -s -X $method "$BASE$endpoint" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json"
  else
    curl -s -X $method "$BASE$endpoint" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "$data"
  fi
}

# Helper to check if entity exists
entity_exists() {
  local response=$1
  local check_field=$2
  
  echo "$response" | grep -q "\"$check_field\""
}

# Store created IDs
declare -A PERSON_IDS
declare -A GROUP_IDS
declare -A SONG_IDS
declare -A FUND_IDS

# ============================================
# PART 1: PEOPLE (25-30 members)
# ============================================
echo -e "${BLUE}Creating People...${NC}"

# Array of realistic names
PEOPLE_DATA=(
  '{"first_name":"James","last_name":"Mitchell","email":"james.mitchell@example.com","phone":"912-555-0101","member_status":"member","role":"Pastor"}'
  '{"first_name":"Sarah","last_name":"Thompson","email":"sarah.thompson@example.com","phone":"912-555-0102","member_status":"member"}'
  '{"first_name":"David","last_name":"Thompson","email":"david.thompson@example.com","phone":"912-555-0103","member_status":"member"}'
  '{"first_name":"Maria","last_name":"Rodriguez","email":"maria.rodriguez@example.com","phone":"912-555-0104","member_status":"member"}'
  '{"first_name":"Robert","last_name":"Johnson","email":"robert.johnson@example.com","phone":"912-555-0105","member_status":"member","role":"Elder"}'
  '{"first_name":"Emily","last_name":"Williams","email":"emily.williams@example.com","phone":"912-555-0106","member_status":"member"}'
  '{"first_name":"Michael","last_name":"Brown","email":"michael.brown@example.com","phone":"912-555-0107","member_status":"member","role":"Worship Leader"}'
  '{"first_name":"Jessica","last_name":"Davis","email":"jessica.davis@example.com","phone":"912-555-0108","member_status":"member"}'
  '{"first_name":"Christopher","last_name":"Garcia","email":"chris.garcia@example.com","phone":"912-555-0109","member_status":"member"}'
  '{"first_name":"Amanda","last_name":"Martinez","email":"amanda.martinez@example.com","phone":"912-555-0110","member_status":"member"}'
  '{"first_name":"Joshua","last_name":"Anderson","email":"joshua.anderson@example.com","phone":"912-555-0111","member_status":"regular_visitor"}'
  '{"first_name":"Ashley","last_name":"Taylor","email":"ashley.taylor@example.com","phone":"912-555-0112","member_status":"member"}'
  '{"first_name":"Daniel","last_name":"Thomas","email":"daniel.thomas@example.com","phone":"912-555-0113","member_status":"member","role":"Youth Leader"}'
  '{"first_name":"Stephanie","last_name":"Moore","email":"stephanie.moore@example.com","phone":"912-555-0114","member_status":"member"}'
  '{"first_name":"Matthew","last_name":"Jackson","email":"matthew.jackson@example.com","phone":"912-555-0115","member_status":"regular_visitor"}'
  '{"first_name":"Jennifer","last_name":"White","email":"jennifer.white@example.com","phone":"912-555-0116","member_status":"member"}'
  '{"first_name":"Anthony","last_name":"Harris","email":"anthony.harris@example.com","phone":"912-555-0117","member_status":"member"}'
  '{"first_name":"Lisa","last_name":"Martin","email":"lisa.martin@example.com","phone":"912-555-0118","member_status":"member"}'
  '{"first_name":"Mark","last_name":"Thompson","email":"mark.thompson@example.com","phone":"912-555-0119","member_status":"member"}'
  '{"first_name":"Karen","last_name":"Lee","email":"karen.lee@example.com","phone":"912-555-0120","member_status":"regular_visitor"}'
  '{"first_name":"Steven","last_name":"Walker","email":"steven.walker@example.com","phone":"912-555-0121","member_status":"member"}'
  '{"first_name":"Betty","last_name":"Hall","email":"betty.hall@example.com","phone":"912-555-0122","member_status":"member"}'
  '{"first_name":"Brian","last_name":"Allen","email":"brian.allen@example.com","phone":"912-555-0123","member_status":"new_visitor"}'
  '{"first_name":"Dorothy","last_name":"Young","email":"dorothy.young@example.com","phone":"912-555-0124","member_status":"member"}'
  '{"first_name":"Kevin","last_name":"King","email":"kevin.king@example.com","phone":"912-555-0125","member_status":"member"}'
  '{"first_name":"Nancy","last_name":"Wright","email":"nancy.wright@example.com","phone":"912-555-0126","member_status":"new_visitor"}'
  '{"first_name":"George","last_name":"Lopez","email":"george.lopez@example.com","phone":"912-555-0127","member_status":"member"}'
  '{"first_name":"Sandra","last_name":"Hill","email":"sandra.hill@example.com","phone":"912-555-0128","member_status":"member"}'
  '{"first_name":"Tyler","last_name":"Scott","email":"tyler.scott@example.com","phone":"912-555-0129","member_status":"regular_visitor"}'
  '{"first_name":"Carol","last_name":"Green","email":"carol.green@example.com","phone":"912-555-0130","member_status":"member"}'
)

for person_data in "${PEOPLE_DATA[@]}"; do
  name=$(echo $person_data | grep -o '"first_name":"[^"]*' | sed 's/"first_name":"//')
  last=$(echo $person_data | grep -o '"last_name":"[^"]*' | sed 's/"last_name":"//')
  
  response=$(api_call POST "/api/people" "$person_data")
  
  if entity_exists "$response" "id"; then
    person_id=$(echo $response | grep -o '"id":"[^"]*' | sed 's/"id":"//' | head -1)
    PERSON_IDS["$name $last"]=$person_id
    echo -e "${GREEN}✓${NC} Created: $name $last"
  else
    echo -e "${YELLOW}⊘${NC} Skipped: $name $last (may already exist)"
  fi
done

echo ""

# ============================================
# PART 2: FUNDS
# ============================================
echo -e "${BLUE}Creating Funds...${NC}"

FUNDS_DATA=(
  '{"name":"General Fund","description":"General church operations and ministry","is_default":true}'
  '{"name":"Building Fund","description":"Building maintenance and expansion"}'
  '{"name":"Missions","description":"Local and global missions support"}'
  '{"name":"Youth Ministry","description":"Youth programs and activities"}'
  '{"name":"Benevolence","description":"Helping those in need"}'
)

for fund_data in "${FUNDS_DATA[@]}"; do
  name=$(echo $fund_data | grep -o '"name":"[^"]*' | sed 's/"name":"//')
  
  response=$(api_call POST "/api/giving/funds" "$fund_data")
  
  if entity_exists "$response" "id"; then
    fund_id=$(echo $response | grep -o '"id":"[^"]*' | sed 's/"id":"//' | head -1)
    FUND_IDS["$name"]=$fund_id
    echo -e "${GREEN}✓${NC} Created fund: $name"
  else
    echo -e "${YELLOW}⊘${NC} Skipped fund: $name"
  fi
done

echo ""

# ============================================
# PART 3: SONGS
# ============================================
echo -e "${BLUE}Creating Songs...${NC}"

SONGS_DATA=(
  '{"title":"10,000 Reasons (Bless the Lord)","artist":"Matt Redman","key":"G","tempo":"73","ccli":"6016351"}'
  '{"title":"How Great Is Our God","artist":"Chris Tomlin","key":"C","tempo":"76","ccli":"4348399"}'
  '{"title":"Goodness of God","artist":"Bethel Music","key":"C","tempo":"120","ccli":"7117726"}'
  '{"title":"Way Maker","artist":"Sinach","key":"D","tempo":"68","ccli":"7115744"}'
  '{"title":"Build My Life","artist":"Housefires","key":"C","tempo":"72","ccli":"7070345"}'
  '{"title":"Holy Spirit","artist":"Francesca Battistelli","key":"C","tempo":"69","ccli":"6087919"}'
  '{"title":"What A Beautiful Name","artist":"Hillsong Worship","key":"D","tempo":"68","ccli":"7068424"}'
  '{"title":"Great Are You Lord","artist":"All Sons & Daughters","key":"G","tempo":"66","ccli":"6460220"}'
  '{"title":"Reckless Love","artist":"Cory Asbury","key":"C","tempo":"70","ccli":"7089641"}'
  '{"title":"King of Kings","artist":"Hillsong Worship","key":"D","tempo":"72","ccli":"7127647"}'
  '{"title":"Good Good Father","artist":"Chris Tomlin","key":"A","tempo":"120","ccli":"7036612"}'
  '{"title":"O Come to the Altar","artist":"Elevation Worship","key":"G","tempo":"76","ccli":"7051511"}'
  '{"title":"Amazing Grace (My Chains Are Gone)","artist":"Chris Tomlin","key":"G","tempo":"80","ccli":"4768151"}'
  '{"title":"This Is Amazing Grace","artist":"Phil Wickham","key":"C","tempo":"122","ccli":"6333821"}'
  '{"title":"Oceans (Where Feet May Fail)","artist":"Hillsong United","key":"D","tempo":"72","ccli":"6428767"}'
)

for song_data in "${SONGS_DATA[@]}"; do
  title=$(echo $song_data | grep -o '"title":"[^"]*' | sed 's/"title":"//')
  
  response=$(api_call POST "/api/worship/songs" "$song_data")
  
  if entity_exists "$response" "id"; then
    song_id=$(echo $response | grep -o '"id":"[^"]*' | sed 's/"id":"//' | head -1)
    SONG_IDS["$title"]=$song_id
    echo -e "${GREEN}✓${NC} Created song: $title"
  else
    echo -e "${YELLOW}⊘${NC} Skipped song: $title"
  fi
done

echo ""

# ============================================
# PART 4: SERVICES
# ============================================
echo -e "${BLUE}Creating Services...${NC}"

# Helper to get date N days ago
days_ago() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    date -v-${1}d +%Y-%m-%d
  else
    date -d "$1 days ago" +%Y-%m-%d
  fi
}

# Last 3 Sundays
for i in 0 7 14; do
  service_date=$(days_ago $i)
  
  service_data="{\"date\":\"${service_date}T10:00:00Z\",\"type\":\"Sunday Morning\",\"notes\":\"Worship service\"}"
  
  response=$(api_call POST "/api/services" "$service_data")
  
  if entity_exists "$response" "id"; then
    echo -e "${GREEN}✓${NC} Created service: Sunday $(days_ago $i)"
  else
    echo -e "${YELLOW}⊘${NC} Skipped service: Sunday $(days_ago $i)"
  fi
done

# Wednesday service
wed_date=$(days_ago 3)
wed_data="{\"date\":\"${wed_date}T19:00:00Z\",\"type\":\"Wednesday Night Prayer\",\"notes\":\"Prayer meeting\"}"
response=$(api_call POST "/api/services" "$wed_data")

if entity_exists "$response" "id"; then
  echo -e "${GREEN}✓${NC} Created service: Wednesday Prayer"
else
  echo -e "${YELLOW}⊘${NC} Skipped service: Wednesday Prayer"
fi

echo ""

# ============================================
# PART 5: GROUPS
# ============================================
echo -e "${BLUE}Creating Groups...${NC}"

GROUPS_DATA=(
  '{"name":"Sunday Morning Bible Study","type":"Small Group","status":"active","description":"Weekly Bible study group","meeting_day":"Sunday","meeting_time":"09:00"}'
  '{"name":"Youth Group","type":"Ministry Team","status":"active","description":"Middle and high school ministry","meeting_day":"Wednesday","meeting_time":"18:30"}'
  '{"name":"Women'\''s Prayer Circle","type":"Small Group","status":"active","description":"Women'\''s prayer and fellowship","meeting_day":"Tuesday","meeting_time":"10:00"}'
  '{"name":"Worship Team","type":"Ministry Team","status":"active","description":"Sunday worship band and vocalists","meeting_day":"Thursday","meeting_time":"19:00"}'
  '{"name":"Building Committee","type":"Committee","status":"active","description":"Facilities planning and maintenance"}'
)

for group_data in "${GROUPS_DATA[@]}"; do
  name=$(echo $group_data | grep -o '"name":"[^"]*' | sed 's/"name":"//')
  
  response=$(api_call POST "/api/groups" "$group_data")
  
  if entity_exists "$response" "id"; then
    group_id=$(echo $response | grep -o '"id":"[^"]*' | sed 's/"id":"//' | head -1)
    GROUP_IDS["$name"]=$group_id
    echo -e "${GREEN}✓${NC} Created group: $name"
  else
    echo -e "${YELLOW}⊘${NC} Skipped group: $name"
  fi
done

echo ""

# ============================================
# PART 6: DONATIONS
# ============================================
echo -e "${BLUE}Creating Donations...${NC}"

# Get some person IDs to assign donations
DONOR_NAMES=(
  "James Mitchell"
  "Sarah Thompson"
  "David Thompson"
  "Maria Rodriguez"
  "Robert Johnson"
  "Emily Williams"
  "Michael Brown"
  "Jessica Davis"
  "Christopher Garcia"
  "Amanda Martinez"
)

# Random amounts and funds
AMOUNTS=(50 75 100 125 150 200 250 300 350 400 450 500)
FUND_NAMES=("General Fund" "Building Fund" "Missions" "Youth Ministry")

for i in {1..20}; do
  # Random donor
  donor_idx=$((RANDOM % ${#DONOR_NAMES[@]}))
  donor="${DONOR_NAMES[$donor_idx]}"
  person_id="${PERSON_IDS[$donor]}"
  
  # Random amount
  amount_idx=$((RANDOM % ${#AMOUNTS[@]}))
  amount="${AMOUNTS[$amount_idx]}"
  
  # Random fund
  fund_idx=$((RANDOM % ${#FUND_NAMES[@]}))
  fund_name="${FUND_NAMES[$fund_idx]}"
  fund_id="${FUND_IDS[$fund_name]}"
  
  # Random date in last 60 days
  days_back=$((RANDOM % 60))
  donation_date=$(days_ago $days_back)
  
  if [ -n "$person_id" ] && [ -n "$fund_id" ]; then
    donation_data="{\"person_id\":\"$person_id\",\"amount\":$amount,\"fund_id\":\"$fund_id\",\"date\":\"${donation_date}T12:00:00Z\",\"method\":\"check\"}"
    
    response=$(api_call POST "/api/giving/donations" "$donation_data")
    
    if entity_exists "$response" "id"; then
      echo -e "${GREEN}✓${NC} Created donation: \$$amount from ${donor} to ${fund_name}"
    fi
  fi
done

echo ""

# ============================================
# PART 7: EVENTS
# ============================================
echo -e "${BLUE}Creating Events...${NC}"

# Helper to get date N days from now
days_from_now() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    date -v+${1}d +%Y-%m-%d
  else
    date -d "$1 days" +%Y-%m-%d
  fi
}

EVENTS_DATA=(
  "{\"title\":\"Sunday Service\",\"description\":\"Weekly worship service\",\"start_date\":\"$(days_from_now 0)T10:00:00Z\",\"end_date\":\"$(days_from_now 0)T11:30:00Z\",\"is_recurring\":true,\"recurrence_rule\":\"FREQ=WEEKLY;BYDAY=SU\",\"location\":\"Main Sanctuary\"}"
  "{\"title\":\"Wednesday Night Prayer\",\"description\":\"Midweek prayer meeting\",\"start_date\":\"$(days_from_now 2)T19:00:00Z\",\"end_date\":\"$(days_from_now 2)T20:00:00Z\",\"is_recurring\":true,\"recurrence_rule\":\"FREQ=WEEKLY;BYDAY=WE\",\"location\":\"Fellowship Hall\"}"
  "{\"title\":\"Youth Game Night\",\"description\":\"Pizza and games for youth group\",\"start_date\":\"$(days_from_now 10)T18:00:00Z\",\"end_date\":\"$(days_from_now 10)T21:00:00Z\",\"location\":\"Youth Room\"}"
  "{\"title\":\"Church Potluck\",\"description\":\"Monthly fellowship meal\",\"start_date\":\"$(days_from_now 15)T12:00:00Z\",\"end_date\":\"$(days_from_now 15)T14:00:00Z\",\"location\":\"Fellowship Hall\"}"
)

for event_data in "${EVENTS_DATA[@]}"; do
  title=$(echo $event_data | grep -o '"title":"[^"]*' | sed 's/"title":"//')
  
  response=$(api_call POST "/api/calendar/events" "$event_data")
  
  if entity_exists "$response" "id"; then
    echo -e "${GREEN}✓${NC} Created event: $title"
  else
    echo -e "${YELLOW}⊘${NC} Skipped event: $title"
  fi
done

echo ""

# ============================================
# PART 8: FOLLOW-UPS
# ============================================
echo -e "${BLUE}Creating Follow-ups...${NC}"

FOLLOWUP_PEOPLE=("Brian Allen" "Nancy Wright" "Karen Lee" "Matthew Jackson" "Joshua Anderson")
FOLLOWUP_TYPES=("visit" "call" "email")
FOLLOWUP_STATUSES=("pending" "in_progress" "completed")

for i in {0..5}; do
  person_idx=$((i % ${#FOLLOWUP_PEOPLE[@]}))
  person="${FOLLOWUP_PEOPLE[$person_idx]}"
  person_id="${PERSON_IDS[$person]}"
  
  type_idx=$((RANDOM % ${#FOLLOWUP_TYPES[@]}))
  followup_type="${FOLLOWUP_TYPES[$type_idx]}"
  
  status_idx=$((RANDOM % ${#FOLLOWUP_STATUSES[@]}))
  status="${FOLLOWUP_STATUSES[$status_idx]}"
  
  due_days=$((RANDOM % 14))
  due_date=$(days_from_now $due_days)
  
  if [ -n "$person_id" ]; then
    followup_data="{\"person_id\":\"$person_id\",\"type\":\"$followup_type\",\"notes\":\"Follow up with $person\",\"status\":\"$status\",\"due_date\":\"${due_date}T12:00:00Z\"}"
    
    response=$(api_call POST "/api/followups" "$followup_data")
    
    if entity_exists "$response" "id"; then
      echo -e "${GREEN}✓${NC} Created follow-up: $followup_type for $person ($status)"
    fi
  fi
done

echo ""

# ============================================
# PART 9: PRAYER REQUESTS
# ============================================
echo -e "${BLUE}Creating Prayer Requests...${NC}"

PRAYER_PEOPLE=("Sarah Thompson" "Maria Rodriguez" "Emily Williams" "Jessica Davis" "Stephanie Moore")
PRAYER_REQUESTS=(
  "Prayer for healing and recovery"
  "Guidance in job search"
  "Family member salvation"
  "Traveling mercies"
  "Wisdom in major life decision"
)

for i in {0..4}; do
  person="${PRAYER_PEOPLE[$i]}"
  person_id="${PERSON_IDS[$person]}"
  request="${PRAYER_REQUESTS[$i]}"
  
  is_public=$((RANDOM % 2))
  
  if [ -n "$person_id" ]; then
    prayer_data="{\"person_id\":\"$person_id\",\"request\":\"$request\",\"is_public\":$([[ $is_public -eq 1 ]] && echo 'true' || echo 'false')}"
    
    response=$(api_call POST "/api/prayer-requests" "$prayer_data")
    
    if entity_exists "$response" "id"; then
      echo -e "${GREEN}✓${NC} Created prayer: $request ($([ $is_public -eq 1 ] && echo 'public' || echo 'private'))"
    fi
  fi
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Demo seed data created successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Summary:"
echo "  • ${#PERSON_IDS[@]} people"
echo "  • ${#FUND_IDS[@]} funds"
echo "  • ${#SONG_IDS[@]} songs"
echo "  • ${#GROUP_IDS[@]} groups"
echo "  • 4 services"
echo "  • ~20 donations"
echo "  • 4 events"
echo "  • 6 follow-ups"
echo "  • 5 prayer requests"
echo ""
