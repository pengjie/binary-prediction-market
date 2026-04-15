#!/bin/bash

# Binary Prediction Market - Complete API Test Script
# Run this after starting the server: go run ./cmd/server/main.go

BASE_URL="http://localhost:8080"
echo "=== Binary Prediction Market API Test ==="
echo

# Step 1: Create a new user
echo "Step 1: Creating new user..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser"}')
echo "Response: $RESPONSE"
echo

# Extract API key from response
API_KEY=$(echo $RESPONSE | python3 -c "import json; data = json.loads('$RESPONSE'); print(data['data']['api_key'])")
USER_ID=$(echo $RESPONSE | python3 -c "import json; data = json.loads('$RESPONSE'); print(data['data']['id'])")
echo "User created:"
echo "  User ID: $USER_ID"
echo "  API Key: $API_KEY"
echo

# Step 2: Deposit funds
echo "Step 2: Depositing 10000 units..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/user/deposit" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{"amount": 10000}')
echo "Response: $RESPONSE"
echo

# Step 3: Check account balance
echo "Step 3: Checking account balance..."
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/user/account" \
  -H "X-API-Key: $API_KEY")
echo "Response: $RESPONSE"
echo

# Step 4: Create an event (as admin)
echo "Step 4: Creating new prediction event..."
START_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ" -d "+1 hour")
END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ" -d "+24 hours")
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/admin/events" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"title\": \"Will it rain tomorrow?\",
    \"description\": \"Prediction whether it will rain in Beijing tomorrow\",
    \"start_time\": \"$START_TIME\",
    \"end_time\": \"$END_TIME\",
    \"initial_yes_price\": 0.6,
    \"initial_supply\": 1000
  }")
echo "Response: $RESPONSE"
echo

EVENT_ID=$(echo $RESPONSE | python3 -c "import json; data = json.loads('$RESPONSE'); print(data['data']['id'])")
echo "Event created with ID: $EVENT_ID"
echo

# Step 5: List all events
echo "Step 5: Listing all events..."
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/events/all" \
  -H "X-API-Key: $API_KEY")
echo "Response: $RESPONSE"
echo

# Step 6: Start trading for the event
echo "Step 6: Starting trading..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/admin/events/$EVENT_ID/start" \
  -H "X-API-Key: $API_KEY")
echo "Response: $RESPONSE"
echo

# Step 7: Get event details
echo "Step 7: Getting event details..."
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/events/$EVENT_ID")
echo "Response: $RESPONSE"
echo

# Step 8: Buy 10 YES shares
echo "Step 8: Buying 10 YES shares..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/trade/buy" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"event_id\": $EVENT_ID,
    \"quantity\": 10,
    \"direction\": \"YES\"
  }")
echo "Response: $RESPONSE"
echo

# Step 9: Check account balance after purchase
echo "Step 9: Checking account balance after purchase..."
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/user/account" \
  -H "X-API-Key: $API_KEY")
echo "Response: $RESPONSE"
echo

# Step 10: Buy 5 NO shares
echo "Step 10: Buying 5 NO shares..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/trade/buy" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"event_id\": $EVENT_ID,
    \"quantity\": 5,
    \"direction\": \"NO\"
  }")
echo "Response: $RESPONSE"
echo

# Step 11: Sell 3 YES shares
echo "Step 11: Selling 3 YES shares..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/trade/sell" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{
    \"event_id\": $EVENT_ID,
    \"quantity\": 3,
    \"direction\": \"YES\"
  }")
echo "Response: $RESPONSE"
echo

# Step 12: Check balance after trades
echo "Step 12: Checking final balance before settlement..."
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/user/account" \
  -H "X-API-Key: $API_KEY")
echo "Response: $RESPONSE"
echo

# Step 13: Settle the event as YES
echo "Step 13: Settling event with result YES..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/admin/events/$EVENT_ID/settle" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{"result": "YES"}')
echo "Response: $RESPONSE"
echo

# Step 14: Check final balance after payout
echo "Step 14: Checking final balance after settlement..."
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/user/account" \
  -H "X-API-Key: $API_KEY")
echo "Response: $RESPONSE"
echo

echo "=== API Test Complete ==="
echo
echo "Summary of all endpoints tested:"
echo "  ✓ POST /api/v1/users - Create user"
echo "  ✓ POST /api/v1/user/deposit - Deposit funds"
echo "  ✓ GET /api/v1/user/account - Get account"
echo "  ✓ POST /api/v1/admin/events - Create event"
echo "  ✓ GET /api/v1/events/all - List all events"
echo "  ✓ GET /api/v1/events/:id - Get event"
echo "  ✓ POST /api/v1/admin/events/:id/start - Start trading"
echo "  ✓ POST /api/v1/trade/buy - Buy shares"
echo "  ✓ POST /api/v1/trade/sell - Sell shares"
echo "  ✓ POST /api/v1/admin/events/:id/settle - Settle event"
