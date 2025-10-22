#!/bin/bash

BASE_URL="http://localhost:8080"
TEST_USER1="sse_test_user1"
TEST_USER2="sse_test_user2"
TEST_EMAIL1="sse_test_user1@example.com"
TEST_EMAIL2="sse_test_user2@example.com"
TEST_PASSWORD="password123"

echo "🚀 Starting SSE Real-time Update Test"
echo "====================================="

# Function to make API calls
api_call() {
    local method=$1
    local url=$2
    local data=$3
    local token=$4
    
    if [ -n "$token" ]; then
        curl -s -X "$method" \
             -H "Content-Type: application/json" \
             -H "Authorization: Bearer $token" \
             -d "$data" \
             "$url"
    else
        curl -s -X "$method" \
             -H "Content-Type: application/json" \
             -d "$data" \
             "$url"
    fi
}

# Step 1: Register and login both users
echo "=== Step 1: User Setup ==="
echo "Registering user 1..."
user1_response=$(api_call "POST" "$BASE_URL/api/auth/register" "{\"username\":\"$TEST_USER1\",\"email\":\"$TEST_EMAIL1\",\"password\":\"$TEST_PASSWORD\"}")
echo "User 1 registered: $user1_response"

echo "Registering user 2..."
user2_response=$(api_call "POST" "$BASE_URL/api/auth/register" "{\"username\":\"$TEST_USER2\",\"email\":\"$TEST_EMAIL2\",\"password\":\"$TEST_PASSWORD\"}")
echo "User 2 registered: $user2_response"

echo "Logging in user 1..."
user1_login=$(api_call "POST" "$BASE_URL/api/auth/login" "{\"email\":\"$TEST_EMAIL1\",\"password\":\"$TEST_PASSWORD\"}")
user1_token=$(echo "$user1_login" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "User 1 token: ${user1_token:0:20}..."

echo "Logging in user 2..."
user2_login=$(api_call "POST" "$BASE_URL/api/auth/login" "{\"email\":\"$TEST_EMAIL2\",\"password\":\"$TEST_PASSWORD\"}")
user2_token=$(echo "$user2_login" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "User 2 token: ${user2_token:0:20}..."

# Step 2: Create a post with user 1
echo ""
echo "=== Step 2: Post Creation ==="
echo "Creating post with user 1..."
post_response=$(api_call "POST" "$BASE_URL/api/posts" "{\"title\":\"SSE Test Post\",\"caption\":\"Testing real-time updates\",\"media_type\":\"image\",\"media_url\":\"http://localhost:8080/test/image\"}" "$user1_token")
post_id=$(echo "$post_response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
echo "Post created with ID: $post_id"

# Step 3: Start SSE connection for user 1 in background
echo ""
echo "=== Step 3: SSE Connection Setup ==="
echo "Starting SSE connection for user 1..."
sse_url="$BASE_URL/api/events/stream?token=$user1_token"
echo "SSE URL: $sse_url"

# Start SSE connection in background and capture events
sse_log="/tmp/sse_test.log"
timeout 10s curl -s -N "$sse_url" > "$sse_log" &
sse_pid=$!

# Wait for SSE connection to establish
sleep 2

# Step 4: User 2 likes the post
echo ""
echo "=== Step 4: Like Action ==="
echo "User 2 liking post $post_id..."
like_response=$(api_call "POST" "$BASE_URL/api/posts/$post_id/like" "" "$user2_token")
echo "Like response: $like_response"

# Wait for SSE events to propagate
sleep 3

# Step 5: Check SSE events
echo ""
echo "=== Step 5: Event Verification ==="
if [ -f "$sse_log" ]; then
    sse_events=$(cat "$sse_log")
    echo "SSE events received:"
    echo "$sse_events"
    
    if echo "$sse_events" | grep -q "post_liked"; then
        echo "✅ SUCCESS: User 1 received post_liked event!"
    else
        echo "❌ FAILURE: User 1 did not receive post_liked event"
    fi
else
    echo "❌ FAILURE: No SSE log file found"
fi

# Step 6: Verify feed updates
echo ""
echo "=== Step 6: Feed Verification ==="
echo "Getting feed for user 1..."
user1_feed=$(api_call "GET" "$BASE_URL/api/feed" "" "$user1_token")
echo "User 1 feed: $user1_feed"

echo "Getting feed for user 2..."
user2_feed=$(api_call "GET" "$BASE_URL/api/feed" "" "$user2_token")
echo "User 2 feed: $user2_feed"

# Cleanup
kill $sse_pid 2>/dev/null
rm -f "$sse_log"

echo ""
echo "🎉 SSE Real-time Update Test Completed!"
