# Binary Prediction Market - Complete API Test Script (PowerShell)
# Run this after starting the server: go run ./cmd/server/main.go

$BASE_URL = "http://localhost:8080"
Write-Host "=== Binary Prediction Market API Test ==="
Write-Host ""

# Step 1: Create a new user
Write-Host "Step 1: Creating new user..."
$body = @{
    username = "testuser"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/users -Method Post -Body $body -ContentType "application/json"
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

$API_KEY = $response.data.api_key
$USER_ID = $response.data.id
Write-Host "User created:"
Write-Host "  User ID: $USER_ID"
Write-Host "  API Key: $API_KEY"
Write-Host ""

# Step 2: Deposit funds
Write-Host "Step 2: Depositing 10000 units..."
$body = @{
    amount = 10000
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/user/deposit" -Method Post -Body $body -ContentType "application/json" -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 3: Check account balance
Write-Host "Step 3: Checking account balance..."
$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/user/account" -Method Get -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 4: Create an event (as admin)
Write-Host "Step 4: Creating new prediction event..."
$startTime = (Get-Date).AddHours(1).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$endTime = (Get-Date).AddHours(24).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$body = @{
    title = "Will it rain tomorrow?"
    description = "Prediction whether it will rain in Beijing tomorrow"
    start_time = $startTime
    end_time = $endTime
    initial_yes_price = 0.6
    initial_supply = 1000
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/admin/events" -Method Post -Body $body -ContentType "application/json" -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

$EVENT_ID = $response.data.id
Write-Host "Event created with ID: $EVENT_ID"
Write-Host ""

# Step 5: List all events
Write-Host "Step 5: Listing all events..."
$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/events/all" -Method Get -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 6: Start trading for the event
Write-Host "Step 6: Starting trading..."
$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/admin/events/$EVENT_ID/start" -Method Post -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 7: Get event details
Write-Host "Step 7: Getting event details..."
$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/events/$EVENT_ID" -Method Get
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 8: Buy 10 YES shares
Write-Host "Step 8: Buying 10 YES shares..."
$body = @{
    event_id = $EVENT_ID
    quantity = 10
    direction = "YES"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/trade/buy" -Method Post -Body $body -ContentType "application/json" -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 9: Check account balance after purchase
Write-Host "Step 9: Checking account balance after purchase..."
$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/user/account" -Method Get -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 10: Buy 5 NO shares
Write-Host "Step 10: Buying 5 NO shares..."
$body = @{
    event_id = $EVENT_ID
    quantity = 5
    direction = "NO"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/trade/buy" -Method Post -Body $body -ContentType "application/json" -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 11: Sell 3 YES shares
Write-Host "Step 11: Selling 3 YES shares..."
$body = @{
    event_id = $EVENT_ID
    quantity = 3
    direction = "YES"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/trade/sell" -Method Post -Body $body -ContentType "application/json" -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 12: Check balance after trades
Write-Host "Step 12: Checking final balance before settlement..."
$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/user/account" -Method Get -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 13: Settle the event as YES
Write-Host "Step 13: Settling event with result YES..."
$body = @{
    result = "YES"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/admin/events/$EVENT_ID/settle" -Method Post -Body $body -ContentType "application/json" -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

# Step 14: Check final balance after payout
Write-Host "Step 14: Checking final balance after settlement..."
$response = Invoke-RestMethod -Uri "$BASE_URL/api/v1/user/account" -Method Get -Headers @{"X-API-Key" = $API_KEY}
Write-Host "Response: $($response | ConvertTo-Json)"
Write-Host ""

Write-Host "=== API Test Complete ==="
Write-Host ""
Write-Host "Summary of all endpoints tested:"
Write-Host "  ✓ POST /api/v1/users - Create user"
Write-Host "  ✓ POST /api/v1/user/deposit - Deposit funds"
Write-Host "  ✓ GET /api/v1/user/account - Get account"
Write-Host "  ✓ POST /api/v1/admin/events - Create event"
Write-Host "  ✓ GET /api/v1/events/all - List all events"
Write-Host "  ✓ GET /api/v1/events/:id - Get event"
Write-Host "  ✓ POST /api/v1/admin/events/:id/start - Start trading"
Write-Host "  ✓ POST /api/v1/trade/buy - Buy shares"
Write-Host "  ✓ POST /api/v1/trade/sell - Sell shares"
Write-Host "  ✓ POST /api/v1/admin/events/:id/settle - Settle event"
